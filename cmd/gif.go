/*
Copyright © 2025 Achno
*/
package cmd

import (
	"fmt"

	"github.com/Achno/gowall/internal/image"
	"github.com/Achno/gowall/utils"
	"github.com/spf13/cobra"
)

var delay int
var loop int
var gifCmd = &cobra.Command{
	Use:   "gif",
	Short: "Create a gif Animation out of Images",
	Long:  `Create a gif Animation out of Images specifying the delay between frames, if the gif loops forever and other options`,
	Run: func(cmd *cobra.Command, args []string) {
		switch {
		case cmd.Flags().Changed("batch"):
			fmt.Println("Creating Gif...")

			options := []image.GifOption{}
			if cmd.Flags().Changed("delay") {
				options = append(options, image.WithDelay(delay))
			}
			if cmd.Flags().Changed("loop") {
				options = append(options, image.WithLoop(loop))
			}
			if cmd.Flags().Changed("output") {
				options = append(options, image.WithOutputName(outputName))
			}

			expandedFiles := utils.ExpandHomeDirectory(shared.BatchFiles)
			err := image.CreateGif(expandedFiles, options...)
			utils.HandleError(err)

		default:
			fmt.Println("Error: requires at least 1 option `-b` where you specify the filePaths, only received 0")
			fmt.Println("Use: gowall gif -b <file,file>")
			_ = cmd.Usage()
		}
	},
}

func init() {
	rootCmd.AddCommand(gifCmd)
	gifCmd.Flags().StringSliceVarP(&shared.BatchFiles, "batch", "b", nil, "Usage: --batch file1.png,file2.png ...")
	gifCmd.Flags().IntVarP(&delay, "delay", "d", 200, "Frame delay (ms)")
	gifCmd.Flags().IntVarP(&loop, "loop", "l", 0, "Loop=0 (loops forever), Loop=-1 shows frames only 1 time, Loop=n (shows frames n+1)")
	gifCmd.Flags().StringVarP(&outputName, "output", "o", "", "Output filename")
}
