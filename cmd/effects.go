/*
Copyright © 2025 Achno <EMAIL ADDRESS>
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

func BuildEffectsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "effects [EFFECT] [INPUT] [OPTIONAL OUTPUT]",
		Short: "Apply various effects to your images",
		Long:  `Apply various effects to your images like flip,mirror,grayscale,br(brightness),and more`,
		Run: func(cmd *cobra.Command, args []string) {
			logger.Print("Please specify an effect to apply")
			err := cmd.Usage()
			utils.HandleError(err)
		},
	}

	cmd.AddCommand(BuildFlipCmd())
	cmd.AddCommand(BuildMirrorCmd())
	cmd.AddCommand(BuildGrayscaleCmd())
	cmd.AddCommand(BuildBrightnessCmd())

	addGlobalFlags(cmd)

	return cmd
}

// Flip Command
func BuildFlipCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "flip [INPUT] [OPTIONAL OUTPUT]",
		Short: "Flips the image horizontally",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return ValidateParseFlipCmd(cmd, shared, args)
		},
		Run: RunFlipCmd,
	}
	return cmd
}

func RunFlipCmd(cmd *cobra.Command, args []string) {
	logger.Print("Processing image...")

	imageOps, err := imageio.DetermineImageOperations(shared, args, cmd)
	utils.HandleError(err, "Error")

	processor := &image.FlipProcessor{}
	processedImages, err := image.ProcessImgs(processor, imageOps, "")
	utils.HandleError(err, "Error")

	openImageInViewer(shared, args, processedImages[0])
}

func ValidateParseFlipCmd(cmd *cobra.Command, flags config.GlobalSubCommandFlags, args []string) error {
	if err := validateInput(flags, args); err != nil {
		return err
	}
	return nil
}

// Mirror Command
func BuildMirrorCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mirror [INPUT] [OPTIONAL OUTPUT]",
		Short: "Mirrors the image horizontally",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return ValidateParseMirrorCmd(cmd, shared, args)
		},
		Run: RunMirrorCmd,
	}
	return cmd
}

func RunMirrorCmd(cmd *cobra.Command, args []string) {
	logger.Print("Processing image...")

	imageOps, err := imageio.DetermineImageOperations(shared, args, cmd)
	utils.HandleError(err, "Error")

	processor := &image.MirrorProcessor{}
	processedImages, err := image.ProcessImgs(processor, imageOps, "")
	utils.HandleError(err, "Error")

	openImageInViewer(shared, args, processedImages[0])
}

func ValidateParseMirrorCmd(cmd *cobra.Command, flags config.GlobalSubCommandFlags, args []string) error {
	if err := validateInput(flags, args); err != nil {
		return err
	}
	return nil
}

// Grayscale Command
func BuildGrayscaleCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "grayscale [INPUT] [OPTIONAL OUTPUT]",
		Short: "Converts image to grayscale (shades of gray)",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return ValidateParseGrayscaleCmd(cmd, shared, args)
		},
		Run: RunGrayscaleCmd,
	}
	return cmd
}

func RunGrayscaleCmd(cmd *cobra.Command, args []string) {
	logger.Print("Processing image...")

	imageOps, err := imageio.DetermineImageOperations(shared, args, cmd)
	utils.HandleError(err, "Error")

	processor := &image.GrayScaleProcessor{}
	processedImages, err := image.ProcessImgs(processor, imageOps, "")
	utils.HandleError(err, "Error")

	openImageInViewer(shared, args, processedImages[0])
}

func ValidateParseGrayscaleCmd(cmd *cobra.Command, flags config.GlobalSubCommandFlags, args []string) error {
	if err := validateInput(flags, args); err != nil {
		return err
	}
	return nil
}

// Brightness Command
func BuildBrightnessCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "br [INPUT] [OPTIONAL OUTPUT] [--flags]",
		Short: "Increases/Decreases the brightness",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return ValidateParseBrightnessCmd(cmd, shared, args)
		},
		Run: RunBrightnessCmd,
	}

	flags := cmd.Flags()
	var factor float64
	flags.Float64VarP(&factor, "factor", "f", 1.1, "1.2 increases brightness by 20%, 0.8 decreases brightness by 20%. Default 1.1")

	return cmd
}

func RunBrightnessCmd(cmd *cobra.Command, args []string) {
	logger.Print("Processing image...")

	imageOps, err := imageio.DetermineImageOperations(shared, args, cmd)
	utils.HandleError(err, "Error")

	factor, err := cmd.Flags().GetFloat64("factor")
	utils.HandleError(err, "Error")

	processor := &image.BrightnessProcessor{Factor: factor}
	processedImages, err := image.ProcessImgs(processor, imageOps, "")
	utils.HandleError(err, "Error")

	openImageInViewer(shared, args, processedImages[0])
}

func ValidateParseBrightnessCmd(cmd *cobra.Command, flags config.GlobalSubCommandFlags, args []string) error {
	if err := validateInput(flags, args); err != nil {
		return err
	}

	factor, _ := cmd.Flags().GetFloat64("factor")
	if factor <= 0.0 || factor > 10.0 {
		return fmt.Errorf("factor must be in range (0.0, 10.0], got: %.2f", factor)
	}

	return nil
}

func init() {
	rootCmd.AddCommand(BuildEffectsCmd())
}
