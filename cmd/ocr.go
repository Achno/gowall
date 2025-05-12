/*
Copyright © 2025 Achno <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"fmt"

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
		fmt.Println("ocr called")
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
		n, err := providers.NewOCRProvider(providers.Config{
			VisionLLMProvider: "openrouter",
			VisionLLMModel:    "qwen/qwen2.5-vl-72b-instruct:free",
			VisionLLMPrompt:   "Extract all visible text from this image in english,Do not summarize, paraphrase, or infer missing text,Retain all spacing, punctuation, and formatting exactly as in the image,Include all text, even if it seems irrelevant or repeated.",
			// VisionLLMPrompt: "turn code to text",
		})
		utils.HandleError(err)

		img, err := image.LoadImage(imageio.FileReader{Path: args[0]})
		utils.HandleError(err)

		res, err := n.OCR(context.Background(), img)
		utils.HandleError(err)

		fmt.Println(res.Text)
		fmt.Println(res.Metadata)

		// img2, err := image.LoadImage(imageio.FileReader{Path: args[1]})

		// imgarray := []image2.Image{img, img2}
		// res := []string{}

		// var wg sync.WaitGroup
		// var mu sync.RWMutex

		// for i := 0; i < len(imgarray); i++ {
		// 	i := i
		// 	wg.Add(1)
		// 	go func() {
		// 		defer wg.Done()
		// 		result, err := n.OCR(context.Background(), imgarray[i])
		// 		utils.HandleError(err)

		// 		mu.Lock()
		// 		res = append(res, result.Text)
		// 		mu.Unlock()
		// 	}()
		// }
		// wg.Wait()

		// for _, item := range res {
		// 	fmt.Println(item)
		// }

	},
}

func init() {
	rootCmd.AddCommand(ocrCmd)
}
