package config

const (
	Version            = "v0.2.1"
	OutputFolder       = "Pictures/gowall"
	configFile         = "config.yml"
	OCRSchemaFile      = "schema.yml"
	WallOfTheDayUrl    = "https://www.reddit.com/r/wallpaper/top/"
	HexCodeVisualUrl   = "https://lawlesscreation.github.io/hex-color-visualiser/"
	UpscalerBinaryName = "realesrgan-ncnn-vulkan"
	EnvFilePath        = ".gowall/.env"
)

var (
	EnableImagePreviewingDefault = true
	InlineImagePreviewDefault    = false
	ImagePreviewBackend          = ""
	ThemesDefault                = []themeWrapper{}
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
