package providers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	cf "github.com/Achno/gowall/config"
	imageio "github.com/Achno/gowall/internal/image_io"
)

// OllamaProvider implements the Provider Interface
type OllamaProvider struct {
	config Config
	host   string
}

type OllamaRequest struct {
	Model    string          `json:"model"`
	Messages []OllamaMessage `json:"messages"`
	Stream   bool            `json:"stream"`
	Options  map[string]any  `json:"options"`
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

func NewOllamaProvider(config Config) (OCRProvider, error) {
	host := cf.GowallConfig.EnvConfig.OLLAMA_HOST
	if host == "" {
		host = "http://127.0.0.1:11434"
	}

	return &OllamaProvider{
		config: config,
		host:   host,
	}, nil
}

func (o *OllamaProvider) OCR(ctx context.Context, input OCRInput) (*OCRResult, error) {
	imgBase64, err := imageio.ImageToBase64(input.Image)
	if err != nil {
		return nil, fmt.Errorf("failed to convert image to base64: %w", err)
	}

	prompt := o.config.OCR.Prompt
	prompt = BuildPrompt(prompt, input.Filename, o.config.OCR.Format)

	ollamaOptions, err := o.ApplyOptions()
	if err != nil {
		return nil, fmt.Errorf("failed to apply options: %w", err)
	}

	req := OllamaRequest{
		Model: o.config.OCR.Model,
		Messages: []OllamaMessage{
			{
				Role:    "user",
				Content: prompt,
				Images:  []string{imgBase64},
			},
		},
		Stream: false,
		Options: map[string]any{
			"num_predict": ollamaOptions.MaxTokens,
			"temperature": ollamaOptions.Temperature,
		},
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
			"model":          o.config.OCR.Model,
			"total_duration": fmt.Sprintf("%d", ollamaResp.TotalDuration),
		},
	}

	return result, nil
}

func (o *OllamaProvider) Complete(ctx context.Context, text string) (string, error) {
	prompt := o.config.OCR.Prompt

	ollamaOptions, err := o.ApplyOptions()
	if err != nil {
		return "", fmt.Errorf("failed to apply options: %w", err)
	}

	fmt.Println("ollamaOptions", ollamaOptions.MaxTokens, ollamaOptions.Temperature)
	fmt.Println("prompt", prompt)
	fmt.Println("text", text)
	req := OllamaRequest{
		Model: o.config.OCR.Model,
		Messages: []OllamaMessage{
			{
				Role:    "user",
				Content: prompt + "\n\n" + text,
			},
		},
		Stream: false,
		Options: map[string]any{
			"num_predict": ollamaOptions.MaxTokens,
			"temperature": ollamaOptions.Temperature,
		},
	}

	reqBody, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	// Send request to Ollama
	resp, err := http.Post(o.host+"/api/chat", "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		return "", fmt.Errorf("failed to send request to Ollama: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("ollama API error (status %d): %s", resp.StatusCode, string(body))
	}

	// Parse response
	var ollamaResp OllamaResponse
	if err := json.NewDecoder(resp.Body).Decode(&ollamaResp); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	return ollamaResp.Message.Content, nil
}

func (o *OllamaProvider) GetConfig() Config {
	return o.config
}

func (o *OllamaProvider) SupportsPDF() bool {
	return false
}

// ApplyOptions merges the schema.yml ollama specific options with the default options
func (o *OllamaProvider) ApplyOptions() (*OllamaOptions, error) {
	ollamaOptions := o.config.OllamaOptions
	if ollamaOptions == nil {
		ollamaOptions = &OllamaOptions{}
	}

	optionsRaw, err := ollamaOptions.Apply(o.getOllamaOptions(), o.config)
	if err != nil {
		return nil, fmt.Errorf("while merging ollama options: %w", err)
	}
	ollamaOptions, ok := optionsRaw.(*OllamaOptions)
	if !ok {
		return nil, fmt.Errorf("options are not a OllamaOptions")
	}

	return ollamaOptions, nil
}

func (o *OllamaProvider) getOllamaOptions() *OllamaOptions {
	return &OllamaOptions{
		MaxTokens:   4096,
		Temperature: 0.7,
	}
}
