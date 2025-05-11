package providers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"image"
	"net/http"
	"os"
	"strings"
)

// OllamaProvider implements the Provider Interface
type MistralProvider struct {
	config  Config
	client  *http.Client
	baseURL string
	apiKey  string
}

// Mistral single Image OCR request
type MistralOcrResquest struct {
	Model              string             `json:"model"`
	ID                 string             `json:"id,omitempty"`
	Document           MistralOcrDocument `json:"document"`
	Pages              []int              `json:"pages,omitempty"`
	IncludeImageBase64 bool               `json:"include_image_base64"`
	ImageLimit         int                `json:"image_limit,omitempty"`
	ImageMinSize       int                `json:"image_min_size,omitempty"`
}

type MistralOcrDocument struct {
	ImageURL string `json:"image_url"`
	Type     string `json:"type"`
}

// Mistral single OCR response
type MistralOcrResponse struct {
	Pages     []MistralOcrPage `json:"pages"`
	Model     string           `json:"model"`
	UsageInfo map[string]any   `json:"usage_info"`
}

type MistralOcrPage struct {
	Index      int                  `json:"index"`
	Markdown   string               `json:"markdown"`
	Images     []MistralOcrImage    `json:"images"`
	Dimensions MistralOcrDimensions `json:"dimensions"`
}

type MistralOcrImage struct {
	ID           string `json:"id"`
	TopLeftX     int    `json:"top_left_x"`
	TopLeftY     int    `json:"top_left_y"`
	BottomRightX int    `json:"bottom_right_x"`
	BottomLeftX  int    `json:"bottom_left_x"`
	ImageBase64  string `json:"image_base64"`
}

type MistralOcrDimensions struct {
	DPI    int `json:"dpi"`
	Height int `json:"height"`
	Width  int `json:"width"`
}

func (m *MistralProvider) OCR(ctx context.Context, img image.Image) (*OCRResult, error) {

	model := "mistral-ocr-latest"
	if m.config.VisionLLMModel != "" {
		model = m.config.VisionLLMModel
	}

	base64Image, err := imageToBase64(img)
	if err != nil {
		return nil, err
	}

	payload := MistralOcrResquest{
		Model:              model,
		Document:           MistralOcrDocument{Type: "image_url", ImageURL: fmt.Sprintf("data:image/jpeg;base64,%s", base64Image)},
		IncludeImageBase64: true,
	}

	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to JSON-encode payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, m.baseURL+"/ocr", bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+m.apiKey)

	res, err := m.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request error: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %s", res.Status)
	}

	var respData MistralOcrResponse
	if err := json.NewDecoder(res.Body).Decode(&respData); err != nil {
		return nil, fmt.Errorf("failed to decode response JSON: %w", err)
	}

	return MistralToOCRResult(&respData)
}

// MistralToOCRResult maps the external MistralOcrResponse into our internal OCRResult
func MistralToOCRResult(res *MistralOcrResponse) (*OCRResult, error) {
	if res == nil {
		return nil, fmt.Errorf("MistralToOCRResult: input response is nil")
	}

	var textParts []string
	var allImages []MistralOcrImage
	for _, page := range res.Pages {
		textParts = append(textParts, page.Markdown)
		allImages = append(allImages, page.Images...)
	}
	combinedText := strings.Join(textParts, "\n\n")

	meta := make(map[string]string, len(res.UsageInfo)+1)
	meta["model"] = res.Model
	for k, v := range res.UsageInfo {
		meta[k] = fmt.Sprint(v)
	}

	return &OCRResult{
		Text: combinedText,
		Images: OCRImage{
			MistralImages: allImages,
		},
		Metadata: meta,
	}, nil
}

func NewMistralProvider(config Config) (OCRProvider, error) {

	apiKey := os.Getenv("MISTRAL_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("MISTRAL_API_KEY env is not set")
	}

	return &MistralProvider{
		config:  config,
		client:  &http.Client{},
		baseURL: "https://api.mistral.ai/v1",
		apiKey:  apiKey,
	}, nil
}
