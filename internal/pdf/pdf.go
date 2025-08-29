package pdf

import (
	"fmt"
	"image"

	"github.com/gen2brain/go-fitz"
)

type ConvertOptions struct {

	// MaxPages limits the number of pages to convert (0 = all pages)
	MaxPages int

	// SkipFirstNPages skips the first N pages
	SkipFirstNPages int

	// DPI is the resolution of the image
	DPI float64
}

func DefaultOptions() ConvertOptions {
	return ConvertOptions{
		MaxPages:        0,
		SkipFirstNPages: 0,
		DPI:             120.0,
	}
}

// ConvertPDFToImages converts a PDF file to a []image.Image. MuPdf is not thread safe.
// There is no point in goroutines here.
func ConvertPDFToImages(pdf []byte, opts ConvertOptions) ([]image.Image, error) {

	doc, err := fitz.NewFromMemory(pdf)
	if err != nil {
		return nil, fmt.Errorf("failed to open PDF: %w", err)
	}
	defer doc.Close()

	// Get the number of pages
	pageCount := doc.NumPage()
	if pageCount == 0 {
		return nil, fmt.Errorf("PDF has no pages")
	}

	// Calculate which pages to process
	startPage := opts.SkipFirstNPages
	if startPage >= pageCount {
		return nil, fmt.Errorf("skip pages (%d) exceeds total pages (%d)", startPage, pageCount)
	}

	endPage := pageCount
	if opts.MaxPages > 0 {
		requestedEndPage := startPage + opts.MaxPages
		if requestedEndPage < endPage {
			endPage = requestedEndPage
		}
	}

	// Convert specified pages to images
	pagesToProcess := endPage - startPage
	images := make([]image.Image, 0, pagesToProcess)

	for i := startPage; i < endPage; i++ {
		img, err := doc.ImageDPI(i, opts.DPI)
		if err != nil {
			return nil, fmt.Errorf("failed to render page %d: %w", i+1, err)
		}

		images = append(images, img)
	}

	return images, nil
}
