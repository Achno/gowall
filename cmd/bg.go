/*
Copyright © 2024 Achno <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"github.com/Achno/gowall/config"
	"github.com/Achno/gowall/internal/image"
	imageio "github.com/Achno/gowall/internal/image_io"
	"github.com/Achno/gowall/internal/logger"
	"github.com/Achno/gowall/utils"
	"github.com/spf13/cobra"
)

func BuildBgCmd() *cobra.Command {
	methods := image.GetBgStrategyNames()

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
		method      string
		maxIter     int
		convergence float64
		sampleRate  float64
		numRoutines int
	)

	flags.StringVarP(&method, "method", "m", "kmeans", "Background removal method. Available methods: "+fmt.Sprint(methods))
	flags.IntVarP(&maxIter, "iterations", "i", 100, "Maximum iterations for background removal")
	flags.IntVarP(&numRoutines, "routines", "r", 4, "Number of goroutines to use")
	flags.Float64VarP(&convergence, "conv", "c", 0.001, "Convergence threshold")
	flags.Float64VarP(&sampleRate, "sRate", "s", 0.5, "Sample rate")

	addGlobalFlags(cmd)

	return cmd
}

func RunBgCmd(cmd *cobra.Command, args []string) {
	// Force png as the output for transparency for now
	//TODO : figure out why jpeg,webp are not preserving transparency
	shared.Format = "png"
	imageOps, err := imageio.DetermineImageOperations(shared, args, cmd)
	utils.HandleError(err, "Error")

	method, err := cmd.Flags().GetString("method")
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

	strategy, cleanup, err := image.GetBgStrategy(method, maxIter, convergence, sampleRate, numRoutines)
	utils.HandleError(err, "Error")
	if cleanup != nil {
		defer cleanup()
	}

	processor := image.NewBackgroundProcessor(strategy)

	processedImages, err := image.ProcessImgs(processor, imageOps, image.ProcessOptions{
		Theme:      "",
		OnComplete: nil,
	})
	utils.HandleError(err, "Error")

	if err != nil {
		logger.Error(err, "The following images had errors while processing")
	}

	openImageInViewer(cmd, shared, args, processedImages[0])
}

func ValidateParseBgCmd(cmd *cobra.Command, flags config.GlobalSubCommandFlags, args []string) error {
	if err := validateInput(flags, args); err != nil {
		return err
	}

	method, err := cmd.Flags().GetString("method")
	if err != nil {
		return err
	}

	if !image.IsValidBgStrategy(method) {
		return fmt.Errorf("invalid background removal method %q", method)
	}

	if method == "bria-rmbg" && isInputBatch(flags) {
		prompt := fmt.Sprintf(
			`%s ◈ Buddy are you sure you want to run --dir or --batch with bria-rmbg? Dont be surprised when your CPU spikes to 100%% and your screen freezes in the next 3 seconds when you select 'y'.I recommend just using it on one image at a time. %s`,
			utils.BlueColor,
			utils.ResetColor,
		)
		if !utils.Confirm(prompt) {
			return fmt.Errorf("background removal declined by user")
		}
	}

	return nil
}

func init() {
	rootCmd.AddCommand(BuildBgCmd())
}
