/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"strings"

	"github.com/Achno/gowall/config"
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
			return ValidateParseCompressCmd(cmd, shared, args)
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

func ValidateParseCompressCmd(cmd *cobra.Command, flags config.GlobalSubCommandFlags, args []string) error {
	if err := validateInput(flags, args); err != nil {
		return err
	}

	quality, _ := cmd.Flags().GetInt("quality")
	if quality < 0 || quality > 100 {
		return fmt.Errorf("quality must be between 1 and 100, got: %d", quality)
	}

	speed, _ := cmd.Flags().GetInt("speed")
	if speed < 0 || speed > 11 {
		return fmt.Errorf("speed must be between 1 and 10, got: %d", speed)
	}

	return nil
}

func init() {
	rootCmd.AddCommand(BuildCompressCmd())
}
