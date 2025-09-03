package utils

import (
	"archive/tar"
	"archive/zip"
	"compress/bzip2"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// Function to expand the tilde (~) to the full home directory path
// @Example ~/Pictures/flowers.png --> /home/username/Pictures/flowers.png
func ExpandTilde(paths []string) []string {
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

// FindBinary checks that the binary exists in dest or $PATH, preferring $PATH.
// Checks that binary also has executable permissions.
func FindBinary(binaryNames map[string]string, destFolder string) (string, error) {

	binaryName, ok := binaryNames[runtime.GOOS]
	if !ok {
		return "", fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}

	binaryPath := filepath.Join(destFolder, binaryName)

	// NixOS does not allow dynamically linked binaries,so just check $PATH for it instead.
	path, err := exec.LookPath(binaryName)
	if err != nil {
	}

	if path != "" {
		binaryPath = path
	}

	// Check if binary exists and is executable
	info, err := os.Stat(binaryPath)
	if err != nil {
		return "", fmt.Errorf("binary not found at %s: %w", binaryPath, err)
	}

	// check if the file is executable on unix
	if runtime.GOOS != "windows" {
		if info.Mode()&0111 == 0 {
			return "", fmt.Errorf("binary at %s is not executable", binaryPath)
		}
	}

	return binaryPath, nil
}

func ExtractArchiveBinary(src, dest, binary string) error {
	if strings.HasSuffix(strings.ToLower(src), ".tar.bz2") {
		return extractTarBz2Binary(src, dest, binary)
	} else if strings.HasSuffix(strings.ToLower(src), ".zip") {
		return extractZipBinary(src, dest, binary)
	}
	return fmt.Errorf("unsupported format")
}

// extractZipBinary extracts the zip file containing a binary to the dest folder and gives it exec permissions.
func extractZipBinary(src, dest, binary string) error {
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
			targetFileName := binary
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

func extractTarBz2Binary(src, dest, binary string) error {
	file, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open file: %v", err)
	}
	defer file.Close()

	bz2Reader := bzip2.NewReader(file)
	tarReader := tar.NewReader(bz2Reader)

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to read tar header: %v", err)
		}

		filePath := filepath.Join(dest, header.Name)
		// for every dir in the tar create its directory
		if header.Typeflag == tar.TypeDir {
			err := os.MkdirAll(filePath, 0755)
			if err != nil {
				return fmt.Errorf("failed to create directory: %v", err)
			}
		} else if header.Typeflag == tar.TypeReg {
			// create the nested directories
			err := os.MkdirAll(filepath.Dir(filePath), 0755)
			if err != nil {
				return fmt.Errorf("failed to create directory for file: %v", err)
			}
			// open the file for writing only,create it if it doesn't exist,truncate length to 0 if exists
			outFile, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.FileMode(header.Mode))
			if err != nil {
				return fmt.Errorf("failed to create file: %v", err)
			}
			_, err = io.Copy(outFile, tarReader)
			outFile.Close()
			if err != nil {
				return fmt.Errorf("failed to extract file: %v", err)
			}
			// give the binary execute permissions
			targetFileName := binary
			if filepath.Base(header.Name) == targetFileName {
				err := os.Chmod(filePath, os.FileMode(header.Mode)|0755)
				if err != nil {
					return fmt.Errorf("failed to chmod file: %v", err)
				}
			}
		}
	}
	return nil
}
