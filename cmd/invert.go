/*
Copyright © 2024 Achnologia <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"github.com/Achno/gowall/internal/image"
	"github.com/Achno/gowall/utils"
	"github.com/spf13/cobra"
)

// invertCmd represents the invert command
var invertCmd = &cobra.Command{
	Use:   "invert [image path]",
	Short: "Inverts the color's of an image",
	Long: `Inverts the color's of an image , then you can convert the inverted image to your favourite color scheme`,
	Run: func(cmd *cobra.Command, args []string) {

		fmt.Println("invert called")

		switch {

		case len(batchFiles) > 0:
			fmt.Println("Processing batch files...")
			processor := &image.Inverter{}
			expandedFiles := utils.ExpandHomeDirectory(batchFiles)
			image.ProcessBatchImgs(expandedFiles,theme,processor)

		case len(args) > 0:
			fmt.Println("Processing single image...")
			processor := &image.Inverter{}
			expandFile := utils.ExpandHomeDirectory(args)
			image.ProcessImg(expandFile[0], processor,theme)
			
		default:
			fmt.Println("Error: requires at least 1 arg(s), only received 0")
			_ = cmd.Usage()
		}
	},
}

func init() {
	rootCmd.AddCommand(invertCmd)


}
