package image

import (
	"fmt"
	"image"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/Achno/gowall/internal/upscaler"
	"github.com/Achno/gowall/utils"
)

type UpscaleProcessor struct {
	InputFile string
	Scale     int
	ModelName string
}

func (p *UpscaleProcessor) Process(img image.Image, theme string) (image.Image, error) {

	// get upscaler directory
	dirFolder, err := utils.CreateDirectory()
	if err != nil {
		return nil, fmt.Errorf("while creating Directory or getting path : %w", err)
	}
	destFolder := filepath.Join(dirFolder, "upscaler")

	// setup upscaler if it has not been already
	if _, err := os.Stat(destFolder); os.IsNotExist(err) {
		upscaler.SetupUpscaler()
	}

	binary, err := findRealESRGANBinary(destFolder)
	if err != nil {
		return nil, fmt.Errorf("while finding upscaler binary : %w", err)
	}

	// validate params
	p.validateParams()

	// construct outputFile
	name := filepath.Base(p.InputFile)
	outputFile := filepath.Join(dirFolder, name)

	cmd := exec.Command(binary, "-i", p.InputFile, "-o", outputFile, "-s", fmt.Sprintf("%d", p.Scale))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Run()
	if err != nil {
		exitError, ok := err.(*exec.ExitError)
		if ok && exitError.ExitCode() == 255 {
			return nil, nil
		}
		return nil, fmt.Errorf("command failed: %w\n", err)
	}

	return nil, nil
}

func (p *UpscaleProcessor) validateParams() error {

	if _, err := os.Stat(p.InputFile); os.IsNotExist(err) {
		return fmt.Errorf("This path does not exist")
	}

	if p.Scale < 2 || p.Scale > 4 {
		return fmt.Errorf("The upscale ratio is invalid")
	}

	modelNames := map[string]bool{
		"realesr-animevideov3":    true,
		"realesrgan-x4plus":       true,
		"realesrgan-x4plus-anime": true,
		"realesrnet-x4plus":       true,
	}

	_, exists := modelNames[p.ModelName]
	if !exists {
		return fmt.Errorf("Invalid Model name")
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
