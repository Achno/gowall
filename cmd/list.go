/*
Copyright © 2024 Achnologia <EMAIL ADDRESS>
*/
package cmd

import (
	"sort"

	"github.com/Achno/gowall/config"
	"github.com/Achno/gowall/internal/image"
	"github.com/Achno/gowall/internal/logger"
	"github.com/Achno/gowall/utils"
	"github.com/spf13/cobra"
)

func BuildListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "Lists available themes",
		Long:  `List all available themes. This includes the predefined and custom user provided themes in ~/.config/gowall/config.yml`,
		Run:   RunListCmd,
	}

	flags := cmd.Flags()
	var (
		theme       string
		previewFlag bool
	)

	flags.StringVarP(&theme, "theme", "t", "", "Usage : --theme <theme_name>")
	flags.BoolVarP(&previewFlag, "preview", "p", false, "gowall extract -p (opens hex code preview site)")

	return cmd
}

func RunListCmd(cmd *cobra.Command, args []string) {
	theme, _ := cmd.Flags().GetString("theme")
	previewFlag, _ := cmd.Flags().GetBool("preview")

	switch {
	case theme != "":
		colors, err := image.GetThemeColors(theme)
		utils.HandleError(err)

		for _, color := range colors {
			logger.Print(color)
		}

		if previewFlag {
			utils.OpenURL(config.HexCodeVisualUrl)
		}

	default:
		allThemes := image.ListThemes()
		sort.Strings(allThemes)
		for _, theme := range allThemes {
			logger.Print(theme)
		}
	}
}

func init() {
	rootCmd.AddCommand(BuildListCmd())
}
