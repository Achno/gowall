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

var (
	colorsNum   int
	previewFlag bool
)

// extractCmd represents the extract command
var extractCmd = &cobra.Command{
	Use:   "extract [INPUT]",
	Short: "Prints the color pallete of the image you specificed (like pywal)",
	Long:  `Using the colorthief backend ( like pywal ) it prints the color pallete of the image (path) you specified`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		err := validateInput(shared, args)
		if err != nil {
			return err
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		imageOps := imageio.DetermineImageOperations(shared, args)

		NumOfColors, err := cmd.Flags().GetInt("colors")
		utils.HandleError(err, "Error")

		processor := &image.ExtractProcessor{
			NumOfColors: NumOfColors,
		}

		_, err = image.ProcessImgs(processor, imageOps, theme)
		utils.HandleError(err, "Error")

		//TODO TEST THIS WITH STDOUT AND MAKE THE IOWRITER for other formats, check for ways to remove the ::: img completed ::: msg
		// open up hex code preview site
		if previewFlag {
			utils.OpenURL(config.HexCodeVisualUrl)
		}
	},
}

func init() {
	rootCmd.AddCommand(extractCmd)
	extractCmd.Flags().IntVarP(&colorsNum, "colors", "c", 6, "-c <number of colors to return>")
	extractCmd.Flags().BoolVarP(&previewFlag, "preview", "p", false, "gowall extract -p (opens hex code preview site)")
	// extractCmd.PersistentFlags().StringVarP(&shared.OutputDestination, "output", "o", "", "Usage: --output imageName (no extension) Not available in batch proccesing")
	addGlobalFlags(extractCmd)
}
