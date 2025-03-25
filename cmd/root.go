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
	imageio "github.com/Achno/gowall/internal/image_io"
	"github.com/Achno/gowall/internal/logger"
	"github.com/Achno/gowall/utils"
	"github.com/spf13/cobra"
)

var (
	shared           config.GlobalSubCommandFlags
	versionFlag      bool
	wallOfTheDayFlag bool
)

func isInputBatch(flags config.GlobalSubCommandFlags) bool {
	return len(flags.InputFiles) > 0 || len(flags.InputDir) > 0
}

func openImageInViewer(flags config.GlobalSubCommandFlags, args []string, path string) {
	if isInputBatch(shared) || imageio.IsStdoutOutput(flags, args) {
		return
	}
	err := image.OpenImageInViewer(path)
	if err != nil {
		logger.Error(err, "Error opening image")
	}
}

// Exit cli early if conflicting flags are present
func validateFlagsCompatibility(cmd *cobra.Command, args []string) error {
	if len(shared.InputFiles) > 0 && len(shared.InputDir) > 0 {
		return fmt.Errorf("cannot use --batch and --dir flags together, use one or the other")
	}
	if isInputBatch(shared) && len(args) > 0 {
		return fmt.Errorf("you cant use --batch and normal input or stdout")
	}
	if (len(args) == 2 && args[1] == "-") && shared.OutputDestination != "" {
		return fmt.Errorf("cannot use - pseudofile for stdout and --output flag at the same time")
	}
	return nil
}

func validateInput(flags config.GlobalSubCommandFlags, args []string) error {
	if len(args) > 0 || len(flags.InputDir) > 0 || len(flags.InputFiles) > 0 {
		return nil
	}
	return fmt.Errorf("no input was given")
}

// Add common global flags to command
func addGlobalFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().StringSliceVar(&shared.InputFiles, "batch", nil, "Usage: --batch file1.png,file2.png... Batch process individual files")
	cmd.PersistentFlags().StringVar(&shared.InputDir, "dir", "", "Usage --dir [/path/to/dir] Batch process an entire directory")
	cmd.PersistentFlags().StringVar(&shared.OutputDestination, "output", "", "Usage: --output ~/Folder (works on --dir and --batch also) or --output ~/NewDir/img.png")
}

// Configure logger and validates flags
func initCli(cmd *cobra.Command, args []string) error {
	logger.SetQuiet(imageio.IsStdoutOutput(shared, args))
	return validateFlagsCompatibility(cmd, args)
}

// Initialize default configuration and creates default directories
func initConfig() {
	config.LoadConfig()
	image.LoadCustomThemes()
}

var rootCmd = &cobra.Command{
	Use:               "gowall",
	Short:             "A tool to convert an img's color shceme ",
	Long:              `Convert an Image's (ex. Wallpaper) color scheme to another ( ex. Catppuccin ) `,
	PersistentPreRunE: initCli,
	Run: func(cmd *cobra.Command, args []string) {
		switch {

		case versionFlag:
			logger.Printf("gowall version: %s\n", config.Version)

		case wallOfTheDayFlag:
			logger.Print("Fetching wallpaper of the day...")
			url, err := api.GetWallpaperOfTheDay()
			utils.HandleError(err, "Could not fetch wallpaper of the day")

			path, err := image.SaveUrlAsImg(url)
			utils.HandleError(err)

			err = image.OpenImageInViewer(path)
			utils.HandleError(err)

			ok := utils.Confirm("Do you want to download this image?")

			if !ok {
				err = os.Remove(path)
				utils.HandleError(err)
				logger.Print("::Image discarded::")
				return
			}

			logger.Printf("Image saved as %s\n", path)
			return

		default:
			_ = cmd.Usage()
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {

	// Prevents Cobra from printing errors so we can wrap them into our logs
	rootCmd.SilenceErrors = true

	err := rootCmd.Execute()
	if err != nil {
		// os.Exit(1)
		utils.HandleError(err, "Error")
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	rootCmd.Flags().BoolVarP(&versionFlag, "version", "v", false, "show gowall version")
	rootCmd.Flags().BoolVarP(&wallOfTheDayFlag, "wall", "w", false, "fetches the wallpaper of the day!")
}
