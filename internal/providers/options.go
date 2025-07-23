package providers

import (
	"fmt"
	"strconv"

	"dario.cat/mergo"
	"github.com/Achno/gowall/utils"
)

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

	if config.Format == "markdown" {
		optionsMap["to"] = "md"
	} else if config.Format == "text" {
		optionsMap["to"] = "text"
	}

	if config.VisionLLMModel != "" {
		// Check if it's a VLM model
		if config.VisionLLMModel == "smoldocling" ||
			config.VisionLLMModel == "granite_vision" ||
			config.VisionLLMModel == "granite_vision_ollama" {

			optionsMap["vlm_model"] = config.VisionLLMModel
			optionsMap["pipeline"] = "vlm"
			delete(optionsMap, "ocr_engine")

		} else {
			optionsMap["ocr_engine"] = config.VisionLLMModel
		}
	}

	if config.Language != "" {
		optionsMap["ocr_lang"] = config.Language
	}

	// let docling auto-detect
	delete(optionsMap, "from")

	return optionsMap, nil
}
