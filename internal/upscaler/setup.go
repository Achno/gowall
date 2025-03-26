package upscaler

import (
	"archive/zip"
	"fmt"
	"io"
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
	err = extractZip(zipPath, destFolder)
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

// extractZip extracts the zip files containing the model to a specified destination and gives it permissions.
func extractZip(src, dest string) error {
	reader, err := zip.OpenReader(src)
	if err != nil {
		return fmt.Errorf("failed to open zip file: %v", err)
	}
	defer reader.Close()

	// Create the structure. Create all the directories,subdirectories and blank files.
	// then just io.Copy() all the content from the zip to the structure and lastly,
	// give chmod permissions to the binary for the upscaler.

	for _, file := range reader.File {

		filePath := filepath.Join(dest, file.Name)

		// for every dir in the zip create its directory
		if file.FileInfo().IsDir() {

			err := os.MkdirAll(filePath, 0755)
			if err != nil {
				return fmt.Errorf("failed to create directory: %v", err)
			}

		} else {

			// create the nested directories
			err := os.MkdirAll(filepath.Dir(filePath), 0755)
			if err != nil {
				return fmt.Errorf("failed to create directory for file: %v", err)
			}

			// open the file for writing only,create it if it doesn't exist,truncate length to 0 if exists
			outFile, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
			if err != nil {
				return fmt.Errorf("failed to create file: %v", err)
			}

			rc, err := file.Open()
			if err != nil {
				return fmt.Errorf("failed to open zip file entry: %v", err)
			}

			_, err = io.Copy(outFile, rc)
			outFile.Close()
			rc.Close()
			if err != nil {
				return fmt.Errorf("failed to extract file: %v", err)
			}

			// give the binary execute permissions
			targetFileName := config.UpscalerBinaryName
			if filepath.Base(file.Name) == targetFileName {
				err := os.Chmod(filePath, file.Mode()|0755)
				if err != nil {
					return fmt.Errorf("failed to chmod file: %v", err)
				}
			}
		}
	}

	return nil
}
