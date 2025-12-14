/*
Copyright © 2025 Achno <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"slices"

	"github.com/Achno/gowall/config"
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
	cmd.AddCommand(BuildShadesCmd())

	addGlobalFlags(cmd)

	return cmd
}

// Convert Command
func BuildClrConvertCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "convert [COLOR]",
		Short: "Convert a color between different formats",
		Long:  `Convert a color between different formats (hex, rgb, hsl, etc.)`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return ValidateParseConvertCmd(cmd, shared, args)
		},
		Run: RunConvertCmd,
	}

	flags := cmd.Flags()
	var toFormat string
	flags.StringVarP(&toFormat, "to", "t", "rgb", "Target format to convert to (rgb, hsl, hex)")

	return cmd
}

func RunClrConvertCmd(cmd *cobra.Command, args []string) {
	logger.Print("Converting color...")

	// TODO: Implement color conversion logic
	color := args[0]
	toFormat, _ := cmd.Flags().GetString("to")

	logger.Print(fmt.Sprintf("Converting %s to %s format", color, toFormat))
}

func ValidateParseClrConvertCmd(cmd *cobra.Command, flags config.GlobalSubCommandFlags, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("color argument is required")
	}

	toFormat, _ := cmd.Flags().GetString("to")
	validFormats := []string{"rgb", "hsl", "hex", "hsv"}
	valid := slices.Contains(validFormats, toFormat)
	if !valid {
		return fmt.Errorf("invalid format '%s'. Valid formats: rgb, hsl, hex, hsv", toFormat)
	}

	return nil
}

// Light Command
func BuildLightCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "light [COLOR] [AMOUNT]",
		Short: "Lighten a color by a specified amount",
		Long:  `Lighten a color by a specified amount (0.0 to 1.0, where 0.3 means 30% lighter)`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return ValidateParseLightCmd(cmd, shared, args)
		},
		Run: RunLightCmd,
	}

	return cmd
}

func RunLightCmd(cmd *cobra.Command, args []string) {
	logger.Print("Lightening color...")

	// TODO: Implement color lightening logic
	color := args[0]
	amount := args[1]

	logger.Print(fmt.Sprintf("Lightening %s by %s", color, amount))
}

func ValidateParseLightCmd(cmd *cobra.Command, flags config.GlobalSubCommandFlags, args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("color and amount arguments are required")
	}

	// TODO: Add validation for color format and amount range (0.0-1.0)

	return nil
}

// Dark Command
func BuildDarkCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "dark [COLOR] [AMOUNT]",
		Short: "Darken a color by a specified amount",
		Long:  `Darken a color by a specified amount (0.0 to 1.0, where 0.2 means 20% darker)`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return ValidateParseDarkCmd(cmd, shared, args)
		},
		Run: RunDarkCmd,
	}

	return cmd
}

func RunDarkCmd(cmd *cobra.Command, args []string) {
	logger.Print("Darkening color...")

	// TODO: Implement color darkening logic
	color := args[0]
	amount := args[1]

	logger.Print(fmt.Sprintf("Darkening %s by %s", color, amount))
}

func ValidateParseDarkCmd(cmd *cobra.Command, flags config.GlobalSubCommandFlags, args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("color and amount arguments are required")
	}

	// TODO: Add validation for color format and amount range (0.0-1.0)

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
	flags.IntVarP(&numColors, "number", "n", 5, "Number of colors to generate in the blend")

	return cmd
}

func RunBlendCmd(cmd *cobra.Command, args []string) {
	logger.Print("Blending colors...")

	// TODO: Implement color blending logic
	color1 := args[0]
	color2 := args[1]
	numColors, _ := cmd.Flags().GetInt("number")

	logger.Print(fmt.Sprintf("Blending %s and %s into %d colors", color1, color2, numColors))
}

func ValidateParseBlendCmd(cmd *cobra.Command, flags config.GlobalSubCommandFlags, args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("two color arguments are required")
	}

	numColors, _ := cmd.Flags().GetInt("number")
	if numColors < 2 {
		return fmt.Errorf("number of colors must be at least 2, got: %d", numColors)
	}

	return nil
}

// Shades Command
func BuildShadesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "shades [COLOR]",
		Short: "Generate shades of a color",
		Long:  `Generate a specified number of shades (darker variations) of a color`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return ValidateParseShadesCmd(cmd, shared, args)
		},
		Run: RunShadesCmd,
	}

	flags := cmd.Flags()
	var numShades int
	flags.IntVarP(&numShades, "number", "n", 5, "Number of shades to generate")

	return cmd
}

func RunShadesCmd(cmd *cobra.Command, args []string) {
	logger.Print("Generating color shades...")

	// TODO: Implement color shades generation logic
	color := args[0]
	numShades, _ := cmd.Flags().GetInt("number")

	logger.Print(fmt.Sprintf("Generating %d shades of %s", numShades, color))
}

func ValidateParseShadesCmd(cmd *cobra.Command, flags config.GlobalSubCommandFlags, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("color argument is required")
	}

	numShades, _ := cmd.Flags().GetInt("number")
	if numShades < 1 {
		return fmt.Errorf("number of shades must be at least 1, got: %d", numShades)
	}

	return nil
}

func init() {
	rootCmd.AddCommand(BuildColorCmd())
}
