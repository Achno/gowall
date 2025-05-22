package providers

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"image"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	batchJobTimeout = 3 * time.Minute
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

// MistralBatchJob represents a batch OCR job
type MistralBatchJob struct {
	ID                string `json:"id"`
	Status            string `json:"status"`
	TotalRequests     int    `json:"total_requests"`
	FailedRequests    int    `json:"failed_requests"`
	SucceededRequests int    `json:"succeeded_requests"`
	OutputFile        string `json:"output_file"`
}

// MistralBatchFile represents a file uploaded for batch processing
type MistralBatchFile struct {
	ID string `json:"id"`
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

// createBatchJSONL creates a temporary JSONL file with batch OCR entries
func (m *MistralProvider) createBatchJSONL(images []image.Image) (*os.File, error) {
	tempFile, err := os.CreateTemp("", "mistral-batch-*.jsonl")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp file: %w", err)
	}

	for i, img := range images {
		base64Image, err := imageToBase64(img)
		if err != nil {
			tempFile.Close()
			os.Remove(tempFile.Name())
			return nil, fmt.Errorf("failed to encode image %d: %w", i, err)
		}

		entry := map[string]interface{}{
			"custom_id": fmt.Sprintf("%d", i),
			"body": map[string]interface{}{
				"document": map[string]string{
					"type":      "image_url",
					"image_url": fmt.Sprintf("data:image/jpeg;base64,%s", base64Image),
				},
				"include_image_base64": true,
			},
		}

		entryBytes, err := json.Marshal(entry)
		if err != nil {
			tempFile.Close()
			os.Remove(tempFile.Name())
			return nil, fmt.Errorf("failed to marshal batch entry %d: %w", i, err)
		}

		if _, err := tempFile.Write(append(entryBytes, '\n')); err != nil {
			tempFile.Close()
			os.Remove(tempFile.Name())
			return nil, fmt.Errorf("failed to write batch entry %d: %w", i, err)
		}
	}

	return tempFile, nil
}

// uploadBatchFile uploads a file to Mistral's API and returns the file ID
func (m *MistralProvider) uploadBatchFile(ctx context.Context, file *os.File) (string, error) {
	// Ensure we're reading from the start of the file
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return "", fmt.Errorf("failed to seek file: %w", err)
	}

	// Build multipart form
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	// 1) purpose field
	if err := writer.WriteField("purpose", "batch"); err != nil {
		return "", fmt.Errorf("failed to write purpose field: %w", err)
	}

	// 2) file field
	part, err := writer.CreateFormFile("file", filepath.Base(file.Name()))
	if err != nil {
		return "", fmt.Errorf("failed to create form file part: %w", err)
	}
	if _, err := io.Copy(part, file); err != nil {
		return "", fmt.Errorf("failed to copy file contents: %w", err)
	}

	// Close to finalize the multipart boundary
	if err := writer.Close(); err != nil {
		return "", fmt.Errorf("failed to close multipart writer: %w", err)
	}

	// Build request
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, m.baseURL+"/files", &buf)
	if err != nil {
		return "", fmt.Errorf("failed to create upload request: %w", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+m.apiKey)
	// (Optional) hint that you expect JSON back
	req.Header.Set("Accept", "application/json")

	// Send and handle response
	res, err := m.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to upload batch file: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to upload batch file: unexpected status %s", res.Status)
	}

	var fileResp MistralBatchFile
	if err := json.NewDecoder(res.Body).Decode(&fileResp); err != nil {
		return "", fmt.Errorf("failed to decode file upload response: %w", err)
	}
	return fileResp.ID, nil

}

// createBatchJob creates a new batch job and returns the job ID
func (m *MistralProvider) createBatchJob(ctx context.Context, fileID, model string) (*MistralBatchJob, error) {
	jobPayload := map[string]any{
		"input_files": []string{fileID},
		"model":       model,
		"endpoint":    "/v1/ocr",
		"metadata": map[string]string{
			"job_type": "testing",
		},
	}

	jobBytes, err := json.Marshal(jobPayload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal job payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, m.baseURL+"/batch/jobs", bytes.NewReader(jobBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create job request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+m.apiKey)
	req.Header.Set("Accept", "application/json")

	res, err := m.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to create batch job 123: %w", err)
	}
	bodyBytes, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read error response body: %w", err)
	}

	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to create batch job: unexpected status  %s %s", res.Status, string(bodyBytes))
	}

	var job MistralBatchJob
	if err := json.NewDecoder(res.Body).Decode(&job); err != nil {
		return nil, fmt.Errorf("failed to decode job response: %w", err)
	}

	return &job, nil
}

// pollJobStatus polls the job status until completion or failure
func (m *MistralProvider) pollJobStatus(ctx context.Context, jobID string) (*MistralBatchJob, error) {
	for {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, m.baseURL+"/batch/jobs/"+jobID, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create status request: %w", err)
		}

		req.Header.Set("Authorization", "Bearer "+m.apiKey)

		res, err := m.client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("failed to check job status: %w", err)
		}

		if res.StatusCode != http.StatusOK {
			res.Body.Close()
			return nil, fmt.Errorf("failed to check job status: unexpected status %s", res.Status)
		}

		var job MistralBatchJob
		if err := json.NewDecoder(res.Body).Decode(&job); err != nil {
			res.Body.Close()
			return nil, fmt.Errorf("failed to decode job status: %w", err)
		}
		res.Body.Close()

		if job.Status == "COMPLETED" {
			return &job, nil
		} else if job.Status == "FAILED" {
			return nil, fmt.Errorf("batch job failed: %d failed requests", job.FailedRequests)
		}

		// Wait before next poll
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(batchJobTimeout):
		}
	}
}

// downloadBatchResults downloads and parses the batch results
func (m *MistralProvider) downloadBatchResults(ctx context.Context, fileID string) ([]*OCRResult, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, m.baseURL+"/files/"+fileID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create download request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+m.apiKey)

	res, err := m.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to download results: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to download results: unexpected status %s", res.Status)
	}

	scanner := bufio.NewScanner(res.Body)
	var results []*OCRResult

	for scanner.Scan() {
		var batchResp MistralOcrResponse
		if err := json.Unmarshal(scanner.Bytes(), &batchResp); err != nil {
			return nil, fmt.Errorf("failed to parse batch result: %w", err)
		}

		result, err := MistralToOCRResult(&batchResp)
		if err != nil {
			return nil, fmt.Errorf("failed to convert batch result: %w", err)
		}

		results = append(results, result)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to read batch results: %w", err)
	}

	return results, nil
}

// Makes use of Mistral's batch OCR API
func (m *MistralProvider) OCRBatchImages(ctx context.Context, images []image.Image) ([]*OCRResult, error) {
	model := "mistral-ocr-latest"
	if m.config.VisionLLMModel != "" {
		model = m.config.VisionLLMModel
	}

	// Create and write batch JSONL file
	tempFile, err := m.createBatchJSONL(images)
	if err != nil {
		return nil, err
	}
	defer os.Remove(tempFile.Name())
	defer tempFile.Close()

	// Upload the batch file
	fileID, err := m.uploadBatchFile(ctx, tempFile)
	if err != nil {
		return nil, err
	}

	// Create batch job
	job, err := m.createBatchJob(ctx, fileID, model)
	if err != nil {
		return nil, err
	}

	// Poll for job completion
	completedJob, err := m.pollJobStatus(ctx, job.ID)
	if err != nil {
		return nil, err
	}

	// Download and process results
	return m.downloadBatchResults(ctx, completedJob.OutputFile)
}
