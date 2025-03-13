/*
Copyright © 2024 Achno <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/Achno/gowall/internal/image"
	imageio "github.com/Achno/gowall/internal/image_io"
	"github.com/Achno/gowall/internal/logger"
	"github.com/Achno/gowall/utils"
	"github.com/spf13/cobra"
)

var (
	colorB          string
	BorderThickness int
)

var drawCmd = &cobra.Command{
	Use:   "draw [PATH] [OPTIONAL OUTPUT]",
	Short: "draw a border with a color and thickness (currently)",
	Long:  `The draw command allows you to draw a plethora of effects. Currently only drawing a border is supported with more to come`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		err := validateInput(shared, args)
		if err != nil {
			logger.Print("Error: requires at least 1 arg(s) and options, received 0 input")
			return err
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) > 0 {
			logger.Print("Processing single image...")
		}
		if isInputBatch(shared) {
			logger.Print("Processing batch of images...")
		}
		hex, err := cmd.Flags().GetString("color")
		utils.HandleError(err, "Error")

		clr, err := image.HexToRGBA(hex)
		utils.HandleError(err, "Error")

		processor := &image.DrawProcessor{
			Color:           clr,
			BorderThickness: BorderThickness,
		}

		imageOps := imageio.DetermineImageOperations(shared, args)

		logger.Print("Processing images...")

		processedImages, err := image.ProcessImgs(processor, imageOps, theme)

		if len(processedImages) == 0 {
			utils.HandleError(err, "No images were processed")
		}

		if err != nil {
			logger.Error(err, "The following images had errors while processing")
		}

		openImageInViewer(shared, args, processedImages[0])
	},
}

func init() {
	rootCmd.AddCommand(drawCmd)
	drawCmd.Flags().StringVarP(&colorB, "color", "c", "#5D3FD3", "--color #5D3FD3")
	drawCmd.Flags().IntVarP(&BorderThickness, "borderThickness", "b", 5, "-b 5")
	addGlobalFlags(drawCmd)
}
