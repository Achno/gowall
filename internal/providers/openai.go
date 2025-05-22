package providers

import (
	"context"
	"errors"
	"fmt"
	"image"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/Achno/gowall/internal/logger"
	"github.com/Achno/gowall/utils"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

const (
	defaultOpenAIModel = "gpt-4o"
	batchInterval      = 2 * time.Second
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
		return nil, fmt.Errorf("%s v", config.VisionLLMProvider)
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
	} else if o.config.EnableJSON {
		prompt += " Format the output as JSON."
	}

	base64Image, err := imageToBase64(image)
	if err != nil {
		return nil, err
	}

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

// OCRBatchImages OCRs a batch of images in parallel and returns a slice of OCRResults or a joint error
func (o *OpenAIProvider) OCRBatchImages(ctx context.Context, images []image.Image) ([]*OCRResult, error) {

	wg := sync.WaitGroup{}
	results := make([]*OCRResult, len(images))
	errChan := make(chan error, len(images))

	for i, img := range images {
		wg.Add(1)
		go func(i int, img image.Image) {
			defer wg.Done()
			result, err := o.OCR(ctx, img)

			if err != nil {
				errChan <- err
			}
			results[i] = result
			logger.Print(utils.BlueColor + " ➜ OCR Batch Image " + strconv.Itoa(i) + " completed" + utils.ResetColor)
		}(i, img)
	}

	wg.Wait()
	close(errChan)

	if len(errChan) > 0 {
		var errs []error

		for err := range errChan {
			errs = append(errs, err)
		}

		return results, errors.New(utils.FormatErrors(errs))
	}

	return results, nil
}
