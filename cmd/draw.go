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

	gridSize      int
	gridColor     string
	gridThickness int
	gridMask      bool
)

var drawCmd = &cobra.Command{
	Use:   "draw [DRAW-COMMAND] [INPUT] [OPTIONAL OUTPUT]",
	Short: "Draw on images",
	Long:  `Draw a [border,grid] in your images`,
	Run: func(cmd *cobra.Command, args []string) {
		logger.Print("Please specify an draw command to apply")
		err := cmd.Usage()
		utils.HandleError(err)
	},
}

var GridCmd = &cobra.Command{
	Use:   "grid [PATH] [OPTIONAL OUTPUT]",
	Short: "draw a grid with a specific size,color and thickness on the image",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		err := validateInput(shared, args)
		if err != nil {
			return err
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		imageOps, err := imageio.DetermineImageOperations(shared, args)
		utils.HandleError(err)
		logger.Print("Processing images...")

		processor := &image.GridProcessor{}
		processor.SetGridOptions(
			image.WithGridSize(gridSize),
			image.WithGridColor(gridColor),
			image.WithGridThickness(gridThickness),
			image.WithMaskonly(gridMask),
		)
		processedImages, err := image.ProcessImgs(processor, imageOps, theme)
		utils.HandleError(err, "Error")

		if err != nil {
			logger.Error(err, "The following images had errors while processing")
		}
		openImageInViewer(shared, args, processedImages[0])
	},
}

var BorderCmd = &cobra.Command{
	Use:   "border [PATH] [OPTIONAL OUTPUT]",
	Short: "draw a border with a specified  color and thickness on the image",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		err := validateInput(shared, args)
		if err != nil {
			return err
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		hex, err := cmd.Flags().GetString("color")
		utils.HandleError(err, "Error")

		clr, err := image.HexToRGBA(hex)
		utils.HandleError(err, "Error")

		processor := &image.BorderProcessor{
			Color:           clr,
			BorderThickness: BorderThickness,
		}

		imageOps, err := imageio.DetermineImageOperations(shared, args)
		utils.HandleError(err)

		logger.Print("Processing images...")

		processedImages, err := image.ProcessImgs(processor, imageOps, theme)
		utils.HandleError(err, "Error")

		if err != nil {
			logger.Error(err, "The following images had errors while processing")
		}

		openImageInViewer(shared, args, processedImages[0])
	},
}

func init() {
	rootCmd.AddCommand(drawCmd)

	drawCmd.AddCommand(BorderCmd)
	drawCmd.AddCommand(GridCmd)

	BorderCmd.Flags().StringVarP(&colorB, "color", "c", "#5D3FD3", "--color #5D3FD3")
	BorderCmd.Flags().IntVarP(&BorderThickness, "borderThickness", "b", 5, "-b 5")
	GridCmd.Flags().IntVarP(&gridSize, "size", "s", 80, "--size 80")
	GridCmd.Flags().IntVarP(&gridThickness, "thickness", "t", 1, "--thickness 1")
	GridCmd.Flags().StringVarP(&gridColor, "color", "c", "#5D3FD3", "--color #5D3FD3")
	GridCmd.Flags().BoolVarP(&gridMask, "mask", "m", false, "--mask true to use apply the grid only to transparent pixels (background)")

	addGlobalFlags(drawCmd)
}
