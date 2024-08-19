package cleed

import (
	"github.com/spf13/cobra"
)

func (r *Root) initConfig() {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Display or change configuration",
		Long: `Display or change configuration

Examples:
  # Display configuration
  cleed config

  # Disable styling
  cleed config --styling=2

  # Map color 0 to 230 and color 1 to 213
  cleed config --map-colors=0:230,1:213

  # Remove color mapping for color 0
  cleed config --map-colors=0:

  # Clear all color mappings
  cleed config --map-colors=

  # Display color range. Useful for finding colors to map
  cleed config --color-range

  # Enable run summary
  cleed config --summary=1
`,
		RunE: r.RunConfig,
	}

	flags := cmd.Flags()
	flags.Uint8("styling", 0, "disable or enable styling (0: default, 1: enable, 2: disable)")
	flags.Uint8("summary", 0, "disable or enable summary (0: disable, 1: enable)")
	flags.String("map-colors", "", "map colors to other colors, e.g. 0:230,1:213. Use --color-range to check available colors")
	flags.Bool("color-range", false, "display color range. Useful for finding colors to map")

	r.Cmd.AddCommand(cmd)
}

func (r *Root) RunConfig(cmd *cobra.Command, args []string) error {
	if cmd.Flag("styling").Changed {
		styling, err := cmd.Flags().GetUint8("styling")
		if err != nil {
			return err
		}
		return r.feed.SetStyling(styling)
	}
	if cmd.Flag("summary").Changed {
		summary, err := cmd.Flags().GetUint8("summary")
		if err != nil {
			return err
		}
		return r.feed.SetSummary(summary)
	}
	if cmd.Flag("map-colors").Changed {
		return r.feed.UpdateColorMap(cmd.Flag("map-colors").Value.String())
	}
	if cmd.Flag("color-range").Changed {
		r.feed.DisplayColorRange()
		return nil
	}
	return r.feed.DisplayConfig()
}
