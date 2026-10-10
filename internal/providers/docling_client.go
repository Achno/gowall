package providers

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"

	request "github.com/Achno/gowall/pkg/requests"
)

// ╔═════════════════╗
// ║  Docling Client ║
// ╚═════════════════╝

type DoclingClient struct {
	Client  *request.Client
	BaseURL string
}

func WithDoclingBaseURL(baseURL string) func(*DoclingClient) {
	return func(c *DoclingClient) {
		c.BaseURL = baseURL
	}
}

func NewDoclingClient(opts ...func(*DoclingClient)) *DoclingClient {
	client := &DoclingClient{
		Client:  request.NewURLClient(0),
		BaseURL: doclingDefaultBaseURL,
	}

	for _, opt := range opts {
		opt(client)
	}

	return client
}

func (d *DoclingClient) HealthCheck(ctx context.Context) error {
	healthResponse, err := request.Get[DoclingHealthResponse](d.Client, d.BaseURL+doclingHealthPath, request.WithContext(ctx))
	if err != nil {
		var statusErr *request.StatusError
		if errors.As(err, &statusErr) {
			return fmt.Errorf("health check failed with status %d: %s", statusErr.Code, string(statusErr.Body))
		}
		return err
	}

	if healthResponse.Status != "ok" {
		return fmt.Errorf("service not healthy, status: %s", healthResponse.Status)
	}
	return nil
}

func (d *DoclingClient) ProcessFile(ctx context.Context, imageBytes []byte, filename string, options map[string]string) (*DoclingConvertDocumentResponse, error) {
	reqURL := d.BaseURL + doclingConvertPath
	convertResponse, err := request.Post[DoclingConvertDocumentResponse](d.Client, reqURL, options,
		request.WithContext(ctx),
		request.WithHeader(request.Header{"Accept": "application/json"}),
		request.WithFile("files", filename, bytes.NewReader(imageBytes)),
	)
	if err != nil {
		var statusErr *request.StatusError
		if errors.As(err, &statusErr) {
			return nil, fmt.Errorf("req to %s failed with status %d: %s", reqURL, statusErr.Code, string(statusErr.Body))
		}
		return nil, fmt.Errorf("request to %s failed: %w", reqURL, err)
	}

	if convertResponse.Status != "success" && convertResponse.Status != "partial_success" {
		return nil, fmt.Errorf("unmarshalling failed status: %s", convertResponse.Status)
	}

	return convertResponse, nil
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

func (c *DoclingCliClient) IsAvailable() bool {
	return c.Available
}

// optionsToCliArgs converts options map to CLI arguments generically
func (c *DoclingCliClient) optionsToCliArgs(options map[string]string) []string {
	var args []string

	for key, value := range options {
		if value == "" {
			continue // Skip empty values
		}

		// add -- prefix to the keys and replace any _ with -
		flagName := "--" + strings.ToLower(strings.ReplaceAll(key, "_", "-"))

		if value == "true" {
			args = append(args, flagName)
		} else if value == "false" {
			if key == "ocr" {
				args = append(args, "--no-ocr")
			}
			if key == "force_ocr" {
				args = append(args, "--no-force-ocr")
			}
		} else {
			args = append(args, flagName, value)
		}
	}

	return args
}

// ProcessFile processes a file using the  docling CLI
func (c *DoclingCliClient) ProcessFile(ctx context.Context, fileBytes []byte, filename string, options map[string]string, outputDir string) (*DoclingConvertDocumentResponse, error) {
	if !c.Available {
		return nil, fmt.Errorf("is not available")
	}

	// Create temporary file because docling CLI doesn't support reading from stdin
	// pass that to docling CLI with the correct arguments

	tempDir, err := os.MkdirTemp("", "docling-input-*")
	if err != nil {
		return nil, fmt.Errorf("while creating temp dir: %w", err)
	}
	defer os.RemoveAll(tempDir)

	tempFile := filepath.Join(tempDir, filename)
	if err := os.WriteFile(tempFile, fileBytes, 0644); err != nil {
		return nil, fmt.Errorf("while writing temp file: %w", err)
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
				return nil, fmt.Errorf("failed with exit code %d: %s", status.ExitStatus(), stderr.String())
			}
		}
		return nil, fmt.Errorf("execution failed: %w, stderr: %s", err, stderr.String())
	}

	outputFormat := options["to"]
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
		return nil, fmt.Errorf("while reading CLI output file %s: %w", outputPath, err)
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
