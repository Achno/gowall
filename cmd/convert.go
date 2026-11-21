/*
Copyright © 2024 Achnologia <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
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

func BuildConvertCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "convert [INPUT]",
		Short: "Convert an img's color scheme",
		Long:  `Convert an img's color scheme or its format ie from webp to png etc`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return ValidateParseConvertCmd(cmd, shared, args)
		},
		Run: RunConvertCmd,
	}

	flags := cmd.Flags()
	var (
		theme     string
		colorPair []string
		mode      string
		monoColor string
	)

	flags.StringVarP(&theme, "theme", "t", "", "Usage : --theme [ThemeName] or [PATH to Json file containing theme]")
	flags.StringVarP(&shared.Format, "format", "f", "", "Usage : --format [image format] png,webp,jpg,jpeg")
	flags.StringSliceVarP(&colorPair, "replace", "r", nil, "Usage: --replace #FromColor,#ToColor")
	flags.StringVarP(&mode, "mode", "m", image.ModeShading, "Usage: --mode "+strings.Join(image.ListThemeConverterModes(), "|")+", shading keeps every pixel's lightness and takes the --theme colors, palette moves every color toward the --theme colors, lightness included, mono keeps the lightness and uses one --color")
	flags.StringVarP(&monoColor, "color", "c", "", "Usage: --mode mono --color #cba6f7, the color mono mode uses")

	cmd.RegisterFlagCompletionFunc("theme", themeCompletion)
	cmd.RegisterFlagCompletionFunc("mode", modeCompletion)

	addGlobalFlags(cmd)

	return cmd
}

func RunConvertCmd(cmd *cobra.Command, args []string) {
	imageOps, err := imageio.DetermineImageOperations(shared, args, cmd)
	utils.HandleError(err, "Error")

	theme, err := cmd.Flags().GetString("theme")
	utils.HandleError(err, "Error")
	colorPair, err := cmd.Flags().GetStringSlice("replace")
	utils.HandleError(err, "Error")
	mode, err := cmd.Flags().GetString("mode")
	utils.HandleError(err, "Error")
	monoColor, err := cmd.Flags().GetString("color")
	utils.HandleError(err, "Error")

	// the nn backend set in the config has no shading, so it keeps using the palette mode unless --mode is given
	if !cmd.Flags().Changed("mode") && config.GowallConfig.ColorCorrectionBackend == "nn" {
		mode = image.ModePalette
	}

	var processor image.ImageProcessor

	// Determine which processor to use
	if len(theme) > 0 || mode == image.ModeMono {
		processor = &image.ThemeConverter{Mode: mode, Color: monoColor}
	} else if len(colorPair) > 0 {
		logger.Print("Replacing color...")
		processor = &image.ReplaceProcessor{}

		// Configure ReplaceProcessor if color pairs are provided
		if len(colorPair) < 2 {
			utils.HandleError(fmt.Errorf("specify both the color to be replaced and the replacement color"), "Error")
		}

		replaceProcessor, ok := processor.(*image.ReplaceProcessor)
		if ok {
			replaceProcessor.FromColor = colorPair[0]
			replaceProcessor.ToColor = colorPair[1]
			replaceProcessor.Threshold = 8.5
			if len(colorPair) > 2 {
				threshold, err := strconv.ParseFloat(colorPair[2], 64)
				utils.HandleError(err, "Error: either specify the threshold or remove the comma")
				replaceProcessor.Threshold = threshold
			}
		}
	} else {
		processor = &image.NoOpImageProcessor{}
	}

	logger.Print("Processing images...")
	processedImages, err := image.ProcessImgs(processor, imageOps, image.ProcessOptions{
		Theme:      theme,
		OnComplete: nil,
	})
	utils.HandleError(err, "Error")

	openImageInViewer(cmd, shared, args, processedImages[0])
}

func ValidateParseConvertCmd(cmd *cobra.Command, flags config.GlobalSubCommandFlags, args []string) error {
	if err := validateInput(flags, args); err != nil {
		return err
	}

	theme, _ := cmd.Flags().GetString("theme")
	colorPair, _ := cmd.Flags().GetStringSlice("replace")

	if len(theme) > 0 && len(colorPair) > 0 {
		return fmt.Errorf("cannot use both the --theme and --replace flags together")
	}

	mode, _ := cmd.Flags().GetString("mode")
	monoColor, _ := cmd.Flags().GetString("color")

	switch mode {
	case image.ModePalette, image.ModeShading:
		if cmd.Flags().Changed("mode") && len(theme) == 0 {
			return fmt.Errorf("--mode %s needs a --theme", mode)
		}
		if len(monoColor) > 0 {
			return fmt.Errorf("--color is only used with --mode mono")
		}
	case image.ModeMono:
		if len(monoColor) == 0 {
			return fmt.Errorf("--mode mono needs a --color, ie --color #cba6f7")
		}
		if _, err := cpkg.HexToRGBA(monoColor); err != nil {
			return fmt.Errorf("--color %s is not a hex color: %w", monoColor, err)
		}
		if len(theme) > 0 || len(colorPair) > 0 {
			return fmt.Errorf("--mode mono uses --color only, it cannot be used with --theme or --replace")
		}
	default:
		return fmt.Errorf("unknown --mode %s, use one of: %s", mode, strings.Join(image.ListThemeConverterModes(), ", "))
	}

	return nil
}

func themeCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return image.ListThemes(), cobra.ShellCompDirectiveNoFileComp
}

func modeCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return image.ListThemeConverterModes(), cobra.ShellCompDirectiveNoFileComp
}

func init() {
	rootCmd.AddCommand(BuildConvertCmd())
}
