package providers

import (
	"context"
	"fmt"

	cf "github.com/Achno/gowall/config"
	imageio "github.com/Achno/gowall/internal/image_io"
	"google.golang.org/genai"
)

const (
	geminimodel = "gemini-2.0-flash"
)

type GeminiProvider struct {
	config Config
	client *genai.Client
}

func NewGeminiProvider(config Config) (OCRProvider, error) {

	apiKey := cf.GowallConfig.EnvConfig.GEMINI_API_KEY
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

func (g *GeminiProvider) OCR(ctx context.Context, input OCRInput) (*OCRResult, error) {

	bytes, err := imageio.ImageToBytes(input.Image)
	if err != nil {
		return nil, fmt.Errorf("while converting img to base64 : %w", err)
	}

	prompt := "Extract all text from this image. Return only the extracted text without any additional descriptions or explanations"
	if g.config.VisionLLMPrompt != "" {
		prompt = g.config.VisionLLMPrompt
	}

	model := geminimodel
	if g.config.VisionLLMModel != "" {
		model = g.config.VisionLLMModel
	}

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
