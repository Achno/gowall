package providers

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	cf "github.com/Achno/gowall/config"
	imageio "github.com/Achno/gowall/internal/image_io"
	"github.com/Achno/gowall/internal/logger"
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

type DoclingErrorResponse struct {
	Detail string `json:"detail"`
	Error  string `json:"error"`
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

// DoclingProcessPayload contains the data needed to process a file with docling
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
		logger.Printf("Docling CLI found, using CLI instead of REST API")
		return provider, nil
	}

	if err := provider.Client.HealthCheck(context.Background()); err != nil {
		return nil, fmt.Errorf("docling health check failed, ensure the server is running and if you don't use port 5001 specify the DOCLING_BASE_URL env: %w", err)
	}

	return provider, nil
}

func (p *DoclingProvider) OCR(ctx context.Context, input OCRInput) (*OCRResult, error) {
	ocrEngine := defaultDoclingOCREngine
	if p.Config.VisionLLMModel != "" {
		ocrEngine = p.Config.VisionLLMModel
	}

	var payload *DoclingProcessPayload
	var err error

	switch input.Type {
	case InputTypeImage:
		payload, err = p.WithImage(ctx, input)
		if err != nil {
			return nil, fmt.Errorf("docling: failed to prepare image: %w", err)
		}
	case InputTypePDF:
		if !p.SupportsPDF() {
			return nil, fmt.Errorf("docling provider doesn't support PDF processing")
		}
		payload, err = p.WithPDF(ctx, input)
		if err != nil {
			return nil, fmt.Errorf("docling: failed to prepare PDF: %w", err)
		}
	default:
		return nil, fmt.Errorf("unsupported input type: %v", input.Type)
	}

	// Process the payload using CLI or REST API
	var result *DoclingConvertDocumentResponse
	if p.CliClient.IsAvailable() {
		// Create temporary output directory for CLI
		tempOutputDir, err := os.MkdirTemp("", "docling-output-*")
		if err != nil {
			return nil, fmt.Errorf("failed to create temp output directory: %w", err)
		}
		defer os.RemoveAll(tempOutputDir)

		result, err = p.CliClient.ProcessFile(ctx, payload.FileBytes, payload.Filename, payload.Options, tempOutputDir)
		if err != nil {
			return nil, fmt.Errorf("docling CLI processing failed: %w", err)
		}
	} else {
		result, err = p.Client.ProcessFile(ctx, payload.FileBytes, payload.Filename, payload.Options)
		if err != nil {
			return nil, fmt.Errorf("docling REST processing failed: %w", err)
		}
	}

	text := ""
	if p.Config.Format == "markdown" {
		text = result.Document.MDContent
	} else {
		text = result.Document.TextContent
	}

	return &OCRResult{
		Text: text,
		Metadata: map[string]string{
			"DoclingOCREngine": ocrEngine,
			"DoclingStatus":    result.Status,
		},
	}, nil
}

func (p *DoclingProvider) WithImage(ctx context.Context, input OCRInput) (*DoclingProcessPayload, error) {
	imageBytes, err := imageio.ImageToBytes(input.Image)
	if err != nil {
		return nil, fmt.Errorf("failed to convert image to bytes: %w", err)
	}

	options := mergeConfigOptions(p.getDefaultOptions(), p.Config)

	// Remove 'from_formats' to let docling auto-detect the input format
	delete(options, "from_formats")

	if p.Config.Format == "markdown" {
		options["to_formats"] = "md"
	} else {
		options["to_formats"] = "text"
	}

	return &DoclingProcessPayload{
		FileBytes: imageBytes,
		Filename:  "image.jpg",
		Options:   options,
	}, nil
}

func (p *DoclingProvider) WithPDF(ctx context.Context, input OCRInput) (*DoclingProcessPayload, error) {
	if input.PDFData == nil {
		return nil, fmt.Errorf("PDF data is nil")
	}

	options := mergeConfigOptions(p.getDefaultOptions(), p.Config)

	// Remove 'from_formats' to let docling auto-detect the input format
	delete(options, "from_formats")

	if p.Config.Format == "markdown" {
		options["to_formats"] = "md"
	} else {
		options["to_formats"] = "text"
	}

	return &DoclingProcessPayload{
		FileBytes: input.PDFData,
		Filename:  "document.pdf",
		Options:   options,
	}, nil
}

func (p *DoclingProvider) SupportsPDF() bool {
	return true
}

func (p *DoclingProvider) GetConfig() Config {
	return p.Config
}

// getDefaultOptions returns the default Docling options map
func (p *DoclingProvider) getDefaultOptions() map[string]string {
	options := map[string]string{
		"do_ocr":    "true",
		"force_ocr": "true",
	}
	return options
}
