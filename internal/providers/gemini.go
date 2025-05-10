package providers

import (
	"context"
	"fmt"
	"image"
	"os"

	// "os"

	"google.golang.org/genai"
)

type GeminiProvider struct {
	config Config
	client *genai.Client
}

func NewGeminiProvider(config Config) (OCRProvider, error) {

	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("GEMINI_API_KEY env is not set")
	}
	client, err := genai.NewClient(context.Background(), &genai.ClientConfig{
		APIKey:  apiKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		return nil, fmt.Errorf("while creating client : %w", err)
	}

	return &GeminiProvider{
		client: client,
		config: config,
	}, nil
}

func (g *GeminiProvider) OCR(ctx context.Context, img image.Image) (*OCRResult, error) {

	bytes, err := imageToBytes(img)
	if err != nil {
		return nil, fmt.Errorf("while converting img to base64 : %w", err)
	}

	prompt := "Extract all text from this image. Return only the extracted text without any additional descriptions or explanations"
	if g.config.VisionLLMPrompt != "" {
		prompt = g.config.VisionLLMPrompt
	}

	model := "gemini-2.0-flash"
	if g.config.VisionLLMModel != "" {
		model = g.config.VisionLLMModel
	}

	// Create prompt
	parts := []*genai.Part{
		{Text: prompt},
		{InlineData: &genai.Blob{Data: bytes, MIMEType: "image/jpeg"}},
	}

	res, err := g.client.Models.GenerateContent(ctx, model, []*genai.Content{{Parts: parts}}, nil)
	if err != nil {
		return nil, fmt.Errorf("error from Gemini API: %w", err)
	}

	return &OCRResult{
		Text: res.Text(),
		Metadata: map[string]string{
			"tokens":      string(res.UsageMetadata.PromptTokenCount),
			"model":       res.ModelVersion,
			"codexresult": res.CodeExecutionResult(),
		},
	}, nil
}
