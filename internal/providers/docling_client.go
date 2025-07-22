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
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
)

// ╔═════════════════╗
// ║  Docling Client ║
// ╚═════════════════╝

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

// ╔═════════════════╗
// ║   Docling CLI   ║
// ╚═════════════════╝

// DoclingCliClient holds the docling CLI information
type DoclingCliClient struct {
	Available  bool
	BinaryPath string
}

// NewDoclingCliClient creates a new CLI client and checks if docling CLI is available
func NewDoclingCliClient() *DoclingCliClient {
	client := &DoclingCliClient{
		Available: false,
	}

	// Check if docling CLI is available
	path, err := exec.LookPath("docling")
	if err != nil {
		return client
	}

	client.BinaryPath = path
	client.Available = true
	return client
}

// IsAvailable returns whether the  docling CLI is available
func (c *DoclingCliClient) IsAvailable() bool {
	return c.Available
}

// optionsToCliArgs maps the docling REST API options to the docling CLI flags
func (c *DoclingCliClient) optionsToCliArgs(options map[string]string) []string {
	var args []string

	for key, value := range options {
		switch key {
		case "do_ocr":
			if value == "true" {
				args = append(args, "--ocr")
			} else {
				args = append(args, "--no-ocr")
			}
		case "force_ocr":
			if value == "true" {
				args = append(args, "--force-ocr")
			} else {
				args = append(args, "--no-force-ocr")
			}
		case "ocr_engine":
			if value != "" {
				args = append(args, "--ocr-engine", value)
			}
		case "pipeline":
			if value != "" {
				args = append(args, "--pipeline", value)
			}
		case "vlm_model":
			if value != "" {
				args = append(args, "--vlm-model", value)
			}
		case "image_export_mode":
			if value != "" {
				args = append(args, "--image-export-mode", value)
			}
		}
	}

	return args
}

// ProcessFile processes a file using the  docling CLI
func (c *DoclingCliClient) ProcessFile(ctx context.Context, fileBytes []byte, filename string, options map[string]string, outputDir string) (*DoclingConvertDocumentResponse, error) {
	if !c.Available {
		return nil, fmt.Errorf("docling CLI is not available")
	}

	// Create temporary file because docling CLI doesn't support reading from stdin
	// pass that to docling CLI with the correct arguments

	tempDir, err := os.MkdirTemp("", "docling-input-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	tempFile := filepath.Join(tempDir, filename)
	if err := os.WriteFile(tempFile, fileBytes, 0644); err != nil {
		return nil, fmt.Errorf("failed to write temp file: %w", err)
	}

	args := []string{tempFile}
	args = append(args, c.optionsToCliArgs(options)...)
	args = append(args, "--output", outputDir)

	// logger.Printf("DEBUG: Executing docling CLI: %s %v\n", c.BinaryPath, args)

	cmd := exec.CommandContext(ctx, c.BinaryPath, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err = cmd.Run()
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			if status, ok := exitError.Sys().(syscall.WaitStatus); ok {
				return nil, fmt.Errorf("docling CLI failed with exit code %d: %s", status.ExitStatus(), stderr.String())
			}
		}
		return nil, fmt.Errorf("docling CLI execution failed: %w, stderr: %s", err, stderr.String())
	}

	outputFormat := options["to_formats"]
	if outputFormat == "" {
		outputFormat = "md"
	}

	// read the temp outfile docling produced and return the Docling response
	var outputFilename string
	baseName := strings.TrimSuffix(filename, filepath.Ext(filename))
	switch outputFormat {
	case "md":
		outputFilename = baseName + ".md"
	case "text":
		outputFilename = baseName + ".txt"
	default:
		outputFilename = baseName + ".md"
	}

	outputPath := filepath.Join(outputDir, outputFilename)

	content, err := os.ReadFile(outputPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read CLI output file %s: %w", outputPath, err)
	}

	// Create response compatible with REST API response
	response := &DoclingConvertDocumentResponse{
		Document: DoclingDocumentResponse{
			Filename: filename,
		},
		Status: "success",
	}

	switch outputFormat {
	case "md":
		response.Document.MDContent = string(content)
		response.Document.TextContent = string(content)
	case "text":
		response.Document.TextContent = string(content)
	default:
		response.Document.MDContent = string(content)
		response.Document.TextContent = string(content)
	}

	return response, nil
}
