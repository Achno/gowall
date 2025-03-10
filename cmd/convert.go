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

var colorPair []string

var convertCmd = &cobra.Command{
	Use:   "convert [image path / batch flag]",
	Short: "Convert an img's color scheme",
	Long:  `Convert an img's color scheme`,
	// In a persistent prerun hook we could validate local command logic
	Run: func(cmd *cobra.Command, args []string) {
		ops := imageio.DetermineImageOperations(shared, args)
		switch {
		case len(shared.InputDir) > 0:
			logger.Print("Processing batch files...")
			processor := &image.ThemeConverter{}

			opts := image.ProcessOptions{
				SaveToFile: true,
				OutputExt:  formatFlag,
				OutputName: "",
				OutputDir:  config.GowallConfig.OutputFolder,
			}

			expandedFiles := utils.ExpandHomeDirectory(shared.BatchFiles)
			err := image.ProcessBatchImgs(expandedFiles, shared.Theme, processor, &opts)

			utils.HandleError(err)

		case len(dirInput) > 0:
			fmt.Println("Processing directory...")
			processor := &image.ThemeConverter{}
			files, err := utils.ExpandDirectory(dirInput)
			utils.HandleError(err, "Error expanding directory")
			err = image.ProcessBatchImgs(files, shared.Theme, processor)
			utils.HandleError(err)

		case len(args) > 0 && formatFlag != "":
			fmt.Println("Processing single image...")
			processor := &image.NoOpImageProcessor{}
			expandFile := utils.ExpandHomeDirectory(args)

			opts := image.ProcessOptions{
				SaveToFile: true,
				OutputExt:  formatFlag,
				OutputName: outputName,
			}

			path, _, err := image.ProcessImg(expandFile[0], processor, shared.Theme, opts)
			utils.HandleError(err, "Error Processing Image")

			err = image.OpenImageInViewer(path)
			utils.HandleError(err, "Error opening image")

		case len(args) > 0 && len(colorPair) > 0:
			fmt.Println("Replacing color...")
			expandFile := utils.ExpandHomeDirectory(args)
			processor := &image.ReplaceProcessor{}

			pairSlice, err := cmd.Flags().GetStringSlice("replace")
			utils.HandleError(err)

			if len(pairSlice) < 2 {
				utils.HandleError(fmt.Errorf("specify both the color to be replaced and the replacement color"), "Error")
			}

			processor.FromColor = pairSlice[0]
			processor.ToColor = pairSlice[1]
			processor.Threshold = 8.5
			if len(pairSlice) > 2 {
				processor.Threshold, err = strconv.ParseFloat(pairSlice[2], 64)
				utils.HandleError(err, "Error: either specify the threshold or remove the comma")
			}

			opts := image.ProcessOptions{
				OutputName: outputName,
				SaveToFile: true,
			}

			path, _, err := image.ProcessImg(expandFile[0], processor, shared.Theme, opts)
			utils.HandleError(err, "Error Processing Image")

			err = image.OpenImageInViewer(path)
			utils.HandleError(err, "Error opening image")

		case len(args) > 0:
			fmt.Println("Processing single image...")
			processor := &image.ThemeConverter{}
			expandFile := utils.ExpandHomeDirectory(args)

			opts := image.ProcessOptions{
				OutputName: outputName,
				SaveToFile: true,
			}
			path, _, err := image.ProcessImg(expandFile[0], processor, shared.Theme, opts)

			utils.HandleError(err, "Error Processing Image")
			err = image.OpenImageInViewer(path)

			utils.HandleError(err, "Error opening image")

		default:
			fmt.Println("Error: requires at least 1 arg(s) or at least --batch or --dir flags set correctly, received none")
			_ = cmd.Usage()
		}
	},
}

func themeCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return image.ListThemes(), cobra.ShellCompDirectiveNoFileComp
}

func init() {
	rootCmd.AddCommand(convertCmd)
	convertCmd.Flags().StringVarP(&shared.Theme, "theme", "t", "catppuccin", "Usage : --theme [ThemeName]")
	convertCmd.Flags().StringVarP(&shared.Format, "format", "f", "png", "Usage : --format [image format] format to encode the image")
	convertCmd.Flags().StringSliceVarP(&colorPair, "replace", "r", nil, "Usage: --replace #FromColor,#ToColor")
	convertCmd.RegisterFlagCompletionFunc("theme", themeCompletion)
	addGlobalFlags(convertCmd)
}
