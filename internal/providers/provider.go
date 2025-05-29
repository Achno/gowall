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
	OCRBatchImages(ctx context.Context, images []image.Image) ([]*OCRResult, error)
}

// OCRResult holds the output from OCR processing
type OCRResult struct {
	Text     string // required
	Images   OCRImage
	Metadata map[string]string
}

type OCRImage struct {
	MistralImages []MistralOcrImage
}
type Config struct {
	VisionLLMProvider string // "ollama,openai,mistral,vllm"
	VisionLLMModel    string
	VisionLLMPrompt   string

	Language string

	// OCR output options
	EnableMarkdown bool
	EnableHOCR     bool // only for providers that support it like tesseract
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

func imageToBase64(img image.Image) (string, error) {
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, nil); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(buf.Bytes()), nil
}

func imageToBytes(img image.Image) ([]byte, error) {
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, nil); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
