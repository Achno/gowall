package cmd

import (
	"fmt"
	"slices"
	"strings"

	"github.com/Achno/gowall/config"
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
The mode flag picks the colors: color (keeps the colors of the image), white (white on black), black (black on white)`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return ValidateParseAsciiCmd(cmd, shared, args)
		},
		Run: RunAsciiCmd,
	}

	flags := cmd.Flags()
	var (
		detail int
		mode   string
	)

	flags.IntVarP(&detail, "detail", "d", 50, "Usage: --detail [0-100] (The higher the number == more and smaller characters)")
	flags.StringVarP(&mode, "mode", "m", string(imgbackend.AsciiColor), "Usage: --mode ["+strings.Join(imgbackend.AsciiModes(), ",")+"]")

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

	logger.Print("Converting to ASCII...")
	processor := image.NewAsciiProcessor(
		image.WithAsciiDetail(detail),
		image.WithAsciiMode(imgbackend.AsciiMode(mode)),
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

	return nil
}

func init() {
	rootCmd.AddCommand(BuildAsciiCmd())
}
