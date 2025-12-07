/*
Copyright © 2024 Achno <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"slices"
	"strings"

	"github.com/Achno/gowall/config"
	"github.com/Achno/gowall/internal/image"
	imageio "github.com/Achno/gowall/internal/image_io"
	"github.com/Achno/gowall/internal/logger"
	"github.com/Achno/gowall/utils"
	"github.com/spf13/cobra"
)

func BuildUpscaleCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "upscale [INPUT] [OPTIONAL OUTPUT]",
		Short: "Upscale (or Deblur) images using an Enhanced Super-Resolution Generative Adversarial Network, your GPU must support Vulkan",
		Long:  `Upscale images using an Enhanced Super-Resolution Generative Adversarial Network, your GPU must support Vulkan,if you sea black image after a lot of time then that means that you GPU does not support Vulkan. You can give options that specify thescale and Modelname`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return ValidateParseUpscaleCmd(cmd, shared, args)
		},
		Run: RunUpscaleCmd,
	}

	flags := cmd.Flags()
	var (
		scale     int
		modelName string
	)

	flags.IntVarP(&scale, "scale", "s", 2, "Scale factor for upscaling (2, 3, or 4)")
	flags.StringVarP(&modelName, "model", "m", "realesr-animevideov3",
		`Model to use for upscaling. Available models:
        realesrgan-x4plus (Slower,Better quality,forces -s 4),
        realesrgan-x4plus-anime (optimized for anime small),
		realesr-animevideov3 (Fast model ,animation video (default))`)

	cmd.RegisterFlagCompletionFunc("model", upscaleCompletion)

	addGlobalFlags(cmd)

	return cmd
}

func RunUpscaleCmd(cmd *cobra.Command, args []string) {
	imageOps, err := imageio.DetermineImageOperations(shared, args, cmd)
	utils.HandleError(err, "Error")

	scale, err := cmd.Flags().GetInt("scale")
	utils.HandleError(err, "Error")
	modelName, err := cmd.Flags().GetString("model")
	utils.HandleError(err, "Error")

	logger.Print("Upscaling images...")
	processor := &image.UpscaleProcessor{
		Scale:     scale,
		ModelName: modelName,
	}

	processedImages, err := image.ProcessImgs(processor, imageOps, "")
	utils.HandleError(err, "Error")

	if err != nil {
		logger.Error(err, "The following images had errors while processing")
	}
	openImageInViewer(shared, args, processedImages[0])
}

func ValidateParseUpscaleCmd(cmd *cobra.Command, flags config.GlobalSubCommandFlags, args []string) error {
	if err := validateInput(flags, args); err != nil {
		return err
	}

	if imageio.IsStdoutOutput(flags, args) {
		return fmt.Errorf("the upscale is not compatible with stdout output")
	}

	scale, _ := cmd.Flags().GetInt("scale")
	if scale < 2 || scale > 4 {
		return fmt.Errorf("scale must be 2, 3, or 4, got: %d", scale)
	}

	modelName, _ := cmd.Flags().GetString("model")
	availableModels := image.GetAvailableUpscaleModels()

	validModel := slices.Contains(availableModels, modelName)

	if !validModel {
		return fmt.Errorf("invalid model name: %s, available models: %s", modelName, strings.Join(availableModels, ", "))
	}

	return nil
}

func upscaleCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return image.GetAvailableUpscaleModels(), cobra.ShellCompDirectiveNoFileComp
}

func init() {
	rootCmd.AddCommand(BuildUpscaleCmd())
}
