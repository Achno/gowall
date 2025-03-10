/*
Copyright © 2025 Achno <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"github.com/Achno/gowall/internal/image"
	"github.com/Achno/gowall/utils"
	"github.com/spf13/cobra"
)

var factor float64

var effectsCmd = &cobra.Command{
	Use:   "effects",
	Short: "Apply various effects to your images",
	Long:  `Apply various effects to your images like flip,mirror,grayscale,br(brightness),and more`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Please specify an effect to apply")
		err := cmd.Usage()
		utils.HandleError(err)
	},
}

var flipCmd = &cobra.Command{
	Use:   "flip [image]",
	Short: "Flips the image horizontally",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Processing image...")
		processor := &image.FlipProcessor{}
		expandFile := utils.ExpandHomeDirectory(args)
		path, _, err := image.ProcessImg(expandFile[0], processor, shared.Theme)
		utils.HandleError(err)
		err = image.OpenImageInViewer(path)
		utils.HandleError(err)
	},
}

var mirrorCmd = &cobra.Command{
	Use:   "mirror [image]",
	Short: "Mirrors the image horizontally",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Processing image...")
		processor := &image.MirrorProcessor{}
		expandFile := utils.ExpandHomeDirectory(args)
		path, _, err := image.ProcessImg(expandFile[0], processor, shared.Theme)
		utils.HandleError(err)
		err = image.OpenImageInViewer(path)
		utils.HandleError(err)
	},
}

var grayscaleCmd = &cobra.Command{
	Use:   "grayscale [image]",
	Short: "Converts image to grayscale (shades of gray)",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Processing image...")
		processor := &image.GrayScaleProcessor{}
		expandFile := utils.ExpandHomeDirectory(args)
		path, _, err := image.ProcessImg(expandFile[0], processor, shared.Theme)
		utils.HandleError(err)
		err = image.OpenImageInViewer(path)
		utils.HandleError(err)
	},
}

var brightnessCmd = &cobra.Command{
	Use:   "br [image]",
	Short: "Increases/Decreases the brightness",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Processing image...")
		processor := &image.BrightnessProcessor{Factor: factor}
		expandFile := utils.ExpandHomeDirectory(args)
		path, _, err := image.ProcessImg(expandFile[0], processor, shared.Theme)
		utils.HandleError(err)
		err = image.OpenImageInViewer(path)
		utils.HandleError(err)
	},
}

func init() {
	rootCmd.AddCommand(effectsCmd)

	effectsCmd.AddCommand(flipCmd)
	effectsCmd.AddCommand(mirrorCmd)
	effectsCmd.AddCommand(grayscaleCmd)
	effectsCmd.AddCommand(brightnessCmd)

	brightnessCmd.Flags().Float64VarP(&factor, "factor", "f", 1.1, "1.2 increases brightness by 20%, 0.8 decreases brightness by 20%. Default 1.1")
}
