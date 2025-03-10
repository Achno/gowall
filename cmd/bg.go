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
	Use:   "bg [Input]",
	Short: "Removes the background of the image",
	Long:  `Removes the background of an image. You can modify the options to achieve better results`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if len(args) > 0 || len(shared.InputDir) > 0 || len(shared.InputFiles) > 0 {
			return
		}
		logger.Error("Error: no input was given, use commands args, or --dir or --batch flags")
		_ = cmd.Usage()
	},
	Run: func(cmd *cobra.Command, args []string) {
		imageOps := imageio.DetermineImageOperations(shared, args)
		logger.Print("Removing background...")
		processor := &image.BackgroundProcessor{}
		processor.SetOptions(
			image.WithConvergence(convergence),
			image.WithMaxIter(maxIter),
			image.WithNumRoutines(numRoutines),
			image.WithSampleRate(sampleRate),
		)
		paths, err := image.ProcessImgs(processor, imageOps)
		// Only crash when we couldn't proccess any images
		if len(paths) == 0 {
			utils.HandleError(err, "Error Processing Images")
		}
		// Otherwise print an error message for the unprocessed images
		if err != nil {
			logger.Errorf("the following images had errors while proccesing", err)
		}
		// We should only open images when we are dealing with one image and without
		// stdout as ouput
		err = image.OpenImageInViewer(paths[0])
		logger.Errorf("error opening image", err)
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
