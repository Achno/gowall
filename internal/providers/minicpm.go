package providers

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	cf "github.com/Achno/gowall/config"
)

// MiniCPMOCRProvider recognizes individual page images via a vLLM-compatible server.
type MiniCPMOCRProvider struct {
	endpoint  string
	apiKey    string
	model     string
	maxTokens int
	client    *http.Client
}

// NewMiniCPMOCRProvider configures a self-hosted MiniCPM-OCR endpoint.
func NewMiniCPMOCRProvider(config Config) (OCRProvider, error) {
	env := cf.GowallConfig.EnvConfig
	if env == nil {
		return nil, errors.New("MiniCPM-OCR environment configuration is unavailable")
	}
	endpoint, err := miniCPMEndpoint(env.MINICPM_OCR_BASE_URL)
	if err != nil {
		return nil, err
	}
	if env.MINICPM_OCR_TIMEOUT_SECONDS <= 0 || env.MINICPM_OCR_MAX_TOKENS <= 0 {
		return nil, errors.New("MiniCPM-OCR timeout and max tokens must be positive")
	}
	model := strings.TrimSpace(config.OCR.Model)
	if model == "" {
		model = "minicpm-ocr"
	}
	return &MiniCPMOCRProvider{endpoint: endpoint, apiKey: env.MINICPM_OCR_API_KEY, model: model, maxTokens: env.MINICPM_OCR_MAX_TOKENS, client: &http.Client{Timeout: time.Duration(env.MINICPM_OCR_TIMEOUT_SECONDS) * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}}, nil
}

func miniCPMEndpoint(base string) (string, error) {
	u, err := url.Parse(strings.TrimSpace(base))
	if err != nil || u.Host == "" || (u.Scheme != "http" && u.Scheme != "https") || u.User != nil {
		return "", errors.New("MINICPM_OCR_BASE_URL must be an HTTP(S) URL without userinfo")
	}
	path := strings.TrimRight(u.Path, "/")
	switch {
	case strings.HasSuffix(path, "/chat/completions"):
	case strings.HasSuffix(path, "/v1"):
		path += "/chat/completions"
	default:
		path += "/v1/chat/completions"
	}
	u.Path = path
	u.RawPath = ""
	return u.String(), nil
}

// OCR sends one image and rejects truncated or malformed responses.
func (p *MiniCPMOCRProvider) OCR(ctx context.Context, input OCRInput) (*OCRResult, error) {
	if input.Type != InputTypeImage || input.Image == nil {
		return nil, errors.New("MiniCPM-OCR accepts page images only")
	}
	bounds := input.Image.Bounds()
	if bounds.Empty() {
		return nil, errors.New("empty image")
	}
	img := image.NewRGBA(bounds)
	draw.Draw(img, bounds, &image.Uniform{C: color.White}, image.Point{}, draw.Src)
	draw.Draw(img, bounds, input.Image, bounds.Min, draw.Over)
	var encoded bytes.Buffer
	if err := png.Encode(&encoded, img); err != nil {
		return nil, fmt.Errorf("encoding page image: %w", err)
	}
	payload := map[string]any{
		"model": p.model, "max_tokens": p.maxTokens, "temperature": 0.01, "top_p": 0.001,
		"top_k": 1, "repetition_penalty": 1.0, "no_repeat_ngram_size": 100,
		"chat_template_kwargs": map[string]any{"enable_thinking": false}, "stream": false,
		"messages": []any{map[string]any{"role": "user", "content": []any{
			map[string]any{"type": "image_url", "image_url": map[string]any{"url": "data:image/png;base64," + base64.StdEncoding.EncodeToString(encoded.Bytes())}},
			map[string]any{"type": "text", "text": "Text Recognition:"},
		}}},
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("encoding OCR request: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, errors.New("creating MiniCPM-OCR request failed")
	}
	req.Header.Set("Content-Type", "application/json")
	if p.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+p.apiKey)
	}
	response, err := p.client.Do(req)
	if err != nil {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		return nil, errors.New("MiniCPM-OCR request failed")
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("MiniCPM-OCR returned HTTP %d", response.StatusCode)
	}
	var result struct {
		Choices []struct {
			Message struct {
				Content *string `json:"content"`
				Refusal string  `json:"refusal"`
			} `json:"message"`
			FinishReason string `json:"finish_reason"`
		} `json:"choices"`
	}
	if err := json.NewDecoder(io.LimitReader(response.Body, 8<<20)).Decode(&result); err != nil || len(result.Choices) == 0 {
		return nil, errors.New("invalid MiniCPM-OCR response")
	}
	choice := result.Choices[0]
	if choice.FinishReason == "length" {
		return nil, errors.New("MiniCPM-OCR output truncated; increase MINICPM_OCR_MAX_TOKENS")
	}
	if choice.Message.Content == nil || choice.Message.Refusal != "" {
		return nil, errors.New("MiniCPM-OCR returned a refusal or unsupported content")
	}
	return &OCRResult{Text: strings.TrimSpace(*choice.Message.Content), Metadata: map[string]string{"Model": p.model}}, nil
}
