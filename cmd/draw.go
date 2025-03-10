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

var (
	colorB          string
	BorderThickness int
)

var drawCmd = &cobra.Command{
	Use:   "draw [PATH] ",
	Short: "draw a border with a color and thickness (currently)",
	Long:  `The draw command allows you to draw a plethora of effects. Currently only drawing a border is supported with more to come`,
	Run: func(cmd *cobra.Command, args []string) {
		switch {
		case len(args) > 0:
			fmt.Println("Processing single image...")

			hex, err := cmd.Flags().GetString("color")
			utils.HandleError(err, "Error")

			clr, err := image.HexToRGBA(hex)
			utils.HandleError(err, "Error")

			processor := &image.DrawProcessor{
				Color:           clr,
				BorderThickness: BorderThickness,
			}
			expandFile := utils.ExpandHomeDirectory(args)

			path, _, err := image.ProcessImg(expandFile[0], processor, shared.Theme)
			utils.HandleError(err)

			err = image.OpenImageInViewer(path)
			utils.HandleError(err)

		default:
			fmt.Println("Error: requires at least 1 arg(s) and options, only received 0")
			_ = cmd.Usage()
		}
	},
}

func init() {
	rootCmd.AddCommand(drawCmd)
	drawCmd.Flags().StringVarP(&colorB, "color", "c", "#5D3FD3", "--color #5D3FD3")
	drawCmd.Flags().IntVarP(&BorderThickness, "borderThickness", "b", 5, "-b 5")
}
