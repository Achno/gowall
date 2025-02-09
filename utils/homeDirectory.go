package utils

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/Achno/gowall/config"
)

func CreateDirectory() (dirPath string, err error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	folderName := config.OutputFolder
	if config.GowallConfig.OutputFolder != "" {
		folderName = config.GowallConfig.OutputFolder
	}
	dirPath = filepath.Join(homeDir, folderName)

	err = os.MkdirAll(dirPath, 0777)
	if err != nil {
		return "", fmt.Errorf("while creating ~/OutputFolder: %w", err)
	}

	err = os.MkdirAll(filepath.Join(dirPath, "cluts"), 0755)
	if err != nil {
		return "", fmt.Errorf("while creating ~/OutputFolder/cluts: %w", err)
	}

	err = os.MkdirAll(filepath.Join(dirPath, "gifs"), 0755)
	if err != nil {
		return "", fmt.Errorf("while creating ~/OutputFolder/gifs: %w", err)
	}

	return dirPath, err
}
