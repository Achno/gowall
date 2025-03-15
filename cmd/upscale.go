/*
Copyright © 2024 Achno <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"github.com/Achno/gowall/internal/image"
	imageio "github.com/Achno/gowall/internal/image_io"
	"github.com/Achno/gowall/internal/logger"
	"github.com/Achno/gowall/utils"
	"github.com/spf13/cobra"
)

// This function is tightly coupled to work with user input, so there's no way
// to use --stdout in this command as user input prompts pollute stdout.

var (
	scale     int
	modelName string
)

var upscaleCmd = &cobra.Command{
	Use:   "upscale [INPUT] [OPTIONAL OUTPUT]",
	Short: "Upscale (or Deblur) images using an Enhanced Super-Resolution Generative Adversarial Network, your GPU must support Vulkan",
	Long:  `Upscale images using an Enhanced Super-Resolution Generative Adversarial Network, your GPU must support Vulkan,if you sea black image after a lot of time then that means that you GPU does not support Vulkan. You can give options that specify thescale and Modelname`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		err := validateInput(shared, args)
		if err != nil {
			return err
		}
		if imageio.IsStdoutOutput(shared, args) {
			return fmt.Errorf("this command is not compatible with stdout output")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		logger.Print("Upscaling images...")
		processor := &image.UpscaleProcessor{
			Scale:     scale,
			ModelName: modelName,
		}

		imageOps := imageio.DetermineImageOperations(shared, args)
		processedImages, err := image.ProcessImgs(processor, imageOps, "")
		// Only crash when we couldn't process any images
		if len(processedImages) == 0 {
			utils.HandleError(err, "Error Processing Images")
		}
		// Otherwise print an error message for the unprocessed images
		if err != nil {
			logger.Error(err, "The following images had errors while processing")
		}
		// Open images only when we are in single image mode
		openImageInViewer(shared, args, processedImages[0])
	},
}

func init() {
	rootCmd.AddCommand(upscaleCmd)
	upscaleCmd.Flags().IntVarP(&scale, "scale", "s", 2, "Scale factor for upscaling (2, 3, or 4)")
	upscaleCmd.Flags().StringVarP(&modelName, "model", "m", "realesr-animevideov3",
		`Model to use for upscaling. Available models:
        realesrgan-x4plus (Slower,Better quality,forces -s 4),
        realesrgan-x4plus-anime (optimized for anime small),
		realesr-animevideov3 (Fast model ,animation video (default))`)
	addGlobalFlags(upscaleCmd)
}
