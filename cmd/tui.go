package cmd

import (
	"github.com/Achno/gowall/config"
	"github.com/Achno/gowall/internal/logger"
	"github.com/Achno/gowall/internal/tui"
	"github.com/Achno/gowall/utils"
	"github.com/spf13/cobra"
)

var (
	tuiSplitFlag  bool
	tuiSocketFlag string
)

// BuildTuiCmd constructs the `gowall tui` interactive studio command.
func BuildTuiCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tui [PATH]",
		Short: "Launch the interactive Gowall Studio TUI",
		Long: `Launch an interactive terminal studio for live theme conversion, effects,
color palette extraction, and side-by-side full-resolution previews.

For full-resolution pixel rendering inside Alacritty (or other terminals without
native Kitty/Sixel graphics), run inside termimg:
  termimg gowall tui [PATH]`,
		Args: cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			logger.SetQuiet(true)
			defer logger.SetQuiet(false)

			var initialPath string
			if len(args) > 0 {
				expanded := config.ExpandTilde([]string{args[0]})
				if len(expanded) > 0 {
					initialPath = expanded[0]
				}
			} else if shared.InputDir != "" {
				initialPath = shared.InputDir
			}

			err := tui.Run(tui.Options{
				InitialPath:  initialPath,
				SplitPreview: tuiSplitFlag,
				SocketPath:   tuiSocketFlag,
			})
			utils.HandleError(err, "Error running Gowall TUI")
		},
	}

	cmd.Flags().BoolVarP(&tuiSplitFlag, "split", "s", true, "Start in side-by-side (Original vs Preview) comparison mode")
	cmd.Flags().StringVar(&tuiSocketFlag, "socket", "", "Unix socket path for Limoni UI automation (when built with -tags limoni_debug)")
	addFlags(cmd).WithDir().WithOutput()

	return cmd
}

func init() {
	rootCmd.AddCommand(BuildTuiCmd())
}
