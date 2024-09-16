/*
Copyright © 2024 Achno <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"github.com/Achno/gowall/internal/image"
	"github.com/Achno/gowall/utils"
	"github.com/spf13/cobra"
)

var (
	colorB          string
	BorderThickness int
)

var drawCmd = &cobra.Command{
	Use:   "draw [PATH]",
	Short: "Draw a border with a color and thickness",
	Long: `The draw command allows you to apply various effects to an image.
Currently, it only supports drawing a border around the image. Future updates may include additional effects.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			fmt.Println("Error: Requires at least 1 argument (the image path).")
			_ = cmd.Usage()
			return
		}

		fmt.Println("Processing image...")

		hex, err := cmd.Flags().GetString("color")
		if err != nil {
			fmt.Println("Error retrieving color flag:", err)
			return
		}

		clr, err := image.HexToRGBA(hex)
		if err != nil {
			fmt.Println("Error converting hex to RGBA:", err)
			return
		}

		processor := &image.DrawProcessor{
			Color:           clr,
			BorderThickness: BorderThickness,
		}

		// Wrap args[0] in a slice to pass to ExpandHomeDirectory
		expandFile := utils.ExpandHomeDirectory([]string{args[0]})

		// Use the first element of expandFile (since it’s a []string)
		path, err := image.ProcessImg(expandFile[0], processor, shared.Theme)
		if err != nil {
			fmt.Println("Error processing image:", err)
			return
		}

		err = image.OpenImage(path)
		if err != nil {
			fmt.Println("Error opening image:", err)
			return
		}
	},
}

func init() {
	rootCmd.AddCommand(drawCmd)
	drawCmd.Flags().StringVarP(&colorB, "color", "c", "#5D3FD3", "Specify the border color in hex format (e.g., #5D3FD3).")
	drawCmd.Flags().IntVarP(&BorderThickness, "borderThickness", "b", 5, "Specify the thickness of the border.")
}
