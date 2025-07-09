package providers

import (
	"context"
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/Achno/gowall/internal/image"
	imageio "github.com/Achno/gowall/internal/image_io"
	"github.com/Achno/gowall/internal/pdf"
	"github.com/Achno/gowall/utils"

	ext "github.com/reugn/go-streams/extension"
	"github.com/reugn/go-streams/flow"
	"golang.org/x/time/rate"
)

// ProcessOCR orchestrates the entire OCR workflow for a set of input operations.
// It now takes a config object to manage concurrency and rate-limiting.
func ProcessOCR(ops []imageio.ImageIO, provider OCRProvider) error {
	startTime := time.Now()
	fmt.Printf("Starting OCR processing for %d files.\n", len(ops))

	// 1. Build initial inputs from file paths
	data, err := buildOCRInputs(ops)
	if err != nil {
		return err
	}
	originalInputs := channelToSlice(data)
	fmt.Printf("Built %d initial OCR inputs.\n", len(originalInputs))

	initialItems := make([]*PipelineItem, len(originalInputs))
	for i, input := range originalInputs {
		initialItems[i] = &PipelineItem{
			Input:         input,
			OriginalIndex: i,
			PageIndex:     -1, // -1 indicates it's the original item or a single-page doc
		}
	}

	// 2. Run Pre-processing Pipeline (PDF to Image conversion, Grayscale)
	pipelineStart := time.Now()
	fmt.Println("Starting pre-processing pipeline (PDF expansion + grayscale)...")
	processedItems := runPreprocessingPipeline(initialItems, provider)
	pipelineDuration := time.Since(pipelineStart)
	fmt.Printf("Pre-processing pipeline completed in %v, produced %d items for OCR.\n", pipelineDuration, len(processedItems))

	// 3. Run OCR Batch Processing
	// check if its of RateLimiterProvider type
	var limiter *rate.Limiter
	if rateLimited, ok := provider.(*RateLimitedProvider); ok {
		limiter = rateLimited.GetRateLimiter()
	}

	batchResults, err := ProcessBatch(
		context.Background(),
		processedItems,
		provider.OCR,
		"openrouter",
		limiter,
		10,
	)
	if err != nil {
		// ProcessBatch already logs detailed errors, so we can return a summary error.
		return fmt.Errorf("OCR batch processing failed: %w", err)
	}

	// 4. Stitch Results
	stitchStart := time.Now()
	fmt.Println("Stitching results...")
	finalResults := stitchPipelineResults(originalInputs, batchResults)
	stitchDuration := time.Since(stitchStart)
	fmt.Printf("Result stitching completed in %v.\n", stitchDuration)

	// 5. Output Results
	for i, item := range finalResults {
		fmt.Printf("\n--- Result for: %s ---\n", originalInputs[i].Filename)
		if item != nil {
			fmt.Println(item.Text)
			imageio.SaveText(item.Text, ops[i].ImageOutput)
			fmt.Println("Saved to", ops[i].ImageOutput)
		} else {
			fmt.Println("Processing failed for this file.")
		}
		fmt.Println("###################")
	}

	totalDuration := time.Since(startTime)
	fmt.Printf("\nTotal OCR processing completed in %v.\n", totalDuration)

	return nil
}

// runPreprocessingPipeline handles PDF-to-image and grayscale transformations.
func runPreprocessingPipeline(initialItems []*PipelineItem, provider OCRProvider) []*PipelineItem {
	out := make(chan any, len(initialItems)*10) // Buffer for expanded items
	source := ext.NewChanSource(pipelineItemsToAnyChan(initialItems))

	// Flow to expand PDFs into images if the provider doesn't support PDFs directly.
	pdfToImagesFlow := flow.NewFlatMap(func(item any) []any {
		pipelineItem := item.(*PipelineItem)

		// Check if it's a PDF and if the provider can handle it
		if pipelineItem.Input.Type == InputTypePDF {
			if pdfCapable, ok := provider.(PDFCapable); !ok || !pdfCapable.SupportsPDF() {
				fmt.Printf("Provider does not support PDF directly. Converting PDF to images: %s\n", pipelineItem.Input.Filename)
				images, err := pdf.ConvertPDFToImages(pipelineItem.Input.PDFData, pdf.DefaultOptions())
				if err != nil {
					utils.HandleError(fmt.Errorf("failed to convert PDF '%s' to images: %w", pipelineItem.Input.Filename, err))
					return []any{} // Skip this item on failure
				}

				var expandedItems []any
				for pageIndex, img := range images {
					imageInput := &OCRInput{
						Type:     InputTypeImage,
						Image:    img,
						Filename: fmt.Sprintf("%s-page-%d", filepath.Base(pipelineItem.Input.Filename), pageIndex+1),
					}
					expandedItems = append(expandedItems, &PipelineItem{
						Input:         imageInput,
						OriginalIndex: pipelineItem.OriginalIndex,
						PageIndex:     pageIndex,
					})
				}
				return expandedItems
			}
		}

		// Pass through non-PDFs or PDFs for capable providers
		return []any{pipelineItem}
	}, 5) // Concurrency for PDF conversion

	// Flow to convert images to grayscale
	gp := image.GrayScaleProcessor{}
	grayScaleFlow := flow.NewMap(func(item any) any {
		pipelineItem := item.(*PipelineItem)
		if pipelineItem.Input.Image == nil {
			return pipelineItem
		}

		img, err := gp.Process(pipelineItem.Input.Image, "")
		if err != nil {
			fmt.Printf("Grayscale processing failed for %s: %v\n", pipelineItem.Input.Filename, err)
			return pipelineItem // Return original on failure
		}

		pipelineItem.Input.Image = img
		return pipelineItem
	}, 5) // Concurrency for grayscale conversion

	sink := ext.NewChanSink(out)

	// Chain the flows together
	source.
		Via(pdfToImagesFlow).
		Via(grayScaleFlow).
		To(sink)

	return sinkToPipelineItems(out)
}

// stitchPipelineResults combines results from expanded items (like PDF pages) back into a single result per original file.
func stitchPipelineResults(originalInputs []*OCRInput, batchResults []*BatchResult) []*OCRResult {
	finalResults := make([]*OCRResult, len(originalInputs))
	resultGroups := make(map[int][]*BatchResult)

	// Group successful results by their original file index
	for _, br := range batchResults {
		if br != nil && br.Result != nil {
			originalIndex := br.Item.OriginalIndex
			resultGroups[originalIndex] = append(resultGroups[originalIndex], br)
		}
	}

	// Process each group
	for originalIndex, group := range resultGroups {
		// If it's a single item (image or direct PDF), just take the result
		if len(group) == 1 {
			finalResults[originalIndex] = group[0].Result
			continue
		}

		// If it's a multi-page PDF, sort by page number and stitch
		sort.Slice(group, func(i, j int) bool {
			return group[i].Item.PageIndex < group[j].Item.PageIndex
		})

		var textParts []string
		var combinedImages OCRImage
		var combinedMetadata map[string]string

		for _, br := range group {
			textParts = append(textParts, br.Result.Text)
			if len(br.Result.Images.MistralImages) > 0 {
				combinedImages.MistralImages = append(combinedImages.MistralImages, br.Result.Images.MistralImages...)
			}
			if combinedMetadata == nil { // Take metadata from the first page
				combinedMetadata = br.Result.Metadata
			}
		}

		finalResults[originalIndex] = &OCRResult{
			Text:     strings.Join(textParts, "\n\n"), // Page break
			Images:   combinedImages,
			Metadata: combinedMetadata,
		}
	}

	return finalResults
}

func buildOCRInputs(ops []imageio.ImageIO) (chan *OCRInput, error) {
	inputsChan := make(chan *OCRInput, len(ops))
	var wg sync.WaitGroup

	for _, op := range ops {
		wg.Add(1)
		go func(op imageio.ImageIO) {
			defer wg.Done()

			path := op.ImageInput.String()
			ext := strings.ToLower(filepath.Ext(path))
			var input *OCRInput

			switch ext {
			case ".pdf":
				pdf, err := imageio.LoadFileBytes(op.ImageInput)
				if err != nil {
					utils.HandleError(err)
					return
				}
				input = &OCRInput{Type: InputTypePDF, PDFData: pdf, Filename: path}
			case ".png", ".jpg", ".jpeg", ".webp":
				img, err := imageio.LoadImage(op.ImageInput)
				if err != nil {
					utils.HandleError(err)
					return
				}
				input = &OCRInput{Type: InputTypeImage, Image: img, Filename: path}
			}

			if input != nil {
				inputsChan <- input
			}
		}(op)
	}

	go func() {
		wg.Wait()
		close(inputsChan)
	}()

	return inputsChan, nil
}

// Helper function to convert channel to slice
func channelToSlice(inputChan chan *OCRInput) []*OCRInput {
	var inputs []*OCRInput
	for input := range inputChan {
		inputs = append(inputs, input)
	}
	return inputs
}

// Helper function to convert sink to pipeline items slice
func sinkToPipelineItems(sinkChan chan any) []*PipelineItem {
	var results []*PipelineItem
	for item := range sinkChan {
		if pipelineItem, ok := item.(*PipelineItem); ok && pipelineItem != nil {
			results = append(results, pipelineItem)
		}
	}
	return results
}
