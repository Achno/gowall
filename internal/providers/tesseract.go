package providers

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/jpeg"
	"os/exec"
	"strings"
)

// Impliments the Provider interface
type TesseractProvider struct {
	config Config
	client *TesseractClient
}

func NewTesseractProvider(config Config) (OCRProvider, error) {
	return &TesseractProvider{
		config: config,
		client: &TesseractClient{},
	}, nil
}

func (p *TesseractProvider) OCR(ctx context.Context, input OCRInput) (*OCRResult, error) {

	if !p.client.IsTesseractInstalled() {
		return nil, fmt.Errorf("tesseract is not installed")
	}

	return p.client.OCRImageCmd(ctx, input.Image, p.config.Language)
}

func (p *TesseractProvider) HOCRImage(ctx context.Context, input OCRInput) (*OCRResult, error) {
	return nil, nil
}

type TesseractClient struct{}

func (c *TesseractClient) OCRImageCmd(ctx context.Context, image image.Image, lang string) (*OCRResult, error) {
	buf := new(bytes.Buffer)
	if err := jpeg.Encode(buf, image, nil); err != nil {
		return nil, err
	}

	cmd := c.ConstructTesseractCommand(ctx, []string{"-l", lang})
	out, err := c.RunTesseractCommand(ctx, cmd, buf)
	if err != nil {
		return nil, err
	}

	text := strings.TrimSpace(out)
	return &OCRResult{Text: text}, nil
}

// IsTesseractInstalled checks if tesseract is installed in $PATH
func (c *TesseractClient) IsTesseractInstalled() bool {
	_, err := exec.LookPath("tesseract")
	return err == nil
}

func (c *TesseractClient) ConstructTesseractCommand(ctx context.Context, args []string) *exec.Cmd {
	fullArgs := append([]string{"stdin", "stdout"}, args...)
	return exec.CommandContext(ctx, "tesseract", fullArgs...)
}

func (c *TesseractClient) RunTesseractCommand(ctx context.Context, cmd *exec.Cmd, input *bytes.Buffer) (string, error) {

	cmd.Stdin = input

	var out bytes.Buffer
	cmd.Stdout = &out
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("tesseract error: %v, stderr: %s", err, stderr.String())
	}

	return out.String(), nil
}
