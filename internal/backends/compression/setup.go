package compression

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/Achno/gowall/config"
	"github.com/Achno/gowall/internal/logger"
	"github.com/Achno/gowall/utils"
)

func SetupPngquant() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("while getting home directory : %w", err)
	}
	// Choose archive extension based on OS to match the actual download format
	archiveExt := ".zip"
	switch runtime.GOOS {
	case "linux", "darwin":
		archiveExt = ".tar.bz2"
	case "windows":
		archiveExt = ".zip"
	}

	archivePath := filepath.Join(homeDir, "pngquant"+archiveExt)
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

	err = utils.DownloadUrl(url, archivePath)
	if err != nil {
		return fmt.Errorf("while downloading pngquant : %v", err)
	}

	// create ~/Pictures/gowall/compression/pngquant
	err = os.MkdirAll(destFolder, 0755)
	if err != nil {
		return fmt.Errorf("failed to create folder: %v", err)
	}
	logger.Print(utils.BlueColor + " ➜ Folder created" + utils.ResetColor)

	// Extract based on archive format
	err = utils.ExtractArchiveBinary(archivePath, destFolder, config.PngquantBinaryName)
	if err != nil {
		return fmt.Errorf("while extracting archive : %v", err)
	}

	// Cleanup
	err = os.Remove(archivePath)
	if err != nil {
		return fmt.Errorf("while cleaning up : %v", err)
	}
	logger.Print(utils.BlueColor + " ➜ Cleaning up" + utils.ResetColor)

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
