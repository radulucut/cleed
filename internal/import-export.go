package internal

import (
	"bufio"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/radulucut/cleed/internal/utils"
)

func (f *TerminalFeed) ImportFromFile(path, list string) error {
	fi, err := os.Open(path)
	if err != nil {
		return utils.NewInternalError("failed to open file: " + err.Error())
	}
	defer fi.Close()
	urls := make([]string, 0)
	scanner := bufio.NewScanner(fi)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "#") {
			continue
		}
		urls = append(urls, line)
	}
	err = f.storage.AddToList(urls, list)
	if err != nil {
		return utils.NewInternalError("failed to save feeds: " + err.Error())
	}
	f.printer.Printf("added %s to list: %s\n", utils.Pluralize(int64(len(urls)), "feed"), list)
	return nil
}

func (f *TerminalFeed) ExportToFile(path, list string) error {
	feeds, err := f.storage.GetFeedsFromList(list)
	if err != nil {
		return utils.NewInternalError("failed to list feeds: " + err.Error())
	}
	if len(feeds) == 0 {
		f.printer.Println("no feeds to export")
		return nil
	}
	fo, err := os.Create(path)
	if err != nil {
		return utils.NewInternalError("failed to create file: " + err.Error())
	}
	defer fo.Close()
	for i := range feeds {
		_, err = fo.WriteString(feeds[i].Address + "\n")
		if err != nil {
			return utils.NewInternalError("failed to write to file: " + err.Error())
		}
	}
	f.printer.Printf("exported %s to %s\n", utils.Pluralize(int64(len(feeds)), "feed"), path)
	return nil
}

func (f *TerminalFeed) ImportFromOPML(path, list string) error {
	opml, err := utils.ParseOPMLFile(path)
	if err != nil {
		return utils.NewInternalError("failed to parse OPML: " + err.Error())
	}
	return f.importOPML(opml, list)
}

func (f *TerminalFeed) ExportToOPML(path, list string, cachedOnly bool) error {
	fo, err := os.Create(path)
	if err != nil {
		return utils.NewInternalError("failed to create file: " + err.Error())
	}
	defer fo.Close()
	res, err := f.writeOPML(fo, list, cachedOnly)
	if err != nil {
		return err
	}
	f.printer.Printf("exported %s from %s to %s\n", utils.Pluralize(res.FeedCount, "feed"), utils.Pluralize(res.ListCount, "list"), path)
	return nil
}

type OPMLExportResult struct {
	FeedCount int64
	ListCount int64
}

func (f *TerminalFeed) writeOPML(fo io.Writer, list string, cachedOnly bool) (*OPMLExportResult, error) {
	fmt.Fprint(fo, xml.Header)
	fmt.Fprint(fo, "<opml version=\"1.0\">\n  <head>\n")
	fmt.Fprint(fo, "    <title>Export from ")
	xml.EscapeText(fo, fmt.Appendf(nil, "cleed/v%s (github.com/radulucut/cleed)", f.version))
	fmt.Fprint(fo, "</title>\n")
	fmt.Fprintf(fo, "    <dateCreated>%s</dateCreated>\n", f.time.Now().Format(time.RFC1123Z))
	fmt.Fprint(fo, "  </head>\n  <body>\n")

	var err error
	lists := make([]string, 0)
	if list == "" {
		lists, err = f.storage.LoadLists()
		if err != nil {
			return nil, utils.NewInternalError("failed to load lists: " + err.Error())
		}
		if len(lists) == 0 {
			lists = append(lists, "default")
		}
	} else {
		lists = append(lists, list)
	}

	feedCount := int64(0)
	for i := range lists {
		feeds, err := f.storage.GetFeedsFromList(lists[i])
		if err != nil {
			return nil, utils.NewInternalError("failed to list feeds: " + err.Error())
		}
		if len(feeds) == 0 {
			continue
		}
		fmt.Fprint(fo, "    <outline text=\"")
		xml.EscapeText(fo, []byte(lists[i]))
		fmt.Fprint(fo, "\">\n")
		for _, item := range feeds {
			feed, err := f.parseFeed(item.Address)
			if cachedOnly && feed == nil {
				continue
			}
			fmt.Fprint(fo, "      <outline")
			if err == nil {
				if feed.Title != "" {
					fmt.Fprint(fo, " text=\"")
					xml.EscapeText(fo, []byte(strings.TrimSpace(feed.Title)))
					fmt.Fprint(fo, "\"")
				}
				if feed.Description != "" {
					feed.Description = strings.TrimSpace(feed.Description)
					fmt.Fprint(fo, " description=\"")
					xml.EscapeText(fo, []byte(feed.Description))
					fmt.Fprint(fo, "\"")
				}
			}
			fmt.Fprint(fo, " xmlUrl=\"")
			xml.EscapeText(fo, []byte(item.Address))
			fmt.Fprint(fo, "\" />\n")
			feedCount++
		}
		fmt.Fprint(fo, "    </outline>\n")
	}
	fmt.Fprint(fo, "  </body>\n</opml>")

	return &OPMLExportResult{
		FeedCount: feedCount,
		ListCount: int64(len(lists)),
	}, nil
}

func (f *TerminalFeed) importOPML(opml *utils.OPML, list string) error {
	if len(opml.Body.Outltines) == 0 {
		return utils.NewInternalError("no feeds found in OPML")
	}
	for _, listOutline := range opml.Body.Outltines {
		urls := make([]string, 0, len(listOutline.Outlines))
		for _, feedOutline := range listOutline.Outlines {
			urls = append(urls, feedOutline.XMLURL)
		}
		if len(urls) == 0 {
			continue
		}
		listName := list
		if list == "" {
			if listOutline.Text != "" {
				listName = listOutline.Text
			} else {
				listName = "default"
			}
		}
		err := f.storage.AddToList(urls, listName)
		if err != nil {
			return utils.NewInternalError("failed to save feeds: " + err.Error())
		}
		f.printer.Printf("added %s to list: %s\n", utils.Pluralize(int64(len(urls)), "feed"), listName)
	}
	return nil
}
