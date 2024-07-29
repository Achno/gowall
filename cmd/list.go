/*
Copyright © 2024 Achnologia <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"github.com/Achno/gowall/internal/image"
	"github.com/spf13/cobra"
)

// invertCmd represents the invert command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "Lists available themes",
	Long:  `List all available themes. This includes the predefined and custom user provided themes`,
	Run: func(cmd *cobra.Command, args []string) {
		allThemes := image.ListThemes()
		for _, theme := range allThemes {
			fmt.Println(theme)
		}
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
