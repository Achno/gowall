/*
Copyright © 2024 Achnologia <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"strconv"

	"github.com/Achno/gowall/internal/image"
	imageio "github.com/Achno/gowall/internal/image_io"
	"github.com/Achno/gowall/internal/logger"
	"github.com/Achno/gowall/utils"
	"github.com/spf13/cobra"
)

var (
	colorPair []string
	theme     string
)

var convertCmd = &cobra.Command{
	Use:   "convert [INPUT] [OPTIONAL OUTPUT]",
	Short: "Convert an img's color scheme",
	Long:  `Convert an img's color scheme or its format ie from webp to png etc`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		err := validateInput(shared, args)
		if err != nil {
			return err
		}
		if len(theme) > 0 && len(colorPair) > 0 {
			return fmt.Errorf("cannot use both the --theme and --replace flags together")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
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
			// Type assertion
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

		imageOps, err := imageio.DetermineImageOperations(shared, args)
		utils.HandleError(err)

		logger.Print("Processing images...")
		processedImages, err := image.ProcessImgs(processor, imageOps, theme)
		utils.HandleError(err, "Error")

		// if len(processedImages) == 0 {
		// 	utils.HandleError(err, "No images were processed")
		// 	return
		// }

		openImageInViewer(shared, args, processedImages[0])
	},
}

func themeCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return image.ListThemes(), cobra.ShellCompDirectiveNoFileComp
}

func init() {
	rootCmd.AddCommand(convertCmd)
	convertCmd.Flags().StringVarP(&theme, "theme", "t", "", "Usage : --theme [ThemeName] or [PATH to Json file containing theme]")
	convertCmd.Flags().StringVarP(&shared.Format, "format", "f", "", "Usage : --format [image format] png,webp,jpg,jpeg")
	convertCmd.Flags().StringSliceVarP(&colorPair, "replace", "r", nil, "Usage: --replace #FromColor,#ToColor")
	convertCmd.RegisterFlagCompletionFunc("theme", themeCompletion)
	addGlobalFlags(convertCmd)
}
