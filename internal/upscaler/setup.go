package upscaler

import (
	"fmt"
	"path/filepath"
	"runtime"

	"github.com/Achno/gowall/config"
	"github.com/Achno/gowall/internal/logger"
	"github.com/Achno/gowall/utils"
)

func SetupUpscaler() error {
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

	logger.Print(utils.BlueColor + " ➜ Downloading models sit back and relax, might take a bit" + utils.ResetColor)

	if err := utils.DownloadAndExtract(url, destFolder, config.UpscalerBinaryName, true); err != nil {
		return fmt.Errorf("setup upscaler: %w", err)
	}

	logger.Print(utils.BlueColor + " ➜ Process complete. Upscaler setup" + utils.ResetColor)
	return nil
}
