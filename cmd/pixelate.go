/*
Copyright © 2024 Achno <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/Achno/gowall/internal/image"
	imageio "github.com/Achno/gowall/internal/image_io"
	"github.com/Achno/gowall/internal/logger"
	"github.com/Achno/gowall/utils"
	"github.com/spf13/cobra"
)

var ScaleFactor float64

var pixelateCmd = &cobra.Command{
	Use:   "pixelate [INPUT] [OPTIONAL OUTPUT]",
	Short: "Turns an image to pixel art depending on the scale flag",
	Long: `It can convert an image to pixel art (blocky appearance). The scale flag [1-25] controls how much the image will get pixelated. 
		   The lower the number the more pixel effect is prevalent. 
		   In really large images with huge resolution you may need to set the scale really low [3-8] `,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		err := validateInput(shared, args)
		if err != nil {
			return err
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		logger.Print("Pixelating image...")

		processor := &image.PixelateProcessor{
			Scale: ScaleFactor,
		}

		imageOps, err := imageio.DetermineImageOperations(shared, args, cmd)
		utils.HandleError(err)

		processedImages, err := image.ProcessImgs(processor, imageOps, "")
		utils.HandleError(err, "Error")

		// if len(processedImages) == 0 {
		// 	utils.HandleError(err, "Error Processing Images")
		// }

		if err != nil {
			logger.Error(err, "The following images had errors while processing")
		}

		openImageInViewer(shared, args, processedImages[0])
	},
}

func init() {
	rootCmd.AddCommand(pixelateCmd)
	pixelateCmd.Flags().Float64VarP(&ScaleFactor, "scale", "s", 15, "Usage: --scale [1-25] (The lower the number == more pixelation)")
	addGlobalFlags(pixelateCmd)
}
