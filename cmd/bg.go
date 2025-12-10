/*
Copyright © 2024 Achno <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/Achno/gowall/config"
	"github.com/Achno/gowall/internal/image"
	imageio "github.com/Achno/gowall/internal/image_io"
	"github.com/Achno/gowall/internal/logger"
	"github.com/Achno/gowall/utils"
	"github.com/spf13/cobra"
)

func BuildBgCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "bg [INPUT]",
		Short: "Removes the background of the image",
		Long:  `Removes the background of an image. You can modify the options to achieve better results`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return ValidateParseBgCmd(cmd, shared, args)
		},
		Run: RunBgCmd,
	}

	flags := cmd.Flags()
	var (
		maxIter     int
		convergence float64
		sampleRate  float64
		numRoutines int
	)

	flags.IntVarP(&maxIter, "iterations", "i", 100, "Maximum iterations for background removal")
	flags.IntVarP(&numRoutines, "routines", "r", 4, "Number of goroutines to use")
	flags.Float64VarP(&convergence, "conv", "c", 0.001, "Convergence threshold")
	flags.Float64VarP(&sampleRate, "sRate", "s", 0.5, "Sample rate")

	addGlobalFlags(cmd)

	return cmd
}

func RunBgCmd(cmd *cobra.Command, args []string) {
	imageOps, err := imageio.DetermineImageOperations(shared, args, cmd)
	utils.HandleError(err, "Error")

	maxIter, err := cmd.Flags().GetInt("iterations")
	utils.HandleError(err, "Error")
	numRoutines, err := cmd.Flags().GetInt("routines")
	utils.HandleError(err, "Error")
	convergence, err := cmd.Flags().GetFloat64("conv")
	utils.HandleError(err, "Error")
	sampleRate, err := cmd.Flags().GetFloat64("sRate")
	utils.HandleError(err, "Error")

	logger.Print("Removing background...")

	processor := &image.BackgroundProcessor{}
	processor.SetOptions(
		image.WithConvergence(convergence),
		image.WithMaxIter(maxIter),
		image.WithNumRoutines(numRoutines),
		image.WithSampleRate(sampleRate),
	)

	processedImages, err := image.ProcessImgs(processor, imageOps, image.ProcessOptions{
		Theme:      "",
		OnComplete: nil,
	})
	utils.HandleError(err, "Error")

	if err != nil {
		logger.Error(err, "The following images had errors while processing")
	}

	openImageInViewer(shared, args, processedImages[0])
}

func ValidateParseBgCmd(cmd *cobra.Command, flags config.GlobalSubCommandFlags, args []string) error {
	if err := validateInput(flags, args); err != nil {
		return err
	}

	return nil
}

func init() {
	rootCmd.AddCommand(BuildBgCmd())
}
