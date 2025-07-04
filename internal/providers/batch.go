package providers

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/Achno/gowall/internal/pdf"
	"github.com/Achno/gowall/utils"
	"golang.org/x/time/rate"
)

const defaultImageBatchSize = 3

// ocrFunc is a function type that matches the signature of a single OCR operation
type ocrFunc func(ctx context.Context, img OCRInput) (*OCRResult, error)

// processBatchConcurrently processes a single chunk of OCR tasks.
// It does not fail fast; it attempts all tasks and returns partial results
// Returns a slice of OCRResults with nils for failures and an aggregated error.
func processBatchConcurrently(ctx context.Context, images []OCRInput, singleOcr ocrFunc, providerName string, limiter *rate.Limiter) ([]*OCRResult, error) {
	wg := sync.WaitGroup{}
	results := make([]*OCRResult, len(images))

	// This implementation collects all errors and returns them together at the end.
	var mu sync.Mutex
	var errs []error

	var completed, failed int64
	total := len(images)

	utils.Spinner.Message(fmt.Sprintf("%d/%d computing, 0/%d completed, 0/%d failed", total, total, total, total))

	for i, img := range images {
		wg.Add(1)
		go func(idx int, currentImage OCRInput) {
			defer wg.Done()
			if limiter != nil {
				if err := limiter.Wait(ctx); err != nil {
					mu.Lock()
					errs = append(errs, fmt.Errorf("rate limiter wait interrupted: %w", err))
					mu.Unlock()
					return
				}
			}
			result, err := singleOcr(ctx, currentImage)

			if err != nil {
				processingErr := fmt.Errorf("error processing image %s with %s: %w", currentImage.Filename, providerName, err)
				mu.Lock()
				errs = append(errs, processingErr)
				mu.Unlock()

				failedCount := atomic.AddInt64(&failed, 1)
				completedCount := atomic.LoadInt64(&completed)
				computing := total - int(completedCount) - int(failedCount)
				utils.Spinner.Message(fmt.Sprintf("%d/%d computing, %d/%d completed, %d/%d failed", computing, total, completedCount, total, failedCount, total))
				return
			}
			results[idx] = result
			completedCount := atomic.AddInt64(&completed, 1)
			failedCount := atomic.LoadInt64(&failed)
			computing := total - int(completedCount) - int(failedCount)
			utils.Spinner.Message(fmt.Sprintf("%d/%d computing, %d/%d completed, %d/%d failed", computing, total, completedCount, total, failedCount, total))
		}(i, img)
	}

	wg.Wait()

	if len(errs) > 0 {
		return results, errors.New(utils.FormatErrors(errs))
	}

	return results, nil
}

func ProcessBatchWithPDFFallback(ctx context.Context, provider OCRProvider, singleOcr ocrFunc, inputs []OCRInput, providerName string, batchSize int, limiter *rate.Limiter) ([]*OCRResult, error) {
	if batchSize <= 0 {
		batchSize = defaultImageBatchSize
	}

	// 1. Expand inputs and create a map to track stitching.
	expandedInputs := make([]OCRInput, 0, len(inputs))
	stitchingMap := make([]int, 0, len(inputs))

	for originalIndex, input := range inputs {
		if input.Type == InputTypePDF {
			pdfCapable, ok := provider.(PDFCapable)
			if !ok || !pdfCapable.SupportsPDF() {
				images, err := pdf.ConvertPDFToImages(input.PDFData, pdf.DefaultOptions())
				if err != nil {
					return nil, fmt.Errorf("failed to convert PDF '%s' to images: %w", input.Filename, err)
				}
				for i, img := range images {
					imageInput := OCRInput{
						Type:     InputTypeImage,
						Image:    img,
						Filename: fmt.Sprintf("%s-page-%d", input.Filename, i+1),
					}
					expandedInputs = append(expandedInputs, imageInput)
					stitchingMap = append(stitchingMap, originalIndex)
				}
				continue
			}
		}
		expandedInputs = append(expandedInputs, input)
		stitchingMap = append(stitchingMap, originalIndex)
	}

	// 2. Process all expanded inputs in chunks.
	allExpandedResults := make([]*OCRResult, 0, len(expandedInputs))
	var allErrors []error

	for i := 0; i < len(expandedInputs); i += batchSize {
		end := min(i+batchSize, len(expandedInputs))
		batchChunk := expandedInputs[i:end]

		results, err := processBatchConcurrently(ctx, batchChunk, singleOcr, providerName, limiter)
		allExpandedResults = append(allExpandedResults, results...)
		if err != nil {
			allErrors = append(allErrors, err)
		}
	}

	// 3. Stitch the results back together.
	finalResults := stitchResults(inputs, allExpandedResults, stitchingMap)

	if len(allErrors) > 0 {
		return finalResults, errors.New(utils.FormatErrors(allErrors))
	}

	return finalResults, nil
}

// stitchResults combines results from expanded PDFs back into single OCRResult objects.
// It takes the original inputs, the flat list of processed page results, and a map
// to link them, returns a result list that matches the original input order.
func stitchResults(originalInputs []OCRInput, expandedResults []*OCRResult, stitchingMap []int) []*OCRResult {
	finalResults := make([]*OCRResult, len(originalInputs))
	textBuilders := make([]*strings.Builder, len(originalInputs))
	for i := range textBuilders {
		textBuilders[i] = &strings.Builder{}
	}

	for i, singleResult := range expandedResults {
		if singleResult == nil {
			continue // Skip failed pages.
		}

		originalIndex := stitchingMap[i]

		if finalResults[originalIndex] == nil {
			finalResults[originalIndex] = &OCRResult{
				Metadata: singleResult.Metadata,
				Images:   singleResult.Images,
			}
		}

		// Page break
		if textBuilders[originalIndex].Len() > 0 {
			textBuilders[originalIndex].WriteString("\n")
		}
		textBuilders[originalIndex].WriteString(singleResult.Text)
	}

	for i, builder := range textBuilders {
		if finalResults[i] != nil {
			finalResults[i].Text = builder.String()
		}
	}

	return finalResults
}
