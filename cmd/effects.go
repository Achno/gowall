/*
Copyright © 2025 Achno <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"strings"

	"github.com/Achno/gowall/internal/image"
	"github.com/Achno/gowall/utils"
	"github.com/spf13/cobra"
)

var factor float64

var effectsCmd = &cobra.Command{
	Use:   "effects [effect]",
	Short: "Apply various effects to your images",
	Long:  `Apply various effects to your images like flip,mirror,grayscale,br(brightness),and more`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 2 {
			fmt.Println("Error: requires 1 command and 1arg(s), only received 0")
			_ = cmd.Usage()
			showAvailableEffects()
			return
		}
		switch strings.ToLower(args[0]) {

		case "flip":
			fmt.Println("Processing image...")
			processor := &image.FlipProcessor{}
			expandFile := utils.ExpandHomeDirectory(args)
			path, _, err := image.ProcessImg(expandFile[1], processor, shared.Theme)

			utils.HandleError(err)
			err = image.OpenImage(path)
			utils.HandleError(err)

		case "mirror":
			fmt.Println("Processing image...")
			processor := &image.MirrorProcessor{}
			expandFile := utils.ExpandHomeDirectory(args)
			path, _, err := image.ProcessImg(expandFile[1], processor, shared.Theme)

			utils.HandleError(err)
			err = image.OpenImage(path)
			utils.HandleError(err)

		case "grayscale":
			fmt.Println("Processing image...")
			processor := &image.GrayScaleProcessor{}
			expandFile := utils.ExpandHomeDirectory(args)
			path, _, err := image.ProcessImg(expandFile[1], processor, shared.Theme)

			utils.HandleError(err)
			err = image.OpenImage(path)
			utils.HandleError(err)

		case "br":
			fmt.Println("Processing image...")
			processor := &image.BrightnessProcessor{Factor: factor}
			expandFile := utils.ExpandHomeDirectory(args)
			path, _, err := image.ProcessImg(expandFile[1], processor, shared.Theme)

			utils.HandleError(err)
			err = image.OpenImage(path)
			utils.HandleError(err)

		default:
			fmt.Println("Error: requires at least 1 arg(s), only received 0")
			_ = cmd.Usage()
			showAvailableEffects()
		}
	},
}

func showAvailableEffects() {
	fmt.Println("\nAvailable Effects:")
	fmt.Println("  flip       Flips the image horizontally")
	fmt.Println("  mirror     Mirrors the image horizontally")
	fmt.Println("  grayscale  Converts image to grayscale (shades of gray)")
	fmt.Println("  br         Increases/Decreases the brightness")
}

func init() {
	rootCmd.AddCommand(effectsCmd)
	effectsCmd.Flags().Float64VarP(&factor, "factor", "f", 1.1, "1.2 increases brightness by 20%, 0.8 decreases brightness by 20%. Default 1.1")
}
