package utils

import (
	"archive/tar"
	"archive/zip"
	"compress/bzip2"
	"compress/gzip"
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

// safeExtractPath validates and returns a safe extraction path, preventing zip-slip attacks.
func safeExtractPath(dest, name string) (string, error) {
	targetPath := filepath.Join(dest, name)
	cleanDest := filepath.Clean(dest) + string(os.PathSeparator)
	cleanTarget := filepath.Clean(targetPath)

	if cleanTarget != filepath.Clean(dest) && !strings.HasPrefix(cleanTarget, cleanDest) {
		return "", fmt.Errorf("archive entry %q escapes destination %q", name, dest)
	}

	return cleanTarget, nil
}

// ExtractZip extracts a zip archive to the destination directory.
func ExtractZip(src, dest string) error {
	reader, err := zip.OpenReader(src)
	if err != nil {
		return fmt.Errorf("open zip archive: %w", err)
	}
	defer reader.Close()

	for _, file := range reader.File {
		targetPath, err := safeExtractPath(dest, file.Name)
		if err != nil {
			return err
		}

		if file.FileInfo().IsDir() {
			if err := os.MkdirAll(targetPath, 0o755); err != nil {
				return fmt.Errorf("create directory %q: %w", targetPath, err)
			}
			continue
		}

		if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
			return fmt.Errorf("create directory for %q: %w", targetPath, err)
		}

		if err := extractZipFile(file, targetPath); err != nil {
			return err
		}
	}

	return nil
}

func extractZipFile(file *zip.File, targetPath string) error {
	srcFile, err := file.Open()
	if err != nil {
		return fmt.Errorf("open zip member %q: %w", file.Name, err)
	}
	defer srcFile.Close()

	dstFile, err := os.OpenFile(targetPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, file.Mode())
	if err != nil {
		return fmt.Errorf("create extracted file %q: %w", targetPath, err)
	}
	defer dstFile.Close()

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return fmt.Errorf("extract %q: %w", file.Name, err)
	}

	return nil
}

// ExtractTgz extracts a .tar.gz or .tgz archive to the destination directory.
func ExtractTgz(src, dest string) error {
	file, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("open tar.gz archive: %w", err)
	}
	defer file.Close()

	gzipReader, err := gzip.NewReader(file)
	if err != nil {
		return fmt.Errorf("open gzip stream: %w", err)
	}
	defer gzipReader.Close()

	return extractTar(tar.NewReader(gzipReader), dest)
}

// ExtractBz2 extracts a .tar.bz2 archive to the destination directory.
func ExtractBz2(src, dest string) error {
	file, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("open tar.bz2 archive: %w", err)
	}
	defer file.Close()

	bz2Reader := bzip2.NewReader(file)
	return extractTar(tar.NewReader(bz2Reader), dest)
}

// extractTar is the shared tar extraction logic for both tgz and bz2.
func extractTar(tarReader *tar.Reader, dest string) error {
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("read tar entry: %w", err)
		}

		targetPath, err := safeExtractPath(dest, header.Name)
		if err != nil {
			return err
		}

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(targetPath, 0o755); err != nil {
				return fmt.Errorf("create directory %q: %w", targetPath, err)
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
				return fmt.Errorf("create directory for %q: %w", targetPath, err)
			}

			if err := extractTarFile(tarReader, targetPath, header.Mode); err != nil {
				return err
			}
		}
	}

	return nil
}

func extractTarFile(tarReader *tar.Reader, targetPath string, mode int64) error {
	dstFile, err := os.OpenFile(targetPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, os.FileMode(mode))
	if err != nil {
		return fmt.Errorf("create extracted file %q: %w", targetPath, err)
	}
	defer dstFile.Close()

	if _, err := io.Copy(dstFile, tarReader); err != nil {
		return fmt.Errorf("extract %q: %w", targetPath, err)
	}

	return nil
}

// ExtractArchive auto-detects the archive format and extracts it.
func ExtractArchive(src, dest string) error {
	lower := strings.ToLower(src)
	switch {
	case strings.HasSuffix(lower, ".zip"):
		return ExtractZip(src, dest)
	case strings.HasSuffix(lower, ".tgz"), strings.HasSuffix(lower, ".tar.gz"):
		return ExtractTgz(src, dest)
	case strings.HasSuffix(lower, ".tar.bz2"):
		return ExtractBz2(src, dest)
	default:
		return fmt.Errorf("unsupported archive format: %s", src)
	}
}

// GiveExecutePerms adds execute permissions to a file.
func GiveExecutePerms(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("stat file %q: %w", path, err)
	}
	return os.Chmod(path, info.Mode()|0755)
}

// ExtractBinary extracts an archive and gives execute permissions to the specified binary.
func ExtractBinary(src, dest, binaryName string) error {
	if err := ExtractArchive(src, dest); err != nil {
		return err
	}

	// Find and chmod the binary
	return filepath.WalkDir(dest, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && filepath.Base(path) == binaryName {
			return GiveExecutePerms(path)
		}
		return nil
	})
}

// DownloadAndExtract downloads an archive from url, extracts it to dest.
// If targetName is specified, only that file is kept (others are removed).
// If giveExec is true, execute permissions are added to the target file.
// The archive is cleaned up after extraction.
func DownloadAndExtract(url, dest, targetName string, giveExec bool) error {
	if err := os.MkdirAll(dest, 0755); err != nil {
		return fmt.Errorf("create destination folder: %w", err)
	}

	// Determine archive path from URL
	archiveName := filepath.Base(url)
	archivePath := filepath.Join(dest, archiveName)

	if err := DownloadUrl(url, archivePath); err != nil {
		return fmt.Errorf("download: %w", err)
	}
	defer os.Remove(archivePath)

	if err := ExtractArchive(archivePath, dest); err != nil {
		return err
	}

	if targetName != "" && giveExec {
		return findAndChmod(dest, targetName)
	}

	return nil
}

func findAndChmod(dest, targetName string) error {
	return filepath.WalkDir(dest, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && filepath.Base(path) == targetName {
			return GiveExecutePerms(path)
		}
		return nil
	})
}
