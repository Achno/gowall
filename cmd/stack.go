/*
Copyright © 2026 Achno
*/
package cmd

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/Achno/gowall/config"
	cpkg "github.com/Achno/gowall/internal/backends/color"
	"github.com/Achno/gowall/internal/image"
	imageio "github.com/Achno/gowall/internal/image_io"
	"github.com/Achno/gowall/internal/logger"
	"github.com/Achno/gowall/utils"
	"github.com/spf13/cobra"
)

func BuildStackCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "stack [--batch,--dir]",
		Short: "Stack multiple images into a single image",
		Long:  "Stack multiple images in a horizontal, vertical, or NxM grid layout.",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return ValidateParseStackCmd(cmd, shared, args)
		},
		Run: RunStackCmd,
	}

	flags := cmd.Flags()
	var (
		layout     string
		border     int
		colorStr   string
		resizeMode string
	)

	flags.StringVarP(&layout, "layout", "l", "vertical", "Layout: horizontal, vertical, or NxM grid (e.g., 2x2)")
	flags.IntVarP(&border, "border", "b", 0, "Border thickness in pixels between and around images")
	flags.StringVarP(&colorStr, "color", "c", "#000000", "Border color as hex/rgb/hsl/lab or named color")
	flags.StringVarP(&resizeMode, "resize", "r", image.StackResizeOff, "Resize mode: biggest (resizes all images down to match the smallest image's dimensions) or off")

	cmd.RegisterFlagCompletionFunc("resize", stackResizeCompletion)
	cmd.RegisterFlagCompletionFunc("layout", stackLayoutCompletion)

	addGlobalFlags(cmd)

	return cmd
}

func RunStackCmd(cmd *cobra.Command, args []string) {
	logger.Print("Stacking images...")

	imageOps, err := imageio.DetermineImageOperations(shared, args, cmd)
	utils.HandleError(err, "Error")

	layout, err := cmd.Flags().GetString("layout")
	utils.HandleError(err, "Error")
	border, err := cmd.Flags().GetInt("border")
	utils.HandleError(err, "Error")
	colorStr, err := cmd.Flags().GetString("color")
	utils.HandleError(err, "Error")
	resizeMode, err := cmd.Flags().GetString("resize")
	utils.HandleError(err, "Error")

	color, err := cpkg.HexToRGBA(colorStr)
	utils.HandleError(err, "Error")
	layoutMode, rows, cols, err := parseStackLayout(layout)
	utils.HandleError(err, "Error")

	processor := &image.StackProcessor{
		LayoutMode:      layoutMode,
		Rows:            rows,
		Cols:            cols,
		BorderThickness: border,
		BorderColor:     color,
		ResizeMode:      strings.ToLower(resizeMode),
	}

	outputPath, err := image.MultiProcessImgs(processor, imageOps, image.MultiProcessOptions{
		Theme:      "",
		OnComplete: nil,
	})
	utils.HandleError(err, "Error")

	openImageInViewer(cmd, shared, args, outputPath)
}

func ValidateParseStackCmd(cmd *cobra.Command, flags config.GlobalSubCommandFlags, args []string) error {
	if err := validateInput(flags, args); err != nil {
		return err
	}

	if len(args) > 0 {
		return fmt.Errorf("use --batch or --dir with stack")
	}

	layout, _ := cmd.Flags().GetString("layout")
	_, _, _, err := parseStackLayout(layout)
	if err != nil {
		return err
	}

	border, _ := cmd.Flags().GetInt("border")
	if border < 0 {
		return fmt.Errorf("border must be non-negative, got: %d", border)
	}

	resizeMode, _ := cmd.Flags().GetString("resize")
	resizeMode = strings.ToLower(resizeMode)
	if resizeMode != image.StackResizeOff && resizeMode != image.StackResizeBiggest {
		return fmt.Errorf("invalid resize mode '%s', valid modes: %s, %s", resizeMode, image.StackResizeOff, image.StackResizeBiggest)
	}

	colorStr, _ := cmd.Flags().GetString("color")
	hexColor, err := cpkg.ParseColorToHex(colorStr)
	if err != nil {
		return err
	}
	cmd.Flags().Set("color", hexColor)

	return nil
}

func parseStackLayout(layout string) (string, int, int, error) {
	layout = strings.ToLower(strings.TrimSpace(layout))

	switch layout {
	case image.StackLayoutHorizontal:
		return image.StackLayoutHorizontal, 0, 0, nil
	case image.StackLayoutVertical:
		return image.StackLayoutVertical, 0, 0, nil
	}

	parts := strings.Split(layout, "x")
	if len(parts) != 2 {
		return "", 0, 0, fmt.Errorf("invalid layout '%s', expected: horizontal, vertical, or NxM like 2x2", layout)
	}

	rows, err := strconv.Atoi(parts[0])
	if err != nil || rows <= 0 {
		return "", 0, 0, fmt.Errorf("invalid grid rows in layout '%s'", layout)
	}

	cols, err := strconv.Atoi(parts[1])
	if err != nil || cols <= 0 {
		return "", 0, 0, fmt.Errorf("invalid grid columns in layout '%s'", layout)
	}

	return image.StackLayoutGrid, rows, cols, nil
}

func stackResizeCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return image.StackResizeList(), cobra.ShellCompDirectiveNoFileComp
}

func stackLayoutCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return image.StackLayoutList(), cobra.ShellCompDirectiveNoFileComp
}

func init() {
	rootCmd.AddCommand(BuildStackCmd())
}
