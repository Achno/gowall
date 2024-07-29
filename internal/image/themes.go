package image

import (
	"encoding/hex"
	"errors"
	"image/color"
	"log"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type Theme struct {
	Name   string
	Colors []color.Color
}

type themeWrapper struct {
	Name   string   `yaml:"name"`
	Colors []string `yaml:"colors"`
}

// Available themes
var themes = map[string]Theme{
	"catppuccin":       Catppuccin,
	"nord":             Nord,
	"everforest":       Everforest,
	"solarized":        Solarized,
	"gruvbox":          Gruvbox,
	"dracula":          Dracula,
	"tokyo-moon":       Tokyo_Moon,
	"onedark":          Onedark,
	"srcery" :          Srcery,
    "monokai":          Monokai,
    "material":         Material,
    "atom-one-light":   AtomOneLight,
    "synthwave-84":     Synthwave84,
    "atomdark":         AtomDark,
    "oceanic-next":     OceanicNext,
    "shades-of-purple": ShadesOfPurple,
    "arcdark":          ArcDark,
    "sunset-aurant":    SunsetAurant,
    "sunset-saffron":   SunsetSaffron,
    "sunset-tangerine": SunsetTangerine,
}

func init() {
	loadCustomThemes()
}

func loadCustomThemes() {

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
	configPath := filepath.Join(configDir, "gowall", "config.yml")

	if _, err = os.Stat(configPath); errors.Is(err, os.ErrNotExist) {
		// file doesnt exist skip custom themes
		return
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		log.Printf("error reading config file: %v", err)
		return
	}

	var rawConfig struct {
		Themes []themeWrapper `yaml:"themes"`
	}
	err = yaml.Unmarshal(data, &rawConfig)
	if err != nil {
		log.Printf("error unmarshalling config file: %v", err)
		return
	}

	for _, tw := range rawConfig.Themes {
		valid := true
		if tw.Name == "" || len(tw.Colors) == 0 {
			// skip invalid color
			continue
		}

		theme := Theme{
			Name:   tw.Name,
			Colors: make([]color.Color, len(tw.Colors)),
		}

		for i, hexColor := range tw.Colors {
			col, err := hexToRGBA(hexColor)
			if err != nil {
				log.Printf("invalid color %s in theme %s: %v", hexColor, tw.Name, err)
				valid = false
				break
			}
			theme.Colors[i] = col
		}

		if valid && !themeExists(theme.Name) {

			themes[strings.ToLower(theme.Name)] = theme
		}
	}
}

func hexToRGBA(hexStr string) (color.RGBA, error) {
	if len(hexStr) != 7 || hexStr[0] != '#' {
		return color.RGBA{}, errors.New("invalid hex color format")
	}
	bytes, err := hex.DecodeString(hexStr[1:])
	if err != nil {
		return color.RGBA{}, err
	}
	return color.RGBA{R: bytes[0], G: bytes[1], B: bytes[2], A: 255}, nil
}

func ListThemes() []string {
	allThemes := make([]string, 0, len(themes))
	for theme := range themes {
		allThemes = append(allThemes, theme)
	}
	return allThemes
}

func SelectTheme(theme string) (Theme, error) {
	selectedTheme, exists := themes[theme]

	if !exists {
		return Theme{}, errors.New("unknown theme")
	}

	return selectedTheme, nil
}

func themeExists(theme string) bool{
	
	_, exists := themes[theme]


	return exists
}

var (
	Catppuccin = Theme{
		Name: "Catpuccin",
		Colors: []color.Color{
			color.RGBA{R: 245, G: 224, B: 220, A: 255},
			color.RGBA{R: 242, G: 205, B: 205, A: 255},
			color.RGBA{R: 245, G: 194, B: 231, A: 255},
			color.RGBA{R: 203, G: 166, B: 247, A: 255},
			color.RGBA{R: 243, G: 139, B: 168, A: 255},
			color.RGBA{R: 235, G: 160, B: 172, A: 255},
			color.RGBA{R: 250, G: 179, B: 135, A: 255},
			color.RGBA{R: 249, G: 226, B: 175, A: 255},
			color.RGBA{R: 166, G: 227, B: 161, A: 255},
			color.RGBA{R: 148, G: 226, B: 213, A: 255},
			color.RGBA{R: 137, G: 220, B: 235, A: 255},
			color.RGBA{R: 116, G: 199, B: 236, A: 255},
			color.RGBA{R: 137, G: 180, B: 250, A: 255},
			color.RGBA{R: 180, G: 190, B: 254, A: 255},
			color.RGBA{R: 205, G: 214, B: 244, A: 255},
			color.RGBA{R: 186, G: 194, B: 222, A: 255},
			color.RGBA{R: 166, G: 173, B: 200, A: 255},
			color.RGBA{R: 147, G: 153, B: 178, A: 255},
			color.RGBA{R: 127, G: 132, B: 156, A: 255},
			color.RGBA{R: 108, G: 112, B: 134, A: 255},
			color.RGBA{R: 88, G: 91, B: 112, A: 255},
			color.RGBA{R: 69, G: 71, B: 90, A: 255},
			color.RGBA{R: 49, G: 50, B: 68, A: 255},
			color.RGBA{R: 30, G: 30, B: 46, A: 255},
			color.RGBA{R: 24, G: 24, B: 37, A: 255},
			color.RGBA{R: 17, G: 17, B: 27, A: 255},
		},
	}

	Nord = Theme{
		Name: "Nord",
		Colors: []color.Color{
			color.RGBA{R: 46, G: 52, B: 64, A: 255},
			color.RGBA{R: 59, G: 66, B: 82, A: 255},
			color.RGBA{R: 67, G: 76, B: 94, A: 255},
			color.RGBA{R: 76, G: 86, B: 106, A: 255},
			color.RGBA{R: 216, G: 222, B: 233, A: 255},
			color.RGBA{R: 229, G: 233, B: 240, A: 255},
			color.RGBA{R: 236, G: 239, B: 244, A: 255},
			color.RGBA{R: 143, G: 188, B: 187, A: 255},
			color.RGBA{R: 136, G: 192, B: 208, A: 255},
			color.RGBA{R: 129, G: 161, B: 193, A: 255},
			color.RGBA{R: 94, G: 129, B: 172, A: 255},
			color.RGBA{R: 191, G: 97, B: 106, A: 255},
			color.RGBA{R: 208, G: 135, B: 112, A: 255},
			color.RGBA{R: 235, G: 203, B: 139, A: 255},
			color.RGBA{R: 163, G: 190, B: 140, A: 255},
			color.RGBA{R: 180, G: 142, B: 173, A: 255},
		},
	}

	Everforest = Theme{
		Name: "Everforest",
		Colors: []color.Color{
			color.RGBA{R: 35, G: 42, B: 46, A: 255},
			color.RGBA{R: 45, G: 53, B: 59, A: 255},
			color.RGBA{R: 52, G: 63, B: 68, A: 255},
			color.RGBA{R: 61, G: 72, B: 77, A: 255},
			color.RGBA{R: 71, G: 82, B: 88, A: 255},
			color.RGBA{R: 79, G: 88, B: 94, A: 255},
			color.RGBA{R: 86, G: 99, B: 95, A: 255},
			color.RGBA{R: 84, G: 58, B: 72, A: 255},
			color.RGBA{R: 81, G: 64, B: 69, A: 255},
			color.RGBA{R: 66, G: 80, B: 71, A: 255},
			color.RGBA{R: 58, G: 81, B: 93, A: 255},
			color.RGBA{R: 77, G: 76, B: 67, A: 255},
			color.RGBA{R: 211, G: 198, B: 170, A: 255},
			color.RGBA{R: 230, G: 126, B: 128, A: 255},
			color.RGBA{R: 230, G: 152, B: 117, A: 255},
			color.RGBA{R: 219, G: 188, B: 127, A: 255},
			color.RGBA{R: 167, G: 192, B: 128, A: 255},
			color.RGBA{R: 131, G: 192, B: 146, A: 255},
			color.RGBA{R: 127, G: 187, B: 179, A: 255},
			color.RGBA{R: 214, G: 153, B: 182, A: 255},
			color.RGBA{R: 122, G: 132, B: 120, A: 255},
			color.RGBA{R: 133, G: 146, B: 137, A: 255},
			color.RGBA{R: 157, G: 169, B: 160, A: 255},
		},
	}

	Gruvbox = Theme{
		Name: "Gruvbox",
		Colors: []color.Color{
			color.RGBA{R: 40, G: 40, B: 40, A: 255},
			color.RGBA{R: 29, G: 32, B: 33, A: 255},
			color.RGBA{R: 50, G: 48, B: 47, A: 255},
			color.RGBA{R: 60, G: 56, B: 54, A: 255},
			color.RGBA{R: 80, G: 73, B: 69, A: 255},
			color.RGBA{R: 102, G: 92, B: 84, A: 255},
			color.RGBA{R: 124, G: 111, B: 100, A: 255},
			color.RGBA{R: 235, G: 219, B: 178, A: 255},
			color.RGBA{R: 251, G: 241, B: 199, A: 255},
			color.RGBA{R: 213, G: 196, B: 161, A: 255},
			color.RGBA{R: 189, G: 174, B: 147, A: 255},
			color.RGBA{R: 168, G: 153, B: 132, A: 255},
			color.RGBA{R: 146, G: 131, B: 116, A: 255},
			color.RGBA{R: 204, G: 36, B: 29, A: 255},
			color.RGBA{R: 251, G: 73, B: 52, A: 255},
			color.RGBA{R: 214, G: 93, B: 14, A: 255},
			color.RGBA{R: 254, G: 128, B: 25, A: 255},
			color.RGBA{R: 215, G: 153, B: 33, A: 255},
			color.RGBA{R: 250, G: 189, B: 47, A: 255},
			color.RGBA{R: 152, G: 151, B: 26, A: 255},
			color.RGBA{R: 184, G: 187, B: 38, A: 255},
			color.RGBA{R: 104, G: 157, B: 106, A: 255},
			color.RGBA{R: 142, G: 192, B: 124, A: 255},
			color.RGBA{R: 69, G: 133, B: 136, A: 255},
			color.RGBA{R: 131, G: 165, B: 152, A: 255},
			color.RGBA{R: 177, G: 98, B: 134, A: 255},
			color.RGBA{R: 211, G: 134, B: 155, A: 255},
		},
	}

	Solarized = Theme{
		Name: "Solarized",
		Colors: []color.Color{
			color.RGBA{R: 0, G: 43, B: 54, A: 255},
			color.RGBA{R: 7, G: 54, B: 66, A: 255},
			color.RGBA{R: 88, G: 110, B: 117, A: 255},
			color.RGBA{R: 101, G: 123, B: 131, A: 255},
			color.RGBA{R: 131, G: 148, B: 150, A: 255},
			color.RGBA{R: 147, G: 161, B: 161, A: 255},
			color.RGBA{R: 238, G: 232, B: 213, A: 255},
			color.RGBA{R: 253, G: 246, B: 227, A: 255},
			color.RGBA{R: 181, G: 137, B: 0, A: 255},
			color.RGBA{R: 203, G: 75, B: 22, A: 255},
			color.RGBA{R: 220, G: 50, B: 47, A: 255},
			color.RGBA{R: 211, G: 54, B: 130, A: 255},
			color.RGBA{R: 108, G: 113, B: 196, A: 255},
			color.RGBA{R: 38, G: 139, B: 210, A: 255},
			color.RGBA{R: 42, G: 161, B: 152, A: 255},
			color.RGBA{R: 133, G: 153, B: 0, A: 255},
		},
	}

	Dracula = Theme{
		Name: "Dracula",
		Colors: []color.Color{
			color.RGBA{R: 40, G: 42, B: 54, A: 255},
			color.RGBA{R: 68, G: 71, B: 90, A: 255},
			color.RGBA{R: 248, G: 248, B: 242, A: 255},
			color.RGBA{R: 98, G: 114, B: 164, A: 255},
			color.RGBA{R: 139, G: 233, B: 253, A: 255},
			color.RGBA{R: 80, G: 250, B: 123, A: 255},
			color.RGBA{R: 255, G: 184, B: 108, A: 255},
			color.RGBA{R: 255, G: 121, B: 198, A: 255},
			color.RGBA{R: 189, G: 147, B: 249, A: 255},
			color.RGBA{R: 255, G: 85, B: 85, A: 255},
			color.RGBA{R: 241, G: 250, B: 140, A: 255},
		},
	}

	Tokyo_Moon = Theme{
		Name: "Tokyo_Moon",
		Colors: []color.Color{
			color.RGBA{R: 34, G: 36, B: 54, A: 255},
			color.RGBA{R: 27, G: 29, B: 43, A: 255},
			color.RGBA{R: 130, G: 170, B: 255, A: 255},
			color.RGBA{R: 68, G: 74, B: 115, A: 255},
			color.RGBA{R: 130, G: 170, B: 255, A: 255},
			color.RGBA{R: 134, G: 225, B: 252, A: 255},
			color.RGBA{R: 195, G: 232, B: 141, A: 255},
			color.RGBA{R: 252, G: 167, B: 234, A: 255},
			color.RGBA{R: 255, G: 117, B: 127, A: 255},
			color.RGBA{R: 200, G: 211, B: 245, A: 255},
			color.RGBA{R: 255, G: 199, B: 119, A: 255},
			color.RGBA{R: 200, G: 211, B: 245, A: 255},
			color.RGBA{R: 134, G: 225, B: 252, A: 255},
			color.RGBA{R: 200, G: 211, B: 245, A: 255},
			color.RGBA{R: 195, G: 232, B: 141, A: 255},
			color.RGBA{R: 192, G: 153, B: 255, A: 255},
			color.RGBA{R: 255, G: 117, B: 127, A: 255},
			color.RGBA{R: 45, G: 63, B: 118, A: 255},
			color.RGBA{R: 130, G: 139, B: 184, A: 255},
			color.RGBA{R: 255, G: 199, B: 119, A: 255},
		},
	}

	Onedark = Theme{
		Name: "Onedark",
		Colors: []color.Color{
			color.RGBA{R: 24, G: 26, B: 31, A: 255},
			color.RGBA{R: 40, G: 44, B: 52, A: 255},
			color.RGBA{R: 49, G: 53, B: 63, A: 255},
			color.RGBA{R: 57, G: 63, B: 74, A: 255},
			color.RGBA{R: 59, G: 63, B: 76, A: 255},
			color.RGBA{R: 33, G: 37, B: 43, A: 255},
			color.RGBA{R: 115, G: 184, B: 241, A: 255},
			color.RGBA{R: 235, G: 208, B: 156, A: 255},
			color.RGBA{R: 171, G: 178, B: 191, A: 255},
			color.RGBA{R: 198, G: 120, B: 221, A: 255},
			color.RGBA{R: 152, G: 195, B: 121, A: 255},
			color.RGBA{R: 209, G: 154, B: 102, A: 255},
			color.RGBA{R: 97, G: 175, B: 239, A: 255},
			color.RGBA{R: 229, G: 192, B: 123, A: 255},
			color.RGBA{R: 86, G: 182, B: 194, A: 255},
			color.RGBA{R: 232, G: 102, B: 113, A: 255},
			color.RGBA{R: 92, G: 99, B: 112, A: 255},
			color.RGBA{R: 132, G: 139, B: 152, A: 255},
			color.RGBA{R: 43, G: 111, B: 119, A: 255},
			color.RGBA{R: 153, G: 57, B: 57, A: 255},
			color.RGBA{R: 147, G: 105, B: 29, A: 255},
			color.RGBA{R: 138, G: 63, B: 160, A: 255},
			color.RGBA{R: 49, G: 57, B: 43, A: 255},
			color.RGBA{R: 56, G: 43, B: 44, A: 255},
			color.RGBA{R: 28, G: 52, B: 72, A: 255},
			color.RGBA{R: 44, G: 83, B: 114, A: 255},
		},
	}

	Srcery = Theme{
		Name: "Srcery",
		Colors: []color.Color{
			color.RGBA{R: 28, G: 27, B: 25, A: 255}, // #1C1B19
			color.RGBA{R: 239, G: 47, B: 39, A: 255}, // #EF2F27
			color.RGBA{R: 81, G: 159, B: 80, A: 255}, // #519F50
			color.RGBA{R: 251, G: 184, B: 41, A: 255}, // #FBB829
			color.RGBA{R: 44, G: 120, B: 191, A: 255}, // #2C78BF
			color.RGBA{R: 224, G: 44, B: 109, A: 255}, // #E02C6D
			color.RGBA{R: 10, G: 174, B: 179, A: 255}, // #0AAEB3
			color.RGBA{R: 186, G: 166, B: 127, A: 255}, // #BAA67F
			color.RGBA{R: 145, G: 129, B: 117, A: 255}, // #918175
			color.RGBA{R: 247, G: 83, B: 65, A: 255}, // #F75341
			color.RGBA{R: 152, G: 188, B: 55, A: 255}, // #98BC37
			color.RGBA{R: 254, G: 208, B: 110, A: 255}, // #FED06E
			color.RGBA{R: 104, G: 168, B: 228, A: 255}, // #68A8E4
			color.RGBA{R: 255, G: 92, B: 143, A: 255}, // #FF5C8F
			color.RGBA{R: 43, G: 228, B: 208, A: 255}, // #2BE4D0
			color.RGBA{R: 252, G: 232, B: 195, A: 255}, // #FCE8C3
		},
	}
	
    Monokai = Theme{
		Name: "Monokai",
		Colors: []color.Color{
			color.RGBA{R: 39, G: 40, B: 34, A: 255},
			color.RGBA{R: 248, G: 248, B: 242, A: 255},
			color.RGBA{R: 255, G: 85, B: 85, A: 255},
			color.RGBA{R: 255, G: 121, B: 198, A: 255},
			color.RGBA{R: 189, G: 147, B: 249, A: 255},
			color.RGBA{R: 80, G: 250, B: 123, A: 255},
			color.RGBA{R: 255, G: 184, B: 108, A: 255},
			color.RGBA{R: 241, G: 250, B: 140, A: 255},
			color.RGBA{R: 39, G: 40, B: 34, A: 255},
			color.RGBA{R: 248, G: 248, B: 242, A: 255},
			color.RGBA{R: 255, G: 85, B: 85, A: 255},
			color.RGBA{R: 255, G: 121, B: 198, A: 255},
			color.RGBA{R: 189, G: 147, B: 249, A: 255},
			color.RGBA{R: 80, G: 250, B: 123, A: 255},
			color.RGBA{R: 255, G: 184, B: 108, A: 255},
			color.RGBA{R: 241, G: 250, B: 140, A: 255},
		},
	}

	Material = Theme{
		Name: "Material",
		Colors: []color.Color{
			color.RGBA{R: 38, G: 50, B: 56, A: 255},
			color.RGBA{R: 255, G: 83, B: 112, A: 255},
			color.RGBA{R: 156, G: 39, B: 176, A: 255},
			color.RGBA{R: 103, G: 58, B: 183, A: 255},
			color.RGBA{R: 33, G: 150, B: 243, A: 255},
			color.RGBA{R: 3, G: 169, B: 244, A: 255},
			color.RGBA{R: 0, G: 188, B: 212, A: 255},
			color.RGBA{R: 0, G: 150, B: 136, A: 255},
			color.RGBA{R: 76, G: 175, B: 80, A: 255},
			color.RGBA{R: 139, G: 195, B: 74, A: 255},
			color.RGBA{R: 205, G: 220, B: 57, A: 255},
			color.RGBA{R: 255, G: 235, B: 59, A: 255},
			color.RGBA{R: 255, G: 193, B: 7, A: 255},
			color.RGBA{R: 255, G: 152, B: 0, A: 255},
			color.RGBA{R: 255, G: 87, B: 34, A: 255},
			color.RGBA{R: 121, G: 85, B: 72, A: 255},
		},
	}

    AtomOneLight = Theme{
		Name: "AtomOneLight",
		Colors: []color.Color{
			color.RGBA{R: 245, G: 245, B: 245, A: 255},
			color.RGBA{R: 250, G: 250, B: 250, A: 255},
			color.RGBA{R: 248, G: 248, B: 248, A: 255},
			color.RGBA{R: 245, G: 245, B: 245, A: 255},
			color.RGBA{R: 238, G: 238, B: 238, A: 255},
			color.RGBA{R: 231, G: 231, B: 231, A: 255},
			color.RGBA{R: 219, G: 219, B: 219, A: 255},
			color.RGBA{R: 203, G: 203, B: 203, A: 255},
			color.RGBA{R: 198, G: 198, B: 198, A: 255},
			color.RGBA{R: 121, G: 112, B: 116, A: 255},
			color.RGBA{R: 183, G: 182, B: 168, A: 255},
			color.RGBA{R: 241, G: 241, B: 240, A: 255},
			color.RGBA{R: 170, G: 182, B: 194, A: 255},
			color.RGBA{R: 206, G: 208, B: 213, A: 255},
			color.RGBA{R: 142, G: 170, B: 198, A: 255},
			color.RGBA{R: 170, G: 189, B: 216, A: 255},
		},
	}

    Synthwave84 = Theme{
		Name: "Synthwave84",
		Colors: []color.Color{
			color.RGBA{R: 24, G: 25, B: 31, A: 255},
			color.RGBA{R: 42, G: 43, B: 50, A: 255},
			color.RGBA{R: 52, G: 54, B: 64, A: 255},
			color.RGBA{R: 72, G: 73, B: 83, A: 255},
			color.RGBA{R: 108, G: 108, B: 131, A: 255},
			color.RGBA{R: 139, G: 139, B: 172, A: 255},
			color.RGBA{R: 161, G: 161, B: 191, A: 255},
			color.RGBA{R: 196, G: 196, B: 214, A: 255},
			color.RGBA{R: 255, G: 83, B: 108, A: 255},
			color.RGBA{R: 255, G: 129, B: 137, A: 255},
			color.RGBA{R: 192, G: 128, B: 255, A: 255},
			color.RGBA{R: 127, G: 159, B: 255, A: 255},
			color.RGBA{R: 255, G: 195, B: 70, A: 255},
			color.RGBA{R: 255, G: 255, B: 153, A: 255},
			color.RGBA{R: 255, G: 163, B: 103, A: 255},
			color.RGBA{R: 191, G: 191, B: 222, A: 255},
		},
	}

	AtomDark = Theme{
		Name: "AtomDark",
		Colors: []color.Color{
			color.RGBA{R: 26, G: 32, B: 44, A: 255},
			color.RGBA{R: 204, G: 102, B: 102, A: 255},
			color.RGBA{R: 102, G: 204, B: 102, A: 255},
			color.RGBA{R: 204, G: 204, B: 102, A: 255},
			color.RGBA{R: 102, G: 204, B: 204, A: 255},
			color.RGBA{R: 204, G: 204, B: 204, A: 255},
			color.RGBA{R: 204, G: 102, B: 102, A: 255},
			color.RGBA{R: 204, G: 204, B: 102, A: 255},
			color.RGBA{R: 102, G: 204, B: 102, A: 255},
			color.RGBA{R: 102, G: 204, B: 204, A: 255},
			color.RGBA{R: 204, G: 204, B: 204, A: 255},
			color.RGBA{R: 204, G: 102, B: 102, A: 255},
			color.RGBA{R: 102, G: 204, B: 102, A: 255},
			color.RGBA{R: 204, G: 204, B: 102, A: 255},
			color.RGBA{R: 102, G: 204, B: 204, A: 255},
			color.RGBA{R: 26, G: 32, B: 44, A: 255},
		},
	}

    OceanicNext = Theme{
		Name: "Oceanic Next",
		Colors: []color.Color{
			color.RGBA{R: 28, G: 34, B: 40, A: 255},
			color.RGBA{R: 232, G: 102, B: 97, A: 255},
			color.RGBA{R: 118, G: 195, B: 115, A: 255},
			color.RGBA{R: 248, G: 185, B: 79, A: 255},
			color.RGBA{R: 102, G: 143, B: 220, A: 255},
			color.RGBA{R: 145, G: 151, B: 158, A: 255},
			color.RGBA{R: 102, G: 143, B: 220, A: 255},
			color.RGBA{R: 232, G: 102, B: 97, A: 255},
			color.RGBA{R: 122, G: 136, B: 149, A: 255},
			color.RGBA{R: 145, G: 151, B: 158, A: 255},
			color.RGBA{R: 248, G: 185, B: 79, A: 255},
			color.RGBA{R: 118, G: 195, B: 115, A: 255},
			color.RGBA{R: 102, G: 143, B: 220, A: 255},
			color.RGBA{R: 145, G: 151, B: 158, A: 255},
			color.RGBA{R: 28, G: 34, B: 40, A: 255},
		},
	}

    ShadesOfPurple = Theme{
		Name: "Shades of Purple",
		Colors: []color.Color{
			color.RGBA{R: 25, G: 20, B: 30, A: 255},
			color.RGBA{R: 209, G: 103, B: 139, A: 255},
			color.RGBA{R: 162, G: 195, B: 252, A: 255},
			color.RGBA{R: 209, G: 119, B: 255, A: 255},
			color.RGBA{R: 128, G: 186, B: 249, A: 255},
			color.RGBA{R: 153, G: 134, B: 159, A: 255},
			color.RGBA{R: 128, G: 186, B: 249, A: 255},
			color.RGBA{R: 209, G: 103, B: 139, A: 255},
			color.RGBA{R: 120, G: 106, B: 120, A: 255},
			color.RGBA{R: 153, G: 134, B: 159, A: 255},
			color.RGBA{R: 209, G: 119, B: 255, A: 255},
			color.RGBA{R: 162, G: 195, B: 252, A: 255},
			color.RGBA{R: 128, G: 186, B: 249, A: 255},
			color.RGBA{R: 153, G: 134, B: 159, A: 255},
			color.RGBA{R: 25, G: 20, B: 30, A: 255},
		},
	}

    ArcDark = Theme{
		Name: "Arc Dark",
		Colors: []color.Color{
			color.RGBA{R: 33, G: 33, B: 33, A: 255},
			color.RGBA{R: 255, G: 85, B: 85, A: 255},
			color.RGBA{R: 138, G: 191, B: 80, A: 255},
			color.RGBA{R: 255, G: 186, B: 77, A: 255},
			color.RGBA{R: 63, G: 127, B: 255, A: 255},
			color.RGBA{R: 136, G: 136, B: 136, A: 255},
			color.RGBA{R: 63, G: 127, B: 255, A: 255},
			color.RGBA{R: 255, G: 85, B: 85, A: 255},
			color.RGBA{R: 70, G: 70, B: 70, A: 255},
			color.RGBA{R: 136, G: 136, B: 136, A: 255},
			color.RGBA{R: 255, G: 186, B: 77, A: 255},
			color.RGBA{R: 138, G: 191, B: 80, A: 255},
			color.RGBA{R: 63, G: 127, B: 255, A: 255},
			color.RGBA{R: 136, G: 136, B: 136, A: 255},
			color.RGBA{R: 33, G: 33, B: 33, A: 255},
		},
	}

    SunsetAurant = Theme{
	    Name: "Sunset Aurant",
	    Colors: []color.Color{
		    color.RGBA{R: 0, G: 0, B: 0, A: 255},
		    color.RGBA{R: 255, G: 255, B: 255, A: 255},
		    color.RGBA{R: 201, G: 144, B: 252, A: 255},
		    color.RGBA{R: 214, G: 233, B: 187, A: 255},
		    color.RGBA{R: 200, G: 160, B: 239, A: 255},
		    color.RGBA{R: 198, G: 151, B: 242, A: 255},
		    color.RGBA{R: 47, G: 176, B: 215, A: 255},
		    color.RGBA{R: 211, G: 151, B: 88, A: 255},
		    color.RGBA{R: 201, G: 144, B: 252, A: 255},
		    color.RGBA{R: 247, G: 196, B: 215, A: 255},
		    color.RGBA{R: 251, G: 165, B: 200, A: 255},
		    color.RGBA{R: 224, G: 147, B: 30, A: 255},
		    color.RGBA{R: 56, G: 62, B: 48, A: 255},
		    color.RGBA{R: 86, G: 95, B: 74, A: 255},
		    color.RGBA{R: 123, G: 134, B: 106, A: 255},
		    color.RGBA{R: 165, G: 180, B: 144, A: 255},
		    color.RGBA{R: 243, G: 136, B: 19, A: 255},
	    },
    }

    SunsetSaffron = Theme{
	    Name: "Sunset Saffron",
	    Colors: []color.Color{
		    color.RGBA{R: 29, G: 32, B: 33, A: 255},
		    color.RGBA{R: 251, G: 241, B: 199, A: 255},
		    color.RGBA{R: 254, G: 128, B: 25, A: 255},
		    color.RGBA{R: 142, G: 192, B: 124, A: 255},
		    color.RGBA{R: 211, G: 134, B: 155, A: 255},
		    color.RGBA{R: 250, G: 189, B: 47, A: 255},
		    color.RGBA{R: 131, G: 165, B: 152, A: 255},
		    color.RGBA{R: 254, G: 128, B: 25, A: 255},
		    color.RGBA{R: 29, G: 32, B: 33, A: 255},
		    color.RGBA{R: 40, G: 40, B: 40, A: 255},
		    color.RGBA{R: 60, G: 56, B: 54, A: 255},
		    color.RGBA{R: 146, G: 131, B: 116, A: 255},
		    color.RGBA{R: 80, G: 73, B: 69, A: 255},
		    color.RGBA{R: 102, G: 92, B: 84, A: 255},
		    color.RGBA{R: 124, G: 111, B: 100, A: 255},
		    color.RGBA{R: 168, G: 153, B: 132, A: 255},
		    color.RGBA{R: 0, G: 0, B: 0, A: 255},
		    color.RGBA{R: 251, G: 241, B: 199, A: 255},
	    },
    }

    SunsetTangerine = Theme{
    	Name: "Sunset Tangerine",
    	Colors: []color.Color{
    		color.RGBA{R: 255, G: 87, B: 51, A: 255},
    		color.RGBA{R: 255, G: 218, B: 51, A: 255},
    		color.RGBA{R: 51, G: 255, B: 87, A: 255},
    		color.RGBA{R: 51, G: 138, B: 255, A: 255},
    		color.RGBA{R: 255, G: 51, B: 245, A: 255},
    		color.RGBA{R: 51, G: 230, B: 255, A: 255},
    		color.RGBA{R: 255, G: 87, B: 51, A: 255},
    		color.RGBA{R: 255, G: 133, B: 51, A: 255},
    		color.RGBA{R: 255, G: 207, B: 51, A: 255},
    		color.RGBA{R: 51, G: 255, B: 107, A: 255},
    		color.RGBA{R: 51, G: 166, B: 255, A: 255},
    		color.RGBA{R: 255, G: 51, B: 181, A: 255},
    		color.RGBA{R: 51, G: 247, B: 255, A: 255},
    		color.RGBA{R: 255, G: 87, B: 51, A: 255},
    		color.RGBA{R: 255, G: 168, B: 51, A: 255},
    		color.RGBA{R: 255, G: 217, B: 51, A: 255},
    		color.RGBA{R: 0, G: 0, B: 0, A: 255},
    		color.RGBA{R: 255, G: 255, B: 255, A: 255},
    	},
    }
)
