/*
Copyright © 2025 Achno <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"math"

	"github.com/Achno/gowall/config"
	cpkg "github.com/Achno/gowall/internal/backends/color"
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
		Long:  `Apply various effects to your images like flip,mirror,grayscale,br(brightness),tilt and more`,
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
	cmd.AddCommand(BuildTiltCmd())

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
	processedImages, err := image.ProcessImgs(processor, imageOps, image.ProcessOptions{
		Theme:      "",
		OnComplete: nil, // default
	})
	utils.HandleError(err, "Error")

	openImageInViewer(cmd, shared, args, processedImages[0])
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
	processedImages, err := image.ProcessImgs(processor, imageOps, image.ProcessOptions{
		Theme:      "",
		OnComplete: nil, // default
	})
	utils.HandleError(err, "Error")

	openImageInViewer(cmd, shared, args, processedImages[0])
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
	processedImages, err := image.ProcessImgs(processor, imageOps, image.ProcessOptions{
		Theme:      "",
		OnComplete: nil,
	})
	utils.HandleError(err, "Error")

	openImageInViewer(cmd, shared, args, processedImages[0])
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
	processedImages, err := image.ProcessImgs(processor, imageOps, image.ProcessOptions{
		Theme:      "",
		OnComplete: nil, // default
	})
	utils.HandleError(err, "Error")

	openImageInViewer(cmd, shared, args, processedImages[0])
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

func BuildTiltCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tilt [INPUT]",
		Short: "Apply 3D tilt effect with rounded corners and gradient background",
		Long:  `Apply 3D tilt effect with rounded corners and gradient background. Use --preset for quick configurations or customize with individual flags.`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return ValidateParseTiltCmd(cmd, shared, args)
		},
		Run: RunTiltCmd,
	}

	flags := cmd.Flags()
	var (
		preset       string
		tiltX        float64
		tiltY        float64
		tiltZ        float64
		scale        float64
		cornerRadius float64
		bgStart      string
		bgEnd        string
	)

	flags.StringVarP(&preset, "preset", "p", "", "Use a preset configuration (p1, p2, p3, p4)")
	flags.Float64VarP(&tiltX, "tiltx", "x", 5.0, "Tilt angle on X axis (degrees)")
	flags.Float64VarP(&tiltY, "tilty", "y", -8.0, "Tilt angle on Y axis (degrees)")
	flags.Float64VarP(&tiltZ, "tiltz", "z", 3.0, "Tilt angle on Z axis / rotation (degrees, positive = clockwise)")
	flags.Float64VarP(&scale, "scale", "s", 0.65, "Scale factor (0.1 - 1.0)")
	flags.Float64VarP(&cornerRadius, "radius", "r", 40.0, "Corner radius for rounding")
	flags.StringVarP(&bgStart, "bg-start", "b", "#121212", "Gradient start color (hex)")
	flags.StringVarP(&bgEnd, "bg-end", "e", "#282828", "Gradient end color (hex)")

	cmd.RegisterFlagCompletionFunc("preset", presetCompletion)

	return cmd
}

func RunTiltCmd(cmd *cobra.Command, args []string) {
	logger.Print("Processing image...")

	imageOps, err := imageio.DetermineImageOperations(shared, args, cmd)
	utils.HandleError(err, "Error")

	presetName, err := cmd.Flags().GetString("preset")
	utils.HandleError(err, "Error")

	var preset image.Preset

	if presetName != "" {
		presetConfig, exists := image.TiltPresets[presetName]
		if !exists {
			utils.HandleError(fmt.Errorf("unknown preset: %s", presetName), "Error")
		}
		preset = presetConfig
	} else {
		tiltX, err := cmd.Flags().GetFloat64("tiltx")
		utils.HandleError(err, "Error")
		tiltY, err := cmd.Flags().GetFloat64("tilty")
		utils.HandleError(err, "Error")
		tiltZ, err := cmd.Flags().GetFloat64("tiltz")
		utils.HandleError(err, "Error")
		scale, err := cmd.Flags().GetFloat64("scale")
		utils.HandleError(err, "Error")
		cornerRadius, err := cmd.Flags().GetFloat64("radius")
		utils.HandleError(err, "Error")
		bgStart, err := cmd.Flags().GetString("bg-start")
		utils.HandleError(err, "Error")
		bgEnd, err := cmd.Flags().GetString("bg-end")
		utils.HandleError(err, "Error")

		bgStartColor, err := cpkg.HexToRGBA(bgStart)
		utils.HandleError(err, "Error parsing bg-start color")
		bgEndColor, err := cpkg.HexToRGBA(bgEnd)
		utils.HandleError(err, "Error parsing bg-end color")

		preset = image.Preset{
			BackgroundStart: bgStartColor,
			BackgroundEnd:   bgEndColor,
			TiltX:           tiltX,
			TiltY:           tiltY,
			TiltZ:           tiltZ,
			Scale:           scale,
			CornerRadius:    cornerRadius,
		}
	}

	processor := &image.TiltProcessor{
		Preset: preset,
	}

	processedImages, err := image.ProcessImgs(processor, imageOps, image.ProcessOptions{
		Theme:      "",
		OnComplete: nil,
	})
	utils.HandleError(err, "Error")

	openImageInViewer(cmd, shared, args, processedImages[0])
}

func ValidateParseTiltCmd(cmd *cobra.Command, flags config.GlobalSubCommandFlags, args []string) error {
	if err := validateInput(flags, args); err != nil {
		return err
	}

	presetName, _ := cmd.Flags().GetString("preset")

	// Check if any manual flags were changed
	manualFlagsUsed := cmd.Flags().Changed("tiltx") ||
		cmd.Flags().Changed("tilty") ||
		cmd.Flags().Changed("tiltz") ||
		cmd.Flags().Changed("scale") ||
		cmd.Flags().Changed("radius") ||
		cmd.Flags().Changed("bg-start") ||
		cmd.Flags().Changed("bg-end")

	if presetName != "" && manualFlagsUsed {
		return fmt.Errorf("cannot use preset flag (-p/--preset) together with manual configuration flags (x, y, z, s, r, b, e) either choose a preset or make your own")
	}

	if presetName != "" {
		if _, exists := image.TiltPresets[presetName]; !exists {
			validPresets := image.GetTiltPresetNames()
			return fmt.Errorf("invalid preset '%s'. Valid presets: %v", presetName, validPresets)
		}
		return nil
	}

	scale, _ := cmd.Flags().GetFloat64("scale")
	if scale <= 0.0 || scale > 1.0 {
		return fmt.Errorf("scale must be in range (0.0, 1.0], got: %.2f", scale)
	}

	cornerRadius, _ := cmd.Flags().GetFloat64("radius")
	if cornerRadius < 0 {
		return fmt.Errorf("corner radius must be >= 0, got: %.2f", cornerRadius)
	}

	bgStart, _ := cmd.Flags().GetString("bg-start")
	_, err := cpkg.HexToRGBA(bgStart)
	if err != nil {
		return fmt.Errorf("invalid bg-start color format: %v (expected format: #RRGGBB)", err)
	}

	bgEnd, _ := cmd.Flags().GetString("bg-end")
	_, err = cpkg.HexToRGBA(bgEnd)
	if err != nil {
		return fmt.Errorf("invalid bg-end color format: %v (expected format: #RRGGBB)", err)
	}

	tiltX, _ := cmd.Flags().GetFloat64("tiltx")
	tiltY, _ := cmd.Flags().GetFloat64("tilty")

	combinedAngle := math.Sqrt(tiltX*tiltX + tiltY*tiltY)
	threshold := 30.0 - (scale * 15.0)

	if combinedAngle > threshold {
		return fmt.Errorf(
			"combined tilt angle (%.1f°) exceeds 'estimated' safe (lmao i hope) threshold (%.1f°) for scale %.2f. "+
				"Reduce tiltx/tilty or decrease scale to avoid rendering issues. "+
				"Combined angle = √(tiltx² + tilty²) = √(%.1f² + %.1f²) = %.1f°",
			combinedAngle, threshold, scale, tiltX, tiltY, combinedAngle,
		)
	}

	return nil
}

func presetCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return image.GetTiltPresetNames(), cobra.ShellCompDirectiveNoFileComp
}

func init() {
	rootCmd.AddCommand(BuildEffectsCmd())
}
