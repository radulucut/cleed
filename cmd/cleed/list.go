package cleed

import (
	"github.com/spf13/cobra"
)

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

  # Rename a list
  cleed list mylist --rename newlist

  # Merge a list. Move all feeds from anotherlist to mylist and remove anotherlist
  cleed list mylist --merge anotherlist

  # Remove a list
  cleed list mylist --remove

  # Import feeds from a file
  cleed list mylist --import-from-file feeds.txt

  # Export feeds to a file
  cleed list mylist --export-to-file feeds.txt
`,

		RunE: r.RunList,
		Args: cobra.MaximumNArgs(1),
	}

	flags := cmd.Flags()
	flags.String("rename", "", "rename a list")
	flags.String("merge", "", "merge a list")
	flags.Bool("remove", false, "remove a list")
	flags.String("import-from-file", "", "import feeds from a file. Newline separated URLs")
	flags.String("export-to-file", "", "export feeds to a file. Newline separated URLs")

	r.Cmd.AddCommand(cmd)
}

func (r *Root) RunList(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return r.feed.Lists()
	}
	rename := cmd.Flag("rename").Value.String()
	if rename != "" {
		return r.feed.RenameList(args[0], rename)
	}
	merge := cmd.Flag("merge").Value.String()
	if merge != "" {
		return r.feed.MergeLists(args[0], merge)
	}
	if cmd.Flag("remove").Changed {
		return r.feed.RemoveList(args[0])
	}
	importFromFile := cmd.Flag("import-from-file").Value.String()
	if importFromFile != "" {
		return r.feed.ImportFromFile(importFromFile, args[0])
	}
	exportToFile := cmd.Flag("export-to-file").Value.String()
	if exportToFile != "" {
		return r.feed.ExportToFile(exportToFile, args[0])
	}
	return r.feed.ListFeeds(args[0])
}
