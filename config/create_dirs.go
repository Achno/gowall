package config

import (
	"fmt"
	"os"
	"path/filepath"
)

func CreateDirectory() (dirPath string, err error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	folderName := GowallConfig.OutputFolder
	if folderName == "" {
		folderName = OutputFolder
	}
	dirPath = filepath.Join(homeDir, folderName)

	// Handle XDG_PICTURES_DIR
	env := os.Getenv("XDG_PICTURES_DIR")
	if env != "" && GowallConfig.OutputFolder == "" {
		dirPath = filepath.Join(env, "gowall")
	}

	// Ensure all required directories exist
	subDirs := []string{"cluts", "gifs", "ocr"}
	for _, sub := range subDirs {
		subDir := filepath.Join(dirPath, sub)
		err = os.MkdirAll(subDir, 0755)
		if err != nil {
			return "", fmt.Errorf("while creating %s: %w", subDir, err)
		}
	}

	return dirPath, nil
}
