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
	"github.com/Achno/gowall/internal/logger"
	"github.com/Achno/gowall/internal/pdf"
	"github.com/Achno/gowall/utils"

	"github.com/synoptiq/go-fluxus"
	"golang.org/x/time/rate"
)

// StartOCRPipeline orchestrates the OCR workflow. Accepts the an OCR provider and a list of imageIO operations.
func StartOCRPipeline(ops []imageio.ImageIO, provider OCRProvider) error {
	startTime := time.Now()
	logger.Printf("Starting OCR processing for %d files.\n", len(ops))

	// 1. Load files concurrently from imageIO operations : maintain order and mapping
	originalInputs, inputToOpsMapping, err := buildOCRInputsWithMapping(ops)
	if err != nil {
		return err
	}

	initialItems := make([]*PipelineItem, len(originalInputs))
	for i, input := range originalInputs {
		initialItems[i] = &PipelineItem{
			Input:         input,
			OriginalIndex: i,
			PageIndex:     -1,
		}
	}

	// 2. Run Pre-processing Pipeline
	pipelineStart := time.Now()
	utils.Spinner.Message("Pre-processing items...")
	processedItems, err := runPreprocessingPipeline(initialItems, provider)
	if err != nil {
		return fmt.Errorf("pre-processing pipeline failed: %w", err)
	}
	pipelineDuration := time.Since(pipelineStart)
	logger.Printf("Pre-processing pipeline completed in %v, produced %d items for OCR.\n", pipelineDuration, len(processedItems))

	// 3. Run OCR Batch Processing
	var limiter *rate.Limiter
	if rateLimited, ok := provider.(*RateLimitedProvider); ok {
		limiter = rateLimited.GetRateLimiter()
	}

	batchResults, err := ProcessBatch(context.Background(), processedItems, provider.OCR, limiter, 10)
	if err != nil {
		return fmt.Errorf("OCR batch processing failed: %w", err)
	}

	// 4. Stitch Results
	stitchStart := time.Now()
	logger.Printf("Stitching results...")
	finalResults := stitchPipelineResults(originalInputs, batchResults)
	stitchDuration := time.Since(stitchStart)
	logger.Printf("Result stitching completed in %v.\n", stitchDuration)

	// 5. Use the mapping to save the results to the correct files
	for i, item := range finalResults {
		if opsIndex, exists := inputToOpsMapping[i]; exists {
			logger.Printf("\n--- Result for: %s ---\n", originalInputs[i].Filename)
			if item != nil {
				// fmt.Println(item.Text)
				imageio.SaveText(item.Text, ops[opsIndex].ImageOutput)
				logger.Printf("Saved to %s\n", ops[opsIndex].ImageOutput)
			} else {
				logger.Printf("Processing failed for this file.")
			}
			logger.Printf("###################")
		}
	}

	totalDuration := time.Since(startTime)
	logger.Printf("\nTotal OCR processing completed in %v.\n", totalDuration)

	return nil
}

// runPreprocessingPipeline runs the given items through a pipeline of various stages.
func runPreprocessingPipeline(initialItems []*PipelineItem, provider OCRProvider) ([]*PipelineItem, error) {
	ctx := context.Background()

	pdfExpandStage := NewExpandSinglePdfStage(provider)
	grayScaleStage := NewGrayScaleStage(image.GrayScaleProcessor{})

	flattenStage := fluxus.StageFunc[[]*PipelineItem, []*PipelineItem](func(ctx context.Context, items []*PipelineItem) ([]*PipelineItem, error) {
		var result []*PipelineItem
		for _, item := range items {
			expanded, err := pdfExpandStage(ctx, item)
			if err != nil {
				return nil, err
			}
			result = append(result, expanded...)
		}
		return result, nil
	})

	grayScaleMap := fluxus.NewMap(grayScaleStage).WithConcurrency(5)

	pipeline := fluxus.NewPipeline(
		fluxus.Chain(
			flattenStage,
			grayScaleMap,
		),
	)

	result, err := pipeline.Process(ctx, initialItems)
	if err != nil {
		return nil, fmt.Errorf("pipeline processing failed: %w", err)
	}

	return result, nil
}

// NewExpandSinglePdfStage creates a PDF expansion stage with the given provider
func NewExpandSinglePdfStage(provider OCRProvider) fluxus.StageFunc[*PipelineItem, []*PipelineItem] {
	return func(ctx context.Context, item *PipelineItem) ([]*PipelineItem, error) {

		if item.Input.Type == InputTypePDF {
			if pdfCapable, ok := provider.(PDFCapable); !ok || !pdfCapable.SupportsPDF() {
				cfg := provider.GetConfig()

				images, err := pdf.ConvertPDFToImages(item.Input.PDFData, pdf.ConvertOptions{MaxPages: 0, SkipFirstNPages: 0, DPI: cfg.DPI})
				if err != nil {
					return []*PipelineItem{}, fmt.Errorf("expanding PDF stage failed to convert PDF '%s' to images: %w", item.Input.Filename, err)
				}

				var expandedItems []*PipelineItem
				totalPages := len(images)
				for pageIndex, img := range images {
					baseFilename := filepath.Base(item.Input.Filename)
					filename := fmt.Sprintf("%s-page-%d-of-%d", baseFilename, pageIndex+1, totalPages)

					imageInput := &OCRInput{
						Type:     InputTypeImage,
						Image:    img,
						Filename: filename,
					}
					expandedItems = append(expandedItems, &PipelineItem{
						Input:         imageInput,
						OriginalIndex: item.OriginalIndex,
						PageIndex:     pageIndex,
					})
				}
				return expandedItems, nil
			}
		}

		return []*PipelineItem{item}, nil
	}
}

// NewGrayScaleStage creates a grayscale processing stage with the given processor
func NewGrayScaleStage(processor image.GrayScaleProcessor) fluxus.StageFunc[*PipelineItem, *PipelineItem] {
	return func(ctx context.Context, item *PipelineItem) (*PipelineItem, error) {
		// ignore pdfs
		if item.Input.Image == nil {
			return item, nil
		}

		img, err := processor.Process(item.Input.Image, "")
		if err != nil {
			return item, fmt.Errorf("grayscale stage %s failed: %w", item.Input.Filename, err)
		}

		item.Input.Image = img
		return item, nil
	}
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
		// If single item (image or direct PDF), just take the result
		if len(group) == 1 {
			finalResults[originalIndex] = group[0].Result
			continue
		}

		// If  multi-page PDF, sort by page number and stitch
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
			if combinedMetadata == nil {
				combinedMetadata = br.Result.Metadata
			}
		}

		finalResults[originalIndex] = &OCRResult{
			Text:     strings.Join(textParts, "\n\n"), // <---  page break
			Images:   combinedImages,
			Metadata: combinedMetadata,
		}
	}

	return finalResults
}

// buildOCRInputsWithMapping loads files from the given imageIO operations and returns a list of OCRInputs and a mapping between the OCRInputs and the original imageIO operations.
func buildOCRInputsWithMapping(ops []imageio.ImageIO) ([]*OCRInput, map[int]int, error) {
	var inputs []*OCRInput
	inputToOpsMapping := make(map[int]int)
	var wg sync.WaitGroup
	var mu sync.Mutex

	for i, op := range ops {
		wg.Add(1)
		go func(i int, op imageio.ImageIO) {
			defer wg.Done()

			path := op.ImageInput.String()
			ext := strings.ToLower(filepath.Ext(path))
			var input *OCRInput

			switch ext {
			case ".pdf":
				pdf, err := imageio.LoadFileBytes(op.ImageInput)
				if err != nil {
					utils.HandleError(fmt.Errorf("failed to load PDF %s: %w", path, err))
					return
				}
				input = &OCRInput{Type: InputTypePDF, PDFData: pdf, Filename: path}
			case ".png", ".jpg", ".jpeg", ".webp":
				img, err := imageio.LoadImage(op.ImageInput)
				if err != nil {
					utils.HandleError(fmt.Errorf("failed to load image %s: %w", path, err))
					return
				}
				input = &OCRInput{Type: InputTypeImage, Image: img, Filename: path}
			}

			if input != nil {
				mu.Lock()
				inputIndex := len(inputs)
				inputs = append(inputs, input)
				inputToOpsMapping[inputIndex] = i
				mu.Unlock()
			}
		}(i, op)
	}

	wg.Wait()
	return inputs, inputToOpsMapping, nil
}
