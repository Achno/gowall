package providers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
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
	Client *DoclingClient
	Config Config
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

func NewDoclingProvider(config Config) (OCRProvider, error) {

	baseURL := cf.GowallConfig.EnvConfig.DOCLING_BASE_URL
	if baseURL == "" {
		baseURL = doclingDefaultBaseURL
	}
	baseURL = strings.TrimRight(baseURL, "/")

	client := NewDoclingClient(WithDoclingBaseURL(baseURL))

	provider := &DoclingProvider{
		Client: client,
		Config: config,
	}

	if err := provider.Client.HealthCheck(context.Background()); err != nil {
		return nil, fmt.Errorf("docling health check failed, ensure the server is running and if you don't use port 5001 specify the DOCLING_BASE_URL env: %w", err)
	}

	return provider, nil
}

func (p *DoclingProvider) OCR(ctx context.Context, input OCRInput) (*OCRResult, error) {
	ocrEngine := defaultDoclingOCREngine
	if env := os.Getenv("DOCLING_OCR_ENGINE"); env != "" {
		ocrEngine = env
	}
	if p.Config.VisionLLMModel != "" {
		ocrEngine = p.Config.VisionLLMModel
	}

	var result *DoclingConvertDocumentResponse
	var err error

	switch input.Type {
	case InputTypeImage:
		result, err = p.WithImage(ctx, input)
		if err != nil {
			return nil, fmt.Errorf("docling: failed to process image: %w", err)
		}
	case InputTypePDF:
		if !p.SupportsPDF() {
			return nil, fmt.Errorf("docling provider doesn't support PDF processing")
		}
		result, err = p.WithPDF(ctx, input)
		if err != nil {
			return nil, fmt.Errorf("docling: failed to process PDF: %w", err)
		}
	default:
		return nil, fmt.Errorf("unsupported input type: %v", input.Type)
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

func (p *DoclingProvider) WithImage(ctx context.Context, input OCRInput) (*DoclingConvertDocumentResponse, error) {
	imageBytes, err := imageio.ImageToBytes(input.Image)
	if err != nil {
		return nil, fmt.Errorf("failed to convert image to bytes: %w", err)
	}

	// Submit file and get result synchronously
	options := map[string]string{
		"to_formats":   "md",
		"from_formats": "image",
		"do_ocr":       "true",
		"force_ocr":    "true",
	}

	// Adjust output format based on config
	if p.Config.Format == "markdown" {
		options["to_formats"] = "md"
	} else {
		options["to_formats"] = "text"
	}

	filename := "image.jpg"
	if p.Config.Format == "markdown" {
		filename = "image.md"
	}

	return p.Client.ProcessFile(ctx, imageBytes, filename, options)
}

func (p *DoclingProvider) WithPDF(ctx context.Context, input OCRInput) (*DoclingConvertDocumentResponse, error) {
	if input.PDFData == nil {
		return nil, fmt.Errorf("PDF data is nil")
	}

	// Submit PDF file and get result synchronously
	options := map[string]string{
		"to_formats":   "md",
		"from_formats": "pdf",
		"do_ocr":       "true",
		"force_ocr":    "true",
	}

	// Adjust output format based on config
	if p.Config.Format == "markdown" {
		options["to_formats"] = "md"
	} else {
		options["to_formats"] = "text"
	}

	filename := "document.pdf"

	return p.Client.ProcessFile(ctx, input.PDFData, filename, options)
}

func (p *DoclingProvider) SupportsPDF() bool {
	return true
}

func (p *DoclingProvider) GetConfig() Config {
	return p.Config
}

type DoclingClient struct {
	Client  *http.Client
	BaseURL string
}

func WithDoclingBaseURL(baseURL string) func(*DoclingClient) {
	return func(c *DoclingClient) {
		c.BaseURL = baseURL
	}
}

func NewDoclingClient(opts ...func(*DoclingClient)) *DoclingClient {
	client := &DoclingClient{
		Client:  &http.Client{},
		BaseURL: doclingDefaultBaseURL,
	}

	for _, opt := range opts {
		opt(client)
	}

	return client
}

func (d *DoclingClient) HealthCheck(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, d.BaseURL+doclingHealthPath, nil)
	if err != nil {
		return fmt.Errorf("failed to create health check request: %w", err)
	}

	resp, err := d.Client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("health check failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var healthResponse DoclingHealthResponse
	if err := json.NewDecoder(resp.Body).Decode(&healthResponse); err != nil {
		return fmt.Errorf("failed to decode health check response: %w", err)
	}

	if healthResponse.Status != "ok" {
		return fmt.Errorf("docling service not healthy, status: %s", healthResponse.Status)
	}
	return nil
}

func (d *DoclingClient) ProcessFile(ctx context.Context, imageBytes []byte, filename string, options map[string]string) (*DoclingConvertDocumentResponse, error) {
	// multipart/form-data Fields
	var requestBody bytes.Buffer
	writer := multipart.NewWriter(&requestBody)

	part, err := writer.CreateFormFile("files", filename)
	if err != nil {
		return nil, fmt.Errorf("docling client: failed to create form file for image '%s': %w", filename, err)
	}
	_, err = io.Copy(part, bytes.NewReader(imageBytes))
	if err != nil {
		return nil, fmt.Errorf("docling client: failed to copy image data to form for '%s': %w", filename, err)
	}

	for key, value := range options {
		if err := writer.WriteField(key, value); err != nil {
			return nil, fmt.Errorf("docling client: failed to write field '%s' with value '%s': %w", key, value, err)
		}
	}

	err = writer.Close()
	if err != nil {
		return nil, fmt.Errorf("docling client: failed to close multipart writer: %w", err)
	}

	reqURL := d.BaseURL + doclingConvertPath
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, reqURL, &requestBody)
	if err != nil {
		return nil, fmt.Errorf("docling client: failed to create convert request: %w", err)
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Accept", "application/json")

	resp, err := d.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("docling: convert request to %s failed: %w", reqURL, err)
	}
	defer resp.Body.Close()

	bodyBytes, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return nil, fmt.Errorf("docling client: failed to read convert response body (status %d): %w", resp.StatusCode, readErr)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("docling: convert request to %s failed with status %d: %s", reqURL, resp.StatusCode, string(bodyBytes))
	}

	var convertResponse DoclingConvertDocumentResponse
	if err := json.Unmarshal(bodyBytes, &convertResponse); err != nil {
		return nil, fmt.Errorf("docling client: failed to decode convert response. Body: %s. Error: %w", string(bodyBytes), err)
	}

	if convertResponse.Status != "success" && convertResponse.Status != "partial_success" {
		return nil, fmt.Errorf("docling: conversion failed with status: %s", convertResponse.Status)
	}

	return &convertResponse, nil
}
