package providers

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/Achno/gowall/utils"
)

// ocrFunc is a function type that matches the signature of a single OCR operation
type ocrFunc func(ctx context.Context, img OCRInput) (*OCRResult, error)

func processBatchConcurrently(ctx context.Context, images []OCRInput, singleOcr ocrFunc, providerName string) ([]*OCRResult, error) {
	wg := sync.WaitGroup{}
	results := make([]*OCRResult, len(images))
	errChan := make(chan error, len(images))

	var completed, failed int64
	total := len(images)

	utils.Spinner.Message(fmt.Sprintf("%d/%d computing, 0/%d completed, 0/%d failed", total, total, total, total))

	for i, img := range images {

		wg.Add(1)
		go func(idx int, currentImage OCRInput) {
			defer wg.Done()
			result, err := singleOcr(ctx, currentImage)

			if err != nil {
				errChan <- fmt.Errorf("error processing image %s with %s: %w", currentImage.Filename, providerName, err)
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
	close(errChan)

	var errs []error
	for err := range errChan {
		errs = append(errs, err)
	}

	if len(errs) > 0 {
		return results, errors.New(utils.FormatErrors(errs))
	}

	return results, nil
}
