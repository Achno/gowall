package config

import (
	"os"
	"path/filepath"
)

const (
	Version            = "v0.2.3"
	OutputFolder       = "Pictures/gowall"
	configFile         = "config.yml"
	OCRSchemaFile      = "schema.yml"
	WallOfTheDayUrl    = "https://www.reddit.com/r/wallpaper/top/"
	HexCodeVisualUrl   = "https://lawlesscreation.github.io/hex-color-visualiser/"
	UpscalerBinaryName = "realesrgan-ncnn-vulkan"
	PngquantBinaryName = "pngquant"
	EnvFilePath        = ".gowall/.env"
	OnnxRuntimeVersion = "1.24.4"
)

var (
	EnableImagePreviewingDefault = true
	InlineImagePreviewDefault    = false
	ImagePreviewBackend          = ""
	ThemesDefault                = []themeWrapper{}
	OnnxRuntimeFolderPath        = OnnxRuntimeFolder()
	OnnxModelFolderPath          = OnnxRuntimeFolder()
)

func defaultConfig() Options {
	return Options{
		EnableImagePreviewing: EnableImagePreviewingDefault,
		Themes:                ThemesDefault,
		InlineImagePreview:    InlineImagePreviewDefault,
		ImagePreviewBackend:   ImagePreviewBackend,
		EnvFilePath:           EnvFilePath,
	}
}

func OnnxRuntimeFolder() string {
	homeDir, _ := os.UserHomeDir()
	return filepath.Join(homeDir, ".u2net")
}
