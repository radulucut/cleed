package cleed

import (
	"fmt"
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
	feed.SetAgent(fmt.Sprintf("cleed/v%s (github.com/radulucut/cleed)", Version))
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

	config, err := root.storage.LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %v", err)
	}
	if config.Styling != 0 {
		root.printer.SetStyling(config.Styling == 1)
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

  # Display feeds from a specific list and limit the number of feeds
  cleed --list my-list --limit 10
`,
		Version: version,
		RunE:    root.RunRoot,
	}

	root.Cmd.SetOut(root.printer.OutWriter)
	root.Cmd.SetErr(root.printer.ErrWriter)

	flags := root.Cmd.Flags()
	flags.StringP("list", "L", "", "list to display feeds from")
	flags.Uint("limit", 50, "limit the number of feeds to display")
	flags.String("since", "", "display feeds since the last run (last), a specific date (e.g. 2024-01-01 12:03:04) or duration (e.g. 1d)")

	root.initVersion()
	root.initFollow()
	root.initUnfollow()
	root.initList()
	root.initConfig()

	return root, nil
}

func (r *Root) RunRoot(cmd *cobra.Command, args []string) error {
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
