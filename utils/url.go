package utils

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"strings"
)

// opens a URL in your default browser of your operating system
func OpenURL(url string) error {

	var cmd *exec.Cmd

	switch runtime.GOOS {

	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	case "darwin":
		cmd = exec.Command("open", url)
	case "linux", "freebsd", "openbsd":
		cmd = exec.Command("xdg-open", url)

	default:
		return fmt.Errorf("unsupported platform")
	}

	return cmd.Start()

}

// downloads a file from a url and places it in a specified destination
func DownloadUrl(url, dest string) error {

	res, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to download file: %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download file: status code %d", res.StatusCode)
	}

	out, err := os.Create(dest)
	if err != nil {
		return fmt.Errorf("failed to create file: %v", err)
	}
	defer out.Close()

	_, err = io.Copy(out, res.Body)
	if err != nil {
		return fmt.Errorf("failed to write file: %v", err)
	}

	return nil
}

func GetFileExtensionFromURL(rawurl string) (string, error) {
	parsedURL, err := url.Parse(rawurl)
	if err != nil {
		return "", fmt.Errorf("could not parse URL: %w", err)
	}

	filePath := parsedURL.Path

	// Extract filename
	fileName := path.Base(filePath)

	// Remove query parameters, if any
	if idx := strings.Index(fileName, "?"); idx != -1 {
		fileName = fileName[:idx]
	}

	// Get the file extension
	extension := filepath.Ext(fileName)
	return extension, nil
}
