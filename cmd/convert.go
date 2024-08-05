/*
Copyright © 2024 Achnologia <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"strings"

	"github.com/Achno/gowall/internal/image"
	"github.com/Achno/gowall/utils"
	"github.com/spf13/cobra"
)

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

			err = image.ProcessBatchImgs(files, shared.Theme, processor)

			utils.HandleError(err)

		case len(args) > 0:
			fmt.Println("Processing single image...")
			processor := &image.ThemeConverter{}
			expandFile := utils.ExpandHomeDirectory(args)
			err := image.ProcessImg(expandFile[0], processor, shared.Theme)

			utils.HandleError(err, "Error Processing Image")

		default:
			fmt.Println("Error: requires at least 1 arg(s), only received 0")
			_ = cmd.Usage()
		}
	},
}

func init() {
	rootCmd.AddCommand(convertCmd)
	convertCmd.Flags().StringVarP(&shared.Theme, "theme", "t", "catppuccin", "Usage : --theme [ThemeName-Lowercase]")
	convertCmd.Flags().StringSliceVarP(&shared.BatchFiles, "batch", "b", nil, "Usage: --batch [file1.png,file2.png ...]")

	// Here you will define your flags and configuration settings.
}
