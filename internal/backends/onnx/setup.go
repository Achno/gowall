package onnx

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/Achno/gowall/internal/logger"
	"github.com/Achno/gowall/utils"
)

const onnxRuntimeVersion = "1.24.4"

func OnnxRuntimeFolder() string {
	homeDir, _ := os.UserHomeDir()
	return filepath.Join(homeDir, ".u2net")
}

// SharedLibraryName returns the platform-specific shared library name
func SharedLibraryName() string {
	switch runtime.GOOS {
	case "linux":
		return fmt.Sprintf("libonnxruntime.so.%s", onnxRuntimeVersion)
	case "darwin":
		return fmt.Sprintf("libonnxruntime.%s.dylib", onnxRuntimeVersion)
	case "windows":
		return "onnxruntime.dll"
	default:
		return ""
	}
}

func runtimeDownloadURL() (string, error) {
	urls := map[string]string{
		"linux":   fmt.Sprintf("https://github.com/microsoft/onnxruntime/releases/download/v%s/onnxruntime-linux-x64-%s.tgz", onnxRuntimeVersion, onnxRuntimeVersion),
		"darwin":  fmt.Sprintf("https://github.com/microsoft/onnxruntime/releases/download/v%s/onnxruntime-osx-arm64-%s.tgz", onnxRuntimeVersion, onnxRuntimeVersion),
		"windows": fmt.Sprintf("https://github.com/microsoft/onnxruntime/releases/download/v%s/onnxruntime-win-x64-%s.zip", onnxRuntimeVersion, onnxRuntimeVersion),
	}

	url, exists := urls[runtime.GOOS]
	if !exists {
		return "", fmt.Errorf("unsupported OS: %s. Only available for linux, mac, windows", runtime.GOOS)
	}
	return url, nil
}

// SetupOnnxRuntime downloads and extracts only the ONNX runtime shared library
func SetupOnnxRuntime() error {
	destFolder := OnnxRuntimeFolder()

	url, err := runtimeDownloadURL()
	if err != nil {
		return err
	}

	size := utils.GetRemoteFileSize(url)
	sizeInfo := ""
	if size != "" {
		sizeInfo = fmt.Sprintf(" size: %s,", size)
	}

	logger.Print(fmt.Sprintf("%s ➜ Downloading ONNX Runtime,%s sit back and relax%s", utils.BlueColor, sizeInfo, utils.ResetColor))

	if err := os.MkdirAll(destFolder, 0755); err != nil {
		return fmt.Errorf("create destination folder: %w", err)
	}

	// Download archive to temp location
	archiveName := filepath.Base(url)
	archivePath := filepath.Join(destFolder, archiveName)

	if err := utils.DownloadUrl(url, archivePath); err != nil {
		return fmt.Errorf("download: %w", err)
	}
	defer os.Remove(archivePath)

	// Extract only the shared library
	libName := SharedLibraryName()
	libPath := filepath.Join(destFolder, libName)

	if err := utils.ExtractSingleFile(archivePath, libPath, libName); err != nil {
		return fmt.Errorf("extract shared library: %w", err)
	}

	logger.Print(utils.BlueColor + " ➜ Process complete. ONNX Runtime setup" + utils.ResetColor)
	return nil
}

// CheckOnnxRuntimeInstalled checks if the ONNX runtime shared library is available
func CheckOnnxRuntimeInstalled() (string, error) {
	destFolder := OnnxRuntimeFolder()
	libName := SharedLibraryName()

	if libName == "" {
		return "", fmt.Errorf("unsupported OS: %s", runtime.GOOS)
	}

	// Library should be directly in the folder now
	libPath := filepath.Join(destFolder, libName)

	if _, err := os.Stat(libPath); err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("onnx runtime library not found at %s", libPath)
		}
		return "", fmt.Errorf("error checking library: %w", err)
	}

	return libPath, nil
}
