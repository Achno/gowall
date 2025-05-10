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
	defaultOpenAIModel = "gpt-4-vision-preview"
)

// OpenAIProvider implements the OCRProvider interface
type OpenAIProvider struct {
	client *openai.Client
	model  string
	config Config
}

// NewOpenAIProvider creates a new OpenAI provider
func NewOpenAIProvider(config Config) (OCRProvider, error) {

	baseURL := "https://api.openai.com/v1"
	if config.VisionLLMProvider == "vllm" {
		//http://localhost:8000/v1
		baseURL = os.Getenv("OPENAI_BASE_URL")
	}

	apiKey := os.Getenv("OPENAI_API_KEY")
	if config.VisionLLMProvider == "vllm" {
		apiKey = "x"
	}
	if apiKey == "" {
		return nil, fmt.Errorf("OPENAI_API_KEY env is not set")
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

func (o *OpenAIProvider) OCR(ctx context.Context, image image.Image) (*OCRResult, error) {

	prompt := "Extract all text from this image."
	if o.config.VisionLLMPrompt != "" {
		prompt = o.config.VisionLLMPrompt
	}

	// Add output format instructions
	if o.config.EnableMarkdown {
		prompt += " Format the output in Markdown."
	} else if o.config.EnableJSON {
		prompt += " Format the output as JSON."
	}

	base64Image, err := imageToBase64(image)
	if err != nil {
		return nil, err
	}

	//TODO remove the official lib for this https://github.com/sashabaranov/go-openai/issues/596
	//TODO NVM found the answer here : https://github.com/openai/openai-go/issues/67

	ImgMsg := openai.UserMessage([]openai.ChatCompletionContentPartUnionParam{
		openai.ImageContentPart(openai.ChatCompletionContentPartImageImageURLParam{
			URL:    fmt.Sprintf("data:image/jpeg;base64,%s", base64Image),
			Detail: "auto",
		}),
	})

	chatCompletion, err := o.client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.UserMessage(o.config.VisionLLMPrompt),
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
