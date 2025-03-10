/*
Copyright © 2024 Achno <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"github.com/Achno/gowall/internal/image"
	"github.com/Achno/gowall/utils"
	"github.com/spf13/cobra"
)

var ScaleFactor float64

var pixelateCmd = &cobra.Command{
	Use:   "pixelate [PATH]",
	Short: "Turns an image to pixel art depending on the scale flag",
	Long: `It can convert an image to pixel art (blocky appearance). The scale flag [1-25] controls how much the image will get pixelated. 
		   The lower the number the more pixel effect is prevalent. 
		   In really large images with huge resolution you may need to set the scale really low [3-8] `,
	Run: func(cmd *cobra.Command, args []string) {
		switch {

		case len(args) > 0:
			fmt.Println("Pixelating image...")
			processor := &image.PixelateProcessor{
				Scale: ScaleFactor,
			}
			expandFile := utils.ExpandHomeDirectory(args)

			path, _, err := image.ProcessImg(expandFile[0], processor, shared.Theme)
			utils.HandleError(err, "Error Processing Image")

			err = image.OpenImageInViewer(path)
			utils.HandleError(err, "Error opening image")

		default:
			fmt.Println("Error: requires at least 1 arg(s), only received 0")
			_ = cmd.Usage()
		}
	},
}

func init() {
	rootCmd.AddCommand(pixelateCmd)
	pixelateCmd.Flags().Float64VarP(&ScaleFactor, "scale", "s", 15, "Usage: --scale [1-25] (The lower the number == more pixelation)")
}
