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
		// n, err := providers.NewOCRProvider(providers.Config{
		// 	VisionLLMProvider: "tesseract",
		// 	VisionLLMModel:    "tesseract",
		// 	VisionLLMPrompt:   "X",
		// 	Language:          "eng",
		// })
		//? Docling
		n, err := providers.NewOCRProvider(providers.Config{
			VisionLLMProvider: "docling",
			VisionLLMModel:    "easyocr",
			VisionLLMPrompt:   "X",
			Language:          "en",
		})
		utils.HandleError(err)
		// imagePaths := []string{"/home/achno/Pictures/2024-07-19_23-17.png", "/home/achno/Pictures/2024-07-19_23-53.png"}
		// imagePaths := []string{
		// 	"/home/achno/Pictures/2024-07-19_23-17.png",
		// 	"/home/achno/Pictures/2024-07-19_23-53.png",
		// 	"/home/achno/Pictures/Screenshots/Screenshot_20250410_232143.png",
		// 	"/home/achno/Pictures/Screenshots/Screenshot_20250412_181416.png",
		// 	"/home/achno/Pictures/Screenshots/Screenshot_20250412_181954.png",
		// 	"/home/achno/Pictures/Screenshots/Screenshot_20250412_220107.png",
		// 	"/home/achno/Pictures/Screenshots/Screenshot_20250412_220454.png",
		// 	"/home/achno/Pictures/Screenshots/Screenshot_20250412_220515.png",
		// 	"/home/achno/Pictures/Screenshots/Screenshot_20250412_220549.png",
		// 	"/home/achno/Pictures/Screenshots/Screenshot_20250412_235445.png",
		// 	"/home/achno/Pictures/Screenshots/Screenshot_20250413_191212.png",
		// 	"/home/achno/Pictures/Screenshots/Screenshot_20250415_190327.png",
		// 	"/home/achno/Pictures/Screenshots/Screenshot_20250415_200626.png",
		// 	"/home/achno/Pictures/Screenshots/Screenshot_20250416_194328.png",
		// 	"/home/achno/Pictures/Screenshots/Screenshot_20250416_202429.png",
		// 	"/home/achno/Pictures/Screenshots/Screenshot_20250418_184311.png",
		// 	"/home/achno/Pictures/Screenshots/Screenshot_20250418_190152.png",
		// 	"/home/achno/Pictures/Screenshots/Screenshot_20250418_201028.png",
		// 	"/home/achno/Pictures/Screenshots/Screenshot_20250419_173845.png",
		// 	"/home/achno/Pictures/Screenshots/Screenshot_20250419_173934.png",
		// 	"/home/achno/Pictures/Screenshots/Screenshot_20250419_181252.png",
		// 	"/home/achno/Pictures/Screenshots/Screenshot_20250419_183339.png",
		// 	"/home/achno/Pictures/Screenshots/Screenshot_20250419_191336.png",
		// 	"/home/achno/Pictures/Screenshots/Screenshot_20250419_191944.png",
		// 	"/home/achno/Pictures/Screenshots/Screenshot_20250419_194103.png",
		// 	"/home/achno/Pictures/Screenshots/Screenshot_20250420_162514.png",
		// 	"/home/achno/Pictures/Screenshots/Screenshot_20250420_174814.png",
		// 	"/home/achno/Pictures/Screenshots/Screenshot_20250420_183610.png",
		// 	"/home/achno/Pictures/Screenshots/Screenshot_20250420_200136.png",
		// 	"/home/achno/Pictures/Screenshots/Screenshot_20250421_181418.png",
		// 	"/home/achno/Pictures/Screenshots/Screenshot_20250422_182655.png",
		// 	"/home/achno/Pictures/Screenshots/Screenshot_20250422_184508.png",
		// 	"/home/achno/Pictures/Screenshots/Screenshot_20250422_191235.png",
		// 	"/home/achno/Pictures/Screenshots/Screenshot_20250422_193130.png",
		// 	"/home/achno/Pictures/Screenshots/Screenshot_20250422_210433.png",
		// 	"/home/achno/Pictures/Screenshots/Screenshot_20250423_153204.png",
		// 	"/home/achno/Pictures/Screenshots/Screenshot_20250423_153329.png",
		// 	"/home/achno/Pictures/Screenshots/Screenshot_20250424_193712.png",
		// 	"/home/achno/Pictures/Screenshots/Screenshot_20250425_113919.png",
		// 	"/home/achno/Pictures/Screenshots/Screenshot_20250425_123435.png",
		// 	"/home/achno/Pictures/Screenshots/Screenshot_20250425_125711.png",
		// 	"/home/achno/Pictures/Screenshots/Screenshot_20250425_154558.png",
		// 	"/home/achno/Pictures/Screenshots/Screenshot_20250425_155735.png",
		// 	"/home/achno/Pictures/Screenshots/Screenshot_20250425_172712.png",
		// 	"/home/achno/Pictures/Screenshots/Screenshot_20250425_181415.png",
		// 	"/home/achno/Pictures/Screenshots/Screenshot_20250425_185155.png",
		// 	"/home/achno/Pictures/Screenshots/Screenshot_20250425_190302.png",
		// 	"/home/achno/Pictures/Screenshots/Screenshot_20250425_191321.png",
		// 	"/home/achno/Pictures/Screenshots/Screenshot_20250425_192653.png",
		// 	"/home/achno/Pictures/Screenshots/Screenshot_20250428_163728.png",
		// 	"/home/achno/Pictures/Screenshots/Screenshot_20250501_193924.png",
		// 	"/home/achno/Pictures/Screenshots/Screenshot_20250501_194553.png",
		// 	"/home/achno/Pictures/Screenshots/Screenshot_20250502_214913.png",
		// 	"/home/achno/Pictures/Screenshots/Screenshot_20250502_225801.png",
		// 	"/home/achno/Pictures/Screenshots/Screenshot_20250504_165819.png",
		// 	"/home/achno/Pictures/Screenshots/Screenshot_20250504_171045.png",
		// 	"/home/achno/Pictures/Screenshots/Screenshot_20250504_171358.png",
		// 	"/home/achno/Pictures/Screenshots/Screenshot_20250504_223811.png",
		// 	"/home/achno/Pictures/Screenshots/Screenshot_20250505_160533.png",
		// 	"/home/achno/Pictures/Screenshots/Screenshot_20250506_232352.png",
		// 	"/home/achno/Pictures/Screenshots/Screenshot_20250507_000212.png",
		// 	"/home/achno/Pictures/Screenshots/Screenshot_20250507_000410.png",
		// 	"/home/achno/Pictures/Screenshots/Screenshot_20250507_001905.png",
		// 	"/home/achno/Pictures/Screenshots/Screenshot_20250507_162045.png",
		// 	"/home/achno/Pictures/Screenshots/Screenshot_20250507_162737.png",
		// 	"/home/achno/Pictures/Screenshots/Screenshot_20250507_164102.png",
		// 	"/home/achno/Pictures/Screenshots/Screenshot_20250507_165320.png",
		// 	"/home/achno/Pictures/Screenshots/Screenshot_20250507_170656.png",
		// 	"/home/achno/Pictures/Screenshots/Screenshot_20250507_175853.png",
		// 	"/home/achno/Pictures/Screenshots/Screenshot_20250507_185130.png",
		// 	"/home/achno/Pictures/Screenshots/Screenshot_20250507_193011.png",
		// 	"/home/achno/Pictures/Screenshots/Screenshot_20250507_193313.png",
		// 	"/home/achno/Pictures/Screenshots/Screenshot_20250507_194609.png",
		// 	"/home/achno/Pictures/Screenshots/Screenshot_20250508_212039.png",
		// 	"/home/achno/Pictures/Screenshots/Screenshot_20250508_213354.png",
		// 	"/home/achno/Pictures/Screenshots/Screenshot_20250508_214619.png",
		// 	"/home/achno/Pictures/Screenshots/Screenshot_20250508_221542.png",
		// 	"/home/achno/Pictures/Screenshots/Screenshot_20250508_222446.png",
		// 	"/home/achno/Pictures/Screenshots/Screenshot_20250508_224406.png",
		// 	"/home/achno/Pictures/Screenshots/Screenshot_20250508_225858.png",
		// 	"/home/achno/Pictures/Screenshots/Screenshot_20250513_200223.png",
		// 	"/home/achno/Pictures/Screenshots/Screenshot_20250513_200316.png",
		// 	"/home/achno/Pictures/Screenshots/Screenshot_20250513_200637.png",
		// 	"/home/achno/Pictures/Screenshots/Screenshot_20250513_202542.png",
		// 	"/home/achno/Pictures/Screenshots/Screenshot_20250513_205604.png",
		// 	"/home/achno/Pictures/Screenshots/Screenshot_20250514_000904.png",
		// 	"/home/achno/Pictures/Screenshots/Screenshot_20250514_152800.png",
		// 	"/home/achno/Pictures/Screenshots/Screenshot_20250514_152927.png",
		// 	"/home/achno/Pictures/Screenshots/Screenshot_20250514_155420.png",
		// 	"/home/achno/Pictures/Screenshots/Screenshot_20250514_164134.png",
		// 	"/home/achno/Pictures/Screenshots/Screenshot_20250514_173804.png",
		// 	"/home/achno/Pictures/Screenshots/Screenshot_20250514_174146.png",
		// 	"/home/achno/Pictures/Screenshots/Screenshot_20250515_193452.png",
		// 	"/home/achno/Pictures/Screenshots/Screenshot_20250515_223201.png",
		// 	"/home/achno/Pictures/Screenshots/Screenshot_20250515_223450.png",
		// 	"/home/achno/Pictures/Screenshots/Screenshot_20250516_171035.png",
		// 	"/home/achno/Pictures/Screenshots/Screenshot_20250516_171939.png",
		// 	"/home/achno/Pictures/Screenshots/Screenshot_20250516_223214.png",
		// 	"/home/achno/Pictures/Screenshots/Screenshot_20250518_152824.png",
		// 	"/home/achno/Pictures/Screenshots/Screenshot_20250518_190920.png",
		// 	"/home/achno/Pictures/Screenshots/Screenshot_20250519_132935.png",
		// 	"/home/achno/Pictures/Screenshots/Screenshot_20250523_191645.png",
		// 	"/home/achno/Pictures/Screenshots/Screenshot_20250524_190202.png",
		// 	"/home/achno/Pictures/Screenshots/Screenshot_20250525_195612.png",
		// 	"/home/achno/Pictures/Screenshots/Screenshot_20250525_213135.png",
		// 	"/home/achno/Pictures/Screenshots/Screenshot_20250526_235047.png",
		// 	"/home/achno/Pictures/Screenshots/Screenshot_20250527_161732.png",
		// 	"/home/achno/Pictures/Screenshots/Screenshot_20250527_172823.png",
		// 	"/home/achno/Pictures/Screenshots/Screenshot_20250527_180246.png",
		// 	"/home/achno/Pictures/Screenshots/Screenshot_20250527_182017.png",
		// 	"/home/achno/Pictures/Screenshots/Screenshot_20250528_005937.png",
		// 	"/home/achno/Pictures/Screenshots/Screenshot_20250528_130341.png",
		// 	"/home/achno/Pictures/Screenshots/Screenshot_20250528_131949.png",
		// 	"/home/achno/Pictures/Screenshots/Screenshot_20250528_143226.png",
		// 	"/home/achno/Pictures/Screenshots/Screenshot_20250528_143301.png",
		// 	"/home/achno/Pictures/Screenshots/Screenshot_20250529_183224.png",
		// 	"/home/achno/Pictures/Screenshots/Screenshot_20250529_192026.png",
		// 	"/home/achno/Pictures/Screenshots/Screenshot_20250529_221123.png",
		// 	"/home/achno/Pictures/Screenshots/Screenshot_20250530_000416.png",
		// }

		imagePaths := []string{"/home/achno/Pictures/2024-07-19_23-17.png"}
		// imagePaths := []string{"/home/achno/Pictures/Screenshots/Screenshot_20250528_005937.png", "/home/achno/Pictures/Screenshots/Screenshot_20250530_000416.png"}

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
