/*
Copyright © 2024 Achnologia <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"path/filepath"
	"slices"
	"strings"

	"github.com/Achno/gowall/config"
	"github.com/Achno/gowall/internal/image"
	imageio "github.com/Achno/gowall/internal/image_io"
	"github.com/Achno/gowall/internal/logger"
	"github.com/Achno/gowall/utils"
	"github.com/spf13/cobra"
)

func BuildExtractCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "extract [INPUT]",
		Short: "Prints the color pallete of the image you specificed (like pywal)",
		Long:  `Using the colorthief backend ( like pywal ) it prints the color pallete of the image (path) you specified`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return ValidateParseExtractCmd(cmd, shared, args)
		},
		Run: RunExtractCmd,
	}

	flags := cmd.Flags()
	var (
		colorsNum   int
		previewFlag bool
		jsonFlag    bool
	)

	flags.IntVarP(&colorsNum, "colors", "c", 6, "-c <number of colors to return>")
	flags.BoolVarP(&previewFlag, "preview", "p", false, "gowall extract -p (opens hex code preview site)")
	flags.BoolVar(&jsonFlag, "json", false, "Save the colors as a theme json named after the image (use --output - for stdout)")

	addGlobalFlags(cmd)

	return cmd
}

func RunExtractCmd(cmd *cobra.Command, args []string) {
	jsonFlag, err := cmd.Flags().GetBool("json")
	utils.HandleError(err, "Error")
	if jsonFlag {
		shared.Format = "json"
	}

	imageOps, err := imageio.DetermineImageOperations(shared, args, cmd)
	utils.HandleError(err, "Error")

	numOfColors, err := cmd.Flags().GetInt("colors")
	utils.HandleError(err, "Error")
	previewFlag, err := cmd.Flags().GetBool("preview")
	utils.HandleError(err, "Error")

	processor := &image.ExtractProcessor{
		NumOfColors: numOfColors,
	}

	// One image at a time, so each json theme is named after its own image
	for _, op := range imageOps {
		if jsonFlag {
			name := filepath.Base(op.ImageInput.String())
			processor.JsonName = strings.TrimSuffix(name, filepath.Ext(name))
		}
		_, err = image.ProcessImgs(processor, []imageio.ImageIO{op}, image.ProcessOptions{
			OnComplete: func(outputPath string, remaining int) {
				if jsonFlag {
					logger.Print("::: Colors saved in " + outputPath + " :::")
				}
			},
		})
		utils.HandleError(err, "Error")
	}

	if previewFlag {
		utils.OpenURL(config.HexCodeVisualUrl)
	}
}

func ValidateParseExtractCmd(cmd *cobra.Command, flags config.GlobalSubCommandFlags, args []string) error {
	if err := validateInput(flags, args); err != nil {
		return err
	}

	// With --dir/--batch a trailing "-" lands in args[0], so check for it too
	jsonFlag, _ := cmd.Flags().GetBool("json")
	isBatch := flags.InputDir != "" || len(flags.InputFiles) > 0
	toStdout := imageio.IsStdoutOutput(flags, args) || slices.Contains(args, "-")
	if jsonFlag && isBatch && toStdout {
		return fmt.Errorf("--json can't print several images to stdout, use --output <folder> with --dir/--batch")
	}

	return nil
}

func init() {
	rootCmd.AddCommand(BuildExtractCmd())
}
