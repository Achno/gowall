package providers

import (
	"context"
	"fmt"
	"image"
	"os"
	"strconv"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

const (
	defaultOpenAIModel = "gpt-4o"
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
		"oc":         os.Getenv("OPENAI_BASE_URL"),
	}

	baseURL, ok := urlMap[config.VisionLLMProvider]
	if !ok {
		return nil, fmt.Errorf("%s is not a valid provider,use [vllm,openrouter,openai] or `oc` alongside the OPENAI_BASE_URL env", config.VisionLLMProvider)
	}

	apiMap := map[string]string{
		"vllm":       "x",
		"openrouter": os.Getenv("OPENROUTER_API_KEY"),
		"openai":     os.Getenv("OPENAI_API_KEY"),
		"oc":         os.Getenv("OPENAI_API_COMPATIBLE_SERVICE_API_KEY"),
	}

	apiKey, ok := apiMap[config.VisionLLMProvider]
	if !ok {
		return nil, fmt.Errorf("%s is not a valid provider,use [vllm,openrouter,openai] or `oc` alongside the OPENAI_BASE_URL env", config.VisionLLMProvider)
	}

	if apiKey == "" {
		return nil, fmt.Errorf("your API key env is not set")
	}
	retriesStr := os.Getenv("OPENAI_MAX_RETRIES")
	var retries = 2

	if retriesStr != "" {
		retriesParsed, err := strconv.Atoi(retriesStr)
		if err != nil {
			return nil, fmt.Errorf("failed converting OPENAI_MAX_RETRIES to an int")
		}
		retries = retriesParsed
	}

	model := defaultOpenAIModel
	if config.VisionLLMModel != "" {
		model = config.VisionLLMModel
	}

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
func (o *OpenAIProvider) OCR(ctx context.Context, image image.Image) (*OCRResult, error) {

	prompt := "Extract all text from this image."
	if o.config.VisionLLMPrompt != "" {
		prompt = o.config.VisionLLMPrompt
	}

	// Add output format instructions
	if o.config.EnableMarkdown {
		prompt += " Format the output in Markdown."
	}

	base64Image, err := imageToBase64(image)
	if err != nil {
		return nil, err
	}

	ImgMsg := openai.UserMessage([]openai.ChatCompletionContentPartUnionParam{
		openai.TextContentPart(prompt),
		openai.ImageContentPart(openai.ChatCompletionContentPartImageImageURLParam{
			URL:    fmt.Sprintf("data:image/jpeg;base64,%s", base64Image),
			Detail: "auto",
		}),
	})

	chatCompletion, err := o.client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Messages: []openai.ChatCompletionMessageParamUnion{
			ImgMsg,
		},
		Model: o.model,
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

func (o *OpenAIProvider) OCRBatchImages(ctx context.Context, images []image.Image) ([]*OCRResult, error) {
	return processBatchConcurrently(ctx, images, o.OCR, "openai")
}
