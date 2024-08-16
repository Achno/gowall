/*
Copyright © 2024 Achnologia <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"github.com/Achno/gowall/config"
	"github.com/Achno/gowall/internal/image"
	"github.com/Achno/gowall/utils"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "Lists available themes",
	Long:  `List all available themes. This includes the predefined and custom user provided themes in ~/.config/gowall/config.yml`,
	Run: func(cmd *cobra.Command, args []string) {

		th, _ := cmd.Flags().GetString("theme")

		switch {
		case th != "":
			colors, err := image.GetThemeColors(th)
			utils.HandleError(err)

			for _, color := range colors {
				fmt.Println(color)
			}

			if previewFlag {
				utils.OpenURL(config.HexCodeVisualUrl)
			}

		default:
			allThemes := image.ListThemes()
			for _, theme := range allThemes {
				fmt.Println(theme)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
	listCmd.Flags().StringVarP(&shared.Theme, "theme", "t", "", "Usage : --theme <theme_name>")
	listCmd.Flags().BoolVarP(&previewFlag, "preview", "p", false, "gowall extract -p (opens hex code preview site)")
}
