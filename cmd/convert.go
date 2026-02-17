/*
Copyright © 2024 Achnologia <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"strconv"

	"github.com/Achno/gowall/config"
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
	)

	flags.StringVarP(&theme, "theme", "t", "", "Usage : --theme [ThemeName] or [PATH to Json file containing theme]")
	flags.StringVarP(&shared.Format, "format", "f", "", "Usage : --format [image format] png,webp,jpg,jpeg")
	flags.StringSliceVarP(&colorPair, "replace", "r", nil, "Usage: --replace #FromColor,#ToColor")

	cmd.RegisterFlagCompletionFunc("theme", themeCompletion)

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

	var processor image.ImageProcessor

	// Determine which processor to use
	if len(theme) > 0 {
		processor = &image.ThemeConverter{}
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

	return nil
}

func themeCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return image.ListThemes(), cobra.ShellCompDirectiveNoFileComp
}

func init() {
	rootCmd.AddCommand(BuildConvertCmd())
}
