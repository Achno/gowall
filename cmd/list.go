/*
Copyright © 2024 Achnologia <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"github.com/Achno/gowall/internal/image"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "Lists available themes",
	Long:  `List all available themes. This includes the predefined and custom user-provided themes in ~/.config/gowall/config.yml`,
	Run: func(cmd *cobra.Command, args []string) {
		allThemes := image.ListThemes()
		if len(allThemes) == 0 {
			fmt.Println("No themes available.")
			return
		}
		for _, theme := range allThemes {
			fmt.Println(theme)
		}
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
