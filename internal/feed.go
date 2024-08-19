package internal

import (
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
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
)

type TerminalFeed struct {
	time    utils.Time
	printer *Printer
	storage *storage.LocalStorage
	http    *http.Client
	parser  *gofeed.Parser

	agent string
}

func NewTerminalFeed(
	time utils.Time,
	printer *Printer,
	storage *storage.LocalStorage,
) *TerminalFeed {
	return &TerminalFeed{
		time:    time,
		printer: printer,
		storage: storage,

		http:   &http.Client{},
		parser: gofeed.NewParser(),
	}
}

func (f *TerminalFeed) SetAgent(agent string) {
	f.agent = agent
}

func (f *TerminalFeed) DisplayConfig() error {
	config, err := f.storage.LoadConfig()
	if err != nil {
		return utils.NewInternalError("failed to load config: " + err.Error())
	}
	styling := "default"
	if config.Styling == 0 {
		styling = "enabled"
	} else if config.Styling == 1 {
		styling = "disabled"
	}
	f.printer.Println("Styling:", styling)
	f.printer.Print("Color map:")
	for k, v := range config.ColorMap {
		f.printer.Printf(" %d:%d", k, v)
	}
	f.printer.Println()
	summary := "disabled"
	if config.Summary == 1 {
		summary = "enabled"
	}
	f.printer.Println("Summary:", summary)
	return nil
}

func (f *TerminalFeed) SetStyling(v uint8) error {
	config, err := f.storage.LoadConfig()
	if err != nil {
		return utils.NewInternalError("failed to load config: " + err.Error())
	}
	if v > 2 {
		return utils.NewInternalError("invalid value for styling")
	}
	config.Styling = v
	err = f.storage.SaveConfig()
	if err != nil {
		return utils.NewInternalError("failed to save config: " + err.Error())
	}
	f.printer.Println("styling was updated")
	return nil
}

func (f *TerminalFeed) SetSummary(v uint8) error {
	config, err := f.storage.LoadConfig()
	if err != nil {
		return utils.NewInternalError("failed to load config: " + err.Error())
	}
	if v > 1 {
		return utils.NewInternalError("invalid value for summary")
	}
	config.Summary = v
	err = f.storage.SaveConfig()
	if err != nil {
		return utils.NewInternalError("failed to save config: " + err.Error())
	}
	f.printer.Println("summary was updated")
	return nil
}

func (f *TerminalFeed) UpdateColorMap(mappings string) error {
	config, err := f.storage.LoadConfig()
	if err != nil {
		return utils.NewInternalError("failed to load config: " + err.Error())
	}
	if mappings == "" {
		config.ColorMap = make(map[uint8]uint8)
	} else {
		colors := strings.Split(mappings, ",")
		for i := range colors {
			parts := strings.Split(colors[i], ":")
			if len(parts) == 0 {
				return utils.NewInternalError("failed to parse color mapping: " + colors[i])
			}
			left, err := strconv.Atoi(parts[0])
			if err != nil {
				return utils.NewInternalError("failed to parse color mapping: " + parts[0])
			}
			if len(parts) == 1 || parts[1] == "" {
				delete(config.ColorMap, uint8(left))
			} else {
				right, err := strconv.Atoi(parts[1])
				if err != nil {
					return utils.NewInternalError("failed to parse color mapping: " + parts[1])
				}
				config.ColorMap[uint8(left)] = uint8(right)
			}
		}
	}
	err = f.storage.SaveConfig()
	if err != nil {
		return utils.NewInternalError("failed to save config: " + err.Error())
	}
	f.printer.Println("color map updated")
	return nil
}

func (f *TerminalFeed) DisplayColorRange() {
	styling := f.printer.GetStyling()
	f.printer.SetStyling(true)
	for i := 0; i < 256; i++ {
		f.printer.Print(f.printer.ColorForeground(fmt.Sprintf("%d ", i), uint8(i)))
	}
	f.printer.Println()
	f.printer.SetStyling(styling)
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

type FeedItem struct {
	Feed              *gofeed.Feed
	Item              *gofeed.Item
	PublishedRelative string
	FeedColor         uint8
	IsNew             bool
}

type FeedOptions struct {
	List  string
	Limit int
	Since time.Time
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
	l := len(items)
	if l == 0 {
		f.printer.ErrPrintln("no items to display")
		return nil
	}
	if opts.Limit > 0 {
		l = min(len(items), opts.Limit)
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
	config.LastRun = f.time.Now()
	f.storage.SaveConfig()
	if config.Summary == 1 {
		summary.ItemsShown = l
		f.printSummary(summary)
	}
	return nil
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
	var err error
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
	items := make([]*FeedItem, 0)
	feedColorMap := make(map[string]uint8)
	for url := range feeds {
		ci := cacheInfo[url]
		if ci == nil {
			ci = &storage.CacheInfoItem{
				URL:        url,
				LastCheck:  time.Unix(0, 0),
				FetchAfter: time.Unix(0, 0),
			}
			cacheInfo[url] = ci
		}
		wg.Add(1)
		go func(ci *storage.CacheInfoItem) {
			defer wg.Done()
			res, err := f.fetchFeed(ci)
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
			color, ok := feedColorMap[feed.Title]
			if !ok {
				color = mapColor(uint8(len(feedColorMap)%256), config)
				feedColorMap[feed.Title] = color
			}
			for _, feedItem := range feed.Items {
				if feedItem.PublishedParsed == nil {
					feedItem.PublishedParsed = &time.Time{}
				}
				if !opts.Since.IsZero() && feedItem.PublishedParsed.Before(opts.Since) {
					continue
				}
				items = append(items, &FeedItem{
					Feed:      feed,
					Item:      feedItem,
					FeedColor: color,
					IsNew:     feedItem.PublishedParsed.After(ci.LastCheck),
				})
			}
			if res.Changed {
				ci.ETag = res.ETag
				ci.LastCheck = f.time.Now()
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
	summary.ItemsCount = len(items)
	return items, nil
}

func (f *TerminalFeed) parseFeed(url string) (*gofeed.Feed, error) {
	fc, err := f.storage.OpenFeedCache(url)
	if err != nil {
		return nil, err
	}
	defer fc.Close()
	return f.parser.Parse(fc)
}

type FetchResult struct {
	Changed    bool
	ETag       string
	FetchAfter time.Time
}

func (f *TerminalFeed) fetchFeed(feed *storage.CacheInfoItem) (*FetchResult, error) {
	if feed.FetchAfter.After(f.time.Now()) {
		return &FetchResult{
			Changed: false,
		}, nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, "GET", feed.URL, nil)
	if err != nil {
		return nil, utils.NewInternalError(fmt.Sprintf("failed to create request: %v", err))
	}
	req.Header.Set("User-Agent", f.agent)
	if feed.ETag != "" {
		req.Header.Set("If-None-Match", feed.ETag)
	}
	if !feed.LastCheck.IsZero() {
		req.Header.Set("If-Modified-Since", feed.LastCheck.Format(http.TimeFormat))
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
