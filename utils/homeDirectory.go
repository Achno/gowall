package utils

import (
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
		return "", err
	}

	return dirPath, err
}
