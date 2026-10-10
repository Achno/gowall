package cmd

import (
	imageio "github.com/Achno/gowall/internal/image_io"
	"github.com/Achno/gowall/terminal"
	"github.com/Achno/gowall/utils"
	"github.com/spf13/cobra"
)

func BuildCompareCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "compare [ORIGINAL] [EDITED]",
		Short: "Compare two images with an interactive before/after slider",
		Long:  `Compare two images with an interactive before/after slider, move it with the arrow keys or by dragging with the mouse. Use --side to place them side by side instead. Needs a terminal with the kitty image protocol`,
		Args:  cobra.ExactArgs(2),
		Run:   RunCompareCmd,
	}

	var sideBySide bool
	cmd.Flags().BoolVarP(&sideBySide, "side", "s", false, "Place the images side by side instead of the slider")

	return cmd
}

func RunCompareCmd(cmd *cobra.Command, args []string) {
	original, err := imageio.LoadImage(imageio.FileReader{Path: args[0]})
	utils.HandleError(err, "Error")

	edited, err := imageio.LoadImage(imageio.FileReader{Path: args[1]})
	utils.HandleError(err, "Error")

	sideBySide, err := cmd.Flags().GetBool("side")
	utils.HandleError(err, "Error")

	if sideBySide {
		err = terminal.RenderKittySideBySide(original, edited)
	} else {
		err = terminal.RenderKittySlider(original, edited)
	}
	utils.HandleError(err, "Error")
}

func init() {
	rootCmd.AddCommand(BuildCompareCmd())
}
