package image

import (
	"fmt"
	"image"
	"image/color"

	"github.com/Achno/gowall/internal/backends/colorthief"
	"github.com/Achno/gowall/internal/logger"
)

type ExtractProcessor struct {
	NumOfColors int
}

func (e *ExtractProcessor) Process(img image.Image, theme string) (image.Image, error) {

	clr, err := colorthief.GetPalette(img, e.NumOfColors)
	if err != nil {
		return nil, err
	}

	for _, c := range clr {
		rgba, ok := c.(color.RGBA)

		if !ok {
			return nil, fmt.Errorf("while RGB casting")
		}
		logger.Print(RGBtoHex(rgba))
	}

	return nil, nil
}
