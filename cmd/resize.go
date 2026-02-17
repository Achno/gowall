/*
Copyright © 2025 Achno <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/Achno/gowall/config"
	"github.com/Achno/gowall/internal/image"
	imageio "github.com/Achno/gowall/internal/image_io"
	"github.com/Achno/gowall/internal/logger"
	"github.com/Achno/gowall/utils"
	"github.com/spf13/cobra"
)

func BuildResizeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "resize [INPUT]",
		Short: "Resize an image, Aspect ratio is preserved",
		Long:  "Resize an image with specified width, height, and resampling method. Aspect ratio is preserved.",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			//TODO: refactor all the other commands to use this style of validating flags and parsing/splitting args here.
			return ValidateParseResizeCmd(cmd, shared, args)
		},
		Run: RunResizeCmd,
	}

	flags := cmd.Flags()
	var (
		dimensions string
		method     string
		width      int
		height     int
	)

	flags.StringVarP(&dimensions, "dimensions", "d", "", "Dimensions in format WIDTHxHEIGHT (e.g., 1920x1080)")
	flags.StringVarP(&method, "method", "m", "lanczos", "Resampling method: lanczos, catmullrom")

	// Hidden flags to pass parsed values from PreRunE to Run
	flags.IntVar(&width, "width", 0, "")
	flags.IntVar(&height, "height", 0, "")
	cmd.Flags().MarkHidden("width")
	cmd.Flags().MarkHidden("height")

	addGlobalFlags(cmd)

	return cmd
}

func RunResizeCmd(cmd *cobra.Command, args []string) {

	ops, err := imageio.DetermineImageOperations(shared, args, cmd)
	utils.HandleError(err, "Error")

	width, err := cmd.Flags().GetInt("width")
	utils.HandleError(err, "Error")
	height, err := cmd.Flags().GetInt("height")
	utils.HandleError(err, "Error")
	method, err := cmd.Flags().GetString("method")
	utils.HandleError(err, "Error")

	processor := &image.ResizeProcessor{}
	processor.SetOptions(
		image.WithWidth(width),
		image.WithHeight(height),
		image.WithMethod(method),
	)

	logger.Print("Resizing images...")
	resizedImages, err := image.ProcessImgs(processor, ops, image.ProcessOptions{
		Theme:      "",
		OnComplete: nil,
	})
	utils.HandleError(err, "Error")

	openImageInViewer(cmd, shared, args, resizedImages[0])
}

func ValidateParseResizeCmd(cmd *cobra.Command, flags config.GlobalSubCommandFlags, args []string) error {
	if err := validateInput(flags, args); err != nil {
		return err
	}

	dimensions, _ := cmd.Flags().GetString("dimensions")
	if dimensions == "" {
		return fmt.Errorf("dimensions cannot be empty, use format WIDTHxHEIGHT (e.g., 1920x1080)")
	}

	parts := strings.Split(dimensions, "x")
	if len(parts) != 2 {
		return fmt.Errorf("invalid dimensions format: %s, use format WIDTHxHEIGHT (e.g., 1920x1080)", dimensions)
	}

	width, err := strconv.Atoi(parts[0])
	if err != nil {
		return fmt.Errorf("invalid width value: %s", parts[0])
	}

	height, err := strconv.Atoi(parts[1])
	if err != nil {
		return fmt.Errorf("invalid height value: %s", parts[1])
	}

	// Store parsed values in hidden flags for Run to use
	cmd.Flags().Set("width", strconv.Itoa(width))
	cmd.Flags().Set("height", strconv.Itoa(height))

	return nil
}

func init() {
	rootCmd.AddCommand(BuildResizeCmd())
}
