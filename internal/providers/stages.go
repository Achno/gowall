package providers

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	"github.com/Achno/gowall/internal/image"
	"github.com/Achno/gowall/internal/pdf"
	"github.com/synoptiq/go-fluxus"
)

// NewExpandSinglePdfStage creates a PDF expansion stage with the given provider
func NewExpandSinglePdfStage(service *ProviderService) fluxus.StageFunc[*PipelineItem, []*PipelineItem] {
	return func(ctx context.Context, item *PipelineItem) ([]*PipelineItem, error) {

		if item.Input.Type == InputTypePDF {
			if !service.SupportsPDF() {
				dpi := service.GetConfig().Pipeline.DPI

				images, err := pdf.ConvertPDFToImages(item.Input.PDFData, pdf.ConvertOptions{MaxPages: 0, SkipFirstNPages: 0, DPI: dpi})
				if err != nil {
					return []*PipelineItem{}, fmt.Errorf("expanding PDF stage > converting PDF '%s' to images: %w", item.Input.Filename, err)
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

		img, err := processor.Process(item.Input.Image, "", "")
		if err != nil {
			return item, fmt.Errorf("grayscale stage > %s : %w", item.Input.Filename, err)
		}

		item.Input.Image = img
		return item, nil
	}
}

func NewSingleInputOCRStageWithProgress(ocrFunc ocrFunc, progress *ProgressTracker) fluxus.StageFunc[*PipelineItem, *BatchResult] {
	return func(ctx context.Context, item *PipelineItem) (*BatchResult, error) {
		itemStart := time.Now()

		result, err := ocrFunc(ctx, *item.Input)
		if err != nil {
			processingErr := fmt.Errorf("ocr stage > '%s' (took %v): %w", item.Input.Filename, time.Since(itemStart), err)
			progress.IncrementFailed()
			return &BatchResult{Item: item, Error: processingErr}, nil
		}

		progress.IncrementCompleted()
		return &BatchResult{Item: item, Result: result}, nil
	}
}

// textCorrectionFunc is a function type that matches the signature of the Complete function of the text processor provider.
type textCorrectionFunc func(ctx context.Context, text string) (string, error)

// NewTextCorrectionStage creates a text correction stage for post-processing OCR results
func NewTextCorrectionStage(textCorrectionFunc textCorrectionFunc, progress *ProgressTracker) fluxus.StageFunc[*OCRResult, *OCRResult] {
	return func(ctx context.Context, result *OCRResult) (*OCRResult, error) {
		if result == nil {
			return result, nil
		}

		correctedText, err := textCorrectionFunc(ctx, result.Text)
		if err != nil {
			progress.IncrementFailed()
			return nil, fmt.Errorf("text correction stage > %w", err)
		}

		correctedResult := &OCRResult{
			Text:     correctedText,
			Images:   result.Images,
			Metadata: result.Metadata,
		}
		if correctedResult.Metadata == nil {
			correctedResult.Metadata = make(map[string]string)
		}
		correctedResult.Metadata["TextCorrectionApplied"] = "true"

		progress.IncrementCompleted()
		return correctedResult, nil
	}
}

func NewLastStageProgressStage(progress *ProgressTracker) fluxus.StageFunc[*PipelineItem, *PipelineItem] {
	return func(ctx context.Context, item *PipelineItem) (*PipelineItem, error) {
		progress.IncrementCompleted()
		return item, nil
	}
}
