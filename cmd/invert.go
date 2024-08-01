package cmd

import (
	"fmt"
	"strings"

	"github.com/Achno/gowall/internal/image"
	"github.com/Achno/gowall/utils"
	"github.com/spf13/cobra"
)

// Declare variables specific to invert command
var (
	invertTheme      string
	invertBatchFiles []string
)

var invertCmd = &cobra.Command{
	Use:   "invert [image path]",
	Short: "Inverts the colors of an image",
	Long:  `Inverts the colors of an image. You can then convert the inverted image to your favorite color scheme.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("invert called")

		switch {
		case len(invertBatchFiles) > 0:
			fmt.Println("Processing batch files...")
			processor := &image.Inverter{}
			expandedFiles := utils.ExpandHomeDirectory(invertBatchFiles)
			image.ProcessBatchImgs(expandedFiles, invertTheme, processor)

		case len(args) > 0 && strings.HasSuffix(args[0], "#"):
			fmt.Println("Processing directory...")
			processor := &image.Inverter{}
			path := utils.DiscardLastCharacter(args[0])
			files, err := utils.ExpandHashtag(path)
			if err != nil {
				fmt.Printf("Error Expanding HashTag: %s\n", err)
				return
			}
			image.ProcessBatchImgs(files, invertTheme, processor)

		case len(args) > 0:
			fmt.Println("Processing single image...")
			processor := &image.Inverter{}
			expandedFile := utils.ExpandHomeDirectory(args)
			if len(expandedFile) > 0 {
				image.ProcessImg(expandedFile[0], processor, invertTheme)
			} else {
				fmt.Println("Error: no valid file path provided")
				_ = cmd.Usage()
			}

		default:
			fmt.Println("Error: requires at least 1 argument, only received 0")
			_ = cmd.Usage()
		}
	},
}

func init() {
	rootCmd.AddCommand(invertCmd)
	invertCmd.Flags().StringVarP(&invertTheme, "theme", "t", "default", "Usage: --theme [ThemeName]")
	invertCmd.Flags().StringSliceVarP(&invertBatchFiles, "batch", "b", nil, "Usage: --batch [file1.png,file2.png ...]")
}
