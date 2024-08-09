package config

const (
	Version      = "v0.1.6"
	OutputFolder = "Pictures/gowall"
	configFile   = "config.yml"
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
