package utils

import (
	"os"
	"path/filepath"
)

func CreateDirectory() (dirPath string, err error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	folderName := "Pictures/gowall"
	dirPath = filepath.Join(homeDir, folderName)

	err = os.MkdirAll(dirPath, 0755)
	if err != nil {
		return "", err
	}

	return dirPath, nil
}
