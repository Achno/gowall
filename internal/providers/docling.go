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
	"time"

	cf "github.com/Achno/gowall/config"
	imageio "github.com/Achno/gowall/internal/image_io"
)

const (
	doclingDefaultBaseURL     = "http://localhost:5001"
	doclingHealthPath         = "/health"
	doclingAsyncConvertPath   = "/v1alpha/convert/file/async"
	doclingPollStatusPath     = "/v1alpha/status/poll/"
	doclingResultPath         = "/v1alpha/result/"
	defaultDoclingOCREngine   = "easyocr"
	defaultPollInterval       = 5 * time.Second
	defaultOverallPollTimeout = 1 * time.Hour
)

// DoclingProvider implements the OCRProvider interface
type DoclingProvider struct {
	Client *DoclingClient
	Config Config
}

type DoclingHealthResponse struct {
	Status string `json:"status"`
}

type DoclingTaskStatusResponse struct {
	TaskID       string `json:"task_id"`
	TaskStatus   string `json:"task_status"` // "pending", "started", "failure", "success"
	TaskPosition *int   `json:"task_position"`
	TaskMeta     any    `json:"task_meta"`
	Error        string `json:"error,omitempty"`
}

type DoclingErrorResponse struct {
	Detail string `json:"detail"`
	Error  string `json:"error"`
}

type DoclingDocumentResponse struct {
	Filename       string          `json:"filename"`
	MDContent      *string         `json:"md_content"`
	JSONContent    json.RawMessage `json:"json_content"`
	HTMLContent    *string         `json:"html_content"`
	TextContent    *string         `json:"text_content"`
	DocTagsContent *string         `json:"doctags_content"`
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

	client := NewDoclingClient(
		WithDoclingBaseURL(baseURL),
		WithDoclingPollInterval(defaultPollInterval),
	)

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

	imageBytes, err := imageio.ImageToBytes(input.Image)
	if err != nil {
		return nil, fmt.Errorf("failed to convert image to bytes: %w", err)
	}

	// Submit task
	options := map[string]string{
		"to_formats":   "text",
		"from_formats": "image",
		"do_ocr":       "true",
	}
	filename := "image.jpg"
	if p.Config.EnableMarkdown {
		filename = "image.md"
	}

	taskID, err := p.Client.ProcessFileAsync(ctx, imageBytes, filename, options)
	if err != nil {
		return nil, fmt.Errorf("docling: failed to submit OCR task: %w", err)
	}

	result, err := p.waitForTask(ctx, taskID)
	if err != nil {
		return nil, err
	}

	text := ""
	if result.Document.TextContent != nil {
		text = *result.Document.TextContent
	}

	return &OCRResult{
		Text: text,
		Metadata: map[string]string{
			"DoclingTaskID":    taskID,
			"DoclingOCREngine": ocrEngine,
			"DoclingStatus":    result.Status,
		},
	}, nil
}

func (p *DoclingProvider) waitForTask(ctx context.Context, taskID string) (*DoclingConvertDocumentResponse, error) {
	pollCtx, cancel := context.WithTimeout(ctx, defaultOverallPollTimeout)
	defer cancel()

	ticker := time.NewTicker(p.Client.PollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-pollCtx.Done():
			return nil, fmt.Errorf("docling: polling timeout for task %s", taskID)
		case <-ticker.C:
			status, err := p.Client.PollTaskStatus(pollCtx, taskID)
			if err != nil {
				continue
			}

			switch status.TaskStatus {
			case "success":
				result, err := p.Client.FetchTaskResult(ctx, taskID)
				if err != nil {
					return nil, fmt.Errorf("docling: failed to fetch result: %w", err)
				}

				if result.Status != "success" && result.Status != "partial_success" {
					return nil, fmt.Errorf("docling: task failed with status: %s", result.Status)
				}

				return result, nil

			case "failure", "skipped":
				errorMsg := "unknown error"
				if status.Error != "" {
					errorMsg = status.Error
				}
				return nil, fmt.Errorf("docling: task %s failed: %s", taskID, errorMsg)
			}
		}
	}
}

type DoclingClient struct {
	Client       *http.Client
	BaseURL      string
	PollInterval time.Duration
	PollTimeout  time.Duration
}

func WithDoclingBaseURL(baseURL string) func(*DoclingClient) {
	return func(c *DoclingClient) {
		c.BaseURL = baseURL
	}
}

func WithDoclingPollInterval(pollInterval time.Duration) func(*DoclingClient) {
	return func(c *DoclingClient) {
		c.PollInterval = pollInterval
	}
}

func NewDoclingClient(opts ...func(*DoclingClient)) *DoclingClient {
	client := &DoclingClient{
		Client:       &http.Client{},
		BaseURL:      doclingDefaultBaseURL,
		PollInterval: defaultPollInterval,
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

func (d *DoclingClient) ProcessFileAsync(ctx context.Context, imageBytes []byte, filename string, options map[string]string) (string, error) {
	// multipart/form-data Fields
	var requestBody bytes.Buffer
	writer := multipart.NewWriter(&requestBody)

	part, err := writer.CreateFormFile("files", filename)
	if err != nil {
		return "", fmt.Errorf("docling client: failed to create form file for image '%s': %w", filename, err)
	}
	_, err = io.Copy(part, bytes.NewReader(imageBytes))
	if err != nil {
		return "", fmt.Errorf("docling client: failed to copy image data to form for '%s': %w", filename, err)
	}

	for key, value := range options {
		if err := writer.WriteField(key, value); err != nil {
			return "", fmt.Errorf("docling client: failed to write field '%s' with value '%s': %w", key, value, err)
		}
	}

	err = writer.Close()
	if err != nil {
		return "", fmt.Errorf("docling client: failed to close multipart writer: %w", err)
	}

	reqURL := d.BaseURL + doclingAsyncConvertPath
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, reqURL, &requestBody)
	if err != nil {
		return "", fmt.Errorf("docling client: failed to create async convert request: %w", err)
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Accept", "application/json")

	resp, err := d.Client.Do(req)
	if err != nil {
		return "", fmt.Errorf("docling: async convert request to %s failed: %w", reqURL, err)
	}
	defer resp.Body.Close()

	bodyBytes, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return "", fmt.Errorf("docling client: failed to read async submission response body (status %d): %w", resp.StatusCode, readErr)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("docling: async submission to %s failed with status %d: %s", reqURL, resp.StatusCode, string(bodyBytes))
	}

	var taskResponse DoclingTaskStatusResponse
	if err := json.Unmarshal(bodyBytes, &taskResponse); err != nil {
		return "", fmt.Errorf("docling client: failed to decode task submission response. Body: %s. Error: %w", string(bodyBytes), err)
	}

	if taskResponse.TaskID == "" {
		return "", fmt.Errorf("docling client: task submission response did not include a task_id. Body: %s", string(bodyBytes))
	}

	return taskResponse.TaskID, nil
}

// PollTaskStatus polls the status of an ongoing task
func (d *DoclingClient) PollTaskStatus(ctx context.Context, taskID string) (*DoclingTaskStatusResponse, error) {
	statusReqURL := d.BaseURL + doclingPollStatusPath + taskID
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, statusReqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("docling: failed to create poll status request for task %s: %w", taskID, err)
	}
	req.Header.Set("Accept", "application/json")

	resp, err := d.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("docling: error polling status for task %s from %s: %w", taskID, statusReqURL, err)
	}
	defer resp.Body.Close()

	bodyBytes, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return nil, fmt.Errorf("docling: failed to read poll status response body for task %s (status %d): %w", taskID, resp.StatusCode, readErr)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("docling: poll status for task %s from %s returned status %d: %s", taskID, statusReqURL, resp.StatusCode, string(bodyBytes))
	}

	var currentTaskStatus DoclingTaskStatusResponse
	if err := json.Unmarshal(bodyBytes, &currentTaskStatus); err != nil {
		return nil, fmt.Errorf("docling: failed to decode poll status response for task %s. Body: %s. Error: %w", taskID, string(bodyBytes), err)
	}
	return &currentTaskStatus, nil
}

// FetchTaskResult retrieves the result of a completed task
func (d *DoclingClient) FetchTaskResult(ctx context.Context, taskID string) (*DoclingConvertDocumentResponse, error) {
	resultReqURL := d.BaseURL + doclingResultPath + taskID
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, resultReqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("docling: failed to create fetch result request for task %s: %w", taskID, err)
	}
	req.Header.Set("Accept", "application/json")

	resp, err := d.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("docling: fetch result request for task %s from %s failed: %w", taskID, resultReqURL, err)
	}
	defer resp.Body.Close()

	bodyBytes, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return nil, fmt.Errorf("docling: failed to read fetch result response body for task %s (status %d): %w", taskID, resp.StatusCode, readErr)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("docling: fetch result for task %s from %s failed with status %d: %s", taskID, resultReqURL, resp.StatusCode, string(bodyBytes))
	}

	var convertResponse DoclingConvertDocumentResponse
	if err := json.Unmarshal(bodyBytes, &convertResponse); err != nil {
		return nil, fmt.Errorf("docling: failed to decode convert response for task %s. Body: %s. Error: %w", taskID, string(bodyBytes), err)
	}
	return &convertResponse, nil
}
