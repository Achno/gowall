package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Achno/gowall/config"
	"github.com/Achno/gowall/internal/providers"
)

func TestLoadOCRConfigSchemaRawPrompt(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	dir := filepath.Join(home, ".config", "gowall")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	schema := []byte(`
schemas:
  - name: minicpm
    config:
      ocr:
        provider: vllm
        model: served-name
        prompt: "Text Recognition:"
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
`)
	if err := os.WriteFile(filepath.Join(dir, config.OCRSchemaFile), schema, 0o644); err != nil {
		t.Fatal(err)
	}

	command := BuildOCRCmd()
	if err := command.ParseFlags([]string{"--schema", "minicpm"}); err != nil {
		t.Fatal(err)
	}
	cfg, err := LoadOCRConfig(command)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.OCR.Provider != "vllm" || cfg.OCR.Model != "served-name" || cfg.OCR.Prompt != "Text Recognition:" || !cfg.OCR.RawPrompt {
		t.Fatalf("ocr config = %+v", cfg.OCR)
	}
	if cfg.OpenAIOptions == nil || cfg.OpenAIOptions.MaxTokens == nil || *cfg.OpenAIOptions.MaxTokens != 8192 || cfg.OpenAIOptions.Temperature == nil || *cfg.OpenAIOptions.Temperature != 0.01 {
		t.Fatalf("openai options = %+v", cfg.OpenAIOptions)
	}
	extra := cfg.OpenAIOptions.Extra
	if extra["top_k"] != 1 || extra["no_repeat_ngram_size"] != 100 {
		t.Fatalf("extra = %#v", extra)
	}
	kwargs, ok := extra["chat_template_kwargs"].(map[string]any)
	if !ok || kwargs["enable_thinking"] != false {
		t.Fatalf("chat_template_kwargs = %#v", extra["chat_template_kwargs"])
	}
	if providers.ResolvePrompt(cfg.OCR, "scan.pdf-page-2-of-5") != "Text Recognition:" {
		t.Fatal("schema raw_prompt was not applied")
	}
}
