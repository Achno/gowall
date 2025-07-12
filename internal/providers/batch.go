package providers

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/Achno/gowall/utils"
	"github.com/synoptiq/go-fluxus"
	"golang.org/x/time/rate"
)

const defaultConcurrency = 10

// ocrFunc is a function type that matches the signature of the OCR function of the provider.
type ocrFunc func(ctx context.Context, input OCRInput) (*OCRResult, error)

// BatchResult holds the processed result and its corresponding input item,
// used for stitching results back together in case of PDF files.
type BatchResult struct {
	Item   *PipelineItem
	Result *OCRResult
	Error  error
}

// ProcessBatch processes a collection of pipeline items concurrently using a go-fluxus pipeline,has progress tracking built in.
// 'concurrency' arg controls the throughput of the worker pool
// 'limiter' arg is used to rate limit the OCR calls
// 'ocrFunc' arg is a given function to execute for each item
func ProcessBatch(ctx context.Context, items []*PipelineItem, ocrFunc ocrFunc, limiter *rate.Limiter, concurrency int) ([]*BatchResult, error) {

	startTime := time.Now()
	totalItems := len(items)
	if totalItems == 0 {
		return []*BatchResult{}, nil
	}

	if concurrency <= 0 {
		concurrency = defaultConcurrency
	}

	// --- Progress Tracking ---
	var completed, failed atomic.Int64

	// Start a ticker for periodic progress updates
	progressTicker := time.NewTicker(500 * time.Millisecond)
	defer progressTicker.Stop()

	progressCtx, progressCancel := context.WithCancel(ctx)
	defer progressCancel()

	go func() {
		for {
			select {
			case <-progressCtx.Done():
				return
			case <-progressTicker.C:
				c := completed.Load()
				f := failed.Load()
				processing := int64(totalItems) - c - f

				utils.Spinner.Message(fmt.Sprintf("OCR Progress: %d computing, %d completed, %d failed (Total Items: %d)", processing, c, f, totalItems))
			}
		}
	}()

	// --- Create OCR Processing Stage ---
	// ocrStage := fluxus.StageFunc[*PipelineItem, *BatchResult](func(ctx context.Context, item *PipelineItem) (*BatchResult, error) {
	// 	itemStart := time.Now()

	// 	// 1. Rate Limit each call
	// 	if limiter != nil {
	// 		if err := limiter.Wait(ctx); err != nil {
	// 			failed.Add(1)
	// 			return &BatchResult{Item: item, Error: fmt.Errorf("rate limiter wait interrupted: %w", err)}, nil
	// 		}
	// 	}

	// 	// 2. Call the OCR function of the provider & update progress
	// 	result, err := ocrFunc(ctx, *item.Input)
	// 	if err != nil {
	// 		processingErr := fmt.Errorf("error processing '%s' (took %v): %w", item.Input.Filename, time.Since(itemStart), err)
	// 		failed.Add(1)
	// 		return &BatchResult{Item: item, Error: processingErr}, nil
	// 	}

	// 	completed.Add(1)
	// 	return &BatchResult{Item: item, Result: result}, nil
	// })

	ocrStage := NewSingleInputOCRStage(ocrFunc, limiter, &failed, &completed)

	// --- Create Map Stage for Concurrent Processing ---
	ocrMap := fluxus.NewMap(ocrStage).
		WithConcurrency(concurrency).
		WithCollectErrors(true) // Collect all errors instead of stopping at first error

	pipeline := fluxus.NewPipeline(ocrMap)

	finalResults, err := pipeline.Process(ctx, items)

	// Stop progress updates
	progressCancel()

	if err != nil {
		return nil, fmt.Errorf("pipeline execution failed: %w", err)
	}

	totalDuration := time.Since(startTime)
	utils.Spinner.Stop()
	fmt.Printf("\nBatch processing finished in %v. Completed: %d, Failed: %d\n", totalDuration, completed.Load(), failed.Load())

	// Collect errors from results
	var allErrors []error
	for _, res := range finalResults {
		if res.Error != nil {
			allErrors = append(allErrors, res.Error)
		}
	}

	var aggregatedError error
	if len(allErrors) > 0 {
		aggregatedError = errors.New(utils.FormatErrors(allErrors))
	}

	return finalResults, aggregatedError
}

func NewSingleInputOCRStage(ocrFunc ocrFunc, limiter *rate.Limiter, failed *atomic.Int64, completed *atomic.Int64) fluxus.StageFunc[*PipelineItem, *BatchResult] {
	return func(ctx context.Context, item *PipelineItem) (*BatchResult, error) {
		itemStart := time.Now()

		// 1. Rate Limit each call
		if limiter != nil {
			if err := limiter.Wait(ctx); err != nil {
				failed.Add(1)
				return &BatchResult{Item: item, Error: fmt.Errorf("rate limiter wait interrupted: %w", err)}, nil
			}
		}

		// 2. Call the OCR function of the provider & update progress
		result, err := ocrFunc(ctx, *item.Input)
		if err != nil {
			processingErr := fmt.Errorf("error processing '%s' (took %v): %w", item.Input.Filename, time.Since(itemStart), err)
			failed.Add(1)
			return &BatchResult{Item: item, Error: processingErr}, nil
		}

		completed.Add(1)
		return &BatchResult{Item: item, Result: result}, nil
	}
}
