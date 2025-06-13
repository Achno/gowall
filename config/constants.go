package config

const (
	Version            = "v0.2.1"
	OutputFolder       = "Pictures/gowall"
	configFile         = "config.yml"
	WallOfTheDayUrl    = "https://www.reddit.com/r/wallpaper/top/"
	HexCodeVisualUrl   = "https://lawlesscreation.github.io/hex-color-visualiser/"
	UpscalerBinaryName = "realesrgan-ncnn-vulkan"

	BackendNN      = "nn"
	BackendRBF     = "rbf"
	BackendShepard = "shepard"
)

var (
	EnableImagePreviewingDefault  = true
	InlineImagePreviewDefault     = false
	ImagePreviewBackend           = ""
	ColorCorrectionBackendDefault = BackendRBF
	ThemesDefault                 = []themeWrapper{}

	ShepardOptionsDefault = ShepardOptions{
		Nearest: 30,
		Power:   4.0,
	}
)

func defaultConfig() Options {
	return Options{
		EnableImagePreviewing:  EnableImagePreviewingDefault,
		InlineImagePreview:     InlineImagePreviewDefault,
		ImagePreviewBackend:    ImagePreviewBackend,
		ColorCorrectionBackend: ColorCorrectionBackendDefault,
		Themes:                 ThemesDefault,
		ShepardOptions:         ShepardOptionsDefault,
	}
}
