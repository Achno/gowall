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

var (
	maxIter     int
	convergence float64
	sampleRate  float64
	numRoutines int
)

// bgCmd represents the bg command
var bgCmd = &cobra.Command{
	Use:   "bg [INPUT] [OPTIONAL OUTPUT]",
	Short: "Removes the background of the image",
	Long:  `Removes the background of an image. You can modify the options to achieve better results`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		err := validateInput(shared, args)
		if err != nil {
			return err
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		imageOps, err := imageio.DetermineImageOperations(shared, args, cmd)
		utils.HandleError(err)
		logger.Print("Removing background...")
		processor := &image.BackgroundProcessor{}
		processor.SetOptions(
			image.WithConvergence(convergence),
			image.WithMaxIter(maxIter),
			image.WithNumRoutines(numRoutines),
			image.WithSampleRate(sampleRate),
		)
		processedImages, err := image.ProcessImgs(processor, imageOps, "")
		utils.HandleError(err, "Error")
		// Only crash when we couldn't process any images
		// if len(processedImages) == 0 {
		// 	utils.HandleError(err, "Error Processing Images")
		// }
		// Otherwise print an error message for the unprocessed images
		if err != nil {
			logger.Error(err, "The following images had errors while processing")
		}
		// Open images only when we are in single image mode
		openImageInViewer(shared, args, processedImages[0])
	},
}

func init() {
	rootCmd.AddCommand(bgCmd)
	bgCmd.Flags().IntVarP(&maxIter, "iterations", "i", 100, "")
	bgCmd.Flags().IntVarP(&numRoutines, "routines", "r", 4, "")
	bgCmd.Flags().Float64VarP(&convergence, "conv", "c", 0.001, "")
	bgCmd.Flags().Float64VarP(&sampleRate, "sRate", "s", 0.5, "")
	addGlobalFlags(bgCmd)
}
