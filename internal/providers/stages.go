package providers

import (
	"context"
	"fmt"
	"path/filepath"
	"sync/atomic"
	"time"

	"github.com/Achno/gowall/internal/image"
	"github.com/Achno/gowall/internal/pdf"
	"github.com/synoptiq/go-fluxus"
	"golang.org/x/time/rate"
)

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
