package cmd

import (
	"errors"
	"fmt"

	"github.com/Achno/gowall/config"
	"github.com/Achno/gowall/internal/image"
	imageio "github.com/Achno/gowall/internal/image_io"
	"github.com/Achno/gowall/internal/logger"
	"github.com/Achno/gowall/terminal"
	"github.com/Achno/gowall/utils"
	"github.com/spf13/cobra"
)

func BuildCropCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "crop [INPUT]",
		Short: "Crop an image by selecting a region with the mouse",
		Long:  `Crop an image by dragging a rectangle over it with the mouse, then press enter to save it or q to cancel. Needs a terminal with the kitty image protocol`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return ValidateParseCropCmd(cmd, shared, args)
		},
		Run: RunCropCmd,
	}

	addFlags(cmd).WithOutput().WithPreview()

	return cmd
}

func RunCropCmd(cmd *cobra.Command, args []string) {
	ops, err := imageio.DetermineImageOperations(shared, args, cmd)
	utils.HandleError(err, "Error")

	img, err := imageio.LoadImage(ops[0].ImageInput)
	utils.HandleError(err, "Error")

	region, err := terminal.SelectCropRegion(img)
	if errors.Is(err, terminal.ErrCropCancelled) {
		logger.Print("Crop cancelled")
		return
	}
	utils.HandleError(err, "Error")

	processedImages, err := image.ProcessImgs(&image.CropProcessor{Rect: region}, ops, image.ProcessOptions{
		Theme:      "",
		OnComplete: nil,
	})
	utils.HandleError(err, "Error")

	openImageInViewer(cmd, shared, args, processedImages[0])
}

func ValidateParseCropCmd(cmd *cobra.Command, flags config.GlobalSubCommandFlags, args []string) error {
	if err := validateInput(flags, args); err != nil {
		return err
	}
	if args[0] == "-" || imageio.IsStdoutOutput(flags, args) {
		return fmt.Errorf("crop uses the terminal for the mouse and the image, so it can't read from stdin or write to stdout with -")
	}
	return nil
}

func init() {
	rootCmd.AddCommand(BuildCropCmd())
}
