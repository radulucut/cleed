package internal

import (
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/andybalholm/brotli"
	"github.com/mattn/go-runewidth"
	"github.com/mmcdole/gofeed"
	"github.com/radulucut/cleed/internal/storage"
	"github.com/radulucut/cleed/internal/utils"
	"golang.org/x/net/proxy"
	miniflux "miniflux.app/v2/client"
)

type TerminalFeed struct {
	time     utils.Time
	printer  *Printer
	storage  *storage.LocalStorage
	http     *http.Client
	parser   *gofeed.Parser
	miniflux *miniflux.Client

	version                  string
	defaultExploreRepository string
}

func NewTerminalFeed(
	_time utils.Time,
	printer *Printer,
	storage *storage.LocalStorage,
) *TerminalFeed {
	return &TerminalFeed{
		time:    _time,
		printer: printer,
		storage: storage,

		http:   &http.Client{},
		parser: gofeed.NewParser(),
	}
}

func (f *TerminalFeed) Follow(urls []string, list string) error {
	if len(urls) == 0 {
		return utils.NewInternalError("please provide at least one URL")
	}
	for i := range urls {
		u, err := url.ParseRequestURI(urls[i])
		if err != nil {
			return utils.NewInternalError("failed to parse URL: " + urls[i])
		}
		urls[i] = u.String()
	}
	err := f.storage.AddToList(urls, list)
	if err != nil {
		return utils.NewInternalError("failed to save feeds: " + err.Error())
	}
	f.printer.Printf("added %s to list: %s\n", utils.Pluralize(int64(len(urls)), "feed"), list)
	return nil
}

func (f *TerminalFeed) Unfollow(urls []string, list string) error {
	results, err := f.storage.RemoveFromList(urls, list)
	if err != nil {
		return utils.NewInternalError(err.Error())
	}
	for i := range urls {
		if results[i] {
			f.printer.Print(urls[i] + " was removed from the list\n")
		} else {
			f.printer.Print(f.printer.ColorForeground(urls[i]+" was not found in the list\n", 11))
		}
	}
	return nil
}

func (f *TerminalFeed) Lists() error {
	lists, err := f.storage.LoadLists()
	if err != nil {
		return utils.NewInternalError("failed to list lists: " + err.Error())
	}
	if len(lists) == 0 {
		f.printer.Println("default")
		return nil
	}
	slices.Sort(lists)
	for i := range lists {
		f.printer.Println(lists[i])
	}
	return nil
}

func (f *TerminalFeed) ListFeeds(list string) error {
	feeds, err := f.storage.GetFeedsFromList(list)
	if err != nil {
		return utils.NewInternalError("failed to list feeds: " + err.Error())
	}
	for i := range feeds {
		f.printer.Printf("%s  %s\n", feeds[i].AddedAt.Format("2006-01-02 15:04:05"), feeds[i].Address)
	}
	f.printer.Println("Total: " + utils.Pluralize(int64(len(feeds)), "feed"))
	return nil
}

func (f *TerminalFeed) RenameList(oldName, newName string) error {
	err := f.storage.RenameList(oldName, newName)
	if err != nil {
		return utils.NewInternalError("failed to rename list: " + err.Error())
	}
	f.printer.Printf("list %s was renamed to %s\n", oldName, newName)
	return nil
}

func (f *TerminalFeed) MergeLists(list, otherList string) error {
	err := f.storage.MergeLists(list, otherList)
	if err != nil {
		return utils.NewInternalError("failed to merge lists: " + err.Error())
	}
	f.printer.Printf("list %s was merged with %s. %s was removed\n", list, otherList, otherList)
	return nil
}

func (f *TerminalFeed) RemoveList(list string) error {
	err := f.storage.RemoveList(list)
	if err != nil {
		return utils.NewInternalError("failed to remove list: " + err.Error())
	}
	f.printer.Printf("list %s was removed\n", list)
	return nil
}

func (f *TerminalFeed) ShowConfigPath() error {
	path, err := f.storage.JoinConfigDir("")
	if err != nil {
		return utils.NewInternalError("failed to get config path: " + err.Error())
	}
	f.printer.Println(path)
	return nil
}

func (f *TerminalFeed) ShowCachePath() error {
	path, err := f.storage.JoinCacheDir("")
	if err != nil {
		return utils.NewInternalError("failed to get cache path: " + err.Error())
	}
	f.printer.Println(path)
	return nil
}

func (f *TerminalFeed) ShowCacheInfo() error {
	cacheInfo, err := f.storage.LoadCacheInfo()
	if err != nil {
		return utils.NewInternalError("failed to load cache info: " + err.Error())
	}
	cellMax := [1]int{}
	items := make([]*storage.CacheInfoItem, 0, len(cacheInfo))
	for k, v := range cacheInfo {
		cellMax[0] = max(cellMax[0], len(k))
		items = append(items, v)
	}
	f.printer.Print(runewidth.FillRight("URL", cellMax[0]))
	f.printer.Println("  Last fetch           Fetch after")
	slices.SortFunc(items, func(a, b *storage.CacheInfoItem) int {
		return strings.Compare(a.URL, b.URL)
	})
	for i := range items {
		f.printer.Print(runewidth.FillRight(items[i].URL, cellMax[0]))
		f.printer.Printf("  %s  %s\n", items[i].LastFetch.Format("2006-01-02 15:04:05"), items[i].FetchAfter.Format("2006-01-02 15:04:05"))
	}
	return nil
}

type FeedOptions struct {
	List       string
	Query      [][]rune
	Regexp     *regexp.Regexp
	Limit      int
	Since      time.Time
	Proxy      *url.URL
	CachedOnly bool
	// Raw prints each item on its own line: fields tab-separated; timestamps UTC RFC3339 (ISO 8601).
	Raw bool
}

func (f *TerminalFeed) Search(query string, opts *FeedOptions) error {
	summary := &RunSummary{
		Start: f.time.Now(),
	}
	config, err := f.storage.LoadConfig()
	if err != nil {
		return utils.NewInternalError("failed to load config: " + err.Error())
	}
	opts.Query = utils.Tokenize(query, nil)
	if len(opts.Query) == 0 {
		return utils.NewInternalError("query is empty")
	}
	items, err := f.processFeeds(opts, config, summary)
	if err != nil {
		return err
	}
	slices.SortFunc(items, func(a, b *FeedItem) int {
		if a.Score > b.Score {
			return 1
		}
		if a.Score < b.Score {
			return -1
		}
		return 0
	})
	f.outputItems(items, config, summary, opts)
	return nil
}

func (f *TerminalFeed) SearchRegexp(query string, opts *FeedOptions) error {
	summary := &RunSummary{
		Start: f.time.Now(),
	}
	config, err := f.storage.LoadConfig()
	if err != nil {
		return utils.NewInternalError("failed to load config: " + err.Error())
	}
	opts.Regexp, err = regexp.Compile(query)
	if err != nil {
		return utils.NewInternalError("failed to compile expression: " + err.Error())
	}
	items, err := f.processFeeds(opts, config, summary)
	if err != nil {
		return err
	}
	slices.SortFunc(items, func(a, b *FeedItem) int {
		if a.Item.PublishedParsed == nil || b.Item.PublishedParsed == nil {
			return 0
		}
		if a.Item.PublishedParsed.After(*b.Item.PublishedParsed) {
			return -1
		}
		if a.Item.PublishedParsed.Before(*b.Item.PublishedParsed) {
			return 1
		}
		return 0
	})
	f.outputItems(items, config, summary, opts)
	return nil
}

type FeedItem struct {
	Feed              *gofeed.Feed
	Item              *gofeed.Item
	PublishedRelative string
	FeedColor         uint8
	IsNew             bool
	Score             int
}

type RunSummary struct {
	Start        time.Time
	FeedsCount   int
	FeedsCached  int
	FeedsFetched int
	ItemsCount   int
	ItemsShown   int
}

func (f *TerminalFeed) Feed(opts *FeedOptions) error {
	summary := &RunSummary{
		Start: f.time.Now(),
	}
	config, err := f.storage.LoadConfig()
	if err != nil {
		return utils.NewInternalError("failed to load config: " + err.Error())
	}
	items, err := f.processFeeds(opts, config, summary)
	if err != nil {
		return err
	}
	slices.SortFunc(items, func(a, b *FeedItem) int {
		if a.Item.PublishedParsed == nil || b.Item.PublishedParsed == nil {
			return 0
		}
		if a.Item.PublishedParsed.After(*b.Item.PublishedParsed) {
			return -1
		}
		if a.Item.PublishedParsed.Before(*b.Item.PublishedParsed) {
			return 1
		}
		return 0
	})
	config.LastRun = f.time.Now()
	f.storage.SaveConfig()
	f.outputItems(items, config, summary, opts)
	return nil
}

func escapeRawField(s string) string {
	s = strings.ReplaceAll(s, "\t", " ")
	s = strings.ReplaceAll(s, "\r", " ")
	s = strings.ReplaceAll(s, "\n", " ")
	return s
}

// rawTSVField strips HTML from typical feed strings and escapes for a single TSV field.
func rawTSVField(s string) string {
	return escapeRawField(utils.PlainTextFromHTML(s))
}

// rawItemDescription returns plain text from the item body for --raw output: no length cap.
// When both description and content exist, it prefers the longest non-redundant text (full article
// when content starts with the same teaser as description, otherwise both parts joined).
func rawItemDescription(item *gofeed.Item) string {
	desc := strings.TrimSpace(item.Description)
	cont := strings.TrimSpace(item.Content)
	plainD := ""
	if desc != "" {
		plainD = utils.PlainTextFromHTML(desc)
	}
	plainC := ""
	if cont != "" {
		plainC = utils.PlainTextFromHTML(cont)
	}
	switch {
	case plainD == "" && plainC == "":
		return ""
	case plainD == "":
		return escapeRawField(plainC)
	case plainC == "":
		return escapeRawField(plainD)
	case plainC == plainD:
		return escapeRawField(plainD)
	case strings.HasPrefix(plainC, plainD):
		return escapeRawField(plainC)
	case strings.HasPrefix(plainD, plainC):
		return escapeRawField(plainD)
	default:
		return escapeRawField(plainD + " " + plainC)
	}
}

func (f *TerminalFeed) outputItems(
	items []*FeedItem,
	config *storage.Config,
	summary *RunSummary,
	opts *FeedOptions,
) {
	l := len(items)
	if l == 0 {
		f.printer.ErrPrintln("no items to display")
		return
	}
	if opts.Limit > 0 {
		l = min(len(items), opts.Limit)
	}
	if opts.Raw {
		for i := 0; i < l; i++ {
			fi := items[i]
			if strings.HasPrefix(fi.Item.Link, "/") {
				baseURL := strings.TrimSuffix(fi.Feed.Link, "/")
				fi.Item.Link = baseURL + fi.Item.Link
			}
			pub := ""
			if fi.Item.PublishedParsed != nil && !fi.Item.PublishedParsed.IsZero() {
				pub = fi.Item.PublishedParsed.UTC().Format(time.RFC3339)
			}
			feedLink := fi.Feed.Link
			if feedLink == "" {
				feedLink = fi.Feed.FeedLink
			}
			cats := make([]string, 0, len(fi.Item.Categories))
			for _, c := range fi.Item.Categories {
				cats = append(cats, utils.PlainTextFromHTML(c))
			}
			catsJoined := strings.Join(cats, ",")
			fields := []string{
				pub,
				rawTSVField(fi.Feed.Title),
				escapeRawField(feedLink),
				rawTSVField(fi.Feed.Description),
				rawTSVField(fi.Item.Title),
				rawItemDescription(fi.Item),
				escapeRawField(fi.Item.Link),
				escapeRawField(catsJoined),
				escapeRawField(fi.Item.GUID),
			}
			f.printer.Println(strings.Join(fields, "\t"))
		}
		return
	}
	cellMax := [1]int{}
	for i := l - 1; i >= 0; i-- {
		fi := items[i]
		fi.PublishedRelative = utils.Relative(f.time.Now().Unix() - fi.Item.PublishedParsed.Unix())
		cellMax[0] = max(cellMax[0], runewidth.StringWidth(fi.Feed.Title), len(fi.PublishedRelative))
	}
	cellMax[0] = min(cellMax[0], 30)
	secondaryTextColor := mapColor(7, config)
	highlightColor := mapColor(10, config)
	for i := l - 1; i >= 0; i-- {
		fi := items[i]
		newMark := ""
		if fi.IsNew {
			newMark = f.printer.ColorForeground("• ", highlightColor)
		}
		if strings.HasPrefix(fi.Item.Link, "/") {
			baseUrl := strings.TrimSuffix(fi.Feed.Link, "/")
			fi.Item.Link = baseUrl + fi.Item.Link
		}
		f.printer.Print(
			f.printer.ColorForeground(runewidth.FillRight(runewidth.Truncate(fi.Feed.Title, cellMax[0], "..."), cellMax[0]), fi.FeedColor),
			"  ",
			newMark+fi.Item.Title,
			"\n",
			f.printer.ColorForeground(runewidth.FillRight(fi.PublishedRelative, cellMax[0]), secondaryTextColor),
			"  ",
			f.printer.ColorForeground(fi.Item.Link, secondaryTextColor),
			"\n\n",
		)
	}
	if config.Summary == 1 {
		summary.ItemsShown = l
		f.printSummary(summary)
	}
}

func (f *TerminalFeed) printSummary(s *RunSummary) {
	f.printer.Printf("Displayed %s from %s (%d cached, %d fetched) with %s in %.2fs\n",
		utils.Pluralize(int64(s.ItemsShown), "item"),
		utils.Pluralize(int64(s.FeedsCount), "feed"),
		s.FeedsCached,
		s.FeedsFetched,
		utils.Pluralize(int64(s.ItemsCount), "item"),
		f.time.Now().Sub(s.Start).Seconds(),
	)
}

func (f *TerminalFeed) processFeeds(opts *FeedOptions, config *storage.Config, summary *RunSummary) ([]*FeedItem, error) {
	f.http.Timeout = time.Duration(config.Timeout) * time.Second
	err := f.setProxy(opts)
	if err != nil {
		return nil, err
	}
	lists := make([]string, 0)
	if opts.List != "" {
		lists = append(lists, opts.List)
	} else {
		lists, err = f.storage.LoadLists()
		if err != nil {
			return nil, utils.NewInternalError("failed to load lists: " + err.Error())
		}
		if len(lists) == 0 {
			return nil, utils.NewInternalError("no feeds to display")
		}
	}
	feeds := make(map[string]*storage.ListItem)
	for i := range lists {
		f.storage.LoadFeedsFromList(feeds, lists[i])
	}
	summary.FeedsCount = len(feeds)
	cacheInfo, err := f.storage.LoadCacheInfo()
	if err != nil {
		return nil, utils.NewInternalError("failed to load cache info: " + err.Error())
	}
	mx := sync.Mutex{}
	wg := sync.WaitGroup{}
	sem := make(chan struct{}, config.BatchSize)
	items := make([]*FeedItem, 0)
	feedColorMap := make(map[string]uint8)
	for url := range feeds {
		sem <- struct{}{}
		ci := cacheInfo[url]
		if ci == nil {
			ci = &storage.CacheInfoItem{
				URL:        url,
				LastFetch:  time.Unix(0, 0),
				FetchAfter: time.Unix(0, 0),
			}
			cacheInfo[url] = ci
		}
		wg.Add(1)
		go func(ci *storage.CacheInfoItem) {
			defer wg.Done()
			defer func() {
				<-sem
			}()
			if opts.CachedOnly {
				feed, err := f.parseFeed(url)
				if err != nil {
					return
				}
				mx.Lock()
				defer mx.Unlock()
				items = f.processFeedItems(feed, items, config, opts, summary, feedColorMap, ci)
				summary.FeedsCached++
				return
			}
			res, err := f.fetchFeed(ci, config)
			if err != nil {
				f.printer.ErrPrintf("failed to fetch feed: %s: %v\n", ci.URL, err)
				return
			}
			feed, err := f.parseFeed(url)
			if err != nil {
				f.printer.ErrPrintf("failed to parse feed: %s: %v\n", ci.URL, err)
				return
			}
			mx.Lock()
			defer mx.Unlock()
			items = f.processFeedItems(feed, items, config, opts, summary, feedColorMap, ci)
			if res.Changed {
				ci.ETag = res.ETag
				ci.LastFetch = f.time.Now()
				summary.FeedsFetched++
			} else {
				summary.FeedsCached++
			}
			if res.FetchAfter.After(ci.FetchAfter) {
				ci.FetchAfter = res.FetchAfter
			}
		}(ci)
	}
	wg.Wait()
	err = f.storage.SaveCacheInfo(cacheInfo)
	if err != nil {
		f.printer.ErrPrintln("failed to save cache informaton:", err)
	}
	return items, nil
}

func (f *TerminalFeed) tokenizeItem(item *gofeed.Item) [][]rune {
	tokens := utils.Tokenize(item.Title, nil)
	for i := range item.Categories {
		tokens = utils.Tokenize(item.Categories[i], tokens)
	}
	return tokens
}

func (f *TerminalFeed) parseFeed(url string) (*gofeed.Feed, error) {
	fc, err := f.storage.OpenFeedCache(url)
	if err != nil {
		return nil, err
	}
	defer fc.Close()
	return f.parser.Parse(fc)
}

func (f *TerminalFeed) processFeedItems(
	feed *gofeed.Feed,
	items []*FeedItem,
	config *storage.Config,
	opts *FeedOptions,
	summary *RunSummary,
	feedColorMap map[string]uint8,
	ci *storage.CacheInfoItem,
) []*FeedItem {
	summary.ItemsCount += len(feed.Items)
	color, ok := feedColorMap[feed.Title]
	if !ok {
		color = mapColor(uint8(len(feedColorMap)%256), config)
		feedColorMap[feed.Title] = color
	}
	currentTime := f.time.Now()
	for _, feedItem := range feed.Items {
		if feedItem.PublishedParsed == nil {
			feedItem.PublishedParsed = &time.Time{}
		}
		if !opts.Since.IsZero() && feedItem.PublishedParsed.Before(opts.Since) {
			continue
		}
		if config.HideFutureItems && feedItem.PublishedParsed.After(currentTime) {
			continue
		}
		// default search
		score := 0
		if len(opts.Query) > 0 {
			score = utils.Score(opts.Query, f.tokenizeItem(feedItem))
			if score == -1 {
				continue
			}
		}
		// regex search
		if opts.Regexp != nil && !opts.Regexp.MatchString(feedItem.Title) {
			continue
		}
		items = append(items, &FeedItem{
			Feed:      feed,
			Item:      feedItem,
			FeedColor: color,
			IsNew:     feedItem.PublishedParsed.After(ci.LastFetch),
			Score:     score,
		})
	}
	return items
}

type FetchResult struct {
	Changed    bool
	ETag       string
	FetchAfter time.Time
}

func (f *TerminalFeed) fetchFeed(feed *storage.CacheInfoItem, config *storage.Config) (*FetchResult, error) {
	if feed.FetchAfter.After(f.time.Now()) {
		return &FetchResult{
			Changed: false,
		}, nil
	}
	req, err := http.NewRequest("GET", feed.URL, nil)
	if err != nil {
		return nil, utils.NewInternalError(fmt.Sprintf("failed to create request: %v", err))
	}
	if config.UserAgent != "-" {
		req.Header.Set("User-Agent", config.UserAgent)
	}
	if feed.ETag != "" {
		req.Header.Set("If-None-Match", feed.ETag)
	}
	if !feed.LastFetch.IsZero() {
		req.Header.Set("If-Modified-Since", feed.LastFetch.Format(http.TimeFormat))
	}
	req.Header.Set("Accept", "application/rss+xml, application/atom+xml, application/xml, application/json, text/xml")
	req.Header.Set("Accept-Encoding", "br, gzip")
	res, err := f.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode == http.StatusNotModified {
		return &FetchResult{
			Changed:    false,
			FetchAfter: f.time.Now().Add(parseMaxAge(res.Header.Get("Cache-Control"))),
		}, nil
	}
	if res.StatusCode == http.StatusTooManyRequests || res.StatusCode == http.StatusServiceUnavailable {
		return &FetchResult{
			Changed:    false,
			FetchAfter: f.parseRetryAfter(res.Header.Get("Retry-After")),
		}, nil
	}
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", res.StatusCode)
	}
	var bodyReader io.Reader = res.Body
	contentEncoding := res.Header.Get("Content-Encoding")
	if contentEncoding == "br" {
		bodyReader = brotli.NewReader(res.Body)
	} else if contentEncoding == "gzip" {
		bodyReader, err = gzip.NewReader(res.Body)
		if err != nil {
			return nil, err
		}
	}
	err = f.storage.SaveFeedCache(bodyReader, feed.URL)
	return &FetchResult{
		Changed:    true,
		ETag:       res.Header.Get("ETag"),
		FetchAfter: f.time.Now().Add(parseMaxAge(res.Header.Get("Cache-Control"))),
	}, err
}

func (f *TerminalFeed) parseRetryAfter(retryAfter string) time.Time {
	if retryAfter == "" {
		return f.time.Now().Add(5 * time.Minute)
	}
	retryAfterSeconds, err := strconv.Atoi(retryAfter)
	if err == nil {
		return f.time.Now().Add(time.Duration(retryAfterSeconds) * time.Second)
	}
	retryAfterTime, err := time.Parse(time.RFC1123, retryAfter)
	if err == nil {
		return retryAfterTime
	}
	return f.time.Now().Add(5 * time.Minute)
}

func (f *TerminalFeed) setProxy(opts *FeedOptions) error {
	if opts.Proxy == nil {
		return nil
	}
	var transport *http.Transport
	if strings.HasPrefix(strings.ToLower(opts.Proxy.Scheme), "socks5") {
		var auth *proxy.Auth
		if opts.Proxy.User != nil {
			password, _ := opts.Proxy.User.Password()
			auth = &proxy.Auth{
				User:     opts.Proxy.User.Username(),
				Password: password,
			}
		}
		dialer, err := proxy.SOCKS5("tcp", opts.Proxy.Host, auth, proxy.Direct)
		if err != nil {
			return utils.NewInternalError("failed to create SOCKS5 dialer: " + err.Error())
		}
		transport = &http.Transport{
			Dial: dialer.Dial,
		}
	} else {
		transport = &http.Transport{
			Proxy: http.ProxyURL(opts.Proxy),
		}
	}
	f.http.Transport = transport
	return nil
}

func parseMaxAge(cacheControl string) time.Duration {
	if cacheControl == "" {
		return 60 * time.Second
	}
	parts := strings.Split(cacheControl, ",")
	for i := range parts {
		part := strings.TrimSpace(parts[i])
		if strings.HasPrefix(part, "max-age=") {
			seconds, err := strconv.ParseInt(part[8:], 10, 64)
			if err == nil {
				return time.Duration(max(seconds, 60)) * time.Second
			}
			break
		}
	}
	return 60 * time.Second
}

func mapColor(color uint8, config *storage.Config) uint8 {
	if c, ok := config.ColorMap[color]; ok {
		return c
	}
	return color
}
