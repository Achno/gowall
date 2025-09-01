package image

import (
	"fmt"
	"image"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/Achno/gowall/config"
	imageio "github.com/Achno/gowall/internal/image_io"
	types "github.com/Achno/gowall/internal/types"
	"github.com/Achno/gowall/internal/upscaler"
	"github.com/Achno/gowall/utils"
)

type UpscaleProcessor struct {
	Scale     int
	ModelName string
}

func (p *UpscaleProcessor) Process(img image.Image, theme string, format string) (image.Image, types.ImageMetadata, error) {
	destFolder := filepath.Join(config.GowallConfig.OutputFolder, "upscaler")
	// setup upscaler if it has not been already
	if _, err := os.Stat(destFolder); os.IsNotExist(err) {

		ok := utils.Confirm(utils.BlueColor + "◈ It seems that the upscaler is not setup yet, would you like for gowall to set it up" + utils.ResetColor)
		if !ok {
			return nil, types.ImageMetadata{}, fmt.Errorf("the upscaler has not been setup")
		}
		upscaler.SetupUpscaler()
	}

	binaryNames := map[string]string{
		"windows": "realesrgan-ncnn-vulkan.exe",
		"darwin":  "realesrgan-ncnn-vulkan",
		"linux":   "realesrgan-ncnn-vulkan",
	}

	binary, err := utils.FindBinary(binaryNames, destFolder)
	if err != nil {
		return nil, types.ImageMetadata{}, fmt.Errorf("while finding upscaler binary : %w", err)
	}
	// Create temporary files for input and output
	tempDir, err := os.MkdirTemp("", "gowall-upscale-*")
	if err != nil {
		return nil, types.ImageMetadata{}, fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	inputPath := filepath.Join(tempDir, "input.png")
	outputPath := filepath.Join(tempDir, "output.png")
	if err := imageio.SaveImage(img, imageio.FileWriter{Path: inputPath}, "png", types.ImageMetadata{}); err != nil {
		return nil, types.ImageMetadata{}, fmt.Errorf("failed to save temp input image: %w", err)
	}
	// Validate params
	if err := p.validateParams(inputPath); err != nil {
		return nil, types.ImageMetadata{}, fmt.Errorf("while validating parameters: %w", err)
	}
	cmd := exec.Command(binary, "-i", inputPath, "-o", outputPath, "-s", fmt.Sprintf("%d", p.Scale))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Run()
	if err != nil {
		exitError, ok := err.(*exec.ExitError)
		if ok && exitError.ExitCode() == 255 {
			return nil, types.ImageMetadata{}, nil
		}
		return nil, types.ImageMetadata{}, fmt.Errorf("command failed: %w", err)
	}
	imgUpscaled, err := imageio.LoadImage(imageio.FileReader{Path: outputPath})
	if err != nil {
		return nil, types.ImageMetadata{}, fmt.Errorf("could not open upscaled image after processing in %s", outputPath)
	}

	return imgUpscaled, types.ImageMetadata{}, nil
}

func (p *UpscaleProcessor) validateParams(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Errorf("this path does not exist")
	}

	if p.Scale < 2 || p.Scale > 4 {
		return fmt.Errorf("the upscale ratio is invalid")
	}

	modelNames := map[string]bool{
		"realesr-animevideov3":    true,
		"realesrgan-x4plus":       true,
		"realesrgan-x4plus-anime": true,
		"realesrnet-x4plus":       true,
	}

	_, exists := modelNames[p.ModelName]
	if !exists {
		return fmt.Errorf("invalid Model name")
	}

	return nil
}
