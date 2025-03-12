/*
Copyright © 2025 Achno
*/
package cmd

import (
	"fmt"

	"github.com/Achno/gowall/internal/image"
	imageio "github.com/Achno/gowall/internal/image_io"
	"github.com/Achno/gowall/internal/logger"
	"github.com/Achno/gowall/utils"
	"github.com/spf13/cobra"
)

var (
	delay  int
	loop   int
	gifCmd = &cobra.Command{
		Use:   "gif -b [PATHS] --flags",
		Short: "Create a gif Animation out of Images",
		Long:  `Create a gif Animation out of Images specifying the delay between frames, if the gif loops forever and other options`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			err := validateInput(shared, args)
			if err != nil {
				logger.Print("Error: requires at least 1 option `-b` where you specify the filePaths, or `--dir`")
				logger.Print("Use: gowall gif -b <file,file>")
				logger.Print("Use: gowall gif -d </path/to/dir>")
				cmd.Usage()
				return err
			}
			if len(args) > 0 {
				return fmt.Errorf("cannot use positional args in the gif command")
			}
			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			logger.Print("Creating Gif...")

			options := []image.GifOption{}
			if cmd.Flags().Changed("delay") {
				options = append(options, image.WithDelay(delay))
			}
			if cmd.Flags().Changed("loop") {
				options = append(options, image.WithLoop(loop))
			}
			if cmd.Flags().Changed("output") {
				options = append(options, image.WithOutputName(shared.OutputDestination))
			}
			imageOps := imageio.DetermineImageOperations(shared, args)
			err := image.CreateGif(imageOps, options...)
			utils.HandleError(err)
		},
	}
)

func init() {
	rootCmd.AddCommand(gifCmd)
	gifCmd.Flags().IntVarP(&delay, "delay", "d", 200, "Frame delay (ms)")
	gifCmd.Flags().IntVarP(&loop, "loop", "l", 0, "Loop=0 (loops forever), Loop=-1 shows frames only 1 time, Loop=n (shows frames n+1)")
	addGlobalFlags(gifCmd)
}
