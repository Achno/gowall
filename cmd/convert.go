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

// Declare variables specific to convert command
var (
	convertTheme      string
	convertBatchFiles []string
)

var convertCmd = &cobra.Command{
	Use:   "convert [image path / batch flag]",
	Short: "Convert an image's color scheme",
	Long:  `Convert an image's color scheme`,
	Run: func(cmd *cobra.Command, args []string) {
		switch {
		case len(convertBatchFiles) > 0:
			fmt.Println("Processing batch files...")
			processor := &image.ThemeConverter{}
			expandedFiles := utils.ExpandHomeDirectory(convertBatchFiles)
			image.ProcessBatchImgs(expandedFiles, convertTheme, processor)

		case len(args) > 0 && strings.HasSuffix(args[0], "#"):
			fmt.Println("Processing directory...")
			processor := &image.ThemeConverter{}
			path := utils.DiscardLastCharacter(args[0])
			files, err := utils.ExpandHashtag(path)

			if err != nil {
				fmt.Printf("Error Expanding HashTag: %s\n", err)
				return
			}
			image.ProcessBatchImgs(files, convertTheme, processor)

		case len(args) > 0:
			fmt.Println("Processing single image...")
			processor := &image.ThemeConverter{}
			expandFile := utils.ExpandHomeDirectory(args)
			image.ProcessImg(expandFile[0], processor, convertTheme)

		default:
			fmt.Println("Error: requires at least 1 arg(s), only received 0")
			_ = cmd.Usage()
		}
	},
}

func init() {
	rootCmd.AddCommand(convertCmd)
	convertCmd.Flags().StringVarP(&convertTheme, "theme", "t", "catppuccin", "Usage: --theme [ThemeName-Lowercase]")
	convertCmd.Flags().StringSliceVarP(&convertBatchFiles, "batch", "b", nil, "Usage: --batch [file1.png,file2.png ...]")
}
