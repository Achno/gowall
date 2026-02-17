/*
Copyright © 2024 Achno <EMAIL ADDRESS>
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

func BuildPixelateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pixelate [INPUT] [OPTIONAL OUTPUT]",
		Short: "Turns an image to pixel art depending on the scale flag",
		Long: `It can convert an image to pixel art (blocky appearance). The scale flag [1-25] controls how much the image will get pixelated. 
		   The lower the number the more pixel effect is prevalent. 
		   In really large images with huge resolution you may need to set the scale really low [3-8] `,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return ValidateParsePixelateCmd(cmd, shared, args)
		},
		Run: RunPixelateCmd,
	}

	flags := cmd.Flags()
	var scale float64

	flags.Float64VarP(&scale, "scale", "s", 15, "Usage: --scale [1-25] (The lower the number == more pixelation)")

	addGlobalFlags(cmd)

	return cmd
}

func RunPixelateCmd(cmd *cobra.Command, args []string) {
	imageOps, err := imageio.DetermineImageOperations(shared, args, cmd)
	utils.HandleError(err, "Error")

	scale, err := cmd.Flags().GetFloat64("scale")
	utils.HandleError(err, "Error")

	logger.Print("Pixelating image...")
	processor := &image.PixelateProcessor{
		Scale: scale,
	}

	processedImages, err := image.ProcessImgs(processor, imageOps, image.ProcessOptions{
		Theme:      "",
		OnComplete: nil,
	})
	utils.HandleError(err, "Error")

	if err != nil {
		logger.Error(err, "The following images had errors while processing")
	}

	openImageInViewer(cmd, shared, args, processedImages[0])
}

func ValidateParsePixelateCmd(cmd *cobra.Command, flags config.GlobalSubCommandFlags, args []string) error {
	if err := validateInput(flags, args); err != nil {
		return err
	}

	scale, _ := cmd.Flags().GetFloat64("scale")
	if scale < 1 || scale > 25 {
		return fmt.Errorf("scale must be between 1 and 25, got: %.2f", scale)
	}

	return nil
}

func init() {
	rootCmd.AddCommand(BuildPixelateCmd())
}
