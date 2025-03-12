/*
Copyright © 2024 Achnologia <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/Achno/gowall/internal/image"
	imageio "github.com/Achno/gowall/internal/image_io"
	"github.com/Achno/gowall/internal/logger"
	"github.com/Achno/gowall/utils"
	"github.com/spf13/cobra"
)

// invertCmd represents the invert command
var invertCmd = &cobra.Command{
	Use:   "invert [image path]",
	Short: "Inverts the color's of an image",
	Long:  `Inverts the color's of an image , then you can convert the inverted image to your favourite color scheme`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		err := validateInput(shared, args)
		if err != nil {
			cmd.Usage()
			return err
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		if isInputBatch(shared) {
			logger.Print("Processing batch files...")
		} else {
			logger.Print("Processing single image...")
		}

		processor := &image.Inverter{}

		imageOps := imageio.DetermineImageOperations(shared, args)

		processedImages, err := image.ProcessImgs(processor, imageOps, "")

		if len(processedImages) == 0 {
			utils.HandleError(err, "Error Processing Images")
		}
		if err != nil {
			logger.Error(err, "The following images had errors while processing")
		}
		openImageInViewer(shared, args, processedImages[0])
	},
}

func init() {
	rootCmd.AddCommand(invertCmd)
	addGlobalFlags(invertCmd)
}
