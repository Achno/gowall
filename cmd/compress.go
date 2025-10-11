/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"strings"

	"github.com/Achno/gowall/internal/image"
	imageio "github.com/Achno/gowall/internal/image_io"
	"github.com/Achno/gowall/internal/logger"
	"github.com/Achno/gowall/utils"
	"github.com/spf13/cobra"
)

func BuildCompressCmd() *cobra.Command {
	strategies := image.NewCompressionProcessor().GetAllStrategiesNames()
	cmd := &cobra.Command{
		Use:   "compress [INPUT]",
		Short: "Compress an image, using  ",
		Long:  "Compress an image",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			err := validateInput(shared, args)
			if err != nil {
				return err
			}
			return nil
		},
		Run: RunCompressCmd,
	}

	flags := cmd.Flags()
	var (
		method  string
		quality int
		speed   int
	)

	flags.StringVarP(&method, "method", "m", "", "Available methods: "+strings.Join(strategies, ", "))
	flags.IntVarP(&quality, "quality", "q", 80, "Quality to use for compression")
	flags.IntVarP(&speed, "speed", "s", 4, "Speed to use for compression")

	addGlobalFlags(cmd)

	return cmd
}

func RunCompressCmd(cmd *cobra.Command, args []string) {

	ops, err := imageio.DetermineImageOperations(shared, args, cmd)
	utils.HandleError(err, "Error")

	method, err := cmd.Flags().GetString("method")
	utils.HandleError(err, "Error")
	quality, err := cmd.Flags().GetInt("quality")
	utils.HandleError(err, "Error")
	speed, err := cmd.Flags().GetInt("speed")
	utils.HandleError(err, "Error")

	processor := image.NewCompressionProcessor(
		image.WithStrategy(method),
		image.WithQuality(quality),
		image.WithSpeed(speed),
	)

	logger.Print("Compressing images...")
	compressedImages, err := image.ProcessImgs(processor, ops, "")
	utils.HandleError(err, "Error")

	openImageInViewer(shared, args, compressedImages[0])
}

func init() {
	rootCmd.AddCommand(BuildCompressCmd())
}
