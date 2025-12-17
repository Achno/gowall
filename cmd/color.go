/*
Copyright © 2025 Achno <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"slices"
	"sort"

	"github.com/Achno/gowall/config"
	cpkg "github.com/Achno/gowall/internal/backends/color"
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
// func BuildWheelCmd() *cobra.Command {
// 	cmd := &cobra.Command{
// 		Use:   "wheel [COLOR]",
// 		Short: "Around the color wheel (Complementary,Constrast,Triadic,Quadratic,Analogous,SplitComplementary)",
// 		Long:  `Generate a color wheel with a specified number of variants of a color (shades, tints, tones)`,
// 		PreRunE: func(cmd *cobra.Command, args []string) error {
// 			return ValidateParseWheelCmd(cmd, shared, args)
// 		},
// 		Run: RunVariantsCmd,
// 	}

// 	flags := cmd.Flags()
// 	var numShades int
// 	var variantType string
// 	flags.StringVarP(&variantType, "type", "t", "shades", "Type of variant to generate (shades, tints, tones)")
// 	flags.IntVarP(&numShades, "number", "n", 5, "Number of variants to generate")

// 	cmd.RegisterFlagCompletionFunc("type", variantsCompletion)

// 	return cmd
// }

// func RunWheelCmd(cmd *cobra.Command, args []string) {
// 	inputColor := args[0]
// 	numShades, _ := cmd.Flags().GetInt("number")
// 	variantType, _ := cmd.Flags().GetString("type")

// 	hexColor, err := cpkg.ParseColorToHex(inputColor)
// 	utils.HandleError(err, "Error")
// 	variationMap := GetvariationMap()
// 	f := variationMap[variantType]
// 	variants, err := f(hexColor, numShades)
// 	utils.HandleError(err, "Error")

// 	t, err := cpkg.NewTransformation([]string{inputColor}, variants)
// 	utils.HandleError(err, "Error creating transformation")

// 	t.Print()
// }

// func ValidateParseWheelCmd(cmd *cobra.Command, flags config.GlobalSubCommandFlags, args []string) error {
// 	if err := validateInput(flags, args); err != nil {
// 		return err
// 	}

// 	variationMap := GetvariationMap()
// 	variantType, _ := cmd.Flags().GetString("type")
// 	_, ok := variationMap[variantType]
// 	if !ok {
// 		return fmt.Errorf("invalid variant type '%s'", variantType)
// 	}

// 	numShades, _ := cmd.Flags().GetInt("number")
// 	if numShades < 1 {
// 		return fmt.Errorf("number of shades must be at least 1, got: %d", numShades)
// 	}

// 	return nil
// }

// func GetvariationMap() map[string]func(string, int) ([]string, error) {
// 	return map[string]func(string, int) ([]string, error){
// 		"shades":     cpkg.GenerateShades,
// 		"tints":      cpkg.GenerateTints,
// 		"tones":      cpkg.GenerateTones,
// 		"monochrome": cpkg.GenerateMonochromatic,
// 	}
// }

// func variantsCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
// 	variationMap := GetvariationMap()
// 	variantTypes := make([]string, 0, len(variationMap))
// 	for variantType := range variationMap {
// 		variantTypes = append(variantTypes, variantType)
// 	}
// 	sort.Strings(variantTypes)
// 	return variantTypes, cobra.ShellCompDirectiveNoFileComp
// }

func init() {
	rootCmd.AddCommand(BuildColorCmd())
}
