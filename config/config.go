package config

import (
	"errors"
	"log"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

var SupportedExtensions = map[string]bool{
	".png":  true,
	".jpeg": true,
	".jpg":  true,
	".webp": true,
}

type GlobalSubCommandFlags struct {
	OutputDestination string
	InputDir          string
	InputFiles        []string
	Theme             string
	Format            string
}

type themeWrapper struct {
	Name   string   `yaml:"name"`
	Colors []string `yaml:"colors"`
}

type Options struct {
	EnableImagePreviewing  bool           `yaml:"EnableImagePreviewing"`
	InlineImagePreview     bool           `yaml:"InlineImagePreview"`
	ColorCorrectionBackend string         `yaml:"ColorCorrectionBackend"`
	OutputFolder           string         `yaml:"OutputFolder"`
	Themes                 []themeWrapper `yaml:"themes"`
}

// global config object, used when config is needed
var GowallConfig = defaultConfig()

func init() {
	// look for $HOME/.config/gowall/config.yml
	configDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("Error could not get Home directory")
	}
	configPath := filepath.Join(configDir, ".config", "gowall", configFile)

	if _, err = os.Stat(configPath); errors.Is(err, os.ErrNotExist) {
		// file doesnt exist skip config file
		return
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		log.Printf("error reading config file: %v", err)
		return
	}

	err = yaml.Unmarshal(data, &GowallConfig)
	if err != nil {
		log.Printf("error unmarshalling config file: %v", err)
		return
	}
}
