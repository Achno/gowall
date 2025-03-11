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

func isInputBatch() bool {
	return len(shared.InputFiles) > 0 || len(shared.InputDir) > 0
}

// Exit cli early if conflicting flags are present
func validateFlagsCompatibility(_ *cobra.Command, args []string) {
	if len(shared.InputFiles) > 0 && len(shared.InputDir) > 0 {
		utils.HandleError(fmt.Errorf("cannot use --batch and --dir flags together, use one or the other"))
	}
	if isInputBatch() && len(shared.OutputDestination) > 0 {
		utils.HandleError(fmt.Errorf("cannot use --output flag with --batch or --dir flags"))
	}
	if isInputBatch() && len(args) > 0 {
		utils.HandleError(fmt.Errorf("cannot use positional args for input and batch file flags at the same time ie: --dir or --batch"))
	}
	// We could just ignore more args instead of erroring
	if len(args) > 2 {
		utils.HandleError(fmt.Errorf("more than two io args provided, only 0, 1 or 2 args are valid"))
	}
	if (len(args) == 2 && args[1] == "-") && shared.OutputDestination != "" {
		utils.HandleError(fmt.Errorf("cannot use - pseudofile for stdout and --output flag at the same time"))
	}
}

// Add common global flags to command
func addGlobalFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().StringSliceVarP(&shared.InputFiles, "batch", "b", nil, "Usage: --batch file1.png,file2.png... Batch proccess individual files")
	cmd.PersistentFlags().StringVarP(&shared.InputDir, "dir", "d", "", "Usage --dir [/path/to/dir] Batch proccess entire directory")
	cmd.PersistentFlags().StringVarP(&shared.OutputDestination, "output", "o", "", "Usage: --output imageName (no extension) Not available in batch proccesing")
}

// Configure logger and validates flags
func initCli(cmd *cobra.Command, args []string) {
	logger.SetQuiet(imageio.IsStdoutOutput(shared, args))
	validateFlagsCompatibility(cmd, args)
}

// Initialize default configuration and creates default directories
func initConfig() {
	config.LoadConfig()
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:              "gowall",
	Short:            "A tool to convert an img's color shceme ",
	Long:             `Convert an Image's (ex. Wallpaper) color scheme to another ( ex. Catppuccin ) `,
	PersistentPreRun: initCli,
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

			err = image.OpenImageInViewer(path)
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
	cobra.OnInitialize(initConfig)
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	rootCmd.Flags().BoolVarP(&versionFlag, "version", "v", false, "show gowall version")
	rootCmd.Flags().BoolVarP(&wallOfTheDayFlag, "wall", "w", false, "fetches the wallpaper of the day!")
}
