package main

import (
	"fmt"
	"os"
	"time"

	"github.com/radulucut/cleed/internal"
	"github.com/radulucut/cleed/internal/storage"
	"github.com/radulucut/cleed/internal/utils"
	"github.com/spf13/cobra"
)

var version = "0.1.0"

func main() {
	time := utils.NewTime()
	printer := internal.NewPrinter(os.Stdin, os.Stdout, os.Stderr)
	storage := storage.NewLocalStorage("cleed", time)
	feed := internal.NewTerminalFeed(time, printer, storage)
	feed.SetAgent(fmt.Sprintf("cleed/v%s (github.com/radulucut/cleed)", version))
	root, err := NewRoot(version, time, printer, storage, feed)
	if err != nil {
		printer.ErrPrintf("Error: %v\n", err)
	}

	err = root.Cmd.Execute()
	if err != nil {
		printer.ErrPrintf("Error: %v\n", err)
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

	root.Cmd = &cobra.Command{
		Use:   "cleed",
		Short: "A command line feed reader",
		Long: `A command line feed reader

Examples:
  # Display feeds from all lists
  cleed

  # Display feeds from a specific list
  cleed --list my-list

  # Display feeds from the last 1 day
  cleed --last 1d

  # Display feeds since a specific date
  cleed --since "2024-01-01 12:03:04"

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
	flags.Int("limit", 50, "number of feeds to display")
	flags.String("last", "", "only display feeds from the last duration (e.g. 1d1h1m)")
	flags.String("since", "", "only display feeds since a specific date (e.g. 2024-01-01 12:03:04)")

	root.initVersion()
	root.initFollow()
	root.initUnfollow()
	root.initList()

	return root, nil
}

func (r *Root) RunRoot(cmd *cobra.Command, args []string) error {
	limit, err := cmd.Flags().GetInt("limit")
	if err != nil {
		return err
	}
	lastFlag := cmd.Flag("last").Value.String()
	sinceFlag := cmd.Flag("since").Value.String()
	since := time.Time{}
	if lastFlag != "" {
		d, err := utils.ParseDuration(lastFlag)
		if err != nil {
			return err
		}
		since = r.time.Now().Add(-d)
	} else if sinceFlag != "" {
		since, err = utils.ParseDateTime(sinceFlag)
		if err != nil {
			return err
		}
	}
	opts := &internal.FeedOptions{
		List:  cmd.Flag("list").Value.String(),
		Limit: limit,
		Since: since,
	}
	return r.feed.Feed(opts)
}
