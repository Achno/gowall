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
	textProcessor         TextProcessor
	rateLimiter           *RateLimiter
	correctionTextLimiter *RateLimiter
	config                Config
}

func NewProviderService(provider OCRProvider, config Config) *ProviderService {
	service := &ProviderService{
		ocr:                   provider,
		rateLimiter:           NewRateLimiter(config.RateLimit),
		correctionTextLimiter: NewRateLimiter(config.TextCorrection.RateLimit),
		config:                config,
	}

	// Create separate text correction provider if enabled
	if config.TextCorrection.Enabled {
		if textCorrectionProvider, err := NewTextCorrectionProvider(config); err == nil {
			service.textProcessor = textCorrectionProvider
		}
	}

	return service
}

func (s *ProviderService) OCR(ctx context.Context, input OCRInput) (*OCRResult, error) {
	if err := s.rateLimiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("rate limit wait failed: %w", err)
	}
	return s.ocr.OCR(ctx, input)
}

func (s *ProviderService) Complete(ctx context.Context, text string) (string, error) {
	if s.textProcessor == nil {
		return "", fmt.Errorf("provider doesn't support text completion")
	}

	if err := s.correctionTextLimiter.Wait(ctx); err != nil {
		return "", fmt.Errorf("rate limit wait failed: %w", err)
	}

	return s.textProcessor.Complete(ctx, text)
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

// getProviderFactories returns the map of available provider factories
func getProviderFactories() map[string]func(config Config) (OCRProvider, error) {
	return map[string]func(config Config) (OCRProvider, error){
		"ollama":     NewOllamaProvider,
		"vllm":       NewOpenAIProvider,
		"openai":     NewOpenAIProvider,
		"gemini":     NewGeminiProvider,
		"mistral":    NewMistralProvider,
		"openrouter": NewOpenAIProvider,
		"tesseract":  NewTesseractProvider,
		"docling":    NewDoclingProvider,
	}
}

func NewProvider(providerName, model string, config Config) (OCRProvider, error) {
	if model == "" || providerName == "" {
		return nil, fmt.Errorf("didn't specify model & provider")
	}

	providers := getProviderFactories()
	providerFunc, ok := providers[providerName]
	if !ok {
		return nil, fmt.Errorf("provider '%s' is not supported", providerName)
	}

	return providerFunc(config)
}

func NewOCRProvider(config Config) (OCRProvider, error) {
	return NewProvider(config.OCR.Provider, config.OCR.Model, config)
}

func NewTextCorrectionProvider(config Config) (TextProcessor, error) {
	textCorrectionConfig := createTextCorrectionConfig(config)

	provider, err := NewProvider(
		config.TextCorrection.Provider.Provider,
		config.TextCorrection.Provider.Model,
		textCorrectionConfig,
	)
	if err != nil {
		return nil, fmt.Errorf("while creating completionprovider: %w", err)
	}

	textProcessor, ok := provider.(TextProcessor)
	if !ok {
		return nil, fmt.Errorf("provider '%s' doesn't support text completion", config.TextCorrection.Provider.Provider)
	}

	return textProcessor, nil
}

func createTextCorrectionConfig(config Config) Config {
	textCorrectionConfig := config

	textCorrectionConfig.OCR = ProviderConfig{
		Provider:    config.TextCorrection.Provider.Provider,
		Model:       config.TextCorrection.Provider.Model,
		Prompt:      config.TextCorrection.Provider.Prompt,
		Language:    config.TextCorrection.Provider.Language,
		Format:      config.TextCorrection.Provider.Format,
		SupportsPDF: config.TextCorrection.Provider.SupportsPDF,
	}

	textCorrectionConfig.RateLimit = config.TextCorrection.RateLimit

	return textCorrectionConfig
}
