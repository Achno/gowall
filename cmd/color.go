/*
Copyright © 2025 Achno <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"slices"
	"sort"
	"strconv"
	"strings"

	"github.com/Achno/gowall/config"
	cpkg "github.com/Achno/gowall/internal/backends/color"
	"github.com/Achno/gowall/internal/image"
	imageio "github.com/Achno/gowall/internal/image_io"
	"github.com/Achno/gowall/internal/logger"
	"github.com/Achno/gowall/utils"
	"github.com/spf13/cobra"
)

func BuildColorCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "color [COMMAND]",
		Short: "Work with colors - generate palletes, convert, lighten, darken, blend, and generate shades",
		Long:  `Work with colors - generate palletes, convert between formats, lighten/darken colors, blend between colors, and generate color shades`,
		Run: func(cmd *cobra.Command, args []string) {
			logger.Print("Please specify a color command")
			err := cmd.Usage()
			utils.HandleError(err)
		},
	}

	cmd.AddCommand(BuildClrConvertCmd())
	cmd.AddCommand(BuildLightCmd())
	cmd.AddCommand(BuildDarkCmd())
	cmd.AddCommand(BuildBlendCmd())
	cmd.AddCommand(BuildVariantsCmd())
	cmd.AddCommand(BuildWheelCmd())
	cmd.AddCommand(BuildGradientCmd())

	addGlobalFlags(cmd)

	return cmd
}

// Convert Command
func BuildClrConvertCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "convert [COLOR]",
		Short: "Convert a color between different formats",
		Long:  `Convert a color between different formats (hex, rgb, hsl, lab)`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return ValidateParseClrConvertCmd(cmd, shared, args)
		},
		Run: RunClrConvertCmd,
	}

	flags := cmd.Flags()
	var toFormat string
	flags.StringVarP(&toFormat, "to", "t", "rgb", "Target format to convert to (rgb, hsl, hex, lab)")

	return cmd
}

func RunClrConvertCmd(cmd *cobra.Command, args []string) {
	inputColor := args[0]
	toFormat, _ := cmd.Flags().GetString("to")

	hexColor, err := cpkg.ParseColorToHex(inputColor)
	utils.HandleError(err, "Error")

	outputStr, _, err := cpkg.ConvertHexToFormat(hexColor, toFormat)
	utils.HandleError(err, "Error")

	t, err := cpkg.NewTransformation([]string{inputColor}, []string{outputStr})
	utils.HandleError(err, "Error creating transformation")

	t.Print()
}

func ValidateParseClrConvertCmd(cmd *cobra.Command, flags config.GlobalSubCommandFlags, args []string) error {
	if err := validateInput(flags, args); err != nil {
		return err
	}

	toFormat, _ := cmd.Flags().GetString("to")
	validFormats := cpkg.ValidFormats()
	valid := slices.Contains(validFormats, toFormat)
	if !valid {
		return fmt.Errorf("invalid format '%s'. Valid formats: %v", toFormat, validFormats)
	}

	return nil
}

// Light Command
func BuildLightCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "light [COLOR]",
		Short: "Lighten a color by a specified amount",
		Long:  `Lighten a color by a specified amount (0.0 to 1.0, where 0.3 means 30% lighter)`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return ValidateParseLightCmd(cmd, shared, args)
		},
		Run: RunLightCmd,
	}

	flags := cmd.Flags()
	var factor float64
	flags.Float64VarP(&factor, "factor", "f", 0.3, "Amount to lighten the color by (0.0 to 1.0, where 0.3 means 30% lighter)")

	return cmd
}

func RunLightCmd(cmd *cobra.Command, args []string) {
	inputColor := args[0]
	factor, _ := cmd.Flags().GetFloat64("factor")

	hexColor, err := cpkg.ParseColorToHex(inputColor)
	utils.HandleError(err, "Error")

	lightenedColor, err := cpkg.LightenColor(hexColor, factor)
	utils.HandleError(err, "Error")

	t, err := cpkg.NewTransformation([]string{inputColor}, []string{lightenedColor})
	utils.HandleError(err, "Error creating transformation")
	t.Print()
}

func ValidateParseLightCmd(cmd *cobra.Command, flags config.GlobalSubCommandFlags, args []string) error {
	if err := validateInput(flags, args); err != nil {
		return err
	}

	amount, _ := cmd.Flags().GetFloat64("amount")
	if amount < 0.0 || amount > 1.0 {
		return fmt.Errorf("amount must be between 0.0 and 1.0, got: %.2f", amount)
	}

	return nil
}

// Dark Command
func BuildDarkCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "dark [COLOR]",
		Short: "Darken a color by a specified amount",
		Long:  `Darken a color by a specified amount (0.0 to 1.0, where 0.2 means 20% darker)`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return ValidateParseDarkCmd(cmd, shared, args)
		},
		Run: RunDarkCmd,
	}

	flags := cmd.Flags()
	var factor float64
	flags.Float64VarP(&factor, "factor", "f", 0.3, "Amount to lighten the color by (0.0 to 1.0, where 0.3 means 30% lighter)")

	return cmd
}

func RunDarkCmd(cmd *cobra.Command, args []string) {
	inputColor := args[0]
	factor, _ := cmd.Flags().GetFloat64("factor")

	hexColor, err := cpkg.ParseColorToHex(inputColor)
	utils.HandleError(err, "Error")

	lightenedColor, err := cpkg.DarkenColor(hexColor, factor)
	utils.HandleError(err, "Error")

	t, err := cpkg.NewTransformation([]string{inputColor}, []string{lightenedColor})
	utils.HandleError(err, "Error creating transformation")
	t.Print()
}

func ValidateParseDarkCmd(cmd *cobra.Command, flags config.GlobalSubCommandFlags, args []string) error {
	if err := validateInput(flags, args); err != nil {
		return err
	}

	amount, _ := cmd.Flags().GetFloat64("amount")
	if amount < 0.0 || amount > 1.0 {
		return fmt.Errorf("amount must be between 0.0 and 1.0, got: %.2f", amount)
	}
	return nil
}

// Blend Command
func BuildBlendCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "blend [COLOR1] [COLOR2]",
		Short: "Blend between two colors",
		Long:  `Blend between two colors generating a specified number of intermediate colors`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return ValidateParseBlendCmd(cmd, shared, args)
		},
		Run: RunBlendCmd,
	}

	flags := cmd.Flags()
	var numColors int
	flags.IntVarP(&numColors, "number", "n", 3, "Number of colors to generate in the blend")

	return cmd
}

func RunBlendCmd(cmd *cobra.Command, args []string) {
	inputColor1 := args[0]
	inputColor2 := args[1]
	numColors, _ := cmd.Flags().GetInt("number")

	hexColor1, err := cpkg.ParseColorToHex(inputColor1)
	utils.HandleError(err, "Error")
	hexColor2, err := cpkg.ParseColorToHex(inputColor2)
	utils.HandleError(err, "Error")
	blendedColors, err := cpkg.BlendColors(hexColor1, hexColor2, numColors)
	utils.HandleError(err, "Error")

	t, err := cpkg.NewTransformation([]string{inputColor1, inputColor2}, blendedColors)
	utils.HandleError(err, "Error creating transformation")
	t.Print()
}

func ValidateParseBlendCmd(cmd *cobra.Command, flags config.GlobalSubCommandFlags, args []string) error {
	if err := validateInput(flags, args); err != nil {
		return err
	}

	if len(args) != 2 {
		return fmt.Errorf("two color arguments are required")
	}

	numColors, _ := cmd.Flags().GetInt("number")
	if numColors < 2 {
		return fmt.Errorf("number of colors must be at least 2, got: %d", numColors)
	}

	return nil
}

// Variants Command
func BuildVariantsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "variants [COLOR]",
		Short: "Generate variants (Shades, Tints, Tones) of a color",
		Long:  `Generate a specified number of variants of a color (shades, tints, tones)`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return ValidateParseVariantsCmd(cmd, shared, args)
		},
		Run: RunVariantsCmd,
	}

	flags := cmd.Flags()
	var numShades int
	var variantType string
	flags.StringVarP(&variantType, "type", "t", "shades", "Type of variant to generate (shades, tints, tones)")
	flags.IntVarP(&numShades, "number", "n", 5, "Number of variants to generate")

	cmd.RegisterFlagCompletionFunc("type", variantsCompletion)

	return cmd
}

func RunVariantsCmd(cmd *cobra.Command, args []string) {
	inputColor := args[0]
	numShades, _ := cmd.Flags().GetInt("number")
	variantType, _ := cmd.Flags().GetString("type")

	hexColor, err := cpkg.ParseColorToHex(inputColor)
	utils.HandleError(err, "Error")
	variationMap := GetvariationMap()
	f := variationMap[variantType]
	variants, err := f(hexColor, numShades)
	utils.HandleError(err, "Error")

	t, err := cpkg.NewTransformation([]string{inputColor}, variants)
	utils.HandleError(err, "Error creating transformation")

	t.Print()
}

func ValidateParseVariantsCmd(cmd *cobra.Command, flags config.GlobalSubCommandFlags, args []string) error {
	if err := validateInput(flags, args); err != nil {
		return err
	}

	variationMap := GetvariationMap()
	variantType, _ := cmd.Flags().GetString("type")
	_, ok := variationMap[variantType]
	if !ok {
		return fmt.Errorf("invalid variant type '%s'", variantType)
	}

	numShades, _ := cmd.Flags().GetInt("number")
	if numShades < 1 {
		return fmt.Errorf("number of shades must be at least 1, got: %d", numShades)
	}

	return nil
}

func GetvariationMap() map[string]func(string, int) ([]string, error) {
	return map[string]func(string, int) ([]string, error){
		"shades":     cpkg.GenerateShades,
		"tints":      cpkg.GenerateTints,
		"tones":      cpkg.GenerateTones,
		"monochrome": cpkg.GenerateMonochromatic,
	}
}

func variantsCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	variationMap := GetvariationMap()
	variantTypes := make([]string, 0, len(variationMap))
	for variantType := range variationMap {
		variantTypes = append(variantTypes, variantType)
	}
	sort.Strings(variantTypes)
	return variantTypes, cobra.ShellCompDirectiveNoFileComp
}

// Wheel Command
func BuildWheelCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "wheel [COLOR]",
		Short: "Functions around the color wheel (triadic, quadratic, analogous, split-complementary)",
		Long:  `Functions around the color wheel (triadic, quadratic, analogous, split-complementary)`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return ValidateParseWheelCmd(cmd, shared, args)
		},
		Run: RunWheelCmd,
	}

	flags := cmd.Flags()
	var wheelType string
	flags.StringVarP(&wheelType, "type", "t", "triadic", "Type of color wheel scheme (triadic, quadratic, analogous, split-complementary)")

	cmd.RegisterFlagCompletionFunc("type", wheelCompletion)

	return cmd
}

func RunWheelCmd(cmd *cobra.Command, args []string) {
	inputColor := args[0]
	wheelType, _ := cmd.Flags().GetString("type")

	hexColor, err := cpkg.ParseColorToHex(inputColor)
	utils.HandleError(err, "Error")

	wheelMap := GetWheelMap()
	f := wheelMap[wheelType]
	wheelColors, err := f(hexColor)
	utils.HandleError(err, "Error")

	t, err := cpkg.NewTransformation([]string{inputColor}, wheelColors)
	utils.HandleError(err, "Error creating transformation")

	t.Print()
}

func ValidateParseWheelCmd(cmd *cobra.Command, flags config.GlobalSubCommandFlags, args []string) error {
	if err := validateInput(flags, args); err != nil {
		return err
	}

	wheelMap := GetWheelMap()
	wheelType, _ := cmd.Flags().GetString("type")
	_, ok := wheelMap[wheelType]
	if !ok {
		return fmt.Errorf("invalid wheel type '%s'", wheelType)
	}

	return nil
}

func GetWheelMap() map[string]func(string) ([]string, error) {
	return map[string]func(string) ([]string, error){
		"triadic":             cpkg.GenerateTriadic,
		"quadratic":           cpkg.GenerateQuadratic,
		"analogous":           cpkg.GenerateAnalogous,
		"split-complementary": cpkg.GenerateSplitComplementary,
	}
}

func wheelCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	wheelMap := GetWheelMap()
	wheelTypes := make([]string, 0, len(wheelMap))
	for wheelType := range wheelMap {
		wheelTypes = append(wheelTypes, wheelType)
	}
	sort.Strings(wheelTypes)
	return wheelTypes, cobra.ShellCompDirectiveNoFileComp
}

// Gradient Command
func BuildGradientCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "gradient [COLORS]",
		Short: "Generate a gradient image from multiple colors",
		Long:  `Generate a gradient image from 2 or more colors (comma-separated) with specified dimensions and direction. Example: gradient "#ff0000,#00ff00,#0000ff"`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return ValidateParseGradientCmd(cmd, shared, args)
		},
		Run: RunGradientCmd,
	}

	flags := cmd.Flags()
	var (
		dimensions string
		direction  string
		width      int
		height     int
		method     string
	)

	flags.StringVarP(&dimensions, "dimensions", "d", "1920x1080", "Dimensions in format WIDTHxHEIGHT (e.g., 1920x1080)")
	flags.StringVarP(&direction, "direction", "r", "vertical", "Gradient direction: vertical or horizontal")
	flags.StringVarP(&method, "method", "m", "rgb", "Gradient method: rgb, hcl, lab, hsv, luv, luvlch")
	// Hidden flags to pass parsed values from PreRunE to Run
	flags.IntVar(&width, "width", 0, "")
	flags.IntVar(&height, "height", 0, "")
	cmd.Flags().MarkHidden("width")
	cmd.Flags().MarkHidden("height")

	addGlobalFlags(cmd)

	return cmd
}

func RunGradientCmd(cmd *cobra.Command, args []string) {
	ops, err := imageio.DetermineImageOperations(shared, args, cmd)
	utils.HandleError(err, "Error")

	width, err := cmd.Flags().GetInt("width")
	utils.HandleError(err, "Error")
	height, err := cmd.Flags().GetInt("height")
	utils.HandleError(err, "Error")
	direction, err := cmd.Flags().GetString("direction")
	utils.HandleError(err, "Error")
	method, err := cmd.Flags().GetString("method")
	utils.HandleError(err, "Error")

	// Parse colors from first argument (comma-separated)
	var hexColors []string
	if len(args) > 0 && args[0] != "-" {
		colorStrings := strings.SplitSeq(args[0], ",")
		for colorStr := range colorStrings {
			colorStr = strings.TrimSpace(colorStr)
			if colorStr == "" {
				continue
			}
			hexColor, err := cpkg.ParseColorToHex(colorStr)
			utils.HandleError(err, "Error parsing color: "+colorStr)
			hexColors = append(hexColors, hexColor)
		}
	}

	processor := &image.GradientProcessor{}
	processor.SetOptions(
		image.WithColors(hexColors),
		image.WithGradientWidth(width),
		image.WithGradientHeight(height),
		image.WithDirection(direction),
		image.WithGradientMethod(method),
	)

	logger.Print("Generating gradient...")
	gradientImages, err := image.ProcessImgs(processor, ops, image.ProcessOptions{
		Theme:      "",
		OnComplete: nil,
	})
	utils.HandleError(err, "Error")

	openImageInViewer(cmd, shared, args, gradientImages[0])
}

func ValidateParseGradientCmd(cmd *cobra.Command, flags config.GlobalSubCommandFlags, args []string) error {
	// Check that we have at least one argument with colors
	if len(args) == 0 {
		return fmt.Errorf("gradient requires colors as first argument (comma-separated, e.g., \"#ff0000,#00ff00,#0000ff\")")
	}

	// Parse and count colors
	colorStrings := strings.Split(args[0], ",")
	validColors := 0
	for _, colorStr := range colorStrings {
		if strings.TrimSpace(colorStr) != "" {
			validColors++
		}
	}

	if validColors < 2 {
		return fmt.Errorf("gradient requires at least 2 colors, got %d", validColors)
	}

	dimensions, _ := cmd.Flags().GetString("dimensions")
	if dimensions == "" {
		return fmt.Errorf("dimensions cannot be empty, use format WIDTHxHEIGHT (e.g., 1920x1080)")
	}

	parts := strings.Split(dimensions, "x")
	if len(parts) != 2 {
		return fmt.Errorf("invalid dimensions format: %s, use format WIDTHxHEIGHT (e.g., 1920x1080)", dimensions)
	}

	width, err := strconv.Atoi(parts[0])
	if err != nil {
		return fmt.Errorf("invalid width value: %s", parts[0])
	}

	height, err := strconv.Atoi(parts[1])
	if err != nil {
		return fmt.Errorf("invalid height value: %s", parts[1])
	}

	if width <= 0 || height <= 0 {
		return fmt.Errorf("width and height must be positive integers")
	}

	direction, _ := cmd.Flags().GetString("direction")
	if direction != "vertical" && direction != "horizontal" {
		return fmt.Errorf("direction must be either 'vertical' or 'horizontal', got: %s", direction)
	}

	cmd.Flags().Set("width", strconv.Itoa(width))
	cmd.Flags().Set("height", strconv.Itoa(height))

	return nil
}

func init() {
	rootCmd.AddCommand(BuildColorCmd())
}
