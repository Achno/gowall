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
	)
	flags.StringVarP(&provider, "provider", "p", "", "Provider to use for OCR")
	flags.StringVarP(&model, "model", "m", "", "Model to use for OCR")
	flags.StringVarP(&uprompt, "uprompt", "u", "", "User prompt to use for OCR")
	flags.StringVarP(&language, "language", "l", "", "Language to use for OCR")
	flags.StringVarP(&schema, "schema", "s", "", "The schema name to use for OCR")

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

	n = providers.WithRateLimit(n, 0, 12)

	utils.Spinner.Start()
	utils.Spinner.Message("Starting OCR...")

	err = providers.StartOCRPipeline(ops, n)
	utils.HandleError(err, "Error")
}

func LoadOCRConfig(cmd *cobra.Command) (providers.Config, error) {

	type Schema struct {
		Name   string           `yaml:"name"`
		Config providers.Config `yaml:"config"`
	}

	type OCRConfig struct {
		Schemas []Schema `yaml:"schemas"`
	}

	var cfg providers.Config
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
				cfg = schema.Config
				schemaFound = true
				break
			}
		}
		if !schemaFound {
			return providers.Config{}, fmt.Errorf("schema name '%s' not found in config file", schemaName)
		}

	} else {
		// If no schema, start with empty config
		cfg = providers.Config{}
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

	//TODO remove this later when you find a centralized prompt
	// cfg.VisionLLMPrompt = "Extract all visible text from this image in english,Do not summarize, paraphrase, or infer missing text,Retain all spacing, punctuation, and formatting exactly as in the image,Include all text, even if it seems irrelevant or repeated"
	cfg.VisionLLMPrompt = "Extract all visible text from this image and format the output as markdown. Follow these rules: 1) Include only the text content with no explanations. 3) If text appears to be a continuation of content (not starting a new major topic), use the sub-heading ##. 4) Preserve the exact text content. Preserve spaces/formatting if it is code. 5) If the image is empty, return an empty string."

	return cfg, nil
}
