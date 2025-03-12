/*
Copyright © 2025 Achno <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/Achno/gowall/internal/image"
	imageio "github.com/Achno/gowall/internal/image_io"
	"github.com/Achno/gowall/internal/logger"
	"github.com/Achno/gowall/utils"
	"github.com/spf13/cobra"
)

var factor float64

var effectsCmd = &cobra.Command{
	Use:   "effects",
	Short: "Apply various effects to your images",
	Long:  `Apply various effects to your images like flip,mirror,grayscale,br(brightness),and more`,
	Run: func(cmd *cobra.Command, args []string) {
		logger.Print("Please specify an effect to apply")
		err := cmd.Usage()
		utils.HandleError(err)
	},
}

var flipCmd = &cobra.Command{
	Use:   "flip [image]",
	Short: "Flips the image horizontally",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		err := validateInput(shared, args)
		if err != nil {
			cmd.Usage()
			return err
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		logger.Print("Processing image...")
		processor := &image.FlipProcessor{}
		imageOps := imageio.DetermineImageOperations(shared, args)
		processedImages, err := image.ProcessImgs(processor, imageOps, "")
		if len(processedImages) == 0 {
			utils.HandleError(err, "Error Processing Images")
		}
		if err != nil {
			logger.Error(err, "The following images had errors while processing")
		}
		openImageInViewer(shared, args, processedImages[0])
	},
}

var mirrorCmd = &cobra.Command{
	Use:   "mirror [image]",
	Short: "Mirrors the image horizontally",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		err := validateInput(shared, args)
		if err != nil {
			cmd.Usage()
			return err
		}
		return nil
	},

	Run: func(cmd *cobra.Command, args []string) {
		logger.Print("Processing image...")
		processor := &image.MirrorProcessor{}
		imageOps := imageio.DetermineImageOperations(shared, args)
		processedImages, err := image.ProcessImgs(processor, imageOps, "")
		if len(processedImages) == 0 {
			utils.HandleError(err, "Error Processing Images")
		}
		if err != nil {
			logger.Error(err, "The following images had errors while processing")
		}
		openImageInViewer(shared, args, processedImages[0])
	},
}

var grayscaleCmd = &cobra.Command{
	Use:   "grayscale [image]",
	Short: "Converts image to grayscale (shades of gray)",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		err := validateInput(shared, args)
		if err != nil {
			cmd.Usage()
			return err
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		logger.Print("Processing image...")
		processor := &image.GrayScaleProcessor{}
		imageOps := imageio.DetermineImageOperations(shared, args)
		processedImages, err := image.ProcessImgs(processor, imageOps, "")
		if len(processedImages) == 0 {
			utils.HandleError(err, "Error Processing Images")
		}
		if err != nil {
			logger.Error(err, "The following images had errors while processing")
		}
		openImageInViewer(shared, args, processedImages[0])
	},
}

var brightnessCmd = &cobra.Command{
	Use:   "br [image]",
	Short: "Increases/Decreases the brightness",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		err := validateInput(shared, args)
		if err != nil {
			cmd.Usage()
			return err
		}
		return nil
	},

	Run: func(cmd *cobra.Command, args []string) {
		logger.Print("Processing image...")
		processor := &image.BrightnessProcessor{Factor: factor}
		imageOps := imageio.DetermineImageOperations(shared, args)
		processedImages, err := image.ProcessImgs(processor, imageOps, "")
		if len(processedImages) == 0 {
			utils.HandleError(err, "Error Processing Images")
		}
		if err != nil {
			logger.Error(err, "The following images had errors while processing")
		}
		openImageInViewer(shared, args, processedImages[0])
	},
}

func init() {
	rootCmd.AddCommand(effectsCmd)

	effectsCmd.AddCommand(flipCmd)
	effectsCmd.AddCommand(mirrorCmd)
	effectsCmd.AddCommand(grayscaleCmd)
	effectsCmd.AddCommand(brightnessCmd)
	brightnessCmd.Flags().Float64VarP(&factor, "factor", "f", 1.1, "1.2 increases brightness by 20%, 0.8 decreases brightness by 20%. Default 1.1")
	addGlobalFlags(effectsCmd)
}
