package providers

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"image"
	"image/jpeg"
)

type OCRProvider interface {
	OCR(ctx context.Context, image image.Image) (*OCRResult, error)
}

// OCRResult holds the output from OCR processing
type OCRResult struct {
	Text     string // required
	HOCR     string // optional if provider supports it.
	Markdown string
	Metadata map[string]string
}

type Config struct {
	ProviderName string // "ollama,openai,mistral"
	OutputFormat string // txt,json,markdown

	VisionLLMProvider string
	VisionLLMModel    string
	VisionLLMPrompt   string

	// OCR output options
	EnableMarkdown bool
	EnableJSON     bool
}

func NewOCRProvider(config Config) (OCRProvider, error) {

	if config.VisionLLMModel == "" || config.VisionLLMProvider == "" {
		return nil, fmt.Errorf("missing OCR model,provider configuration")
	}

	providers := map[string]func(config Config) OCRProvider{
		"ollama": NewOllamaProvider,
		"vllm":   NewOpenAIProvider,
	}

	provider, ok := providers[config.ProviderName]
	if !ok {
		return nil, fmt.Errorf("you have not entered a valid provider")
	}

	return provider(config), nil
}

func imageToBase64(img image.Image) (string, error) {
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, nil); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(buf.Bytes()), nil
}
