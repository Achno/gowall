package cmd

import (
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/Achno/gowall/config"
	cpkg "github.com/Achno/gowall/internal/backends/color"
	imgbackend "github.com/Achno/gowall/internal/backends/image"
	"github.com/Achno/gowall/internal/image"
	imageio "github.com/Achno/gowall/internal/image_io"
	"github.com/Achno/gowall/internal/logger"
	"github.com/Achno/gowall/utils"
	"github.com/spf13/cobra"
)

func BuildAsciiCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ascii [INPUT] [OPTIONAL OUTPUT]",
		Short: "Turns an image or an animated gif into ASCII art",
		Long: `Turns an image or an animated gif into ASCII art, gifs stay animated. The output keeps the size of the input.
The detail flag [0-100] controls how many characters are used, the higher the number the more and smaller characters.
The mode flag picks the colors: color (keeps the colors of the image), white (white on black), black (black on white).
The bg flag overrides the background color of the mode.
The size flag sets the output size, e.g. --size 1920x1080 to make a wallpaper out of any image or gif.
The scale flag decides how the image is placed on that size: fit (the whole image centered, the rest is background)`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return ValidateParseAsciiCmd(cmd, shared, args)
		},
		Run: RunAsciiCmd,
	}

	flags := cmd.Flags()
	var (
		detail int
		mode   string
		bg     string
		size   string
		scale  string
		width  int
		height int
	)

	flags.IntVarP(&detail, "detail", "d", 50, "Usage: --detail [0-100] (The higher the number == more and smaller characters)")
	flags.StringVarP(&mode, "mode", "m", string(imgbackend.AsciiColor), "Usage: --mode ["+strings.Join(imgbackend.AsciiModes(), ",")+"]")
	flags.StringVarP(&bg, "bg", "b", cpkg.RGBtoHex(imgbackend.AsciiDark), "Usage: --bg [#hex] background color, --mode black defaults to "+cpkg.RGBtoHex(imgbackend.AsciiLight))

	flags.StringVarP(&size, "size", "s", "", "Usage: --size [WIDTHxHEIGHT] output size in px (e.g. 1920x1080), keeps the size of the input when empty")
	flags.StringVar(&scale, "scale", string(imgbackend.AsciiFit), "Usage: --scale ["+strings.Join(imgbackend.AsciiScales(), ",")+"] how the image is placed on --size")

	// Hidden flags to pass the parsed --size from PreRunE to Run, 0 keeps the size of the input
	flags.IntVar(&width, "width", 0, "")
	flags.IntVar(&height, "height", 0, "")
	cmd.Flags().MarkHidden("width")
	cmd.Flags().MarkHidden("height")

	cmd.RegisterFlagCompletionFunc("mode", asciiModeCompletion)
	cmd.RegisterFlagCompletionFunc("scale", asciiScaleCompletion)

	addGlobalFlags(cmd)

	return cmd
}

func RunAsciiCmd(cmd *cobra.Command, args []string) {
	imageOps, err := imageio.DetermineImageOperations(shared, args, cmd)
	utils.HandleError(err, "Error")

	detail, err := cmd.Flags().GetInt("detail")
	utils.HandleError(err, "Error")
	mode, err := cmd.Flags().GetString("mode")
	utils.HandleError(err, "Error")
	bgStr, err := cmd.Flags().GetString("bg")
	utils.HandleError(err, "Error")

	bg, err := cpkg.HexToRGBA(bgStr)
	utils.HandleError(err, "Error")
	width, err := cmd.Flags().GetInt("width")
	utils.HandleError(err, "Error")
	height, err := cmd.Flags().GetInt("height")
	utils.HandleError(err, "Error")
	scale, err := cmd.Flags().GetString("scale")
	utils.HandleError(err, "Error")

	logger.Print("Converting to ASCII...")
	processor := image.NewAsciiProcessor(
		image.WithAsciiDetail(detail),
		image.WithAsciiMode(imgbackend.AsciiMode(mode)),
		image.WithAsciiBackground(bg),
		image.WithAsciiSize(width, height),
		image.WithAsciiScale(imgbackend.AsciiScale(scale)),
	)

	processedImages, err := image.ProcessImgs(processor, imageOps, image.ProcessOptions{
		Theme:      "",
		OnComplete: nil,
	})
	utils.HandleError(err, "Error")

	openImageInViewer(cmd, shared, args, processedImages[0])
}

func ValidateParseAsciiCmd(cmd *cobra.Command, flags config.GlobalSubCommandFlags, args []string) error {
	if err := validateInput(flags, args); err != nil {
		return err
	}

	detail, _ := cmd.Flags().GetInt("detail")
	if detail < 0 || detail > 100 {
		return fmt.Errorf("detail must be between 0 and 100, got: %d", detail)
	}

	mode, _ := cmd.Flags().GetString("mode")
	if !slices.Contains(imgbackend.AsciiModes(), mode) {
		return fmt.Errorf("unsupported mode %q, use one of: %s", mode, strings.Join(imgbackend.AsciiModes(), ", "))
	}

	if mode == string(imgbackend.AsciiBlack) && !cmd.Flags().Changed("bg") {
		cmd.Flags().Set("bg", cpkg.RGBtoHex(imgbackend.AsciiLight))
	}

	bg, _ := cmd.Flags().GetString("bg")
	if _, err := cpkg.HexToRGBA(bg); err != nil {
		return fmt.Errorf("--bg must be a hex color like #1E1E2E, got: %s", bg)
	}

	scale, _ := cmd.Flags().GetString("scale")
	if !slices.Contains(imgbackend.AsciiScales(), scale) {
		return fmt.Errorf("unsupported scale %q, use one of: %s", scale, strings.Join(imgbackend.AsciiScales(), ", "))
	}

	size, _ := cmd.Flags().GetString("size")
	if size == "" {
		if cmd.Flags().Changed("scale") {
			return fmt.Errorf("--scale only works together with --size")
		}
		return nil
	}

	parts := strings.Split(strings.ToLower(size), "x")
	if len(parts) != 2 {
		return fmt.Errorf("invalid size format: %s, use format WIDTHxHEIGHT (e.g., 1920x1080)", size)
	}

	// gif sizes are 16 bit
	width, err := strconv.Atoi(parts[0])
	if err != nil || width < 1 || width > 65535 {
		return fmt.Errorf("invalid width value: %s", parts[0])
	}

	height, err := strconv.Atoi(parts[1])
	if err != nil || height < 1 || height > 65535 {
		return fmt.Errorf("invalid height value: %s", parts[1])
	}

	// Store parsed values in hidden flags for Run to use
	cmd.Flags().Set("width", strconv.Itoa(width))
	cmd.Flags().Set("height", strconv.Itoa(height))

	return nil
}

func asciiModeCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return imgbackend.AsciiModes(), cobra.ShellCompDirectiveNoFileComp
}

func asciiScaleCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return imgbackend.AsciiScales(), cobra.ShellCompDirectiveNoFileComp
}

func init() {
	rootCmd.AddCommand(BuildAsciiCmd())
}
