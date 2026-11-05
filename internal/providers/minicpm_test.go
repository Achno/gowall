package providers

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"image"
	"image/color"
	"image/png"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	cf "github.com/Achno/gowall/config"
)

func TestMiniCPMEndpoint(t *testing.T) {
	cases := []struct {
		in, want string
		bad      bool
	}{
		{"http://localhost:8000", "http://localhost:8000/v1/chat/completions", false},
		{"https://host/v1/?token=abc", "https://host/v1/chat/completions?token=abc", false},
		{"http://host/v1/chat/completions", "http://host/v1/chat/completions", false},
		{"http://u:p@host", "", true}, {"ftp://host", "", true}, {"http:///v1", "", true},
	}
	for _, tc := range cases {
		got, err := miniCPMEndpoint(tc.in)
		if tc.bad {
			if err == nil {
				t.Errorf("%s: expected error", tc.in)
			}
			continue
		}
		if err != nil || got != tc.want {
			t.Errorf("%s: got %q, %v", tc.in, got, err)
		}
	}
}

func TestMiniCPMRequest(t *testing.T) {
	var got map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/chat/completions" || r.Header.Get("Authorization") != "Bearer test-key" {
			t.Errorf("unexpected request")
		}
		if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
			t.Error(err)
		}
		_, _ = w.Write([]byte(`{"choices":[{"message":{"content":"Hello"},"finish_reason":"stop"}]}`))
	}))
	defer srv.Close()
	endpoint, _ := miniCPMEndpoint(srv.URL)
	p := &MiniCPMOCRProvider{endpoint: endpoint, model: "minicpm-ocr", apiKey: "test-key", maxTokens: 1024, client: srv.Client()}
	img := image.NewNRGBA(image.Rect(0, 0, 2, 1))
	img.SetNRGBA(1, 0, color.NRGBA{A: 255})
	result, err := p.OCR(context.Background(), OCRInput{Type: InputTypeImage, Image: img})
	if err != nil || result.Text != "Hello" {
		t.Fatalf("result %v error %v", result, err)
	}
	if got["model"] != "minicpm-ocr" || got["max_tokens"] != float64(1024) || got["temperature"] != 0.01 || got["top_p"] != 0.001 || got["top_k"] != float64(1) || got["no_repeat_ngram_size"] != float64(100) {
		t.Fatalf("wrong request parameters: %#v", got)
	}
	messages := got["messages"].([]any)
	content := messages[0].(map[string]any)["content"].([]any)
	if content[1].(map[string]any)["text"] != "Text Recognition:" {
		t.Error("incorrect prompt")
	}
	data := content[0].(map[string]any)["image_url"].(map[string]any)["url"].(string)
	pngBytes, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(data, "data:image/png;base64,"))
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := png.Decode(bytes.NewReader(pngBytes))
	if err != nil {
		t.Fatal(err)
	}
	if got := color.RGBAModel.Convert(decoded.At(0, 0)).(color.RGBA); got != (color.RGBA{255, 255, 255, 255}) {
		t.Errorf("transparent background = %v", got)
	}
	if got := color.RGBAModel.Convert(decoded.At(1, 0)).(color.RGBA); got != (color.RGBA{0, 0, 0, 255}) {
		t.Errorf("text = %v", got)
	}
}

func TestMiniCPMErrors(t *testing.T) {
	for _, tc := range []struct {
		name, body string
		status     int
	}{
		{"http", "error", 500}, {"invalid json", "not-json", 200}, {"empty choices", `{"choices":[]}`, 200},
		{"truncated", `{"choices":[{"message":{"content":"partial"},"finish_reason":"length"}]}`, 200},
		{"refusal", `{"choices":[{"message":{"content":null,"refusal":"no"}}]}`, 200},
	} {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tc.status)
				_, _ = w.Write([]byte(tc.body))
			}))
			defer srv.Close()
			p := &MiniCPMOCRProvider{endpoint: srv.URL, model: "m", maxTokens: 4, client: srv.Client()}
			if _, err := p.OCR(context.Background(), OCRInput{Image: image.NewRGBA(image.Rect(0, 0, 1, 1))}); err == nil {
				t.Fatal("expected error")
			}
		})
	}
	p := &MiniCPMOCRProvider{client: &http.Client{Timeout: time.Second}}
	if _, err := p.OCR(context.Background(), OCRInput{Type: InputTypePDF}); err == nil {
		t.Fatal("PDF should be rejected")
	}
}

func TestMiniCPMRegistration(t *testing.T) {
	old := cf.GowallConfig.EnvConfig
	defer func() { cf.GowallConfig.EnvConfig = old }()
	cf.GowallConfig.EnvConfig = &cf.EnvConfig{MINICPM_OCR_BASE_URL: "http://localhost:8000", MINICPM_OCR_TIMEOUT_SECONDS: 12, MINICPM_OCR_MAX_TOKENS: 512}
	p, err := NewOCRProvider(Config{OCR: ProviderConfig{Provider: "minicpm", Model: "minicpm-ocr"}})
	if err != nil {
		t.Fatal(err)
	}
	mini, ok := p.(*MiniCPMOCRProvider)
	if !ok || mini.maxTokens != 512 || mini.client.Timeout != 12*time.Second {
		t.Fatalf("unexpected provider %v", p)
	}
	if service := NewProviderService(p, Config{}); service.SupportsPDF() {
		t.Fatal("PDF must be rendered as pages")
	}
}
