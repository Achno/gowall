package upscaler

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/Achno/gowall/config"
	"github.com/Achno/gowall/internal/logger"
	"github.com/Achno/gowall/utils"
)

func SetupUpscaler() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("while getting home directory : %w", err)
	}
	zipPath := filepath.Join(homeDir, "tmp.zip")
	destFolder := filepath.Join(config.GowallConfig.OutputFolder, "upscaler")

	// urls of the ESRGAN portable model depending on the operating system
	urls := map[string]string{
		"linux":   "https://github.com/xinntao/Real-ESRGAN/releases/download/v0.2.5.0/realesrgan-ncnn-vulkan-20220424-ubuntu.zip",
		"windows": "https://github.com/xinntao/Real-ESRGAN/releases/download/v0.2.5.0/realesrgan-ncnn-vulkan-20220424-windows.zip",
		"darwin":  "https://github.com/xinntao/Real-ESRGAN/releases/download/v0.2.5.0/realesrgan-ncnn-vulkan-20220424-macos.zip",
	}

	url, exists := urls[runtime.GOOS]

	if !exists {
		return fmt.Errorf("unsupported OS: %s\n Only available for linux,mac,windows", runtime.GOOS)
	}

	logger.Print(utils.BlueColor + " ➜ Downloading models sit back and relax,might take a bit" + utils.ResetColor)
	// download model
	err = utils.DownloadUrl(url, zipPath)
	if err != nil {
		return fmt.Errorf("while downloading model : %v", err)
	}

	// create ~/Pictures/gowall/upscaler
	err = os.MkdirAll(destFolder, 0755)
	if err != nil {
		return fmt.Errorf("failed to create folder: %v", err)
	}
	logger.Print(utils.BlueColor + " ➜ Folder created" + utils.ResetColor)

	// Extract  zip
	err = utils.ExtractZipBinary(zipPath, destFolder, config.UpscalerBinaryName)
	if err != nil {
		return fmt.Errorf("while extracting zip : %v", err)
	}

	// Cleanup
	err = os.Remove(zipPath)
	if err != nil {
		return fmt.Errorf("while cleaning up : %v", err)
	}
	logger.Print(utils.BlueColor + " ➜ Cleaning up" + utils.ResetColor)

	logger.Print(utils.BlueColor + " ➜ Process complete. Upscaler setup" + utils.ResetColor)
	return nil
}
