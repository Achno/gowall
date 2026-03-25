/*
Copyright © 2025 Achno <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"image/color"
	"math"
	"strings"

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
	cmd.AddCommand(BuildContrastCmd())
	cmd.AddCommand(BuildGammaCmd())
	cmd.AddCommand(BuildSaturationCmd())
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

// Contrast Command
func BuildContrastCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "contrast [INPUT]",
		Short: "Adjust image contrast (normal or sigmoidal mode)",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return ValidateParseContrastCmd(cmd, shared, args)
		},
		Run: RunContrastCmd,
	}

	flags := cmd.Flags()
	var (
		mode          string
		factor        float64
		midpoint      float64
		sigmoidFactor float64
	)

	flags.StringVarP(&mode, "mode", "m", image.ContrastModeNormal, "Contrast mode: normal or sigmoid")
	flags.Float64VarP(&factor, "factor", "f", 0.0, "Normal contrast percentage in range (-100.0, 100.0)")
	flags.Float64VarP(&midpoint, "midpoint", "p", 0.5, "Sigmoid midpoint in range [0.0, 1.0]")
	flags.Float64VarP(&sigmoidFactor, "sigmoid-factor", "s", 0.0, "Sigmoid contrast factor in range (-10.0, 10.0)")

	cmd.RegisterFlagCompletionFunc("mode", contrastModeCompletion)

	return cmd
}

func RunContrastCmd(cmd *cobra.Command, args []string) {
	logger.Print("Processing image...")

	imageOps, err := imageio.DetermineImageOperations(shared, args, cmd)
	utils.HandleError(err, "Error")

	mode, err := cmd.Flags().GetString("mode")
	utils.HandleError(err, "Error")
	factor, err := cmd.Flags().GetFloat64("factor")
	utils.HandleError(err, "Error")
	midpoint, err := cmd.Flags().GetFloat64("midpoint")
	utils.HandleError(err, "Error")
	sigmoidFactor, err := cmd.Flags().GetFloat64("sigmoid-factor")
	utils.HandleError(err, "Error")

	processor := &image.ContrastProcessor{
		Mode:          strings.ToLower(mode),
		Factor:        factor,
		Midpoint:      midpoint,
		SigmoidFactor: sigmoidFactor,
	}

	processedImages, err := image.ProcessImgs(processor, imageOps, image.ProcessOptions{
		Theme:      "",
		OnComplete: nil,
	})
	utils.HandleError(err, "Error")

	openImageInViewer(cmd, shared, args, processedImages[0])
}

func ValidateParseContrastCmd(cmd *cobra.Command, flags config.GlobalSubCommandFlags, args []string) error {
	if err := validateInput(flags, args); err != nil {
		return err
	}

	mode, _ := cmd.Flags().GetString("mode")
	mode = strings.ToLower(mode)

	if mode != image.ContrastModeNormal && mode != image.ContrastModeSigmoid {
		return fmt.Errorf("invalid mode '%s', valid modes: %s, %s", mode, image.ContrastModeNormal, image.ContrastModeSigmoid)
	}

	if mode == image.ContrastModeNormal {
		if cmd.Flags().Changed("midpoint") || cmd.Flags().Changed("sigmoid-factor") {
			return fmt.Errorf("--midpoint and --sigmoid-factor can only be used with --mode sigmoid")
		}

		factor, _ := cmd.Flags().GetFloat64("factor")
		if factor <= -100.0 || factor >= 100.0 {
			return fmt.Errorf("factor must be in range (-100.0, 100.0), got: %.2f", factor)
		}

		return nil
	}

	if cmd.Flags().Changed("factor") {
		return fmt.Errorf("--factor can only be used with --mode normal")
	}

	midpoint, _ := cmd.Flags().GetFloat64("midpoint")
	if midpoint < 0.0 || midpoint > 1.0 {
		return fmt.Errorf("midpoint must be in range [0.0, 1.0], got: %.2f", midpoint)
	}

	sigmoidFactor, _ := cmd.Flags().GetFloat64("sigmoid-factor")
	if sigmoidFactor <= -10.0 || sigmoidFactor >= 10.0 {
		return fmt.Errorf("sigmoid-factor must be in range (-10.0, 10.0), got: %.2f", sigmoidFactor)
	}

	return nil
}

func contrastModeCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return []string{image.ContrastModeNormal, image.ContrastModeSigmoid}, cobra.ShellCompDirectiveNoFileComp
}

// Gamma Command
func BuildGammaCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "gamma [INPUT] [OPTIONAL OUTPUT] [--flags]",
		Short: "Apply gamma correction",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return ValidateParseGammaCmd(cmd, shared, args)
		},
		Run: RunGammaCmd,
	}

	flags := cmd.Flags()
	var gamma float64
	flags.Float64VarP(&gamma, "gamma", "g", 1.0, "Gamma correction factor (> 0.0)")

	return cmd
}

func RunGammaCmd(cmd *cobra.Command, args []string) {
	logger.Print("Processing image...")

	imageOps, err := imageio.DetermineImageOperations(shared, args, cmd)
	utils.HandleError(err, "Error")

	gamma, err := cmd.Flags().GetFloat64("gamma")
	utils.HandleError(err, "Error")

	processor := &image.GammaProcessor{
		Gamma: gamma,
	}

	processedImages, err := image.ProcessImgs(processor, imageOps, image.ProcessOptions{
		Theme:      "",
		OnComplete: nil,
	})
	utils.HandleError(err, "Error")

	openImageInViewer(cmd, shared, args, processedImages[0])
}

func ValidateParseGammaCmd(cmd *cobra.Command, flags config.GlobalSubCommandFlags, args []string) error {
	if err := validateInput(flags, args); err != nil {
		return err
	}

	gamma, _ := cmd.Flags().GetFloat64("gamma")
	if gamma <= 0.0 {
		return fmt.Errorf("gamma must be > 0.0, got: %.2f", gamma)
	}

	return nil
}

// Saturation Command
func BuildSaturationCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "saturation [INPUT] [OPTIONAL OUTPUT] [--flags]",
		Short: "Adjust image saturation",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return ValidateParseSaturationCmd(cmd, shared, args)
		},
		Run: RunSaturationCmd,
	}

	flags := cmd.Flags()
	var percentage float64
	flags.Float64VarP(&percentage, "percentage", "p", 0.0, "Saturation percentage in range [-100.0, 100.0]")

	return cmd
}

func RunSaturationCmd(cmd *cobra.Command, args []string) {
	logger.Print("Processing image...")

	imageOps, err := imageio.DetermineImageOperations(shared, args, cmd)
	utils.HandleError(err, "Error")

	percentage, err := cmd.Flags().GetFloat64("percentage")
	utils.HandleError(err, "Error")

	processor := &image.SaturationProcessor{
		Percentage: percentage,
	}

	processedImages, err := image.ProcessImgs(processor, imageOps, image.ProcessOptions{
		Theme:      "",
		OnComplete: nil,
	})
	utils.HandleError(err, "Error")

	openImageInViewer(cmd, shared, args, processedImages[0])
}

func ValidateParseSaturationCmd(cmd *cobra.Command, flags config.GlobalSubCommandFlags, args []string) error {
	if err := validateInput(flags, args); err != nil {
		return err
	}

	percentage, _ := cmd.Flags().GetFloat64("percentage")
	if percentage < -100.0 || percentage > 100.0 {
		return fmt.Errorf("percentage must be in range [-100.0, 100.0], got: %.2f", percentage)
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
		bgImage      string
	)

	flags.StringVarP(&preset, "preset", "p", "", "Use a preset configuration (p1, p2, p3, p4)")
	flags.Float64VarP(&tiltX, "tiltx", "x", 5.0, "Tilt angle on X axis (degrees)")
	flags.Float64VarP(&tiltY, "tilty", "y", -8.0, "Tilt angle on Y axis (degrees)")
	flags.Float64VarP(&tiltZ, "tiltz", "z", 3.0, "Tilt angle on Z axis / rotation (degrees, positive = clockwise)")
	flags.Float64VarP(&scale, "scale", "s", 0.65, "Scale factor (0.1 - 1.0)")
	flags.Float64VarP(&cornerRadius, "radius", "r", 40.0, "Corner radius for rounding")
	flags.StringVarP(&bgStart, "bg-start", "b", "#121212", "Gradient start color (hex)")
	flags.StringVarP(&bgEnd, "bg-end", "e", "#282828", "Gradient end color (hex)")
	flags.StringVarP(&bgImage, "bg-image", "i", "", "Background image path (overrides gradient)")

	cmd.RegisterFlagCompletionFunc("preset", presetCompletion)

	return cmd
}

func RunTiltCmd(cmd *cobra.Command, args []string) {
	logger.Print("Processing image...")

	imageOps, err := imageio.DetermineImageOperations(shared, args, cmd)
	utils.HandleError(err, "Error")

	preset, err := buildTiltPreset(cmd)
	utils.HandleError(err, "Error")

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

func buildTiltPreset(cmd *cobra.Command) (image.Preset, error) {
	presetName, err := cmd.Flags().GetString("preset")
	if err != nil {
		return image.Preset{}, err
	}

	preset, err := image.GetTiltPreset(presetName)
	if err != nil {
		return image.Preset{}, err
	}

	return applyTiltFlagOverrides(cmd, preset)
}

func applyTiltFlagOverrides(cmd *cobra.Command, preset image.Preset) (image.Preset, error) {
	floatOverrides := []struct {
		flag  string
		field *float64
	}{
		{"tiltx", &preset.TiltX},
		{"tilty", &preset.TiltY},
		{"tiltz", &preset.TiltZ},
		{"scale", &preset.Scale},
		{"radius", &preset.CornerRadius},
	}

	for _, o := range floatOverrides {
		if cmd.Flags().Changed(o.flag) {
			val, err := cmd.Flags().GetFloat64(o.flag)
			if err != nil {
				return preset, err
			}
			*o.field = val
		}
	}

	// Color overrides
	colorOverrides := []struct {
		flag  string
		field *color.RGBA
	}{
		{"bg-start", &preset.BackgroundStart},
		{"bg-end", &preset.BackgroundEnd},
	}

	for _, o := range colorOverrides {
		if cmd.Flags().Changed(o.flag) {
			val, err := cmd.Flags().GetString(o.flag)
			if err != nil {
				return preset, err
			}
			clr, err := cpkg.HexToRGBA(val)
			if err != nil {
				return preset, fmt.Errorf("invalid %s color format: %v (expected format: #RRGGBB)", o.flag, err)
			}
			*o.field = clr
		}
	}

	// Background image override
	if cmd.Flags().Changed("bg-image") {
		bgImagePath, err := cmd.Flags().GetString("bg-image")
		if err != nil {
			return preset, err
		}
		bgImg, err := imageio.LoadImage(imageio.FileReader{Path: bgImagePath})
		if err != nil {
			return preset, fmt.Errorf("failed to load background image: %v", err)
		}
		preset.BackgroundImage = bgImg
	}

	return preset, nil
}

func ValidateParseTiltCmd(cmd *cobra.Command, flags config.GlobalSubCommandFlags, args []string) error {
	if err := validateInput(flags, args); err != nil {
		return err
	}

	// Validate bg-image and bg-start/bg-end are mutually exclusive
	if cmd.Flags().Changed("bg-image") && (cmd.Flags().Changed("bg-start") || cmd.Flags().Changed("bg-end")) {
		return fmt.Errorf("cannot use --bg-image together with --bg-start or --bg-end; choose either a background image or gradient colors")
	}

	preset, err := buildTiltPreset(cmd)
	if err != nil {
		return err
	}

	if preset.Scale <= 0.0 || preset.Scale > 1.0 {
		return fmt.Errorf("scale must be in range (0.0, 1.0], got: %.2f", preset.Scale)
	}

	if preset.CornerRadius < 0 {
		return fmt.Errorf("corner radius must be >= 0, got: %.2f", preset.CornerRadius)
	}

	combinedAngle := math.Sqrt(preset.TiltX*preset.TiltX + preset.TiltY*preset.TiltY)
	threshold := 30.0 - (preset.Scale * 15.0)

	if combinedAngle > threshold {
		return fmt.Errorf(
			"combined tilt angle (%.1f°) exceeds 'estimated' safe (lmao i hope) threshold (%.1f°) for scale %.2f. "+
				"Reduce tiltx/tilty or decrease scale to avoid rendering issues. "+
				"Combined angle = √(tiltx² + tilty²) = √(%.1f² + %.1f²) = %.1f°",
			combinedAngle, threshold, preset.Scale, preset.TiltX, preset.TiltY, combinedAngle,
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
