package providers

import (
	"context"
	"encoding/json"
	"fmt"
	"image"
	"math"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"gopkg.in/yaml.v3"
)

func TestOpenAIOCRRawPromptAndExtra(t *testing.T) {
	const schema = `
ocr:
  provider: vllm
  model: served-name
  prompt: "Text Recognition:"
  append_prompt: "IGNORED"
  format: md
  raw_prompt: true
openai_options:
  max_tokens: 8192
  temperature: 0.01
  extra:
    top_k: 1
    top_p: 0.001
    no_repeat_ngram_size: 100
    chat_template_kwargs:
      enable_thinking: false
`
	var cfg Config
	if err := yaml.Unmarshal([]byte(schema), &cfg); err != nil {
		t.Fatal(err)
	}

	var warnings []string
	restore := swapWarnf(func(format string, args ...any) {
		warnings = append(warnings, fmt.Sprintf(format, args...))
	})
	defer restore()

	bodies := []map[string]any{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasSuffix(r.URL.Path, "/chat/completions") {
			t.Errorf("path = %s", r.URL.Path)
		}
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Errorf("decode body: %v", err)
		}
		bodies = append(bodies, body)
		finish := "length"
		if len(bodies) > 1 {
			finish = "stop"
		}
		writeChatCompletion(w, finish, "recognized text")
	}))
	defer srv.Close()

	p := openAIProvider(srv, cfg)
	result, err := p.OCR(context.Background(), OCRInput{
		Type:     InputTypeImage,
		Image:    image.NewRGBA(image.Rect(0, 0, 2, 2)),
		Filename: "scan.pdf-page-2-of-5",
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.Text != "recognized text" {
		t.Fatalf("text = %q", result.Text)
	}
	if len(warnings) != 1 || !strings.Contains(warnings[0], "finish_reason=length") {
		t.Fatalf("warnings = %#v", warnings)
	}

	body := bodies[0]
	if got := messageText(t, body); got != "Text Recognition:" {
		t.Fatalf("prompt = %q", got)
	}
	assertNumber(t, body, "max_tokens", 8192)
	assertNumber(t, body, "temperature", 0.01)
	assertNumber(t, body, "top_k", 1)
	assertNumber(t, body, "top_p", 0.001)
	assertNumber(t, body, "no_repeat_ngram_size", 100)
	kwargs, ok := body["chat_template_kwargs"].(map[string]any)
	if !ok || kwargs["enable_thinking"] != false {
		t.Fatalf("chat_template_kwargs = %#v", body["chat_template_kwargs"])
	}

	corrected, err := p.Complete(context.Background(), "page text")
	if err != nil {
		t.Fatal(err)
	}
	if corrected != "recognized text" {
		t.Fatalf("complete = %q", corrected)
	}
	if len(bodies) != 2 {
		t.Fatalf("requests = %d", len(bodies))
	}
	assertNumber(t, bodies[1], "max_tokens", 8192)
	assertNumber(t, bodies[1], "temperature", 0.01)
	assertNumber(t, bodies[1], "top_k", 1)
	if len(warnings) != 1 {
		t.Fatalf("stop should not warn, warnings = %#v", warnings)
	}
}

func TestOpenAIOCRBuildsPromptAndDefaults(t *testing.T) {
	cfg := Config{OCR: ProviderConfig{
		Provider: "vllm", Model: "served-name", Prompt: "Read this.", AppendPrompt: "Keep line breaks.", Format: "md",
	}}
	var bodies []map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Errorf("decode body: %v", err)
		}
		bodies = append(bodies, body)
		writeChatCompletion(w, "stop", "ok")
	}))
	defer srv.Close()

	p := openAIProvider(srv, cfg)
	if _, err := p.OCR(context.Background(), OCRInput{
		Type: InputTypeImage, Image: image.NewRGBA(image.Rect(0, 0, 1, 1)), Filename: "scan.pdf-page-2-of-5",
	}); err != nil {
		t.Fatal(err)
	}
	got := messageText(t, bodies[0])
	for _, want := range []string{"Read this.", "Format the output in Markdown.", "PAGE 2", "Keep line breaks."} {
		if !strings.Contains(got, want) {
			t.Fatalf("prompt %q missing %q", got, want)
		}
	}
	assertAbsent(t, bodies[0], "max_tokens")
	assertAbsent(t, bodies[0], "temperature")
	if _, ok := bodies[0]["top_k"]; ok {
		t.Fatalf("unexpected extra field: %#v", bodies[0]["top_k"])
	}

	if _, err := p.Complete(context.Background(), "page text"); err != nil {
		t.Fatal(err)
	}
	if len(bodies) != 2 {
		t.Fatalf("requests = %d", len(bodies))
	}
	assertNumber(t, bodies[1], "max_tokens", 4096)
	assertNumber(t, bodies[1], "temperature", 0.7)
}

func TestOpenAIOCRExplicitSampling(t *testing.T) {
	const schema = `
ocr:
  provider: vllm
  model: served-name
openai_options:
  max_tokens: 0
  temperature: 0
`
	var cfg Config
	if err := yaml.Unmarshal([]byte(schema), &cfg); err != nil {
		t.Fatal(err)
	}
	var body map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Errorf("decode body: %v", err)
		}
		writeChatCompletion(w, "stop", "ok")
	}))
	defer srv.Close()

	if _, err := openAIProvider(srv, cfg).OCR(context.Background(), OCRInput{
		Type: InputTypeImage, Image: image.NewRGBA(image.Rect(0, 0, 1, 1)),
	}); err != nil {
		t.Fatal(err)
	}
	assertNumber(t, body, "max_tokens", 0)
	assertNumber(t, body, "temperature", 0)
}

func TestTruncationWarning(t *testing.T) {
	if got := truncationWarning("stop"); got != "" {
		t.Fatalf("stop warning = %q", got)
	}
	if got := truncationWarning("length"); !strings.Contains(got, "max_tokens") {
		t.Fatalf("length warning = %q", got)
	}
}

func TestMiniCPMProviderRemoved(t *testing.T) {
	_, err := NewOCRProvider(Config{OCR: ProviderConfig{Provider: "minicpm", Model: "minicpm-ocr"}})
	if err == nil || !strings.Contains(err.Error(), "not supported") {
		t.Fatalf("minicpm error = %v", err)
	}
}

func openAIProvider(srv *httptest.Server, cfg Config) *OpenAIProvider {
	client := openai.NewClient(
		option.WithAPIKey("test"),
		option.WithBaseURL(srv.URL+"/v1"),
		option.WithMaxRetries(0),
	)
	return &OpenAIProvider{client: &client, model: cfg.OCR.Model, config: cfg}
}

func writeChatCompletion(w http.ResponseWriter, finish, content string) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"id": "c", "object": "chat.completion", "created": 1, "model": "served-name",
		"choices": []any{map[string]any{
			"index": 0, "finish_reason": finish,
			"message": map[string]any{"role": "assistant", "content": content},
		}},
	})
}

func messageText(t *testing.T, body map[string]any) string {
	t.Helper()
	messages, ok := body["messages"].([]any)
	if !ok || len(messages) == 0 {
		t.Fatalf("messages = %#v", body["messages"])
	}
	content := messages[0].(map[string]any)["content"]
	switch c := content.(type) {
	case string:
		return c
	case []any:
		var b strings.Builder
		for _, part := range c {
			if text, ok := part.(map[string]any)["text"].(string); ok {
				b.WriteString(text)
			}
		}
		return b.String()
	default:
		t.Fatalf("content = %#v", content)
		return ""
	}
}

func assertAbsent(t *testing.T, body map[string]any, key string) {
	t.Helper()
	if _, ok := body[key]; ok {
		t.Fatalf("%s = %#v, want omitted", key, body[key])
	}
}

func assertNumber(t *testing.T, body map[string]any, key string, want float64) {
	t.Helper()
	got, ok := body[key].(float64)
	if !ok || math.Abs(got-want) > 1e-9 {
		t.Fatalf("%s = %#v, want %v", key, body[key], want)
	}
}

func swapWarnf(fn func(string, ...any)) func() {
	prev := warnf
	warnf = fn
	return func() { warnf = prev }
}
