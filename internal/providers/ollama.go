package providers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"io"
	"net/http"
	"os"
	"strconv"
	"sync"

	"github.com/Achno/gowall/internal/logger"
	"github.com/Achno/gowall/utils"
)

// OllamaProvider implements the Provider Interface
type OllamaProvider struct {
	config Config
	host   string
}

// Ollama chat with images API request structure
type OllamaRequest struct {
	Model    string          `json:"model"`
	Messages []OllamaMessage `json:"messages"`
	Stream   bool            `json:"stream"`
}

type OllamaMessage struct {
	Role    string   `json:"role"`
	Content string   `json:"content"`
	Images  []string `json:"images,omitempty"` // Array of base64 encoded images
}

type OllamaResponse struct {
	Model     string `json:"model"`
	CreatedAt string `json:"created_at"`
	Message   struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	} `json:"message"`
	Done          bool  `json:"done"`
	TotalDuration int64 `json:"total_duration"`
}

func (o *OllamaProvider) OCR(ctx context.Context, img image.Image) (*OCRResult, error) {
	// Convert image to base64
	imgBase64, err := imageToBase64(img)
	if err != nil {
		return nil, fmt.Errorf("failed to convert image to base64: %w", err)
	}

	// Create request payload
	prompt := "Extract all text from this image."
	if o.config.VisionLLMPrompt != "" {
		prompt = o.config.VisionLLMPrompt
	}

	if o.config.EnableMarkdown {
		prompt += "the output format should be Markdown"
	}

	if o.config.EnableJSON {
		prompt += "the output format should be JSON"
	}

	req := OllamaRequest{
		Model: o.config.VisionLLMModel,
		Messages: []OllamaMessage{
			{
				Role:    "user",
				Content: prompt,
				Images:  []string{imgBase64},
			},
		},
		Stream: false,
	}

	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Send request to Ollama
	resp, err := http.Post(o.host+"/api/chat", "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to send request to Ollama: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("ollama API error (status %d): %s", resp.StatusCode, string(body))
	}

	// Parse response
	var ollamaResp OllamaResponse
	if err := json.NewDecoder(resp.Body).Decode(&ollamaResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Extract text and metadata
	result := &OCRResult{
		Text: ollamaResp.Message.Content,
		Metadata: map[string]string{
			"provider":       "ollama",
			"model":          o.config.VisionLLMModel,
			"total_duration": fmt.Sprintf("%d", ollamaResp.TotalDuration),
		},
	}

	return result, nil
}

func NewOllamaProvider(config Config) (OCRProvider, error) {
	host := os.Getenv("OLLAMA_HOST")
	if host == "" {
		host = "http://127.0.0.1:11434"
	}

	return &OllamaProvider{
		config: config,
		host:   host,
	}, nil
}

func (o *OllamaProvider) OCRBatchImages(ctx context.Context, images []image.Image) ([]*OCRResult, error) {
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
