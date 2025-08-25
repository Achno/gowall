package providers

import (
	"context"
	"fmt"
)

type OCRProvider interface {
	OCR(ctx context.Context, input OCRInput) (*OCRResult, error)
}

type TextProcessor interface {
	Complete(ctx context.Context, text string) (string, error)
}

type PDFCapable interface {
	SupportsPDF() bool
}

type Configurable interface {
	GetConfig() Config
}

// ProviderOptionsInterface allows each provider to define their own options struct
type ProviderOptionsInterface interface {
	// Apply merges default options with schema options and returns a config type the provider can use.
	Apply(defaults any, config Config) (any, error)
}

// ProviderService composes providers with their dependencies
type ProviderService struct {
	ocr                   OCRProvider
	rateLimiter           *RateLimiter
	correctionTextLimiter *RateLimiter
	config                Config
}

func NewProviderService(provider OCRProvider, config Config) *ProviderService {
	return &ProviderService{
		ocr:                   provider,
		rateLimiter:           NewRateLimiter(config.RateLimit),
		correctionTextLimiter: NewRateLimiter(config.TextCorrection.RateLimit),
		config:                config,
	}
}

func (s *ProviderService) OCR(ctx context.Context, input OCRInput) (*OCRResult, error) {
	if err := s.rateLimiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("rate limit wait failed: %w", err)
	}
	return s.ocr.OCR(ctx, input)
}

func (s *ProviderService) Complete(ctx context.Context, text string) (string, error) {
	textProcessor, ok := s.ocr.(TextProcessor)
	if !ok {
		return "", fmt.Errorf("provider doesn't support text completion")
	}

	if err := s.correctionTextLimiter.Wait(ctx); err != nil {
		return "", fmt.Errorf("rate limit wait failed: %w", err)
	}

	return textProcessor.Complete(ctx, text)
}

func (s *ProviderService) SupportsPDF() bool {
	if pdfCapable, ok := s.ocr.(PDFCapable); ok {
		return pdfCapable.SupportsPDF()
	}
	return false
}

func (s *ProviderService) GetConfig() Config {
	return s.config
}

func NewOCRProvider(config Config) (OCRProvider, error) {

	if config.OCR.Model == "" || config.OCR.Provider == "" {
		return nil, fmt.Errorf("didn't specify OCR model & provider")
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

	provider, ok := providers[config.OCR.Provider]
	if !ok {
		return nil, fmt.Errorf("you have not entered a valid provider")
	}

	return provider(config)
}
