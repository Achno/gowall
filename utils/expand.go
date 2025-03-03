package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Function to expand the tilde (~) to the full home directory path
// @Example ~/Pictures/flowers.png --> /home/username/Pictures/flowers.png
func ExpandHomeDirectory(paths []string) []FileSource {
	var expandedFiles []FileSource
	homeDir, _ := os.UserHomeDir()

	for _, path := range paths {
		if strings.HasPrefix(path, "~") {
			path = filepath.Join(homeDir, path[1:])
		}
		expandedFiles = append(expandedFiles, FileSource{Path: path})
	}
	return expandedFiles
}

// Function to expand the delimiter '#' to every file under that directory
//
//	Example "~/Pictures/#" -->["Pictures/img1.png","~/Pictures/img2.png","~/Pictures/img3.png"]
func ExpandDirectory(pathToDir string) ([]FileSource, error) {
	// path := DiscardLastCharacter(pathToDir)

	ImgFileSources, err := expandToImgFiles(pathToDir)
	if err != nil {
		return nil, fmt.Errorf("error expanding to image files: %w", err)
	}

	return ImgFileSources, nil
}

// Expands a directory to only image files of type .png .jpeg .jpg .webp
//
//	Example "~/Pictures/" -->["Pictures/img1.png","~/Pictures/img2.png","~/Pictures/img3.png"]
func expandToImgFiles(path string) ([]FileSource, error) {
	filePaths, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	images, err := filterImages(filePaths)
	if err != nil {
		return nil, err
	}

	if len(images) == 0 {
		return nil, fmt.Errorf("no image files in directory")
	}

	contents := make([]FileSource, len(images))
	for i, img := range images {
		fullPath := filepath.Join(path, img)
		contents[i] = FileSource{Path: fullPath}
		// fmt.Println(fullPath) debugging
	}

	return contents, nil
}
