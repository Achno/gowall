package providers

import (
	"context"
	"fmt"
	"image"
	"strings"
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

type OCRProvider interface {
	OCR(ctx context.Context, input OCRInput) (*OCRResult, error)
	GetConfig() Config
}

type TextProcessorProvider interface {
	Complete(ctx context.Context, text string) (string, error)
}

type PDFCapable interface {
	SupportsPDF() bool
}

type RateLimited interface {
	SetRateLimit(rps float64, burst int)
}

// ProviderOptionsInterface allows each provider to define their own options struct
type ProviderOptionsInterface interface {
	// Apply merges default options with schema options and returns a config type the provider can use.
	Apply(defaults any, config Config) (any, error)
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

	// Pipeline settings
	DPI         float64 `yaml:"dpi"` // DPI affects the image resolution in pdf->images conversion
	SupportsPDF bool    `yaml:"supports_pdf"`

	// Used in the post-processing pipeline for the text correction step
	TextCorrectionEnabled  bool    `yaml:"text_correction_enabled"`
	TextCorrectionProvider string  `yaml:"text_correction_provider"`
	TextCorrectionModel    string  `yaml:"text_correction_model"`
	TextCorrectionPrompt   string  `yaml:"text_correction_prompt"`
	TextCorrectionRPS      float64 `yaml:"text_correction_rps"`
	TextCorrectionBurst    int     `yaml:"text_correction_burst"`

	// Provider-specific options
	DoclingOptions *DoclingOptions `yaml:"docling_options,omitempty"`
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

func NewTextCorrectionProvider(config Config) (TextProcessorProvider, error) {
	if !config.TextCorrectionEnabled {
		return nil, fmt.Errorf("text correction is not enabled")
	}

	if config.TextCorrectionModel == "" || config.TextCorrectionProvider == "" {
		return nil, fmt.Errorf("missing text correction model and provider")
	}

	textCorrectionConfig := Config{
		VisionLLMProvider: config.TextCorrectionProvider,
		VisionLLMModel:    config.TextCorrectionModel,
		VisionLLMPrompt:   config.TextCorrectionPrompt,
		RateLimitRPS:      config.TextCorrectionRPS,
		RateLimitBurst:    config.TextCorrectionBurst,
	}

	providers := map[string]func(config Config) (OCRProvider, error){
		"ollama":     NewOllamaProvider,
		"vllm":       NewOpenAIProvider,
		"openai":     NewOpenAIProvider,
		"gemini":     NewGeminiProvider,
		"mistral":    NewMistralProvider,
		"openrouter": NewOpenAIProvider,
	}

	providerFunc, ok := providers[config.TextCorrectionProvider]
	if !ok {
		return nil, fmt.Errorf("you have not entered a valid text correction provider")
	}

	baseProvider, err := providerFunc(textCorrectionConfig)
	if err != nil {
		return nil, err
	}

	//! careful with the casting
	if config.TextCorrectionRPS > 0 {
		rateLimitedProvider := WithRateLimit(baseProvider, config.TextCorrectionRPS, config.TextCorrectionBurst)
		if textProcessor, ok := rateLimitedProvider.(TextProcessorProvider); ok {
			return textProcessor, nil
		}
		return nil, fmt.Errorf("rate-limited provider does not implement TextProcessorProvider interface")
	}

	if textProcessor, ok := baseProvider.(TextProcessorProvider); ok {
		return textProcessor, nil
	}

	return nil, fmt.Errorf("provider does not implement TextProcessorProvider interface")
}

// AddPageContextToPrompt enhances the prompt with page-specific context for multi-page documents.
// It extracts page information from filenames with format "document.pdf-page-2-of-5" or "document.pdf-page-2"
// and adds appropriate context to help the OCR provider understand the document structure and place headings.
func AddPageContextToPrompt(filename, originalPrompt string) string {
	prompt := originalPrompt

	if !strings.Contains(filename, "-page-") {
		return prompt
	}

	// Extract page info from filename like "document.pdf-page-2-of-5"
	parts := strings.Split(filename, "-page-")
	if len(parts) != 2 {
		return prompt
	}

	pageInfo := parts[1] // "2-of-5" or just "2"

	var pageNum, totalPages string
	if strings.Contains(pageInfo, "-of-") {
		pageParts := strings.Split(pageInfo, "-of-")
		pageNum = pageParts[0]
		totalPages = pageParts[1]
	} else {
		pageNum = pageInfo
	}

	if pageNum == "1" {
		if totalPages != "" {
			prompt += fmt.Sprintf(" This is the FIRST PAGE of a %s-page document. Use top-level headings (# and ##) as appropriate for a document beginning.", totalPages)
		} else {
			prompt += " This is the FIRST PAGE of a multi-page document. Use top-level headings (# and ##) as appropriate for a document beginning."
		}
	} else {
		if totalPages != "" {
			prompt += fmt.Sprintf(" This is PAGE %s of %s total pages (NOT the first page). Assume this document has already started with main headings on previous pages. Use continuation-level headings (##) unless you see clear evidence this page starts a major new section.", pageNum, totalPages)
		} else {
			prompt += fmt.Sprintf(" This is PAGE %s of a multi-page document (NOT the first page). Assume this document has already started with main headings on previous pages. Use continuation-level headings (##) unless you see clear evidence this page starts a major new section.", pageNum)
		}
	}

	return prompt
}
