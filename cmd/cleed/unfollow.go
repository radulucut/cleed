package main

import "github.com/spf13/cobra"

func (r *Root) initUnfollow() {
	cmd := &cobra.Command{
		Use:   "unfollow [feed]",
		Short: "Unfollow a feed",
		Long: `Unfollow a feed

Examples:
  # Remove a feed from the default list
  cleed unfollow https://example.com/feed.xml

  # Remove multiple feeds from a list
  cleed unfollow https://example.com/feed.xml https://example2.com/feed --list mylist
`,

		RunE: r.RunUnfollow,
		Args: cobra.MinimumNArgs(1),
	}

	flags := cmd.Flags()
	flags.StringP("list", "L", "default", "the list to remove the feed from")

	r.Cmd.AddCommand(cmd)
}

func (r *Root) RunUnfollow(cmd *cobra.Command, args []string) error {
	list, err := cmd.Flags().GetString("list")
	if err != nil {
		return err
	}
	return r.feed.Unfollow(args, list)
}
