package config

import (
	"errors"
	"log"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Shared struct {
	Theme      string
	BatchFiles []string
}

type themeWrapper struct {
	Name   string   `yaml:"name"`
	Colors []string `yaml:"colors"`
}

type Options struct {
	EnableImagePreviewing bool           `yaml:"EnableImagePreviewing"`
	Themes                []themeWrapper `yaml:"themes"`
}

// global config object, used when config is needed
var GowallConfig = defaultConfig()

func init() {
	// look for $XDG_CONFIG_HOME/gowall/config.yml or $HOME/.config/gowall/config.yml
	configDir, err := os.UserConfigDir()

	if err != nil {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			// cant find home or config just give up
			return
		}
		configDir = filepath.Join(homeDir, ".config")
	}
	configPath := filepath.Join(configDir, "gowall", configFile)

	if _, err = os.Stat(configPath); errors.Is(err, os.ErrNotExist) {
		// file doesnt exist skip custom themes
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
