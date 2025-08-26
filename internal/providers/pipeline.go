package providers

import (
	"context"
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/Achno/gowall/internal/image"
	imageio "github.com/Achno/gowall/internal/image_io"
	"github.com/Achno/gowall/internal/logger"
	"github.com/Achno/gowall/utils"

	"github.com/synoptiq/go-fluxus"
)

// StartOCRPipeline orchestrates the OCR workflow. Accepts the an OCR provider and a list of imageIO operations.
func StartOCRPipeline(ops []imageio.ImageIO, service *ProviderService) error {
	config := service.GetConfig()

	// 1. Load files concurrently from imageIO operations : maintain order and mapping
	originalInputs, inputToOpsMapping, err := MapToOCRInput(ops)
	if err != nil {
		return fmt.Errorf("imageIO mapping failed: %w", err)
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
	progress := WithPrefixProgress(len(initialItems), "Pre-Processing...")
	progress.Start()

	processedItems, err := runPreprocessingPipeline(initialItems, service, progress)
	if err != nil {
		progress.Stop("Pre-Processing failed.")
		return fmt.Errorf("pre-processing failed: %w", err)
	}
	progress.Stop("Pre-Processing completed.")

	// 3. Run OCR Batch Processing
	ocrProgress := WithPrefixProgress(len(processedItems), "OCR Processing")
	batchResults, err := ProcessBatch(context.Background(), processedItems, service.OCR, config.Pipeline.OCRConcurrency, ocrProgress)
	if err != nil {
		return fmt.Errorf("OCR processing failed: %w", err)
	}

	// 4. Stitch Results
	finalResults := stitchPipelineResults(originalInputs, batchResults)

	// 5. Run optional Post-processing Pipeline
	if config.TextCorrection.Enabled {
		postProcessingProgress := WithPrefixProgress(len(finalResults), "Post-Processing")
		postProcessingProgress.Start()
		finalResults, err = runPostprocessingPipeline(finalResults, config, service, postProcessingProgress)
		if err != nil {
			return fmt.Errorf("post-processing failed: %w", err)
		}
		postProcessingProgress.Stop("Post-Processing completed.")
	}

	// 6. Use the mapping to save the results to the correct files
	for i, item := range finalResults {
		if opsIndex, exists := inputToOpsMapping[i]; exists {
			if item != nil {
				imageio.SaveText(item.Text, ops[opsIndex].ImageOutput)
				logger.Printf(fmt.Sprintf(utils.BlueColor+"● Saved to %s\n", ops[opsIndex].ImageOutput.String()+utils.ResetColor))

			}
		}
	}

	return nil
}

// runPreprocessingPipeline runs the given items through a pipeline of various stages.
func runPreprocessingPipeline(initialItems []*PipelineItem, service *ProviderService, progress *ProgressTracker) ([]*PipelineItem, error) {
	ctx := context.Background()
	config := service.GetConfig()

	pdfExpandStage := NewExpandSinglePdfStage(service)
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

		progress.SetTotal(int64(len(result)))
		return result, nil
	})

	grayScaleMap := fluxus.NewMap(grayScaleStage).WithConcurrency(config.Pipeline.Concurrency)
	lastStageMap := fluxus.NewMap(NewLastStageProgressStage(progress)).WithConcurrency(config.Pipeline.Concurrency)

	pipeline := fluxus.NewPipeline(
		fluxus.Chain(
			flattenStage,
			fluxus.Chain(
				grayScaleMap,
				lastStageMap,
			),
		),
	)

	result, err := pipeline.Process(ctx, initialItems)
	if err != nil {
		return nil, fmt.Errorf("pipeline: %w", err)
	}

	return result, nil
}

// runPostprocessingPipeline runs text correction on the stitched OCR results
func runPostprocessingPipeline(ocrResults []*OCRResult, config Config, service *ProviderService, progress *ProgressTracker) ([]*OCRResult, error) {
	ctx := context.Background()

	validResults := 0
	for _, result := range ocrResults {
		if result != nil {
			validResults++
		}
	}

	if validResults == 0 {
		return ocrResults, nil
	}

	textCorrectionStage := NewTextCorrectionStage(service.Complete, progress)

	textCorrectionMap := fluxus.NewMap(textCorrectionStage).
		WithConcurrency(config.Pipeline.Concurrency).
		WithCollectErrors(true)

	pipeline := fluxus.NewPipeline(textCorrectionMap)

	correctedResults, err := pipeline.Process(ctx, ocrResults)
	if err != nil {
		return nil, fmt.Errorf("pipeline: %w", err)
	}

	return correctedResults, nil
}

// stitchPipelineResults combines results from expanded items (like PDF pages) back into a single result per original file.
func stitchPipelineResults(originalInputs []*OCRInput, batchResults []*BatchResult) []*OCRResult {
	finalResults := make([]*OCRResult, len(originalInputs))
	resultGroups := make(map[int][]*BatchResult)
	pageBreaker := "\n\n --- \n\n"

	// Group successful results by their original file index
	for _, br := range batchResults {
		if br != nil && br.Result != nil {
			originalIndex := br.Item.OriginalIndex
			resultGroups[originalIndex] = append(resultGroups[originalIndex], br)
		}
	}

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
			Text:     strings.Join(textParts, pageBreaker),
			Images:   combinedImages,
			Metadata: combinedMetadata,
		}
	}

	return finalResults
}

// MapToOCRInput loads files from the given imageIO operations and returns a list of OCRInputs and a mapping between the OCRInputs and the original imageIO operations.
func MapToOCRInput(ops []imageio.ImageIO) ([]*OCRInput, map[int]int, error) {
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
					utils.HandleError(fmt.Errorf("loading PDF %s: %w", path, err))
					return
				}
				input = &OCRInput{Type: InputTypePDF, PDFData: pdf, Filename: path}
			case ".png", ".jpg", ".jpeg", ".webp":
				img, err := imageio.LoadImage(op.ImageInput)
				if err != nil {
					utils.HandleError(fmt.Errorf("loading image %s: %w", path, err))
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
