package utils

import (
	"os"
	"path/filepath"
	"strings"
)

// Function to expand the tilde (~) to the full home directory path
// @Example ~/Pictures/flowers.png --> /home/username/Pictures/flowers.png
func ExpandHomeDirectory(paths []string) []string {
	var expandedPaths []string
	homeDir, _ := os.UserHomeDir()

	for _, path := range paths {
		if strings.HasPrefix(path, "~") {
			path = filepath.Join(homeDir, path[1:])
		}
		expandedPaths = append(expandedPaths, path)
	}
	return expandedPaths
}