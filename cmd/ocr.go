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
	)
	flags.StringVarP(&provider, "provider", "p", "", "Provider to use for OCR")
	flags.StringVarP(&model, "model", "m", "", "Model to use for OCR")
	flags.StringVarP(&uprompt, "uprompt", "u", "", "User prompt to use for OCR")
	flags.StringVarP(&language, "language", "l", "", "Language to use for OCR")
	flags.Float64VarP(&dpi, "dpi", "d", 120.0, "DPI in pdf->images conversion")
	flags.Float64VarP(&rps, "rps", "r", 0, "Rate limit : requests per second")
	flags.IntVarP(&burst, "burst", "b", 5, "Rate limit burst requests")
	flags.StringVarP(&schema, "schema", "s", "", "The schema name to use for OCR")
	flags.StringVarP(&shared.Format, "format", "f", "", "Output format: 'md' or 'txt'")

	addGlobalFlags(cmd)

	return cmd
}

func init() {
	rootCmd.AddCommand(BuildOCRCmd())
}

func runOCRcmd(cmd *cobra.Command, args []string) {

	cfg, err := LoadOCRConfig(cmd)
	utils.HandleError(err, "Error")

	ops, err := imageio.DetermineImageOperations(shared, args, cmd)
	utils.HandleError(err, "Error")

	n, err := providers.NewOCRProvider(cfg)
	utils.HandleError(err, "Error")

	service := providers.NewProviderService(n, cfg)
	err = providers.StartOCRPipeline(ops, service)
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
				if schema.Config.OCR.Provider != "" {
					cfg.OCR.Provider = schema.Config.OCR.Provider
				}
				if schema.Config.OCR.Model != "" {
					cfg.OCR.Model = schema.Config.OCR.Model
				}
				if schema.Config.OCR.Prompt != "" {
					cfg.OCR.Prompt = schema.Config.OCR.Prompt
				}
				if schema.Config.OCR.Language != "" {
					cfg.OCR.Language = schema.Config.OCR.Language
				}
				if schema.Config.OCR.Format != "" {
					cfg.OCR.Format = schema.Config.OCR.Format
				}
				if schema.Config.Pipeline.DPI != 0 {
					cfg.Pipeline.DPI = schema.Config.Pipeline.DPI
				}
				if schema.Config.RateLimit.RPS != 0 {
					cfg.RateLimit.RPS = schema.Config.RateLimit.RPS
				}
				if schema.Config.RateLimit.Burst != 0 {
					cfg.RateLimit.Burst = schema.Config.RateLimit.Burst
				}
				if schema.Config.Pipeline.Concurrency != 0 {
					cfg.Pipeline.Concurrency = schema.Config.Pipeline.Concurrency
				}
				if schema.Config.Pipeline.OCRConcurrency != 0 {
					cfg.Pipeline.OCRConcurrency = schema.Config.Pipeline.OCRConcurrency
				}
				if schema.Config.TextCorrection.Enabled {
					cfg.TextCorrection.Enabled = schema.Config.TextCorrection.Enabled
				}
				if schema.Config.TextCorrection.Provider.Provider != "" {
					cfg.TextCorrection.Provider.Provider = schema.Config.TextCorrection.Provider.Provider
				}
				if schema.Config.TextCorrection.Provider.Model != "" {
					cfg.TextCorrection.Provider.Model = schema.Config.TextCorrection.Provider.Model
				}
				if schema.Config.TextCorrection.Provider.Prompt != "" {
					cfg.TextCorrection.Provider.Prompt = schema.Config.TextCorrection.Provider.Prompt
				}
				if schema.Config.TextCorrection.RateLimit.RPS != 0 {
					cfg.TextCorrection.RateLimit.RPS = schema.Config.TextCorrection.RateLimit.RPS
				}
				if schema.Config.TextCorrection.RateLimit.Burst != 0 {
					cfg.TextCorrection.RateLimit.Burst = schema.Config.TextCorrection.RateLimit.Burst
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
		cfg.OCR.Provider = v
	}
	if flags.Changed("model") {
		v, err := flags.GetString("model")
		if err != nil {
			return providers.Config{}, err
		}
		cfg.OCR.Model = v
	}
	if flags.Changed("uprompt") {
		v, err := flags.GetString("uprompt")
		if err != nil {
			return providers.Config{}, err
		}
		cfg.OCR.Prompt = v
	}
	if flags.Changed("language") {
		v, err := flags.GetString("language")
		if err != nil {
			return providers.Config{}, err
		}
		cfg.OCR.Language = v
	}
	if flags.Changed("dpi") {
		v, err := flags.GetFloat64("dpi")
		if err != nil {
			return providers.Config{}, err
		}
		cfg.Pipeline.DPI = v
	}
	if flags.Changed("rps") {
		v, err := flags.GetFloat64("rps")
		if err != nil {
			return providers.Config{}, err
		}
		cfg.RateLimit.RPS = v
	}
	if flags.Changed("burst") {
		v, err := flags.GetInt("burst")
		if err != nil {
			return providers.Config{}, err
		}
		cfg.RateLimit.Burst = v
	}
	if flags.Changed("format") {
		v, err := flags.GetString("format")
		if err != nil {
			return providers.Config{}, err
		}
		cfg.OCR.Format = v
	}

	return cfg, nil
}

func setDefaultOCRConfig() providers.Config {
	pdfOpts := pdf.DefaultOptions()
	return providers.Config{
		Pipeline: providers.PipelineConfig{
			DPI:            pdfOpts.DPI,
			Concurrency:    5,
			OCRConcurrency: 10,
		},
		RateLimit: providers.RateLimitConfig{
			RPS:   0,
			Burst: 12,
		},
		OCR: providers.ProviderConfig{
			Format: "md",
			// Prompt: "Extract all visible text from this image **without any changes**. Do not summarize, paraphrase, or infer missing text. Retain all spacing, punctuation, and formatting exactly as in the image. If text is unclear or partially visible, extract as much as possible without guessing. Include all text, even if it seems irrelevant or repeated.",
			Prompt: "Extract all visible text from this image and format the output as markdown. Include only the text content; no explanations or additional text should be included. If the image is empty, return an empty string. Fix any formatting issues or inconsistencies found in the extracted content",
		},
		TextCorrection: providers.TextCorrectionConfig{
			Enabled: false,
			Provider: providers.ProviderConfig{
				Prompt: `
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
			},
		},
	}
}
