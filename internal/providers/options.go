package providers

import (
	"fmt"
	"strconv"

	"dario.cat/mergo"
	"github.com/Achno/gowall/utils"
)

// 	TextCorrectionEnabled  bool    `yaml:"text_correction_enabled"`
// 	TextCorrectionProvider string  `yaml:"text_correction_provider"`
// 	TextCorrectionModel    string  `yaml:"text_correction_model"`
// 	TextCorrectionPrompt   string  `yaml:"text_correction_prompt"`
// 	TextCorrectionRPS      float64 `yaml:"text_correction_rps"`
// 	TextCorrectionBurst    int     `yaml:"text_correction_burst"`

// 	// Provider-specific options
// 	DoclingOptions *DoclingOptions `yaml:"docling_options,omitempty"`
// }

// Core provider configuration
type ProviderConfig struct {
	Provider    string `yaml:"provider"`
	Model       string `yaml:"model"`
	Prompt      string `yaml:"prompt"`
	Language    string `yaml:"language"`
	Format      string `yaml:"format"` // "markdown", "text"
	SupportsPDF bool   `yaml:"supports_pdf"`
}

// Pipeline configuration
type PipelineConfig struct {
	DPI            float64 `yaml:"dpi"`
	Concurrency    int     `yaml:"concurrency"`
	OCRConcurrency int     `yaml:"ocr_concurrency"`
}

// Rate limiting configuration
type RateLimitConfig struct {
	RPS   float64 `yaml:"rps"`
	Burst int     `yaml:"burst"`
}

// Text correction configuration
type TextCorrectionConfig struct {
	Enabled   bool            `yaml:"enabled"`
	Provider  ProviderConfig  `yaml:"provider"`
	RateLimit RateLimitConfig `yaml:"rate_limit"`
}

// Main configuration that composes others
type Config struct {
	OCR            ProviderConfig       `yaml:"ocr"`
	Pipeline       PipelineConfig       `yaml:"pipeline"`
	RateLimit      RateLimitConfig      `yaml:"rate_limit"`
	TextCorrection TextCorrectionConfig `yaml:"text_correction"`

	// Provider-specific options
	DoclingOptions *DoclingOptions `yaml:"docling_options,omitempty"`
}

// DoclingOptions 3 state bool to overcome go unset,specified false bool fields.
type DoclingOptions struct {
	OCR       *bool `yaml:"do_ocr,omitempty"`
	Force_OCR *bool `yaml:"force_ocr,omitempty"`

	Pipeline string `yaml:"pipeline,omitempty"`

	PDF_Backend      string `yaml:"pdf_backend,omitempty"`
	Abort_On_Error   *bool  `yaml:"abort_on_error,omitempty"`
	Document_Timeout string `yaml:"document_timeout,omitempty"`
	Num_Threads      string `yaml:"num_threads,omitempty"`
	Device           string `yaml:"device,omitempty"`
	Verbose          string `yaml:"verbose,omitempty"`
}

// Apply merges schema options with defaults and returns map[string]string
func (d *DoclingOptions) Apply(defaults any, config Config) (any, error) {
	defaultOptions, ok := defaults.(*DoclingOptions)
	if !ok {
		return nil, fmt.Errorf("defaults 'any' is not a 'DoclingOptions'")
	}

	// Start with defaults and merge overrides on top
	merged := *defaultOptions
	if d != nil {
		if err := mergo.Merge(&merged, d, mergo.WithoutDereference, mergo.WithSliceDeepCopy); err != nil {
			return nil, fmt.Errorf("failed to merge DoclingOptions: %w", err)
		}
	}

	optionsMap := make(map[string]string)

	optionsMap["ocr"] = strconv.FormatBool(utils.BoolValue(merged.OCR))
	optionsMap["force_ocr"] = strconv.FormatBool(utils.BoolValue(merged.Force_OCR))
	optionsMap["pipeline"] = merged.Pipeline
	optionsMap["pdf_backend"] = merged.PDF_Backend
	optionsMap["abort_on_error"] = strconv.FormatBool(utils.BoolValue(merged.Abort_On_Error))
	optionsMap["document_timeout"] = merged.Document_Timeout
	optionsMap["num_threads"] = merged.Num_Threads
	optionsMap["device"] = merged.Device
	optionsMap["verbose"] = merged.Verbose

	if config.OCR.Format == "markdown" {
		optionsMap["to"] = "md"
	} else if config.OCR.Format == "text" {
		optionsMap["to"] = "text"
	}

	if config.OCR.Model != "" {
		// Check if it's a VLM model
		if config.OCR.Model == "smoldocling" ||
			config.OCR.Model == "granite_vision" ||
			config.OCR.Model == "granite_vision_ollama" {

			optionsMap["vlm_model"] = config.OCR.Model
			optionsMap["pipeline"] = "vlm"
			delete(optionsMap, "ocr_engine")

		} else {
			optionsMap["ocr_engine"] = config.OCR.Model
		}
	}

	if config.OCR.Language != "" {
		optionsMap["ocr_lang"] = config.OCR.Language
	}

	// let docling auto-detect
	delete(optionsMap, "from")

	return optionsMap, nil
}
