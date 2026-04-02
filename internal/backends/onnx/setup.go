package onnx

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/Achno/gowall/config"
	"github.com/Achno/gowall/internal/logger"
	"github.com/Achno/gowall/utils"
	ort "github.com/yalue/onnxruntime_go"
)

var onnxRuntimeVersion = config.OnnxRuntimeVersion

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

func ensureRuntimeAvailable() (string, error) {
	runtimePath, err := CheckOnnxRuntimeInstalled()
	if err == nil {
		return runtimePath, nil
	}

	prompt := fmt.Sprintf("%s ◈ ONNX Runtime is not installed. Would you like to set it up?%s", utils.BlueColor, utils.ResetColor)
	if !utils.Confirm(prompt) {
		return "", fmt.Errorf("onnx runtime download declined")
	}

	if err := SetupOnnxRuntime(); err != nil {
		return "", err
	}

	return CheckOnnxRuntimeInstalled()
}

func ensureEnvironment(libraryPath string) error {
	ort.SetSharedLibraryPath(libraryPath)
	if err := ort.InitializeEnvironment(); err != nil {
		return fmt.Errorf("initialize onnxruntime: %w", err)
	}

	return nil
}

func ensureModelAvailable(model Model) (string, error) {
	baseDir := config.OnnxModelFolderPath
	modelPath := modelCachePath(baseDir, model)

	if _, err := os.Stat(modelPath); err == nil {
		return modelPath, nil
	} else if !os.IsNotExist(err) {
		return "", fmt.Errorf("check model path %q: %w", modelPath, err)
	}

	prompt := fmt.Sprintf("%s ◈ Model %q is missing. Download it?%s", utils.BlueColor, model.Name(), utils.ResetColor)
	if !utils.Confirm(prompt) {
		return "", fmt.Errorf("model download declined for %q", model.Name())
	}

	// Get file size for progress info
	size := utils.GetRemoteFileSize(model.DownloadURL())
	sizeInfo := ""
	if size != "" {
		sizeInfo = fmt.Sprintf(" size: %s,", size)
	}

	logger.Print(fmt.Sprintf("%s ➜ Downloading %s,%s sit back and relax%s", utils.BlueColor, model.Name(), sizeInfo, utils.ResetColor))

	if err := os.MkdirAll(filepath.Dir(modelPath), 0o755); err != nil {
		return "", fmt.Errorf("create model directory: %w", err)
	}

	if err := model.Download(DownloadOptions{DestPath: modelPath}); err != nil {
		return "", fmt.Errorf("download model %q: %w", model.Name(), err)
	}

	if _, err := os.Stat(modelPath); err != nil {
		return "", fmt.Errorf("model %q was not written to %s: %w", model.Name(), modelPath, err)
	}

	logger.Print(fmt.Sprintf("%s ➜ Process complete. Model %s downloaded%s", utils.BlueColor, model.Name(), utils.ResetColor))

	return modelPath, nil
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
	destFolder := config.OnnxRuntimeFolderPath

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
	destFolder := config.OnnxRuntimeFolderPath
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

// modelCachePath returns ~/.u2net/<model>.onnx
func modelCachePath(baseDir string, model Model) string {
	name := strings.TrimSpace(model.Name())
	fileName := name
	if filepath.Ext(fileName) == "" {
		fileName += ".onnx"
	}
	return filepath.Join(baseDir, fileName)
}

func namesFromIO(infos []ort.InputOutputInfo) []string {
	names := make([]string, 0, len(infos))
	for _, info := range infos {
		names = append(names, info.Name)
	}
	return names
}

func cloneIOInfo(infos []ort.InputOutputInfo) []ort.InputOutputInfo {
	if len(infos) == 0 {
		return nil
	}

	cloned := make([]ort.InputOutputInfo, len(infos))
	copy(cloned, infos)
	return cloned
}

func destroyValues(values []ort.Value) {
	for _, value := range values {
		if value == nil {
			continue
		}
		value.Destroy()
	}
}
