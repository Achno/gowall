/*
Copyright © 2024 Achno <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"github.com/Achno/gowall/internal/image"
	"github.com/Achno/gowall/utils"
	"github.com/spf13/cobra"
)

var (
	maxIter     int
	convergence float64
	sampleRate  float64
	numRoutines int
)

// bgCmd represents the bg command
var bgCmd = &cobra.Command{
	Use:   "bg [PATH]",
	Short: "Removes the background of the image",
	Long:  `Removes the background of an image. You can modify the options to achieve better results `,
	Run: func(cmd *cobra.Command, args []string) {
		switch {

		case len(args) > 0:
			fmt.Println("Removing background...")
			processor := &image.BackgroundProcessor{}
			processor.SetOptions(
				image.WithConvergence(convergence),
				image.WithMaxIter(maxIter),
				image.WithNumRoutines(numRoutines),
				image.WithSampleRate(sampleRate),
			)

			expandFile := utils.ExpandHomeDirectory(args)

			path, _, err := image.ProcessImg(expandFile[0], processor, shared.Theme)
			utils.HandleError(err, "Error Processing Image")

			err = image.OpenImageInViewer(path)
			utils.HandleError(err, "Error opening image")

		default:
			fmt.Println("Error: requires at least 1 arg(s), only received 0")
			_ = cmd.Usage()
		}
	},
}

func init() {
	rootCmd.AddCommand(bgCmd)
	bgCmd.Flags().IntVarP(&maxIter, "iterations", "i", 100, "")
	bgCmd.Flags().IntVarP(&numRoutines, "routines", "r", 4, "")
	bgCmd.Flags().Float64VarP(&convergence, "conv", "c", 0.001, "")
	bgCmd.Flags().Float64VarP(&sampleRate, "sRate", "s", 0.5, "")
}
