package providers

import (
	"context"
	"fmt"

	cf "github.com/Achno/gowall/config"
	"github.com/Achno/gowall/internal/logger"

	imageio "github.com/Achno/gowall/internal/image_io"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

// OpenAIProvider implements the OCRProvider interface
type OpenAIProvider struct {
	client *openai.Client
	model  string
	config Config
}

// NewOpenAIProvider creates a new OpenAI provider
func NewOpenAIProvider(config Config) (OCRProvider, error) {

	urlMap := map[string]string{
		"vllm":       "http://localhost:8000/v1",
		"openrouter": "https://openrouter.ai/api/v1",
		"openai":     "https://api.openai.com/v1",
		"oc":         cf.GowallConfig.EnvConfig.OPENAI_BASE_URL,
	}

	baseURL, ok := urlMap[config.OCR.Provider]
	if !ok {
		return nil, fmt.Errorf("%s is not a valid provider,use [vllm,openrouter,openai] or `oc` with the OPENAI_BASE_URL env", config.OCR.Provider)
	}

	apiMap := map[string]string{
		"vllm":       "x",
		"openrouter": cf.GowallConfig.EnvConfig.OPENROUTER_API_KEY,
		"openai":     cf.GowallConfig.EnvConfig.OPENAI_API_KEY,
		"oc":         cf.GowallConfig.EnvConfig.OPENAI_API_COMPATIBLE_SERVICE_API_KEY,
	}

	apiKey, ok := apiMap[config.OCR.Provider]
	if !ok {
		return nil, fmt.Errorf("%s is not a valid provider,use [vllm,openrouter,openai] or `oc` with the OPENAI_BASE_URL,OPENAI_API_COMPATIBLE_SERVICE_API_KEY envs", config.OCR.Provider)
	}

	if apiKey == "" {
		return nil, fmt.Errorf("the [OpenAI/OpenRouter or OpenAI Compatible] API key env is not set, check that your .env file location is correct inside config.yml and you are properly providing the env's")
	}
	retries := cf.GowallConfig.EnvConfig.OPENAI_MAX_RETRIES

	model := config.OCR.Model

	opts := []option.RequestOption{
		option.WithAPIKey(apiKey),
		option.WithBaseURL(baseURL),
		option.WithMaxRetries(retries),
	}

	client := openai.NewClient(opts...)

	return &OpenAIProvider{
		client: &client,
		model:  model,
		config: config,
	}, nil
}

// OCR OCRs a single image and returns an OCRResult
func (o *OpenAIProvider) OCR(ctx context.Context, input OCRInput) (*OCRResult, error) {

	messages, err := o.InputToMessages(input)
	if err != nil {
		return nil, err
	}

	chatCompletion, err := o.client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Messages: messages,
		Model:    o.model,
	})
	if err != nil {
		return nil, err
	}

	return &OCRResult{
		Text: chatCompletion.Choices[0].Message.Content,
		Metadata: map[string]string{
			"Model":    chatCompletion.Model,
			"RawJSON":  chatCompletion.JSON.Choices.Raw(),
			"RawJSON2": chatCompletion.Usage.PromptTokensDetails.RawJSON(),
		},
	}, nil
}

func (o *OpenAIProvider) Complete(ctx context.Context, text string) (string, error) {
	prompt := o.config.TextCorrection.Provider.Prompt

	messages := []openai.ChatCompletionMessageParamUnion{
		openai.UserMessage(prompt + "\n\n" + text),
	}

	chatCompletion, err := o.client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Messages: messages,
		Model:    o.model,
	})
	if err != nil {
		logger.Warnf("Error correcting text falling back to original: %v", err)
		return text, err
	}

	return chatCompletion.Choices[0].Message.Content, nil
}

func (o *OpenAIProvider) GetConfig() Config {
	return o.config
}

func (o *OpenAIProvider) SupportsPDF() bool {
	bMap := map[string]bool{
		"openai":     false,
		"openrouter": false,
		"vllm":       false,
		"oc":         o.config.OCR.SupportsPDF,
	}
	supported, ok := bMap[o.config.OCR.Provider]
	if !ok {
		return false
	}
	return supported
}

func (o *OpenAIProvider) WithImage(base64Image string, prompt string) openai.ChatCompletionMessageParamUnion {

	return openai.UserMessage([]openai.ChatCompletionContentPartUnionParam{
		openai.TextContentPart(prompt),
		openai.ImageContentPart(openai.ChatCompletionContentPartImageImageURLParam{
			URL:    fmt.Sprintf("data:image/jpeg;base64,%s", base64Image),
			Detail: "auto",
		}),
	})
}

func (o *OpenAIProvider) WithPDF(base64PDF string, prompt string) openai.ChatCompletionMessageParamUnion {
	return openai.UserMessage([]openai.ChatCompletionContentPartUnionParam{
		openai.TextContentPart(prompt),
		openai.FileContentPart(openai.ChatCompletionContentPartFileFileParam{
			FileData: openai.String(fmt.Sprintf("data:application/pdf;base64,%s", base64PDF)),
		}),
	})
}

func (o *OpenAIProvider) InputToMessages(input OCRInput) ([]openai.ChatCompletionMessageParamUnion, error) {

	prompt := o.config.OCR.Prompt
	prompt = BuildPrompt(prompt, input.Filename, o.config.OCR.Format)

	switch input.Type {
	case InputTypeImage:
		base64, err := imageio.ImageToBase64(input.Image)
		if err != nil {
			return nil, err
		}
		return []openai.ChatCompletionMessageParamUnion{o.WithImage(base64, prompt)}, nil
	case InputTypePDF:
		// If the provider supports PDF's directly, just send the pdf
		if o.SupportsPDF() {
			base64, err := imageio.BytesToBase64(input.PDFData)
			if err != nil {
				return nil, err
			}
			return []openai.ChatCompletionMessageParamUnion{o.WithPDF(base64, prompt)}, nil
		}
		return nil, fmt.Errorf("the provider doesn't support PDF's directly")
	default:
		return nil, fmt.Errorf("unsupported input type: %v", input.Type)
	}
}
