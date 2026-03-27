package png

import (
	"fmt"
	"path/filepath"
	"runtime"

	"github.com/Achno/gowall/config"
	"github.com/Achno/gowall/internal/logger"
	"github.com/Achno/gowall/utils"
)

func SetupPngquant() error {
	destFolder := filepath.Join(config.GowallConfig.OutputFolder, "compression", "pngquant")

	urls := map[string]string{
		"linux":   "https://pngquant.org/pngquant-linux.tar.bz2",
		"windows": "https://pngquant.org/pngquant-windows.zip",
		"darwin":  "https://pngquant.org/pngquant.tar.bz2",
	}

	url, exists := urls[runtime.GOOS]
	if !exists {
		return fmt.Errorf("unsupported OS: %s\n Only available for linux,mac,windows", runtime.GOOS)
	}

	logger.Print(utils.BlueColor + " ➜ Downloading pngquant binary, sit back and relax" + utils.ResetColor)

	if err := utils.DownloadAndExtract(url, destFolder, config.PngquantBinaryName, true); err != nil {
		return fmt.Errorf("setup pngquant: %w", err)
	}

	logger.Print(utils.BlueColor + " ➜ Process complete. Pngquant setup" + utils.ResetColor)
	return nil
}

// CheckPngquantInstalled checks if pngquant is available (either installed or downloaded)
func CheckPngquantInstalled() (string, error) {
	binaryNames := map[string]string{
		"linux":   config.PngquantBinaryName,
		"windows": config.PngquantBinaryName + ".exe",
		"darwin":  config.PngquantBinaryName,
	}

	destFolder := filepath.Join(config.GowallConfig.OutputFolder, "compression", "pngquant")

	return utils.FindBinary(binaryNames, destFolder)
}
