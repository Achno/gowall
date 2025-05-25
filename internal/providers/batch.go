package providers

import (
	"context"
	"errors"
	"fmt"
	"image"
	"strconv"
	"sync"

	"github.com/Achno/gowall/internal/logger"
	"github.com/Achno/gowall/utils"
)

// ocrFunc is a function type that matches the signature of a single OCR operation
type ocrFunc func(ctx context.Context, img image.Image) (*OCRResult, error)

func processBatchConcurrently(ctx context.Context, images []image.Image, singleOcr ocrFunc, providerName string) ([]*OCRResult, error) {
	wg := sync.WaitGroup{}
	results := make([]*OCRResult, len(images))
	errChan := make(chan error, len(images))

	for i, img := range images {
		wg.Add(1)
		go func(idx int, currentImage image.Image) {
			defer wg.Done()
			result, err := singleOcr(ctx, currentImage)
			if err != nil {
				errChan <- fmt.Errorf("error processing image %d with %s: %w", idx, providerName, err)
				return
			}
			results[idx] = result
			logger.Print(utils.BlueColor + " ➜ OCR Batch Image " + strconv.Itoa(idx) + " for " + providerName + " completed" + utils.ResetColor)
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
