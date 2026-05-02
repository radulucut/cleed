package cleed

import (
	"fmt"
	"net/url"
	"os"
	"time"

	"github.com/radulucut/cleed/internal"
	"github.com/radulucut/cleed/internal/storage"
	"github.com/radulucut/cleed/internal/utils"
	"github.com/spf13/cobra"
)

var Version string

func Execute() {
	time := utils.NewTime()
	printer := internal.NewPrinter(os.Stdin, os.Stdout, os.Stderr)
	storage := storage.NewLocalStorage("cleed", time)
	feed := internal.NewTerminalFeed(time, printer, storage)
	feed.SetDefaultExploreRepository("https://github.com/radulucut/cleed-explore.git")
	root, err := NewRoot(Version, time, printer, storage, feed)
	if err != nil {
		printer.ErrPrintf("Error: %v\n", err)
		os.Exit(1)
	}

	err = root.Cmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

type Root struct {
	Cmd *cobra.Command

	version string

	time    utils.Time
	printer *internal.Printer
	storage *storage.LocalStorage
	feed    *internal.TerminalFeed
}

func NewRoot(
	version string,
	time utils.Time,
	printer *internal.Printer,
	storage *storage.LocalStorage,
	feed *internal.TerminalFeed,
) (*Root, error) {
	root := &Root{
		version: version,
		time:    time,
		printer: printer,
		storage: storage,
		feed:    feed,
	}
	err := root.storage.Init(root.version)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize storage: %v", err)
	}

	feed.SetVersion(root.version)

	config, err := root.storage.LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %v", err)
	}
	if config.Styling != 0 {
		root.printer.SetStyling(config.Styling == 1)
	}
	if config.UserAgent == "" {
		config.UserAgent = fmt.Sprintf("cleed/v%s (github.com/radulucut/cleed)", root.version)
	}

	root.Cmd = &cobra.Command{
		Use:   "cleed",
		Short: "A command line feed reader",
		Long: `A command line feed reader

Examples:
  # Display feeds from all lists
  cleed

  # Display feeds from a specific list
  cleed --list my-list

  # Display feeds since a specific date
  cleed --since "2024-01-01 12:03:04"

  # Display feeds from the last 1 day
  cleed --since 1d

  # Display feeds since the last run
  cleed --since last

  # Display feeds from a specific list and limit the number of items
  cleed --list my-list --limit 10

  # Search for items
  cleed --search "keyword" --limit 10

  # Search for items in cached feeds
  cleed --search "keyword" -C

  # Search for items using a regular expression (https://github.com/google/re2/wiki/Syntax)
  cleed --searchr "^keyword"

  # Search for items using a regular expression case insensitive
  cleed --searchr "(?i)keyword"

  # Using a proxy
  cleed --proxy socks5://user:password@proxy.example.com:8080

  # One tab-separated line per item; times in ISO-8601 (RFC3339)
  cleed --raw
`,
		Version: version,
		RunE:    root.RunRoot,
	}

	root.Cmd.SetOut(root.printer.OutWriter)
	root.Cmd.SetErr(root.printer.ErrWriter)

	flags := root.Cmd.Flags()
	flags.StringP("list", "L", "", "list to display feeds from")
	flags.Uint("limit", 50, "limit the number of items to display")
	flags.String("since", "", "display feeds since the last run (last), a specific date (e.g. 2024-01-01 12:03:04) or duration (e.g. 1d)")
	flags.String("search", "", "search for items (title, categories)")
	flags.String("searchr", "", "search for items using regular expressions (title)")
	flags.String("proxy", "", "proxy to use for requests")
	flags.BoolP("cached-only", "C", false, "display or search only from cached feeds")
	flags.BoolP("raw", "r", false, "print each item on its own line: tab-separated fields (published, feed title, feed link, feed description, item title, item description, item link, categories, guid); HTML stripped from titles, descriptions, and categories; timestamps UTC RFC3339")
	flags.Bool("config-path", false, "show the path to the config directory")
	flags.Bool("cache-path", false, "show the path to the cache directory")
	flags.Bool("cache-info", false, "show the cache information")

	root.initVersion()
	root.initFollow()
	root.initUnfollow()
	root.initList()
	root.initConfig()
	root.initExplore()
	root.initMiniflux()

	return root, nil
}

func (r *Root) RunRoot(cmd *cobra.Command, args []string) error {
	if cmd.Flag("config-path").Changed {
		return r.feed.ShowConfigPath()
	}
	if cmd.Flag("cache-path").Changed {
		return r.feed.ShowCachePath()
	}
	if cmd.Flag("cache-info").Changed {
		return r.feed.ShowCacheInfo()
	}
	limit, err := cmd.Flags().GetUint("limit")
	if err != nil {
		return err
	}
	since, err := r.parseSinceFlag(cmd.Flag("since").Value.String())
	if err != nil {
		return err
	}
	opts := &internal.FeedOptions{
		List:  cmd.Flag("list").Value.String(),
		Limit: int(limit),
		Since: since,
	}
	opts.CachedOnly, err = cmd.Flags().GetBool("cached-only")
	if err != nil {
		return err
	}
	opts.Raw, err = cmd.Flags().GetBool("raw")
	if err != nil {
		return err
	}
	proxy := cmd.Flag("proxy").Value.String()
	if proxy != "" {
		url, err := url.Parse(proxy)
		if err != nil {
			return fmt.Errorf("failed to parse proxy URL: %v", err)
		}
		opts.Proxy = url
	}
	if cmd.Flag("search").Changed {
		return r.feed.Search(cmd.Flag("search").Value.String(), opts)
	}
	if cmd.Flag("searchr").Changed {
		return r.feed.SearchRegexp(cmd.Flag("searchr").Value.String(), opts)
	}
	return r.feed.Feed(opts)
}

func (r *Root) parseSinceFlag(flag string) (time.Time, error) {
	if flag == "" {
		return time.Time{}, nil
	}
	if flag == "last" {
		config, err := r.storage.LoadConfig()
		if err != nil {
			return time.Time{}, fmt.Errorf("failed to load config: %v", err)
		}
		return config.LastRun, nil
	}
	d, err := utils.ParseDuration(flag)
	if err == nil {
		return r.time.Now().Add(-d), nil
	}
	return utils.ParseDateTime(flag)
}
