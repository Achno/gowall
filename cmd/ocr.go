/*
Copyright © 2025 Achno <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/Achno/gowall/config"
	imageio "github.com/Achno/gowall/internal/image_io"
	"github.com/Achno/gowall/internal/pdf"
	"github.com/Achno/gowall/internal/providers"
	"github.com/Achno/gowall/utils"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

func BuildOCRCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ocr [INPUT]",
		Short: "Input images & pdfs extract the text and return as plain text or markdown.",
		Long:  `Input images & pdfs, apply pre-processing ,connect to various OCR providers regardless if they support pdfs or not, apply post-processing and output as markdown or plain text`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if !cmd.Flags().Changed("schema") {
				if err := cmd.MarkFlagRequired("provider"); err != nil {
					return err
				}
				if err := cmd.MarkFlagRequired("model"); err != nil {
					return err
				}
			}
			return nil
		},
		Run: runOCRcmd,
	}

	flags := cmd.Flags()
	var (
		provider string
		model    string
		uprompt  string
		language string
		schema   string
		dpi      float64
		rps      float64
		burst    int
		format   string
	)
	flags.StringVarP(&provider, "provider", "p", "", "Provider to use for OCR")
	flags.StringVarP(&model, "model", "m", "", "Model to use for OCR")
	flags.StringVarP(&uprompt, "uprompt", "u", "", "User prompt to use for OCR")
	flags.StringVarP(&language, "language", "l", "", "Language to use for OCR")
	flags.Float64VarP(&dpi, "dpi", "d", 120.0, "DPI in pdf->images conversion")
	flags.Float64VarP(&rps, "rps", "r", 0, "Rate limit : requests per second")
	flags.IntVarP(&burst, "burst", "b", 5, "Rate limit burst requests")
	flags.StringVarP(&schema, "schema", "s", "", "The schema name to use for OCR")
	flags.StringVarP(&format, "format", "f", "", "Output format: 'markdown' or 'text'")

	addGlobalFlags(cmd)

	return cmd
}

func init() {
	rootCmd.AddCommand(BuildOCRCmd())
}

func runOCRcmd(cmd *cobra.Command, args []string) {

	cfg, err := LoadOCRConfig(cmd)
	utils.HandleError(err, "Error")

	// n, err := providers.NewOCRProvider(providers.Config{
	// 	ProviderName:      "ollama",
	// 	VisionLLMProvider: "ollama",
	// 	VisionLLMModel:    "minicpm-v",
	// 	VisionLLMPrompt:   "Extract the text in this image, DO not infer programming languages syntax, just write whatever you see, DO NOT WRITE ANYTHING ELSE BUT THE CONTENT inside the image,also keep the format, if they have a new line then write the content in the new line ect...",
	// })
	//? VLLM
	// n, err := providers.NewOCRProvider(providers.Config{
	// 	VisionLLMProvider: "vllm",
	// 	VisionLLMModel:    "ds4sd/SmolDocling-256M-preview",
	// 	// VisionLLMPrompt: "Extract all visible text from this image in english,Do not summarize, paraphrase, or infer missing text,Retain all spacing, punctuation, and formatting exactly as in the image,Include all text, even if it seems irrelevant or repeated.",
	// 	VisionLLMPrompt: "turn code to text",
	// })
	//? OPENAI
	// n, err := providers.NewOCRProvider(providers.Config{
	// 	VisionLLMProvider: "openai",
	// 	VisionLLMModel:    "gpt-4-vision-preview",
	// 	// VisionLLMPrompt: "Extract all visible text from this image in english,Do not summarize, paraphrase, or infer missing text,Retain all spacing, punctuation, and formatting exactly as in the image,Include all text, even if it seems irrelevant or repeated.",
	// 	VisionLLMPrompt: "turn code to text",
	// })
	//? Gemini
	// n, err := providers.NewOCRProvider(providers.Config{
	// 	VisionLLMProvider: "gemini",
	// 	VisionLLMModel:    "gemini-2.5-pro-exp-03-25",
	// 	VisionLLMPrompt:   "Extract all visible text from this image in english,Do not summarize, paraphrase, or infer missing text,Retain all spacing, punctuation, and formatting exactly as in the image,Include all text, even if it seems irrelevant or repeated.",
	// 	// VisionLLMPrompt: "turn code to text",
	// })
	//? Mistral
	// n, err := providers.NewOCRProvider(providers.Config{
	// 	VisionLLMProvider: "mistral",
	// 	VisionLLMModel:    "mistral-ocr-latest",
	// 	VisionLLMPrompt:   "Extract all visible text from this pdf,Do not summarize, paraphrase, or infer missing text,Retain all spacing, punctuation, and formatting exactly as in the image,Include all text, even if it seems irrelevant or repeated.",
	// 	// VisionLLMPrompt: "turn code to text",
	// })
	//? Openrouter
	// n, err := providers.NewOCRProvider(providers.Config{
	// 	VisionLLMProvider: "openrouter",
	// 	VisionLLMModel:    "qwen/qwen2.5-vl-72b-instruct:free",
	// 	VisionLLMPrompt:   "Extract all visible text from this image in english,Do not summarize, paraphrase, or infer missing text,Retain all spacing, punctuation, and formatting exactly as in the image,Include all text, even if it seems irrelevant or repeated.",
	// 	// VisionLLMPrompt: "turn code to text",
	// })
	//? Tesseract
	// n, err := providers.NewOCRProvider(providers.Config{
	// 	VisionLLMProvider: "tesseract",
	// 	VisionLLMModel:    "tesseract",
	// 	VisionLLMPrompt:   "X",
	// 	Language:          "eng",
	// })
	//? Docling
	// n, err := providers.NewOCRProvider(providers.Config{
	// 	VisionLLMProvider: "docling",
	// 	VisionLLMModel:    "easyocr",
	// 	VisionLLMPrompt:   "X",
	// 	Language:          "en",
	// })
	ops, err := imageio.DetermineImageOperations(shared, args, cmd)
	utils.HandleError(err, "Error")

	n, err := providers.NewOCRProvider(cfg)
	utils.HandleError(err, "Error")

	n = providers.WithRateLimit(n, cfg.RateLimitRPS, cfg.RateLimitBurst)

	utils.Spinner.Start()
	utils.Spinner.Message("Starting OCR...")

	err = providers.StartOCRPipeline(ops, n)
	utils.HandleError(err, "Error")
}

// (1) Loads default config
// (2) Merges the schema file if the schema flag is set and overrides the fields that are set
// (3) Overwrites the config with the flags if they are set
func LoadOCRConfig(cmd *cobra.Command) (providers.Config, error) {

	type Schema struct {
		Name   string           `yaml:"name"`
		Config providers.Config `yaml:"config"`
	}

	type OCRConfig struct {
		Schemas []Schema `yaml:"schemas"`
	}

	cfg := setDefaultOCRConfig()
	flags := cmd.Flags()

	// If schema flag is set, load from schema file
	if flags.Changed("schema") {
		schemaName, err := flags.GetString("schema")
		if err != nil {
			return providers.Config{}, err
		}

		home, err := os.UserHomeDir()
		if err != nil {
			return providers.Config{}, err
		}
		schemaPath := filepath.Join(home, ".config", "gowall", config.OCRSchemaFile)
		data, err := os.ReadFile(schemaPath)
		if err != nil {
			return providers.Config{}, fmt.Errorf("%w : You need to create the %s first", err, config.OCRSchemaFile)
		}
		var ocrConfig OCRConfig
		err = yaml.Unmarshal(data, &ocrConfig)
		if err != nil {
			return providers.Config{}, err
		}

		// Check if the schema exists
		schemaFound := false
		for _, schema := range ocrConfig.Schemas {
			if schema.Name == schemaName {
				// Merge schema config with default config (schema values override defaults)
				if schema.Config.VisionLLMProvider != "" {
					cfg.VisionLLMProvider = schema.Config.VisionLLMProvider
				}
				if schema.Config.VisionLLMModel != "" {
					cfg.VisionLLMModel = schema.Config.VisionLLMModel
				}
				if schema.Config.VisionLLMPrompt != "" {
					cfg.VisionLLMPrompt = schema.Config.VisionLLMPrompt
				}
				if schema.Config.Language != "" {
					cfg.Language = schema.Config.Language
				}
				if schema.Config.Format != "" {
					cfg.Format = schema.Config.Format
				}
				if schema.Config.DPI != 0 {
					cfg.DPI = schema.Config.DPI
				}
				if schema.Config.RateLimitRPS != 0 {
					cfg.RateLimitRPS = schema.Config.RateLimitRPS
				}
				if schema.Config.RateLimitBurst != 0 {
					cfg.RateLimitBurst = schema.Config.RateLimitBurst
				}
				if schema.Config.Concurrency != 0 {
					cfg.Concurrency = schema.Config.Concurrency
				}
				if schema.Config.TextCorrectionEnabled {
					cfg.TextCorrectionEnabled = schema.Config.TextCorrectionEnabled
				}
				if schema.Config.TextCorrectionPrompt != "" {
					cfg.TextCorrectionPrompt = schema.Config.TextCorrectionPrompt
				}
				if schema.Config.TextCorrectionProvider != "" {
					cfg.TextCorrectionProvider = schema.Config.TextCorrectionProvider
				}
				if schema.Config.TextCorrectionModel != "" {
					cfg.TextCorrectionModel = schema.Config.TextCorrectionModel
				}
				if schema.Config.TextCorrectionRPS != 0 {
					cfg.TextCorrectionRPS = schema.Config.TextCorrectionRPS
				}
				if schema.Config.TextCorrectionBurst != 0 {
					cfg.TextCorrectionBurst = schema.Config.TextCorrectionBurst
				}
				if schema.Config.DoclingOptions != nil {
					cfg.DoclingOptions = schema.Config.DoclingOptions
				}
				schemaFound = true
				break
			}
		}
		if !schemaFound {
			return providers.Config{}, fmt.Errorf("schema name '%s' not found in config file", schemaName)
		}

	}

	// Overwrite the config with the flags if they are set,otherwise keep default values
	if flags.Changed("provider") {
		v, err := flags.GetString("provider")
		if err != nil {
			return providers.Config{}, err
		}
		cfg.VisionLLMProvider = v
	}
	if flags.Changed("model") {
		v, err := flags.GetString("model")
		if err != nil {
			return providers.Config{}, err
		}
		cfg.VisionLLMModel = v
	}
	if flags.Changed("uprompt") {
		v, err := flags.GetString("uprompt")
		if err != nil {
			return providers.Config{}, err
		}
		cfg.VisionLLMPrompt = v
	}
	if flags.Changed("language") {
		v, err := flags.GetString("language")
		if err != nil {
			return providers.Config{}, err
		}
		cfg.Language = v
	}
	if flags.Changed("dpi") {
		v, err := flags.GetFloat64("dpi")
		if err != nil {
			return providers.Config{}, err
		}
		cfg.DPI = v
	}
	if flags.Changed("rps") {
		v, err := flags.GetFloat64("rps")
		if err != nil {
			return providers.Config{}, err
		}
		cfg.RateLimitRPS = v
	}
	if flags.Changed("burst") {
		v, err := flags.GetInt("burst")
		if err != nil {
			return providers.Config{}, err
		}
		cfg.RateLimitBurst = v
	}
	if flags.Changed("format") {
		v, err := flags.GetString("format")
		if err != nil {
			return providers.Config{}, err
		}
		cfg.Format = v
	}

	return cfg, nil
}

func setDefaultOCRConfig() providers.Config {
	pdfOpts := pdf.DefaultOptions()
	return providers.Config{
		DPI:                   pdfOpts.DPI,
		RateLimitRPS:          0,
		RateLimitBurst:        12,
		Concurrency:           10,
		Format:                "markdown",
		VisionLLMPrompt:       "Extract all visible text from this image **without any changes**. Do not summarize, paraphrase, or infer missing text. Retain all spacing, punctuation, and formatting exactly as in the image. If text is unclear or partially visible, extract as much as possible without guessing. Include all text, even if it seems irrelevant or repeated.",
		TextCorrectionEnabled: false,
		TextCorrectionPrompt: `
			Correct OCR-induced errors in the text, ensuring it flows coherently with the previous context. Follow these guidelines:

			1. Fix OCR-induced typos and errors:
			- Correct words split across line breaks
			- Fix common OCR errors (e.g., 'rn' misread as 'm')
			- Use context and common sense to correct errors
			- Only fix clear errors, don't alter the content unnecessarily
			- Do not add extra periods or any unnecessary punctuation

			2. Maintain original structure:
			- Keep all headings and subheadings intact

			3. Preserve original content:
			- Keep all important information from the original text
			- Do not add any new information not present in the original text
			- Remove unnecessary line breaks within sentences or paragraphs
			- Maintain paragraph breaks
			
			4. Maintain coherence:
			- Ensure the content connects smoothly with the previous context
			- Handle text that starts or ends mid-sentence appropriately

			IMPORTANT: Respond ONLY with the corrected text. Preserve all original formatting, including line breaks. Do not include any introduction, explanation, or metadata
		`,
	}
}
