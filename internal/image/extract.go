package image

import (
	"fmt"
	"image"
	"image/color"

	"github.com/Achno/gowall/internal/backends/colorthief"
	"github.com/Achno/gowall/internal/logger"
	types "github.com/Achno/gowall/internal/types"
)

type ExtractProcessor struct {
	NumOfColors int
}

func (e *ExtractProcessor) Process(img image.Image, theme string, format string) (image.Image, types.ImageMetadata, error) {
	clr, err := colorthief.GetPalette(img, e.NumOfColors)
	if err != nil {
		return nil, types.ImageMetadata{}, err
	}

	for _, c := range clr {
		rgba, ok := c.(color.RGBA)

		if !ok {
			return nil, types.ImageMetadata{}, fmt.Errorf("while RGB casting")
		}
		logger.Print(RGBtoHex(rgba))
	}

	return nil, types.ImageMetadata{}, nil
}
