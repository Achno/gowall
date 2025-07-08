package image

import (
	"fmt"
	"image"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/Achno/gowall/config"
	imageio "github.com/Achno/gowall/internal/image_io"
	"github.com/Achno/gowall/internal/upscaler"
	"github.com/Achno/gowall/utils"
)

type UpscaleProcessor struct {
	Scale     int
	ModelName string
}

func (p *UpscaleProcessor) Process(img image.Image, theme string) (image.Image, error) {
	destFolder := filepath.Join(config.GowallConfig.OutputFolder, "upscaler")
	// setup upscaler if it has not been already
	if _, err := os.Stat(destFolder); os.IsNotExist(err) {

		ok := utils.Confirm(utils.BlueColor + "◈ It seems that the upscaler is not setup yet, would you like for gowall to set it up" + utils.ResetColor)
		if !ok {
			return nil, fmt.Errorf("the upscaler has not been setup")
		}
		upscaler.SetupUpscaler()
	}

	binary, err := findRealESRGANBinary(destFolder)
	if err != nil {
		return nil, fmt.Errorf("while finding upscaler binary : %w", err)
	}
	// Create temporary files for input and output
	tempDir, err := os.MkdirTemp("", "gowall-upscale-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	inputPath := filepath.Join(tempDir, "input.png")
	outputPath := filepath.Join(tempDir, "output.png")
	if err := imageio.SaveImage(img, imageio.FileWriter{Path: inputPath}, "png"); err != nil {
		return nil, fmt.Errorf("failed to save temp input image: %w", err)
	}
	// Validate params
	if err := p.validateParams(inputPath); err != nil {
		return nil, fmt.Errorf("while validating parameters: %w", err)
	}
	cmd := exec.Command(binary, "-i", inputPath, "-o", outputPath, "-s", fmt.Sprintf("%d", p.Scale))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Run()
	if err != nil {
		exitError, ok := err.(*exec.ExitError)
		if ok && exitError.ExitCode() == 255 {
			return nil, nil
		}
		return nil, fmt.Errorf("command failed: %w", err)
	}
	imgUpscaled, err := imageio.LoadImage(imageio.FileReader{Path: outputPath})
	if err != nil {
		return nil, fmt.Errorf("could not open upscaled image after processing in %s", outputPath)
	}

	return imgUpscaled, nil
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

func findRealESRGANBinary(destFolder string) (string, error) {
	binaryNames := map[string]string{
		"windows": "realesrgan-ncnn-vulkan.exe",
		"darwin":  "realesrgan-ncnn-vulkan",
		"linux":   "realesrgan-ncnn-vulkan",
	}

	binaryName, ok := binaryNames[runtime.GOOS]
	if !ok {
		return "", fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}

	binaryPath := filepath.Join(destFolder, binaryName)

	// NixOS does not allow dynamically linked binaries,so just check $PATH for it instead.
	path, err := exec.LookPath(binaryName)
	if err != nil {
	}

	if path != "" {
		binaryPath = path
	}

	// Check if binary exists and is executable
	info, err := os.Stat(binaryPath)
	if err != nil {
		return "", fmt.Errorf("binary not found at %s: %w", binaryPath, err)
	}

	// check if the file is executable on unix
	if runtime.GOOS != "windows" {
		if info.Mode()&0111 == 0 {
			return "", fmt.Errorf("binary at %s is not executable", binaryPath)
		}
	}

	return binaryPath, nil
}
