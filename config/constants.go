package config

const (
	Version          = "v0.1.7"
	OutputFolder     = "Pictures/gowall"
	configFile       = "config.yml"
	WallOfTheDayUrl  = "https://www.reddit.com/r/wallpaper/top/"
	HexCodeVisualUrl = "https://lawlesscreation.github.io/hex-color-visualiser/"
)

var (
	EnableImagePreviewingDefault = true
	ThemesDefault                = []themeWrapper{}
)

func defaultConfig() Options {
	return Options{
		EnableImagePreviewing: EnableImagePreviewingDefault,
		Themes:                ThemesDefault,
	}
}
