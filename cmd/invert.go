/*
Copyright © 2024 Achnologia <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// invertCmd represents the invert command
var invertCmd = &cobra.Command{
	Use:   "invert [image path]",
	Short: "Inverts the color's of an image",
	Long: `Inverts the color's of an image , then you can convert the inverted image to your favourite color scheme`,
	Run: func(cmd *cobra.Command, args []string) {

		fmt.Println("invert called")
	},
}

func init() {
	rootCmd.AddCommand(invertCmd)


}
