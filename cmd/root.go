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

var shared config.Shared
var versionFlag bool

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "gowall",
	Short: "A tool to convert an img's color shceme ",
	Long:  `Convert an Image's (ex. Wallpaper) color scheme to another ( ex. Catppuccin ) `,
	Run: func(cmd *cobra.Command, args []string) {

		switch {

		case versionFlag:
			fmt.Printf("gowall version: %s\n", config.Version)

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
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.gowall.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	rootCmd.Flags().BoolVarP(&versionFlag, "version", "v", false, "show gowall version")
}
