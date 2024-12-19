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
	dirPath = filepath.Join(homeDir, folderName)

	err = os.MkdirAll(dirPath, 0777)
	if err != nil {
		return "", fmt.Errorf("while creating ~/Pictures/gowall: %w", err)
	}

	err = os.MkdirAll(filepath.Join(dirPath, "cluts"), 0755)
	if err != nil {
		return "", fmt.Errorf("while creating ~/Pictures/gowal/cluts: %w", err)
	}

	return dirPath, err
}
