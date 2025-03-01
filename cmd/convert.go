/*
Copyright © 2024 Achnologia <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/Achno/gowall/internal/image"
	"github.com/Achno/gowall/utils"
	"github.com/spf13/cobra"
)

var formatFlag string
var colorPair []string
var outputName string

var convertCmd = &cobra.Command{
	Use:   "convert [image path / batch flag]",
	Short: "Convert an img's color scheme",
	Long:  `Convert an img's color scheme`,
	// Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		switch {

		case len(shared.BatchFiles) > 0:
			fmt.Println("Processing batch files...")
			processor := &image.ThemeConverter{}
			expandedFiles := utils.ExpandHomeDirectory(shared.BatchFiles)
			err := image.ProcessBatchImgs(expandedFiles, shared.Theme, processor)

			utils.HandleError(err)

		case len(args) > 0 && strings.HasSuffix(args[0], "#"):
			fmt.Println("Processing directory...")
			processor := &image.ThemeConverter{}
			path := utils.DiscardLastCharacter(args[0])
			files, err := utils.ExpandHashtag(path)

			utils.HandleError(err, "Error ExpandingHashTag")

			if outputName != "" {
				utils.HandleError(fmt.Errorf("You cannot use the '-o' flag and Batch conversion together"), "Error")
			}

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

			err = image.OpenImage(path)
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

			err = image.OpenImage(path)
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
			err = image.OpenImage(path)

			utils.HandleError(err, "Error opening image")

		default:
			fmt.Println("Error: requires at least 1 arg(s), only received 0")
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
	convertCmd.Flags().StringSliceVarP(&shared.BatchFiles, "batch", "b", nil, "Usage: --batch file1.png,file2.png ...")
	convertCmd.Flags().StringVarP(&formatFlag, "format", "f", "", "Usage: --format [Extension]")
	convertCmd.Flags().StringSliceVarP(&colorPair, "replace", "r", nil, "Usage: --replace #FromColor,#ToColor")
	convertCmd.Flags().StringVarP(&outputName, "output", "o", "", "Usage: --output imageName (no extension) Can only be used alongside with -t,-r,-f flags")

	convertCmd.RegisterFlagCompletionFunc("theme", themeCompletion)
}
