/*
Copyright © 2025 Achno <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"fmt"
	image2 "image"

	"github.com/Achno/gowall/internal/image"
	imageio "github.com/Achno/gowall/internal/image_io"
	"github.com/Achno/gowall/internal/providers"
	"github.com/Achno/gowall/utils"
	"github.com/spf13/cobra"
)

// ocrCmd represents the ocr command
var ocrCmd = &cobra.Command{
	Use:   "ocr",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
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
		// 	VisionLLMPrompt:   "Extract all visible text from this image in english,Do not summarize, paraphrase, or infer missing text,Retain all spacing, punctuation, and formatting exactly as in the image,Include all text, even if it seems irrelevant or repeated.",
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
		n, err := providers.NewOCRProvider(providers.Config{
			VisionLLMProvider: "tesseract",
			VisionLLMModel:    "tesseract",
			VisionLLMPrompt:   "X",
			Language:          "eng",
		})
		utils.HandleError(err)
		// imagePaths := []string{"/home/achno/Pictures/2024-07-19_23-17.png", "/home/achno/Pictures/2024-07-19_23-53.png"}
		imagePaths := []string{"/home/achno/Pictures/2024-07-19_23-17.png"}

		// load the images from the imagePaths
		utils.Spinner.Start()
		utils.Spinner.Message("uploading data")
		imgs := []image2.Image{}
		for _, path := range imagePaths {
			img, err := image.LoadImage(imageio.FileReader{Path: path})
			utils.HandleError(err)
			imgs = append(imgs, img)
		}
		res, err := n.OCRBatchImages(context.Background(), imgs)
		utils.HandleError(err)
		utils.Spinner.Stop()

		for _, item := range res {
			fmt.Println(item.Text)
			fmt.Println("###################")
		}

	},
}

func init() {
	rootCmd.AddCommand(ocrCmd)
}
