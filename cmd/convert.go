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

var (
	theme      string
	batchFiles []string
)

var convertCmd = &cobra.Command{
	Use:   "convert [image path / batch flag]",
	Short: "Convert an img's color scheme",
	Long:  `Convert an img's color scheme`,
	// Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		switch {

		case len(batchFiles) > 0:
			fmt.Println("Processing batch files...")
			processor := &image.ThemeConverter{}
			expandedFiles := utils.ExpandHomeDirectory(batchFiles)
			image.ProcessBatchImgs(expandedFiles, theme, processor)
			
		case strings.HasSuffix(args[0],"#") :
			fmt.Println("Processing directory...")
			processor := &image.ThemeConverter{}
			path := utils.DiscardLastCharacter(args[0])
			files ,err := utils.ExpandHashtag(path)

			if err != nil {
				fmt.Printf("Error ExpandingHashTag: %s\n",err)
				return
			}
			image.ProcessBatchImgs(files,theme,processor)

		case len(args) > 0:
			fmt.Println("Processing single image...")
			processor := &image.ThemeConverter{}
			expandFile := utils.ExpandHomeDirectory(args)
			image.ProcessImg(expandFile[0], processor, theme)

		default:
			fmt.Println("Error: requires at least 1 arg(s), only received 0")
			_ = cmd.Usage()
		}
	},
}

func init() {
	rootCmd.AddCommand(convertCmd)
	convertCmd.Flags().StringVarP(&theme, "theme", "t", "catppuccin", "Usage : --theme [ThemeName-Lowercase]")
	convertCmd.Flags().StringSliceVarP(&batchFiles, "batch", "b", nil, "Usage: --batch [file1.png,file2.png ...]")

	// Here you will define your flags and configuration settings.
}
