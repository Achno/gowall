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

// invertCmd represents the invert command
var invertCmd = &cobra.Command{
	Use:   "invert [image path]",
	Short: "Inverts the color's of an image",
	Long:  `Inverts the color's of an image , then you can convert the inverted image to your favourite color scheme`,
	Run: func(cmd *cobra.Command, args []string) {

		switch {

		case len(shared.BatchFiles) > 0:
			fmt.Println("Processing batch files...")
			processor := &image.Inverter{}
			expandedFiles := utils.ExpandHomeDirectory(shared.BatchFiles)
			err := image.ProcessBatchImgs(expandedFiles, shared.Theme, processor)

			utils.HandleError(err)

		case len(args) > 0 && strings.HasSuffix(args[0], "#"):
			fmt.Println("Processing directory...")
			processor := &image.Inverter{}
			path := utils.DiscardLastCharacter(args[0])
			files, err := utils.ExpandHashtag(path)

			utils.HandleError(err, "Error ExpandingHashTag")

			err = image.ProcessBatchImgs(files, shared.Theme, processor)
			utils.HandleError(err)

		case len(args) > 0:
			fmt.Println("Processing single image...")
			processor := &image.Inverter{}
			expandFile := utils.ExpandHomeDirectory(args)
			err := image.ProcessImg(expandFile[0], processor, shared.Theme)

			utils.HandleError(err)

		default:
			fmt.Println("Error: requires at least 1 arg(s), only received 0")
			_ = cmd.Usage()
		}
	},
}

func init() {
	rootCmd.AddCommand(invertCmd)

}
