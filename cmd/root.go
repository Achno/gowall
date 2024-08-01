/*
Copyright © 2024 Achnologia <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"

	"github.com/Achno/gowall/config"
	"github.com/spf13/cobra"
)

var versionFlag bool

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "gowall",
	Short: "A tool to convert an image's color scheme",
	Long:  `Convert an image's (e.g., wallpaper) color scheme to another (e.g., Catppuccin)`,
	Run: func(cmd *cobra.Command, args []string) {
		if versionFlag {
			fmt.Printf("gowall version: %s\n", config.Version)
		} else {
			cmd.Help()
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err) // Print the error to stderr
		os.Exit(1)
	}
}

func init() {
	// Define your flags and configuration settings here.

	// Persistent flags are global for your application.
	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.gowall.yaml)")

	// Local flags only apply to this command.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	rootCmd.Flags().BoolVarP(&versionFlag, "version", "v", false, "Show gowall version")
}
