package providers

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/Achno/gowall/utils"
	ext "github.com/reugn/go-streams/extension"
	"github.com/reugn/go-streams/flow"
	"golang.org/x/time/rate"
)

const defaultConcurrency = 10

// ocrFunc is a function type that matches the signature of a single OCR operation.
type ocrFunc func(ctx context.Context, input OCRInput) (*OCRResult, error)

// BatchResult holds the processed result and its corresponding input item,
// which is crucial for stitching results back together (e.g., PDF pages).
type BatchResult struct {
	Item   *PipelineItem
	Result *OCRResult
	Error  error
}

// ProcessBatch processes a collection of pipeline items concurrently using a go-streams pipeline.
// It offers high throughput via a worker pool and supports rate limiting.
func ProcessBatch(
	ctx context.Context,
	items []*PipelineItem, // Input items, potentially expanded from original files.
	ocrFunc ocrFunc, // The function to execute for each item, e.g., provider.OCR
	providerName string,
	limiter *rate.Limiter,
	concurrency int,
) ([]*BatchResult, error) {

	startTime := time.Now()
	totalItems := len(items)
	if totalItems == 0 {
		return []*BatchResult{}, nil
	}

	if concurrency <= 0 {
		concurrency = defaultConcurrency
	}

	fmt.Printf(
		"Starting batch OCR for %d items using provider '%s' with %d workers.\n",
		totalItems, providerName, concurrency,
	)

	// --- Progress Tracking ---
	var completed, failed atomic.Int64
	updateProgress := func() {
		c := completed.Load()
		f := failed.Load()
		processing := int64(totalItems) - c - f

		utils.Spinner.Message(fmt.Sprintf(
			"[%s] OCR Progress: %d computing, %d completed, %d failed (Total Items: %d)",
			providerName, processing, c, f, totalItems,
		))
	}
	updateProgress()

	// --- Pipeline Setup ---
	source := ext.NewChanSource(pipelineItemsToAnyChan(items))

	// The ocrFlow is a map operation that acts as our worker pool.
	ocrFlow := flow.NewMap(func(itemAny any) any {
		item := itemAny.(*PipelineItem)
		itemStart := time.Now()

		// 1. Rate Limiting: Blocks until the limiter allows proceeding.
		if limiter != nil {
			if err := limiter.Wait(ctx); err != nil {
				failed.Add(1)
				updateProgress()
				return &BatchResult{Item: item, Error: fmt.Errorf("rate limiter wait interrupted: %w", err)}
			}
		}

		// 2. OCR Execution: Calls the provider's OCR function.
		result, err := ocrFunc(ctx, *item.Input)

		// 3. Result Wrapping & Progress Update
		if err != nil {
			processingErr := fmt.Errorf("error processing '%s' with %s (took %v): %w",
				item.Input.Filename, providerName, time.Since(itemStart), err)
			failed.Add(1)
			updateProgress()
			return &BatchResult{Item: item, Error: processingErr}
		}

		completed.Add(1)
		updateProgress()
		return &BatchResult{Item: item, Result: result}

	}, concurrency)

	resultsChan := make(chan any, totalItems)
	sink := ext.NewChanSink(resultsChan)

	// --- Run Pipeline ---
	source.Via(ocrFlow).To(sink)

	// --- Collect Results & Errors ---
	finalResults := make([]*BatchResult, 0, totalItems)
	var allErrors []error
	for resAny := range resultsChan {
		if res, ok := resAny.(*BatchResult); ok {
			finalResults = append(finalResults, res)
			if res.Error != nil {
				allErrors = append(allErrors, res.Error)
			}
		}
	}

	totalDuration := time.Since(startTime)
	// Clear spinner and print final status
	utils.Spinner.Stop()
	fmt.Printf("\nBatch processing finished in %v. Completed: %d, Failed: %d\n", totalDuration, completed.Load(), failed.Load())

	var aggregatedError error
	if len(allErrors) > 0 {
		aggregatedError = errors.New(utils.FormatErrors(allErrors))
	}

	return finalResults, aggregatedError
}

// pipelineItemsToAnyChan converts a slice of PipelineItem pointers to a channel for the go-streams source.
func pipelineItemsToAnyChan(items []*PipelineItem) chan any {
	out := make(chan any, len(items))
	go func() {
		for _, item := range items {
			out <- item
		}
		close(out)
	}()
	return out
}
