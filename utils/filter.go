package utils

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"
	"unicode/utf8"
)

// Filters out all files other than .png, .jpeg, .jpg, .webp in a directory
func filterImages(entries []fs.DirEntry) ([]string, error) {
	if len(entries) == 0 {
		return nil, fmt.Errorf("directory is empty")
	}

	var imageFiles []string

	supportedExtensions := map[string]bool{
		".png":  true,
		".jpeg": true,
		".jpg":  true,
		".webp": true,
	}

	for _, entry := range entries {
		if !entry.IsDir() && supportedExtensions[strings.ToLower(filepath.Ext(entry.Name()))] {
			imageFiles = append(imageFiles, entry.Name())
		}
	}

	return imageFiles, nil
}

// Discards the last character of a string
func DiscardLastCharacter(s string) string {
	if len(s) == 0 {
		return s
	}

	// Decode the last rune
	_, size := utf8.DecodeLastRuneInString(s)

	// Exclude the last character
	return s[:len(s)-size]
}
