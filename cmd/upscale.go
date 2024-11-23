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

var upscaleCmd = &cobra.Command{
	Use:   "upscale [PATH]",
	Short: "[TODO]",
	Long:  `[TODO]`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("upscale called")
		switch {

		case len(args) > 0:
			fmt.Println("Upscaling image...")
			processor := &image.UpscaleProcessor{}
			expandFile := utils.ExpandHomeDirectory(args)

			opts := image.ProcessOptions{
				SaveToFile: false,
			}
			_, _, err := image.ProcessImg(expandFile[0], processor, shared.Theme, opts)
			utils.HandleError(err, "Error Processing Image")

		// err = image.OpenImage(path)
		// utils.HandleError(err, "Error opening image")

		default:
			fmt.Println("Error: requires at least 1 arg(s), only received 0")
			_ = cmd.Usage()
		}
	},
}

func init() {
	rootCmd.AddCommand(upscaleCmd)
}
