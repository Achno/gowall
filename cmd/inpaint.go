package cmd

import (
	"fmt"
	"slices"
	"strings"

	"github.com/Achno/gowall/config"
	ipkg "github.com/Achno/gowall/internal/backends/image"
	"github.com/Achno/gowall/internal/image"
	imageio "github.com/Achno/gowall/internal/image_io"
	"github.com/Achno/gowall/internal/logger"
	"github.com/Achno/gowall/terminal"
	"github.com/Achno/gowall/utils"
	"github.com/spf13/cobra"
)

func BuildInpaintCmd() *cobra.Command {
	strategies := image.NewInpaintProcessor().GetAllStrategiesNames()
	cmd := &cobra.Command{
		Use:   "inpaint [INPUT]",
		Short: "Erase areas of an image, like watermarks or text, and fill them in from their surroundings",
		Long:  `Erase areas of an image by dragging rectangles over them with the mouse (u undoes the last one), then press enter to fill them in or q to cancel. Needs a terminal with the kitty image protocol. Use --rect to mark areas without the mouse, which also works with stdin/stdout`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return ValidateParseInpaintCmd(cmd, shared, args)
		},
		Run: RunInpaintCmd,
	}

	flags := cmd.Flags()
	var (
		rects  []string
		method string
	)
	flags.StringArrayVarP(&rects, "rect", "r", nil, "Area to erase without the mouse, in pixels: x,y,width,height (e.g. 100,50,800,600). Repeat it for more areas")
	flags.StringVarP(&method, "method", "m", "", "Available methods: "+strings.Join(strategies, ", "))

	cmd.RegisterFlagCompletionFunc("method", inpaintCompletion)

	addFlags(cmd).WithBatch().WithDir().WithOutput().WithPreview()

	return cmd
}

func RunInpaintCmd(cmd *cobra.Command, args []string) {
	ops, err := imageio.DetermineImageOperations(shared, args, cmd)
	utils.HandleError(err, "Error")

	rectFlags, err := cmd.Flags().GetStringArray("rect")
	utils.HandleError(err, "Error")
	method, err := cmd.Flags().GetString("method")
	utils.HandleError(err, "Error")

	rects, err := ipkg.ParseRects(rectFlags...)
	utils.HandleError(err, "Error")
	if len(rects) == 0 {
		img, err := imageio.LoadImage(ops[0].ImageInput)
		utils.HandleError(err, "Error")

		rects, err = terminal.SelectRegions(img)
		utils.HandleError(err, "Error")
	}

	processor := image.NewInpaintProcessor(
		image.WithRects(rects),
		image.WithInpaintStrategy(method),
	)

	logger.Print("Inpainting...")
	processedImages, err := image.ProcessImgs(processor, ops, image.ProcessOptions{
		Theme:      "",
		OnComplete: nil,
	})
	utils.HandleError(err, "Error")

	openImageInViewer(cmd, shared, args, processedImages[0])
}

func ValidateParseInpaintCmd(cmd *cobra.Command, flags config.GlobalSubCommandFlags, args []string) error {
	if err := validateInput(flags, args); err != nil {
		return err
	}

	// checked here and not only when processing, so a wrong method fails before the user marks the areas
	method, _ := cmd.Flags().GetString("method")
	if strategies := image.NewInpaintProcessor().GetAllStrategiesNames(); method != "" && !slices.Contains(strategies, method) {
		return fmt.Errorf("inpainting method %q not available, use one of %v", method, strategies)
	}

	rects, _ := cmd.Flags().GetStringArray("rect")
	if _, err := ipkg.ParseRects(rects...); err != nil {
		return err
	}
	if len(rects) == 0 && ((len(args) > 0 && args[0] == "-") || imageio.IsStdoutOutput(flags, args)) {
		return fmt.Errorf("inpaint uses the terminal for the mouse and the image, so it can't read from stdin or write to stdout with -, use --rect for that")
	}

	return nil
}

func inpaintCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return image.NewInpaintProcessor().GetAllStrategiesNames(), cobra.ShellCompDirectiveNoFileComp
}

func init() {
	rootCmd.AddCommand(BuildInpaintCmd())
}
