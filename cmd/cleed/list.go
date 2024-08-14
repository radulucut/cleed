package cleed

import "github.com/spf13/cobra"

func (r *Root) initList() {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "Show all lists or feeds in a list",
		Long: `Show all lists or feeds in a list

Examples:
  # Show all lists
  cleed list

  # Show all feeds in a list
  cleed list mylist
`,

		RunE: r.RunList,
		Args: cobra.MaximumNArgs(1),
	}

	r.Cmd.AddCommand(cmd)
}

func (r *Root) RunList(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return r.feed.Lists()
	}
	return r.feed.ListFeeds(args[0])
}
