package providers

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Achno/gowall/internal/logger"
	"github.com/Achno/gowall/utils"
	"github.com/synoptiq/go-fluxus"
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
func ProcessBatch(ctx context.Context, items []*PipelineItem, ocrFunc ocrFunc, concurrency int, progress *ProgressTracker) ([]*BatchResult, error) {

	startTime := time.Now()
	totalItems := len(items)
	if totalItems == 0 {
		return []*BatchResult{}, nil
	}

	if concurrency <= 0 {
		concurrency = defaultConcurrency
	}

	// Set up progress tracker for OCR processing
	progress.SetTotal(int64(totalItems))
	progress.SetPrefix("OCR Processing")
	progress.Start()

	ocrStage := NewSingleInputOCRStageWithProgress(ocrFunc, progress)

	// --- Create Map Stage for Concurrent Processing ---
	ocrMap := fluxus.NewMap(ocrStage).
		WithConcurrency(concurrency).
		WithCollectErrors(true) // Collect all errors instead of stopping at first error

	pipeline := fluxus.NewPipeline(ocrMap)

	finalResults, err := pipeline.Process(ctx, items)

	if err != nil {
		progress.Stop("OCR Processing failed.")
		return nil, fmt.Errorf("pipeline execution failed: %w", err)
	}

	totalDuration := time.Since(startTime)
	_, completed, failed, _ := progress.GetCounters()
	progress.Stop("OCR Processing completed.")
	logger.Printf("\nBatch processing finished in %v. Completed: %d, Failed: %d\n", totalDuration, completed, failed)

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
