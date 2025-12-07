/*
Copyright © 2024 Achnologia <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/Achno/gowall/config"
	"github.com/Achno/gowall/internal/image"
	imageio "github.com/Achno/gowall/internal/image_io"
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
	)

	flags.IntVarP(&colorsNum, "colors", "c", 6, "-c <number of colors to return>")
	flags.BoolVarP(&previewFlag, "preview", "p", false, "gowall extract -p (opens hex code preview site)")

	addGlobalFlags(cmd)

	return cmd
}

func RunExtractCmd(cmd *cobra.Command, args []string) {
	imageOps, err := imageio.DetermineImageOperations(shared, args, cmd)
	utils.HandleError(err, "Error")

	numOfColors, err := cmd.Flags().GetInt("colors")
	utils.HandleError(err, "Error")
	previewFlag, err := cmd.Flags().GetBool("preview")
	utils.HandleError(err, "Error")

	processor := &image.ExtractProcessor{
		NumOfColors: numOfColors,
	}

	_, err = image.ProcessImgs(processor, imageOps, theme)
	utils.HandleError(err, "Error")

	if previewFlag {
		utils.OpenURL(config.HexCodeVisualUrl)
	}
}

func ValidateParseExtractCmd(cmd *cobra.Command, flags config.GlobalSubCommandFlags, args []string) error {
	if err := validateInput(flags, args); err != nil {
		return err
	}
	return nil
}

func init() {
	rootCmd.AddCommand(BuildExtractCmd())
}
