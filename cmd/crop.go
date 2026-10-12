package cmd

import (
	"errors"
	"fmt"

	"github.com/Achno/gowall/config"
	ipkg "github.com/Achno/gowall/internal/backends/image"
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
		Long:  `Crop an image by dragging a rectangle over it with the mouse, then press enter to save it or q to cancel. Needs a terminal with the kitty image protocol. Use --rect to crop without the mouse, which also works with stdin/stdout`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return ValidateParseCropCmd(cmd, shared, args)
		},
		Run: RunCropCmd,
	}

	var rect string
	cmd.Flags().StringVarP(&rect, "rect", "r", "", "Crop without the mouse, in pixels: x,y,width,height (e.g. 100,50,800,600)")

	addFlags(cmd).WithOutput().WithPreview()

	return cmd
}

func RunCropCmd(cmd *cobra.Command, args []string) {
	ops, err := imageio.DetermineImageOperations(shared, args, cmd)
	utils.HandleError(err, "Error")

	rect, err := cmd.Flags().GetString("rect")
	utils.HandleError(err, "Error")

	processor := &image.CropProcessor{}
	if rect != "" {
		processor.Rect, err = ipkg.ParseCropRect(rect)
	} else {
		img, loadErr := imageio.LoadImage(ops[0].ImageInput)
		utils.HandleError(loadErr, "Error")

		processor.Rect, err = terminal.SelectCropRegion(img)
	}

	if errors.Is(err, terminal.ErrCropCancelled) {
		logger.Print("Crop cancelled")
		return
	}
	utils.HandleError(err, "Error")

	processedImages, err := image.ProcessImgs(processor, ops, image.ProcessOptions{
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
	rect, _ := cmd.Flags().GetString("rect")
	if rect != "" {
		_, err := ipkg.ParseCropRect(rect)
		return err
	}
	if args[0] == "-" || imageio.IsStdoutOutput(flags, args) {
		return fmt.Errorf("crop uses the terminal for the mouse and the image, so it can't read from stdin or write to stdout with -, use --rect for that")
	}
	return nil
}

func init() {
	rootCmd.AddCommand(BuildCropCmd())
}
