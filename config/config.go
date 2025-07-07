package config

import (
	"errors"
	"log"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

var SupportedImageExtensions = map[string]bool{
	".png":  true,
	".jpeg": true,
	".jpg":  true,
	".webp": true,
}

var SupportedTextExtensions = map[string]bool{
	".pdf": true,
}

type GlobalSubCommandFlags struct {
	OutputDestination string
	InputDir          string
	InputFiles        []string
	Format            string
}

type themeWrapper struct {
	Name   string   `yaml:"name"`
	Colors []string `yaml:"colors"`
}

type Options struct {
	EnableImagePreviewing  bool           `yaml:"EnableImagePreviewing"`
	InlineImagePreview     bool           `yaml:"InlineImagePreview"`
	ImagePreviewBackend    string         `yaml:"ImagePreviewBackend"`
	ColorCorrectionBackend string         `yaml:"ColorCorrectionBackend"`
	OutputFolder           string         `yaml:"OutputFolder"`
	Themes                 []themeWrapper `yaml:"themes"`
	EnvConfig              *EnvConfig
	EnvFilePath            string `yaml:"EnvFilePath"`
}

var GowallConfig = defaultConfig()

func LoadConfig() {
	configDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("Error: Could not get home directory")
	}

	configPath := filepath.Join(configDir, ".config", "gowall", configFile)
	configFolder := filepath.Dir(configPath)

	err = os.MkdirAll(configFolder, 0755)
	if err != nil {
		log.Fatalf("could not create config directory %v", err)
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		log.Printf("Error reading config file: %v", err)
		return
	}

	err = yaml.Unmarshal(data, &GowallConfig)
	if err != nil {
		log.Printf("Error unmarshalling config file: %v", err)
		return
	}

	// err = os.MkdirAll(configFolder, 0755)
	// if err != nil {
	// 	log.Fatalf("Error: Could not create config directory: %v", err)
	// }

	defaultDir, err := CreateDirectory()
	if err != nil {
		log.Fatalf("Error: Could not create output directories: %v", err)
	}
	GowallConfig.OutputFolder = defaultDir
	if _, err := os.Stat(configPath); errors.Is(err, os.ErrNotExist) {
		return
	}

	// data, err = os.ReadFile(configPath)
	// if err != nil {
	// 	log.Printf("Error reading config file: %v", err)
	// 	return
	// }

	// err = yaml.Unmarshal(data, &GowallConfig)
	// if err != nil {
	// 	log.Printf("Error unmarshalling config file: %v", err)
	// 	return
	// }

	EnvConfig := GetEnvConfig(filepath.Join(configDir, GowallConfig.EnvFilePath))
	GowallConfig.EnvConfig = EnvConfig
}
