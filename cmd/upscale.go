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

// TODO add do you want to setup the upscaler, say downloading models....
func UpscaleCmd() *cobra.Command {

	var (
		scale     int
		modelName string
	)

	upscaleCmd := &cobra.Command{
		Use:   "upscale [PATH]",
		Short: "Upscale (or Deblur) images using an Enhanced Super-Resolution Generative Adversarial Network, your GPU must support Vulkan",
		Long:  `Upscale images using an Enhanced Super-Resolution Generative Adversarial Network, your GPU must support Vulkan,if you sea black image after a lot of time then that means that you GPU does not support Vulkan. You can give options that specify thescale and Modelname`,
		RunE: func(cmd *cobra.Command, args []string) error {

			switch {
			case len(args) > 0:
				fmt.Println("Upscaling image...")

				processor := &image.UpscaleProcessor{
					Scale:     scale,
					ModelName: modelName,
				}

				expandFile := utils.ExpandHomeDirectory(args)
				processor.InputFile = expandFile[0]

				opts := image.ProcessOptions{
					SaveToFile: false,
				}

				_, _, err := image.ProcessImg(expandFile[0], processor, shared.Theme, opts)
				utils.HandleError(err, "Error Processing Image")

				err = image.OpenImage(processor.OutputFile)
				utils.HandleError(err, "Error opening image")

			default:
				fmt.Println("Error: requires at least 1 arg(s), only received 0")
				_ = cmd.Usage()
			}
			return nil
		},
	}

	upscaleCmd.Flags().IntVarP(&scale, "scale", "s", 2, "Scale factor for upscaling (2, 3, or 4)")
	upscaleCmd.Flags().StringVarP(&modelName, "model", "m", "realesr-animevideov3",
		`Model to use for upscaling. Available models:
        realesrgan-x4plus (Slower,Better quality,forces -s 4),
        realesrgan-x4plus-anime (optimized for anime small),
		realesr-animevideov3 (Fast model ,animation video (default))`)

	return upscaleCmd
}

func init() {
	rootCmd.AddCommand(UpscaleCmd())
}
