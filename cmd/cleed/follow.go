package main

import (
	"github.com/spf13/cobra"
)

func (r *Root) initFollow() {
	cmd := &cobra.Command{
		Use:   "follow [feed]",
		Short: "Follow a feed",
		Long: `Follow a feed

Examples:
  # Add a feed to the default list
  cleed follow https://example.com/feed.xml

  # Add multiple feeds to a list
  cleed follow https://example.com/feed.xml https://example2.com/feed --list mylist
`,
		RunE: r.RunFollow,
		Args: cobra.MinimumNArgs(1),
	}

	flags := cmd.Flags()
	flags.StringP("list", "L", "default", "the list to add the feed to")

	r.Cmd.AddCommand(cmd)
}

func (r *Root) RunFollow(cmd *cobra.Command, args []string) error {
	list, err := cmd.Flags().GetString("list")
	if err != nil {
		return err
	}
	return r.feed.Follow(args, list)
}
