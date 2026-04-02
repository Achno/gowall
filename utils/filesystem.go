package utils

import (
	"archive/tar"
	"archive/zip"
	"compress/bzip2"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"net/http"
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

var errArchiveEntryMatched = errors.New("archive entry matched")

// ExtractZip extracts a zip archive to the destination directory.
func ExtractZip(src, dest string) error {
	return walkZipArchive(src, func(file *zip.File) error {
		targetPath, err := safeExtractPath(dest, file.Name)
		if err != nil {
			return err
		}

		if file.FileInfo().IsDir() {
			if err := os.MkdirAll(targetPath, 0o755); err != nil {
				return fmt.Errorf("create directory %q: %w", targetPath, err)
			}
			return nil
		}

		return extractZipFile(file, targetPath)
	})
}

func walkZipArchive(src string, fn func(file *zip.File) error) error {
	reader, err := zip.OpenReader(src)
	if err != nil {
		return fmt.Errorf("open zip archive: %w", err)
	}
	defer reader.Close()

	for _, file := range reader.File {
		if err := fn(file); err != nil {
			return err
		}
	}

	return nil
}

func extractZipFile(file *zip.File, targetPath string) error {
	if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
		return fmt.Errorf("create directory for %q: %w", targetPath, err)
	}

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

	return extractTarArchive(tar.NewReader(gzipReader), dest)
}

// ExtractBz2 extracts a .tar.bz2 archive to the destination directory.
func ExtractBz2(src, dest string) error {
	file, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("open tar.bz2 archive: %w", err)
	}
	defer file.Close()

	bz2Reader := bzip2.NewReader(file)
	return extractTarArchive(tar.NewReader(bz2Reader), dest)
}

// extractTarArchive is the shared tar extraction logic for both tgz and bz2.
func extractTarArchive(tarReader *tar.Reader, dest string) error {
	return walkTarArchive(tarReader, func(header *tar.Header, tarReader *tar.Reader) error {
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
			if err := extractTarFile(tarReader, targetPath, header.Mode); err != nil {
				return err
			}
		}

		return nil
	})
}

func walkTarArchive(tarReader *tar.Reader, fn func(header *tar.Header, tarReader *tar.Reader) error) error {
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("read tar entry: %w", err)
		}

		if err := fn(header, tarReader); err != nil {
			return err
		}
	}

	return nil
}

func extractTarFile(tarReader *tar.Reader, targetPath string, mode int64) error {
	if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
		return fmt.Errorf("create directory for %q: %w", targetPath, err)
	}

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

// GetRemoteFileSize returns the size of a remote file in a human-readable format.
// Returns empty string if size cannot be determined.
func GetRemoteFileSize(url string) string {
	resp, err := http.Head(url)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()

	if resp.ContentLength <= 0 {
		return ""
	}

	return FormatFileSize(resp.ContentLength)
}

// FormatFileSize formats bytes into human-readable format (KB, MB, GB).
func FormatFileSize(bytes int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)

	switch {
	case bytes >= GB:
		return fmt.Sprintf("%.1f GB", float64(bytes)/float64(GB))
	case bytes >= MB:
		return fmt.Sprintf("%.1f MB", float64(bytes)/float64(MB))
	case bytes >= KB:
		return fmt.Sprintf("%.1f KB", float64(bytes)/float64(KB))
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}

// ExtractSingleFile extracts a single file from an archive to destPath.
// It searches for a file matching targetName anywhere in the archive.
func ExtractSingleFile(archivePath, destPath, targetName string) error {
	lower := strings.ToLower(archivePath)
	switch {
	case strings.HasSuffix(lower, ".zip"):
		return extractSingleFromZip(archivePath, destPath, targetName)
	case strings.HasSuffix(lower, ".tgz"), strings.HasSuffix(lower, ".tar.gz"):
		return extractSingleFromTgz(archivePath, destPath, targetName)
	case strings.HasSuffix(lower, ".tar.bz2"):
		return extractSingleFromBz2(archivePath, destPath, targetName)
	default:
		return fmt.Errorf("unsupported archive format: %s", archivePath)
	}
}

func extractSingleFromZip(archivePath, destPath, targetName string) error {
	found := false
	err := walkZipArchive(archivePath, func(file *zip.File) error {
		if file.FileInfo().IsDir() || filepath.Base(file.Name) != targetName {
			return nil
		}

		found = true
		if err := extractZipFile(file, destPath); err != nil {
			return err
		}
		return errArchiveEntryMatched
	})
	if err != nil && !errors.Is(err, errArchiveEntryMatched) {
		return err
	}

	if !found {
		return fmt.Errorf("file %q not found in archive", targetName)
	}

	return nil
}

func extractSingleFromTgz(archivePath, destPath, targetName string) error {
	file, err := os.Open(archivePath)
	if err != nil {
		return fmt.Errorf("open tar.gz archive: %w", err)
	}
	defer file.Close()

	gzipReader, err := gzip.NewReader(file)
	if err != nil {
		return fmt.Errorf("open gzip stream: %w", err)
	}
	defer gzipReader.Close()

	return extractSingleFromTar(tar.NewReader(gzipReader), destPath, targetName)
}

func extractSingleFromBz2(archivePath, destPath, targetName string) error {
	file, err := os.Open(archivePath)
	if err != nil {
		return fmt.Errorf("open tar.bz2 archive: %w", err)
	}
	defer file.Close()

	bz2Reader := bzip2.NewReader(file)
	return extractSingleFromTar(tar.NewReader(bz2Reader), destPath, targetName)
}

func extractSingleFromTar(tarReader *tar.Reader, destPath, targetName string) error {
	found := false
	err := walkTarArchive(tarReader, func(header *tar.Header, tarReader *tar.Reader) error {
		if header.Typeflag != tar.TypeReg || filepath.Base(header.Name) != targetName {
			return nil
		}

		found = true
		if err := extractTarFile(tarReader, destPath, header.Mode); err != nil {
			return err
		}
		return errArchiveEntryMatched
	})
	if err != nil && !errors.Is(err, errArchiveEntryMatched) {
		return err
	}

	if !found {
		return fmt.Errorf("file %q not found in archive", targetName)
	}

	return nil
}
