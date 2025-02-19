/*
Copyright © 2024 Achnologia <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"image/color"

	"github.com/Achno/gowall/config"
	"github.com/Achno/gowall/internal/backends/colorthief"
	"github.com/Achno/gowall/internal/image"
	"github.com/Achno/gowall/utils"
	"github.com/spf13/cobra"
)

var colorsNum int
var previewFlag bool

// extractCmd represents the extract command
var extractCmd = &cobra.Command{
	Use:   "extract [FILE]",
	Short: "Returns the color palette of the image you specified (like pywal)",
	Long:  `Using the colorthief backend ( like pywal ) it returns the color palette of the image (path) you specified`,
	Run: func(cmd *cobra.Command, args []string) {

		switch {
		case len(args) > 0:
			expandFile := utils.ExpandHomeDirectory(args)
			clr, err := colorthief.GetPaletteFromFile(expandFile[0], colorsNum)
			utils.HandleError(err)

			for _, c := range clr {
				rgba, ok := c.(color.RGBA)

				if !ok {
					utils.HandleError(fmt.Errorf("error in RGB casting"))
				}
				fmt.Println(image.RGBtoHex(rgba))
			}

			// open up hex code preview site
			if previewFlag {
				utils.OpenURL(config.HexCodeVisualUrl)
			}

		default:
			fmt.Println("Error: requires at least 1 arg(s), only received 0")
			_ = cmd.Usage()
		}

	},
}

func init() {
	rootCmd.AddCommand(extractCmd)
	extractCmd.Flags().IntVarP(&colorsNum, "colors", "c", 6, "-c <number of colors to return>")
	extractCmd.Flags().BoolVarP(&previewFlag, "preview", "p", false, "gowall extract -p (opens hex code preview site)")

}
