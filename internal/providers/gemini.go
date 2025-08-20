package providers

import (
	"context"
	"fmt"

	cf "github.com/Achno/gowall/config"
	imageio "github.com/Achno/gowall/internal/image_io"
	"google.golang.org/genai"
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

	parts, err := g.InputToMessages(input)
	if err != nil {
		return nil, err
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

func (g *GeminiProvider) GetConfig() Config {
	return g.config
}

func (g *GeminiProvider) InputToMessages(input OCRInput) ([]*genai.Part, error) {
	prompt := g.config.OCR.Prompt

	if g.config.OCR.Format == "md" {
		prompt += " Format the output in Markdown."
		prompt = AddPageContextToPrompt(input.Filename, prompt)
	}

	prompt += " Format the output in plain text"

	switch input.Type {
	case InputTypeImage:
		bytes, err := imageio.ImageToBytes(input.Image)
		if err != nil {
			return nil, fmt.Errorf("while converting img to bytes : %w", err)
		}
		return []*genai.Part{
			{Text: prompt},
			{InlineData: &genai.Blob{Data: bytes, MIMEType: "image/jpeg"}},
		}, nil
	case InputTypePDF:
		return nil, fmt.Errorf("PDF input not supported by Gemini provider")
	default:
		return nil, fmt.Errorf("unsupported input type: %v", input.Type)
	}
}
