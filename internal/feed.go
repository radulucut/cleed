package internal

import (
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"slices"
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
	feeds, err := f.storage.GetFeedFromList(list)
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

func (f *TerminalFeed) Feed(opts *FeedOptions) error {
	items, err := f.processFeeds(opts)
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
	cellMax := [2]int{}
	for i := l - 1; i >= 0; i-- {
		fi := items[i]
		fi.PublishedRelative = utils.Relative(f.time.Now().Unix() - fi.Item.PublishedParsed.Unix())
		cellMax[0] = max(cellMax[0], runewidth.StringWidth(fi.Feed.Title), len(fi.PublishedRelative))
		cellMax[1] = max(cellMax[1], runewidth.StringWidth(fi.Item.Title), runewidth.StringWidth(fi.Item.Link))
	}
	cellMax[0] = min(cellMax[0], 30)
	for i := l - 1; i >= 0; i-- {
		fi := items[i]
		newMark := ""
		if fi.IsNew {
			newMark = f.printer.ColorForeground("• ", 10)
		}
		f.printer.Print(
			f.printer.ColorForeground(runewidth.FillRight(runewidth.Truncate(fi.Feed.Title, cellMax[0], "..."), cellMax[0]), fi.FeedColor),
			"  ",
			newMark+fi.Item.Title,
			"\n",
			f.printer.ColorForeground(runewidth.FillRight(fi.PublishedRelative, cellMax[0]), 7),
			"  ",
			f.printer.ColorForeground(fi.Item.Link, 7),
			"\n\n",
		)
	}
	return nil
}

func (f *TerminalFeed) processFeeds(opts *FeedOptions) ([]*FeedItem, error) {
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
		f.storage.LoadFeedFromList(feeds, lists[i])
	}
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
				URL:       url,
				LastCheck: time.Time{},
			}
			cacheInfo[url] = ci
		}
		wg.Add(1)
		go func(ci *storage.CacheInfoItem) {
			defer wg.Done()
			changed, etag, err := f.fetchFeed(ci)
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
				color = uint8(len(feedColorMap) % 231)
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
			if changed {
				ci.ETag = etag
				ci.LastCheck = f.time.Now()
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

func (f *TerminalFeed) parseFeed(url string) (*gofeed.Feed, error) {
	fc, err := f.storage.OpenFeedCache(url)
	if err != nil {
		return nil, err
	}
	defer fc.Close()
	return f.parser.Parse(fc)
}

func (f *TerminalFeed) fetchFeed(feed *storage.CacheInfoItem) (bool, string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, "GET", feed.URL, nil)
	if err != nil {
		return false, "", utils.NewInternalError(fmt.Sprintf("failed to create request: %v", err))
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
		return false, "", err
	}
	defer res.Body.Close()
	if res.StatusCode == http.StatusNotModified {
		return false, "", nil
	}
	if res.StatusCode != http.StatusOK {
		return false, "", fmt.Errorf("unexpected status code: %d", res.StatusCode)
	}
	var bodyReader io.Reader = res.Body
	contentEncoding := res.Header.Get("Content-Encoding")
	if contentEncoding == "br" {
		bodyReader = brotli.NewReader(res.Body)
	} else if contentEncoding == "gzip" {
		bodyReader, err = gzip.NewReader(res.Body)
		if err != nil {
			return false, "", err
		}
	}
	err = f.storage.SaveFeedCache(bodyReader, feed.URL)
	return true, res.Header.Get("ETag"), err
}
