/*
Copyright © 2024 Achnologia <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/Achno/gowall/config"
	"github.com/Achno/gowall/internal/image"
	imageio "github.com/Achno/gowall/internal/image_io"
	"github.com/Achno/gowall/internal/logger"
	"github.com/Achno/gowall/utils"
	"github.com/spf13/cobra"
)

func BuildInvertCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "invert [INPUT] [OPTIONAL OUTPUT]",
		Short: "Inverts the color's of an image",
		Long:  `Inverts the color's of an image , then you can convert the inverted image to your favourite color scheme`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return ValidateParseInvertCmd(cmd, shared, args)
		},
		Run: RunInvertCmd,
	}

	addGlobalFlags(cmd)

	return cmd
}

func RunInvertCmd(cmd *cobra.Command, args []string) {

	logger.Print("Processing images...")

	imageOps, err := imageio.DetermineImageOperations(shared, args, cmd)
	utils.HandleError(err, "Error")

	processor := &image.Inverter{}

	processedImages, err := image.ProcessImgs(processor, imageOps, image.ProcessOptions{
		Theme:      "",
		OnComplete: nil,
	})
	utils.HandleError(err, "Error")

	openImageInViewer(cmd, shared, args, processedImages[0])
}

func ValidateParseInvertCmd(cmd *cobra.Command, flags config.GlobalSubCommandFlags, args []string) error {
	if err := validateInput(flags, args); err != nil {
		return err
	}

	return nil
}

func init() {
	rootCmd.AddCommand(BuildInvertCmd())
}
