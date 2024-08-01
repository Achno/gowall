package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Function to expand the tilde (~) to the full home directory path
// @Example ~/Pictures/flowers.png --> /home/username/Pictures/flowers.png
func ExpandHomeDirectory(paths []string) []string {
	var expandedPaths []string
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Printf("Error getting home directory: %v\n", err)
		return paths // Return original paths if home directory cannot be determined
	}

	for _, path := range paths {
		if strings.HasPrefix(path, "~") {
			path = filepath.Join(homeDir, path[1:])
		}
		expandedPaths = append(expandedPaths, path)
	}
	return expandedPaths
}

// Function to expand the delimiter '#' to every file under that directory
//  Example "~/Pictures/#" -->["Pictures/img1.png","~/Pictures/img2.png","~/Pictures/img3.png"]
func ExpandHashtag(pathWithHashtag string) ([]string, error) {
	path := DiscardLastCharacter(pathWithHashtag)

	imgPaths, err := expandToImgFiles(path)
	if err != nil {
		return nil, fmt.Errorf("error expanding to image files: %w", err)
	}

	return imgPaths, nil
}

// Expands a directory to only image files of type .png .jpeg .jpg .webp
//  Example "~/Pictures/" -->["Pictures/img1.png","~/Pictures/img2.png","~/Pictures/img3.png"]
func expandToImgFiles(path string) ([]string, error) {
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

	contents := make([]string, len(images))
	for i, img := range images {
		fullPath := filepath.Join(path, img)
		contents[i] = fullPath
		// fmt.Println(fullPath) debugging
	}

	return contents, nil
}
