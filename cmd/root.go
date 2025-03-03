/*
Copyright © 2024 Achnologia <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"

	"github.com/Achno/gowall/config"
	"github.com/Achno/gowall/internal/api"
	"github.com/Achno/gowall/internal/image"
	"github.com/Achno/gowall/utils"
	"github.com/spf13/cobra"
)

var (
	shared           config.Shared
	versionFlag      bool
	wallOfTheDayFlag bool
	formatFlag       string
	outputName       string
	dirInput         string
)

// Exits cli early if conflicting flags are present
func validateFlagsCompatibility(_ *cobra.Command, _ []string) {
	if len(shared.BatchFiles) > 0 && len(dirInput) > 0 {
		utils.HandleError(fmt.Errorf("cannot use --batch and --dir flags together, use one or the other"))
	}
	if (len(shared.BatchFiles) > 0 || len(dirInput) > 0) && len(outputName) > 0 {
		utils.HandleError(fmt.Errorf("cannot use --output flag with --batch or --dir flags"))
	}
}

// Checks whether we should output to stdout
func setOutputSource(args []string) {
	shared.UseSTDOUT = false

	// If there's batch processing do not use stdout
	if len(shared.BatchFiles) > 0 || len(dirInput) > 0 {
		return
	}

	// --output has the highest priority
	if len(outputName) > 0 {
		shared.UseSTDOUT = (outputName == "-")
		return
	}

	// Second argument has next priority
	if len(args) > 1 && args[1] == "-" {
		shared.UseSTDOUT = true
	}
}

// Add batch proccessing flags to command
func addBatchProccesingFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().StringSliceVarP(&shared.BatchFiles, "batch", "b", nil, "Usage: --batch file1.png,file2.png... Batch proccess individual files")
	cmd.PersistentFlags().StringVarP(&dirInput, "dir", "d", "", "Usage --dir [/path/to/dir] Batch proccess entire directory")
}

// Add common global flags to command
func addGlobalFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().StringVarP(&formatFlag, "format", "f", "", "Usage: --format [Extension]")
	cmd.PersistentFlags().StringVarP(&outputName, "output", "o", "", "Usage: --output imageName (no extension) Not available in batch proccesing")
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "gowall",
	Short: "A tool to convert an img's color shceme ",
	Long:  `Convert an Image's (ex. Wallpaper) color scheme to another ( ex. Catppuccin ) `,
	Run: func(cmd *cobra.Command, args []string) {
		switch {

		case versionFlag:
			fmt.Printf("gowall version: %s\n", config.Version)

		case wallOfTheDayFlag:
			fmt.Println("Fetching wallpaper of the day...")
			url, err := api.GetWallpaperOfTheDay()
			utils.HandleError(err, "Could not fetch wallpaper of the day")

			path, err := image.SaveUrlAsImg(url)
			utils.HandleError(err)

			err = image.OpenImage(path)
			utils.HandleError(err)

			ok := utils.Confirm("Do you want to download this image?")

			if !ok {
				err = os.Remove(path)
				utils.HandleError(err)
				fmt.Println("::Image discarded::")
				return
			}

			fmt.Printf("Image saved as %s\n", path)
			return

		default:
			cmd.Help()

		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	rootCmd.Flags().BoolVarP(&versionFlag, "version", "v", false, "show gowall version")
	rootCmd.Flags().BoolVarP(&wallOfTheDayFlag, "wall", "w", false, "fetches the wallpaper of the day!")
}
