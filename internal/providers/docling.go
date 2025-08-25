package providers

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	cf "github.com/Achno/gowall/config"
	imageio "github.com/Achno/gowall/internal/image_io"
)

const (
	doclingDefaultBaseURL   = "http://localhost:5001"
	doclingHealthPath       = "/health"
	doclingConvertPath      = "/v1/convert/file"
	defaultDoclingOCREngine = "easyocr"
)

// DoclingProvider implements the OCRProvider interface
type DoclingProvider struct {
	Client    *DoclingClient
	CliClient *DoclingCliClient
	Config    Config
}

type DoclingHealthResponse struct {
	Status string `json:"status"`
}

type DoclingDocumentResponse struct {
	Filename       string          `json:"filename"`
	MDContent      string          `json:"md_content"`
	JSONContent    json.RawMessage `json:"json_content"`
	HTMLContent    string          `json:"html_content"`
	TextContent    string          `json:"text_content"`
	DocTagsContent string          `json:"doctags_content"`
}

type DoclingConvertDocumentResponse struct {
	Document       DoclingDocumentResponse `json:"document"`
	Status         string                  `json:"status"` // "success", "partial_success", "skipped", "failure"
	Errors         []any                   `json:"errors"`
	ProcessingTime float64                 `json:"processing_time"`
	Timings        map[string]any          `json:"timings"`
}

type DoclingProcessPayload struct {
	FileBytes []byte
	Filename  string
	Options   map[string]string
}

func NewDoclingProvider(config Config) (OCRProvider, error) {

	baseURL := cf.GowallConfig.EnvConfig.DOCLING_BASE_URL
	if baseURL == "" {
		baseURL = doclingDefaultBaseURL
	}
	baseURL = strings.TrimRight(baseURL, "/")

	client := NewDoclingClient(WithDoclingBaseURL(baseURL))
	cliClient := NewDoclingCliClient()

	provider := &DoclingProvider{
		Client:    client,
		CliClient: cliClient,
		Config:    config,
	}

	// Try CLI first, fallback to REST API if not found
	if cliClient.Available {
		return provider, nil
	}

	if err := provider.Client.HealthCheck(context.Background()); err != nil {
		return nil, fmt.Errorf("docling health check failed, ensure the server is running and if you don't use port 5001 specify the DOCLING_BASE_URL env: %w", err)
	}

	return provider, nil
}

func (p *DoclingProvider) OCR(ctx context.Context, input OCRInput) (*OCRResult, error) {

	payload, err := p.InputToMessages(input, ctx)
	if err != nil {
		return nil, fmt.Errorf("docling: failed to prepare input: %w", err)
	}

	// Process the payload using CLI or REST API
	var result *DoclingConvertDocumentResponse
	if p.CliClient.IsAvailable() {
		// Create temporary output directory for CLI
		tempOutputDir, err := os.MkdirTemp("", "docling-output-*")
		if err != nil {
			return nil, fmt.Errorf("docling: failed to create temp output directory: %w", err)
		}
		defer os.RemoveAll(tempOutputDir)

		result, err = p.CliClient.ProcessFile(ctx, payload.FileBytes, payload.Filename, payload.Options, tempOutputDir)
		if err != nil {
			return nil, fmt.Errorf("docling CLI: %w", err)
		}
	} else {
		result, err = p.Client.ProcessFile(ctx, payload.FileBytes, payload.Filename, payload.Options)
		if err != nil {
			return nil, fmt.Errorf("docling REST: %w", err)
		}
	}

	text := ""
	if p.Config.OCR.Format == "md" {
		text = result.Document.MDContent
	} else {
		text = result.Document.TextContent
	}

	return &OCRResult{
		Text: text,
		Metadata: map[string]string{
			"DoclingStatus": result.Status,
		},
	}, nil
}

func (p *DoclingProvider) WithImage(ctx context.Context, input OCRInput) (*DoclingProcessPayload, error) {
	imageBytes, err := imageio.ImageToBytes(input.Image)
	if err != nil {
		return nil, fmt.Errorf("failed to convert image to bytes: %w", err)
	}

	// Safely handle nil DoclingOptions
	var doclingOptions *DoclingOptions
	if p.Config.DoclingOptions != nil {
		doclingOptions = p.Config.DoclingOptions
	} else {
		doclingOptions = &DoclingOptions{}
	}

	options, err := doclingOptions.Apply(p.getDefaultDoclingOptions(), p.Config)
	if err != nil {
		return nil, fmt.Errorf("failed to apply schema options: %w", err)
	}

	optionsMap, ok := options.(map[string]string)
	if !ok {
		return nil, fmt.Errorf("options are not a map[string]string")
	}

	return &DoclingProcessPayload{
		FileBytes: imageBytes,
		Filename:  "image.jpg",
		Options:   optionsMap,
	}, nil
}

func (p *DoclingProvider) WithPDF(ctx context.Context, input OCRInput) (*DoclingProcessPayload, error) {
	if input.PDFData == nil {
		return nil, fmt.Errorf("PDF data is nil")
	}

	var doclingOptions *DoclingOptions
	if p.Config.DoclingOptions != nil {
		doclingOptions = p.Config.DoclingOptions
	} else {
		doclingOptions = &DoclingOptions{}
	}

	options, err := doclingOptions.Apply(p.getDefaultDoclingOptions(), p.Config)
	if err != nil {
		return nil, fmt.Errorf("failed to apply schema options: %w", err)
	}

	optionsMap, ok := options.(map[string]string)
	if !ok {
		return nil, fmt.Errorf("options are not a map[string]string")
	}

	return &DoclingProcessPayload{
		FileBytes: input.PDFData,
		Filename:  "document.pdf",
		Options:   optionsMap,
	}, nil
}

func (p *DoclingProvider) InputToMessages(input OCRInput, ctx context.Context) (*DoclingProcessPayload, error) {

	switch input.Type {
	case InputTypeImage:
		payload, err := p.WithImage(ctx, input)
		if err != nil {
			return nil, fmt.Errorf("failed WithImage: %w", err)
		}
		return payload, nil
	case InputTypePDF:
		if !p.SupportsPDF() {
			return nil, fmt.Errorf("provider doesn't support PDF processing")
		}
		payload, err := p.WithPDF(ctx, input)
		if err != nil {
			return nil, fmt.Errorf("failed WithPDF: %w", err)
		}
		return payload, nil
	default:
		return nil, fmt.Errorf("unsupported input type: %v", input.Type)
	}
}

func (p *DoclingProvider) SupportsPDF() bool {
	return true
}

func (p *DoclingProvider) GetConfig() Config {
	return p.Config
}

func (p *DoclingProvider) getDefaultDoclingOptions() *DoclingOptions {
	// Using pointer to bool for 3 states: non-existent (nil), or a pointer to true/false value
	trueValue := true
	return &DoclingOptions{
		OCR:       &trueValue,
		Force_OCR: &trueValue,
	}
}
