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

const defaultImageBatchSize = 10

// ocrFunc is a function type that matches the signature of a single OCR operation
type ocrFunc func(ctx context.Context, img OCRInput) (*OCRResult, error)

// expandedItem tracks the mapping between original inputs and expanded items
type expandedItem struct {
	input         OCRInput // The original input either a PDF or an image
	originalIndex int
	pageIndex     int // -1 for non-PDF items
}

// processBatchConcurrently processes a batch of OCR tasks concurrently
func processBatchConcurrently(ctx context.Context, items []expandedItem, singleOcr ocrFunc, providerName string, limiter *rate.Limiter, originalTotal int) ([]*OCRResult, error) {
	results := make([]*OCRResult, len(items))
	var wg sync.WaitGroup
	var mu sync.Mutex
	var errs []error

	var completed, failed int64

	// Track completion per original item
	originalCompleted := make(map[int]bool)
	originalFailed := make(map[int]bool)

	updateProgress := func() {
		mu.Lock()
		defer mu.Unlock()

		// Count unique original items that are completed or failed
		completedCount := len(originalCompleted)
		failedCount := len(originalFailed)
		computing := originalTotal - completedCount - failedCount

		utils.Spinner.Message(fmt.Sprintf("%d/%d computing, %d/%d completed, %d/%d failed",
			computing, originalTotal, completedCount, originalTotal, failedCount, originalTotal))
	}

	updateProgress() // Initial message

	for i, item := range items {
		wg.Add(1)
		go func(idx int, currentItem expandedItem) {
			defer wg.Done()

			// Rate limiting
			if limiter != nil {
				if err := limiter.Wait(ctx); err != nil {
					mu.Lock()
					errs = append(errs, fmt.Errorf("rate limiter wait interrupted: %w", err))
					originalFailed[currentItem.originalIndex] = true
					mu.Unlock()
					updateProgress()
					return
				}
			}

			result, err := singleOcr(ctx, currentItem.input)
			if err != nil {
				processingErr := fmt.Errorf("error processing %s with %s: %w",
					currentItem.input.Filename, providerName, err)

				mu.Lock()
				errs = append(errs, processingErr)
				originalFailed[currentItem.originalIndex] = true
				mu.Unlock()

				atomic.AddInt64(&failed, 1)
				updateProgress()
				return
			}

			results[idx] = result
			atomic.AddInt64(&completed, 1)

			// Mark original as completed only when we have a successful result
			mu.Lock()
			originalCompleted[currentItem.originalIndex] = true
			mu.Unlock()

			updateProgress()
		}(i, item)
	}

	wg.Wait()

	var aggregatedError error
	if len(errs) > 0 {
		aggregatedError = errors.New(utils.FormatErrors(errs))
	}

	return results, aggregatedError
}

func ProcessBatchWithPDFFallback(ctx context.Context, provider OCRProvider, singleOcr ocrFunc, inputs []OCRInput, providerName string, limiter *rate.Limiter) ([]*OCRResult, error) {
	batchSize := defaultImageBatchSize

	originalTotal := len(inputs)

	// 1. Expand inputs and create mapping
	var expandedItems []expandedItem

	for originalIndex, input := range inputs {
		if input.Type == InputTypePDF {
			pdfCapable, ok := provider.(PDFCapable)
			if !ok || !pdfCapable.SupportsPDF() {
				// Convert PDF to images
				images, err := pdf.ConvertPDFToImages(input.PDFData, pdf.DefaultOptions())
				if err != nil {
					return nil, fmt.Errorf("failed to convert PDF '%s' to images: %w", input.Filename, err)
				}

				// Add each page as a separate item
				for pageIndex, img := range images {
					imageInput := OCRInput{
						Type:     InputTypeImage,
						Image:    img,
						Filename: fmt.Sprintf("%s-page-%d", input.Filename, pageIndex+1),
					}
					expandedItems = append(expandedItems, expandedItem{
						input:         imageInput,
						originalIndex: originalIndex,
						pageIndex:     pageIndex,
					})
				}
				continue
			}
		}

		expandedItems = append(expandedItems, expandedItem{
			input:         input,
			originalIndex: originalIndex,
			pageIndex:     -1, // Not a PDF page
		})
	}

	// 2. Process all expanded items in batches
	var allResults []*OCRResult
	var allErrors []error

	for i := 0; i < len(expandedItems); i += batchSize {
		end := min(i+batchSize, len(expandedItems))
		batchItems := expandedItems[i:end]

		results, err := processBatchConcurrently(ctx, batchItems, singleOcr, providerName, limiter, originalTotal)
		allResults = append(allResults, results...)
		if err != nil {
			allErrors = append(allErrors, err)
		}
	}

	// 3. Stitch results back together
	finalResults := stitchResults(inputs, allResults, expandedItems)

	// Return combined error if any
	var finalError error
	if len(allErrors) > 0 {
		finalError = errors.New(utils.FormatErrors(allErrors))
	}

	return finalResults, finalError
}

// stitchResults combines results from expanded items back into original structure
func stitchResults(originalInputs []OCRInput, expandedResults []*OCRResult, expandedItems []expandedItem) []*OCRResult {
	finalResults := make([]*OCRResult, len(originalInputs))

	// Group results by original index
	resultGroups := make(map[int][]*OCRResult)
	pageGroups := make(map[int][]int) // Track page indices for proper ordering

	for i, result := range expandedResults {
		if result == nil {
			continue // Skip failed items
		}

		item := expandedItems[i]
		originalIndex := item.originalIndex

		if resultGroups[originalIndex] == nil {
			resultGroups[originalIndex] = make([]*OCRResult, 0)
			pageGroups[originalIndex] = make([]int, 0)
		}

		resultGroups[originalIndex] = append(resultGroups[originalIndex], result)
		pageGroups[originalIndex] = append(pageGroups[originalIndex], item.pageIndex)
	}

	// Combine results for each original input
	for originalIndex, results := range resultGroups {
		if len(results) == 0 {
			continue
		}

		// For single results (non-PDF or direct PDF processing)
		if len(results) == 1 {
			finalResults[originalIndex] = results[0]
			continue
		}

		// For multiple results (expanded PDF), stitch them together
		var textParts []string
		var combinedImages OCRImage
		var combinedMetadata map[string]string

		// Sort results by page index to ensure correct order
		pageIndices := pageGroups[originalIndex]
		sortedResults := make([]*OCRResult, len(results))

		for i, pageIndex := range pageIndices {
			if pageIndex >= 0 && pageIndex < len(results) {
				sortedResults[pageIndex] = results[i]
			}
		}

		// Handle unsorted results (fallback)
		sortedIndex := 0
		for i, result := range sortedResults {
			if result == nil && sortedIndex < len(results) {
				sortedResults[i] = results[sortedIndex]
				sortedIndex++
			}
		}

		// Combine all text and metadata
		for _, result := range sortedResults {
			if result == nil {
				continue
			}

			textParts = append(textParts, result.Text)

			// Combine images
			if len(result.Images.MistralImages) > 0 {
				combinedImages.MistralImages = append(combinedImages.MistralImages, result.Images.MistralImages...)
			}

			// Use metadata from first successful result
			if combinedMetadata == nil {
				combinedMetadata = result.Metadata
			}
		}

		// Create final combined result
		finalResults[originalIndex] = &OCRResult{
			Text:     strings.Join(textParts, "\n\n"), // Use double newline for page breaks
			Images:   combinedImages,
			Metadata: combinedMetadata,
		}
	}

	return finalResults
}
