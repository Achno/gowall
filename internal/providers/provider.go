package providers

import (
	"context"
	"fmt"
	"strings"
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
		return "", fmt.Errorf("provider does not support text processing")
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

	provider, ok := providers[config.OCR.Provider]
	if !ok {
		return nil, fmt.Errorf("you have not entered a valid provider")
	}

	return provider(config)
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
