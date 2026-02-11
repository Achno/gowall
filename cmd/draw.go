/*
Copyright © 2024 Achno <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"github.com/Achno/gowall/config"
	cpkg "github.com/Achno/gowall/internal/backends/color"
	"github.com/Achno/gowall/internal/image"
	imageio "github.com/Achno/gowall/internal/image_io"
	"github.com/Achno/gowall/internal/logger"
	"github.com/Achno/gowall/utils"
	"github.com/spf13/cobra"
)

func BuildDrawCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "draw [DRAW-COMMAND] [INPUT]",
		Short: "Draw on images",
		Long:  `Draw a [border,grid,round] in your images`,
		Run: func(cmd *cobra.Command, args []string) {
			logger.Print("Please specify an draw command to apply")
			err := cmd.Usage()
			utils.HandleError(err)
		},
	}

	cmd.AddCommand(BuildBorderCmd())
	cmd.AddCommand(BuildGridCmd())
	cmd.AddCommand(BuildRoundCmd())

	addGlobalFlags(cmd)

	return cmd
}

// Border Command
func BuildBorderCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "border [PATH]",
		Short: "draw a border with a specified  color and thickness on the image",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return ValidateParseBorderCmd(cmd, shared, args)
		},
		Run: RunBorderCmd,
	}

	flags := cmd.Flags()
	var (
		color           string
		borderThickness int
		cornerRadius    float64
	)

	flags.StringVarP(&color, "color", "c", "", "--color #5D3FD3")
	flags.IntVarP(&borderThickness, "borderThickness", "b", 5, "-b 5")
	flags.Float64VarP(&cornerRadius, "radius", "r", 0, "Corner radius (0 = no rounding)")

	return cmd
}

func RunBorderCmd(cmd *cobra.Command, args []string) {
	logger.Print("Processing images...")

	imageOps, err := imageio.DetermineImageOperations(shared, args, cmd)
	utils.HandleError(err, "Error")

	hex, err := cmd.Flags().GetString("color")
	utils.HandleError(err, "Error")
	borderThickness, err := cmd.Flags().GetInt("borderThickness")
	utils.HandleError(err, "Error")
	cornerRadius, err := cmd.Flags().GetFloat64("radius")
	utils.HandleError(err, "Error")

	clr, err := cpkg.HexToRGBA(hex)
	utils.HandleError(err, "Error")

	processor := &image.BorderProcessor{
		Color:           clr,
		BorderThickness: borderThickness,
		CornerRadius:    cornerRadius,
	}

	processedImages, err := image.ProcessImgs(processor, imageOps, image.ProcessOptions{
		Theme:      "",
		OnComplete: nil,
	})
	utils.HandleError(err, "Error")

	openImageInViewer(shared, args, processedImages[0])
}

func ValidateParseBorderCmd(cmd *cobra.Command, flags config.GlobalSubCommandFlags, args []string) error {
	if err := validateInput(flags, args); err != nil {
		return err
	}

	colorStr, _ := cmd.Flags().GetString("color")
	if colorStr == "" {
		return fmt.Errorf("color must be specified (use --color or -c with a hex color like #5D3FD3)")
	}

	_, err := cpkg.HexToRGBA(colorStr)
	if err != nil {
		return fmt.Errorf("invalid color format: %v (expected format: #RRGGBB, e.g., #5D3FD3)", err)
	}

	cornerRadius, _ := cmd.Flags().GetFloat64("radius")
	if cornerRadius < 0 {
		return fmt.Errorf("corner radius must be greater than 0")
	}
	return nil
}

func BuildGridCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "grid [PATH]",
		Short: "draw a grid with a specific size,color and thickness on the image",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return ValidateParseGridCmd(cmd, shared, args)
		},
		Run: RunGridCmd,
	}

	flags := cmd.Flags()
	var (
		gridSize      int
		gridColor     string
		gridThickness int
		gridMask      bool
	)

	flags.IntVarP(&gridSize, "size", "s", 80, "--size 80")
	flags.IntVarP(&gridThickness, "thickness", "t", 1, "--thickness 1")
	flags.StringVarP(&gridColor, "color", "c", "#5D3FD3", "--color #5D3FD3")
	flags.BoolVarP(&gridMask, "mask", "m", false, "--mask true to use apply the grid only to transparent pixels (background)")

	return cmd
}

func RunGridCmd(cmd *cobra.Command, args []string) {
	logger.Print("Processing images...")

	imageOps, err := imageio.DetermineImageOperations(shared, args, cmd)
	utils.HandleError(err, "Error")

	gridSize, err := cmd.Flags().GetInt("size")
	utils.HandleError(err, "Error")
	gridThickness, err := cmd.Flags().GetInt("thickness")
	utils.HandleError(err, "Error")
	gridColor, err := cmd.Flags().GetString("color")
	utils.HandleError(err, "Error")
	gridMask, err := cmd.Flags().GetBool("mask")
	utils.HandleError(err, "Error")

	processor := &image.GridProcessor{}
	processor.SetGridOptions(
		image.WithGridSize(gridSize),
		image.WithGridColor(gridColor),
		image.WithGridThickness(gridThickness),
		image.WithMaskonly(gridMask),
	)

	processedImages, err := image.ProcessImgs(processor, imageOps, image.ProcessOptions{
		Theme:      "",
		OnComplete: nil,
	})
	utils.HandleError(err, "Error")

	openImageInViewer(shared, args, processedImages[0])
}

func ValidateParseGridCmd(cmd *cobra.Command, flags config.GlobalSubCommandFlags, args []string) error {
	if err := validateInput(flags, args); err != nil {
		return err
	}
	return nil
}

func BuildRoundCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "round [PATH]",
		Short: "round the corners of an image with a specified radius",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return ValidateParseRoundCmd(cmd, shared, args)
		},
		Run: RunRoundCmd,
	}

	flags := cmd.Flags()
	var cornerRadius float64

	flags.Float64VarP(&cornerRadius, "radius", "r", 30, "Corner radius for rounding")

	return cmd
}

func RunRoundCmd(cmd *cobra.Command, args []string) {
	logger.Print("Processing images...")

	imageOps, err := imageio.DetermineImageOperations(shared, args, cmd)
	utils.HandleError(err, "Error")

	cornerRadius, err := cmd.Flags().GetFloat64("radius")
	utils.HandleError(err, "Error")

	processor := &image.RoundProcessor{
		CornerRadius: cornerRadius,
	}

	processedImages, err := image.ProcessImgs(processor, imageOps, image.ProcessOptions{
		Theme:      "",
		OnComplete: nil,
	})
	utils.HandleError(err, "Error")

	openImageInViewer(shared, args, processedImages[0])
}

func ValidateParseRoundCmd(cmd *cobra.Command, flags config.GlobalSubCommandFlags, args []string) error {
	if err := validateInput(flags, args); err != nil {
		return err
	}

	cornerRadius, _ := cmd.Flags().GetFloat64("radius")
	if cornerRadius <= 0 {
		return fmt.Errorf("corner radius must be greater than 0")
	}
	return nil
}

func init() {
	rootCmd.AddCommand(BuildDrawCmd())
}
