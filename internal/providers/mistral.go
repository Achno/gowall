package providers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	cf "github.com/Achno/gowall/config"
)

const (
	model = "mistral-ocr-latest"
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
	ImageURL     string `json:"image_url,omitempty"`
	DocumentURL  string `json:"document_url,omitempty"`
	DocumentName string `json:"document_name,omitempty"`
	Type         string `json:"type"`
}

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

type MistralBatchJob struct {
	ID                string `json:"id"`
	Status            string `json:"status"`
	TotalRequests     int    `json:"total_requests"`
	FailedRequests    int    `json:"failed_requests"`
	SucceededRequests int    `json:"succeeded_requests"`
	OutputFile        string `json:"output_file"`
}

type MistralBatchFile struct {
	ID string `json:"id"`
}

func NewMistralProvider(config Config) (OCRProvider, error) {

	apiKey := cf.GowallConfig.EnvConfig.MISTRAL_API_KEY
	if apiKey == "" {
		return nil, fmt.Errorf("MISTRAL_API_KEY env is not set,check that your .env file location is correct inside config.yml or you are properly providing the env's")
	}

	return &MistralProvider{
		config:  config,
		client:  &http.Client{},
		baseURL: "https://api.mistral.ai/v1",
		apiKey:  apiKey,
	}, nil
}

func (m *MistralProvider) OCR(ctx context.Context, input OCRInput) (*OCRResult, error) {

	payload, err := m.InputToMessages(input)
	if err != nil {
		return nil, fmt.Errorf("failed to convert input to messages: %w", err)
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

func (m *MistralProvider) SupportsPDF() bool {
	return true
}

func (m *MistralProvider) InputToMessages(input OCRInput) (MistralOcrResquest, error) {

	model := model
	if m.config.VisionLLMModel != "" {
		model = m.config.VisionLLMModel
	}

	switch input.Type {
	case InputTypeImage:
		base64Image, err := imageToBase64(input.Image)
		if err != nil {
			return MistralOcrResquest{}, err
		}
		return MistralOcrResquest{
			Model:              model,
			Document:           MistralOcrDocument{Type: "image_url", ImageURL: fmt.Sprintf("data:image/jpeg;base64,%s", base64Image)},
			IncludeImageBase64: true,
		}, nil
	case InputTypePDF:
		if m.SupportsPDF() {
			base64PDF, err := bytesToBase64(input.PDFData)
			if err != nil {
				return MistralOcrResquest{}, err
			}
			return MistralOcrResquest{
				Model:              model,
				Document:           MistralOcrDocument{Type: "document_url", DocumentURL: fmt.Sprintf("data:application/pdf;base64,%s", base64PDF), DocumentName: input.Filename},
				IncludeImageBase64: true,
			}, nil
		}
		return MistralOcrResquest{}, fmt.Errorf("MistralProvider does not support PDF's directly")
	default:
		return MistralOcrResquest{}, fmt.Errorf("unsupported input type: %v", input.Type)
	}

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
