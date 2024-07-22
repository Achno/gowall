/*
Copyright © 2024 Achnologia <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/Achno/gowall/internal/image"
	"github.com/spf13/cobra"
)

var theme string

var convertCmd = &cobra.Command{
	Use:   "convert [image path]",
	Short: "convert an img's color shceme",
	Long: `convert an img's color shceme`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {

		image.ProcessImg(args[0],theme)

	},
}

func init() {
	rootCmd.AddCommand(convertCmd)
	convertCmd.Flags().StringVarP(&theme,"theme","t","catpuccin","Usage : --theme [ThemeName-Lowercase]")

	// Here you will define your flags and configuration settings.

}
