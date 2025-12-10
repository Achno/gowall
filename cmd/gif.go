/*
Copyright © 2025 Achno
*/
package cmd

import (
	"fmt"

	"github.com/Achno/gowall/config"
	"github.com/Achno/gowall/internal/image"
	imageio "github.com/Achno/gowall/internal/image_io"
	"github.com/Achno/gowall/internal/logger"
	"github.com/Achno/gowall/utils"
	"github.com/spf13/cobra"
)

func BuildGifCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "gif [--batch,--dir] [PATH(S)]",
		Short: "Create a gif Animation out of Images",
		Long:  `Create a gif Animation out of Images specifying the delay between frames, if the gif loops forever and other options`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return ValidateParseGifCmd(cmd, shared, args)
		},
		Run: RunGifCmd,
	}

	flags := cmd.Flags()
	var (
		delay  int
		loop   int
		resize int
	)

	flags.IntVarP(&delay, "delay", "d", 200, "Frame delay (ms)")
	flags.IntVarP(&resize, "resize", "r", image.Resize, "Automatically resizes all images to the same dimensions")
	flags.IntVarP(&loop, "loop", "l", 0, "Loop=0 (loops forever), Loop=-1 shows frames only 1 time, Loop=n (shows frames n+1)")

	addGlobalFlags(cmd)

	return cmd
}

func RunGifCmd(cmd *cobra.Command, args []string) {
	logger.Print("Creating Gif...")

	imageOps, err := imageio.DetermineImageOperations(shared, args, cmd)
	utils.HandleError(err, "Error")

	delay, err := cmd.Flags().GetInt("delay")
	utils.HandleError(err, "Error")
	loop, err := cmd.Flags().GetInt("loop")
	utils.HandleError(err, "Error")
	resize, err := cmd.Flags().GetInt("resize")
	utils.HandleError(err, "Error")

	options := []image.GifOption{}
	if cmd.Flags().Changed("delay") {
		options = append(options, image.WithDelay(delay))
	}
	if cmd.Flags().Changed("loop") {
		options = append(options, image.WithLoop(loop))
	}
	// if cmd.Flags().Changed("output") {
	// 	options = append(options, image.WithOutputName(shared.OutputDestination))
	// }
	if cmd.Flags().Changed("resize") {
		options = append(options, image.WithMode(resize))
	}

	err = image.CreateGif(imageOps, options...)
	utils.HandleError(err, "Error")
}

func ValidateParseGifCmd(cmd *cobra.Command, flags config.GlobalSubCommandFlags, args []string) error {
	if err := validateInput(flags, args); err != nil {
		return err
	}

	if len(args) > 0 {
		return fmt.Errorf("use --batch or --dir with gif with multiple images")
	}

	delay, _ := cmd.Flags().GetInt("delay")
	if delay < 0 {
		return fmt.Errorf("delay must be non-negative, got: %d", delay)
	}

	return nil
}

func init() {
	rootCmd.AddCommand(BuildGifCmd())
}
