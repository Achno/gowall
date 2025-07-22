package providers

import (
	"context"
	"fmt"
	"image"
	"maps"
)

type InputType int

const (
	InputTypeImage InputType = iota
	InputTypePDF
)

type OCRInput struct {
	Type     InputType
	Image    image.Image
	PDFData  []byte
	Filename string
}

type OCRResult struct {
	Text     string
	Images   OCRImage
	Metadata map[string]string
}

type OCRImage struct {
	MistralImages []MistralOcrImage
}

// PipelineItem tracks an item through the processing pipeline.
// It maps expanded items (like PDF pages) back to their original source file.
type PipelineItem struct {
	Input         *OCRInput
	OriginalIndex int
	PageIndex     int // -1 for non-PDF pages or single-page docs
}

// Core OCR interface (simplified)
type OCRProvider interface {
	OCR(ctx context.Context, input OCRInput) (*OCRResult, error)
	GetConfig() Config
}

// Capability interfaces
type PDFCapable interface {
	SupportsPDF() bool
}

type RateLimited interface {
	SetRateLimit(rps float64, burst int)
}

// Configuration for providers
type Config struct {
	VisionLLMProvider string `yaml:"provider"` // "openai", "openrouter", "mistral", "vllm" ...
	VisionLLMModel    string `yaml:"model"`
	VisionLLMPrompt   string `yaml:"prompt"`
	Language          string `yaml:"language"` // depends on provider,llms don't need it, docling & tesseract do

	Format string `yaml:"format"` // "markdown", "text"

	// Concurrency and Rate limiting
	Concurrency    int     `yaml:"concurrency"` // Worker pool size
	RateLimitRPS   float64 `yaml:"rps"`         // requests per second
	RateLimitBurst int     `yaml:"burst"`       // burst size

	// Provider-specific settings
	DPI         float64 `yaml:"dpi"` // DPI affects the image resolution in pdf->images conversion
	SupportsPDF bool    `yaml:"supports_pdf"`

	// Provider-specific options
	ProviderOptions map[string]string `yaml:"provider_options"`
}

func NewOCRProvider(config Config) (OCRProvider, error) {

	if config.VisionLLMModel == "" || config.VisionLLMProvider == "" {
		return nil, fmt.Errorf("missing OCR model,provider configuration")
	}

	providers := map[string]func(config Config) (OCRProvider, error){
		"ollama":     NewOllamaProvider,
		"vllm":       NewOpenAIProvider,
		"openai":     NewOpenAIProvider,
		"gemini":     NewGeminiProvider,
		"mistral":    NewMistralProvider,
		"openrouter": NewOpenAIProvider,
		"tesseract":  NewTesseractProvider,
		"docling":    NewDoclingProvider,
	}

	provider, ok := providers[config.VisionLLMProvider]
	if !ok {
		return nil, fmt.Errorf("you have not entered a valid provider")
	}

	return provider(config)
}

// mergeConfigOptions merges default options with schema-provided options
func mergeConfigOptions(defaults map[string]string, config Config) map[string]string {
	merged := make(map[string]string)

	// Copy defaults
	maps.Copy(merged, defaults)

	// Override with schema options
	if config.ProviderOptions != nil {
		maps.Copy(merged, config.ProviderOptions)
	}

	return merged
}
