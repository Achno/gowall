package image

import (
	"fmt"
	"image"
	"image/color"
	"sync/atomic"

	cpkg "github.com/Achno/gowall/internal/backends/color"
	"github.com/Achno/gowall/internal/backends/pixels"
	types "github.com/Achno/gowall/internal/types"
)

type ReplaceProcessor struct {
	FromColor string
	ToColor   string
	Threshold float64
}

func (r *ReplaceProcessor) Process(img image.Image, theme string, format string) (image.Image, types.ImageMetadata, error) {

	from, err := cpkg.HexToRGBA(r.FromColor)

	if err != nil {
		return nil, types.ImageMetadata{}, err
	}

	to, err := cpkg.HexToRGBA(r.ToColor)

	if err != nil {
		return nil, types.ImageMetadata{}, err
	}
	newimage, err := replaceColor(img, from, to, r.Threshold)

	if err != nil {
		return nil, types.ImageMetadata{}, fmt.Errorf("replacing color failed : %w", err)
	}

	return newimage, types.ImageMetadata{}, nil
}

// replaces every pixel from the "from" color over to the "to" color in the image
func replaceColor(img image.Image, from, to color.Color, threshold float64) (image.Image, error) {
	src := pixels.ToRGBA(img)
	newImg := image.NewRGBA(src.Rect)
	width, height := src.Rect.Dx(), src.Rect.Dy()

	fromRGBA := cpkg.RGBA8(from)
	toRGBA := cpkg.RGBA8(to)

	var replacementMade atomic.Bool

	pixels.ParallelRows(height, func(y0, y1 int) {
		replaced := false
		for y := y0; y < y1; y++ {
			srcRow := src.Pix[y*src.Stride : y*src.Stride+width*4]
			dstRow := newImg.Pix[y*newImg.Stride : y*newImg.Stride+width*4]
			for i := 0; i < len(srcRow); i += 4 {
				originalColor := color.RGBA{srcRow[i], srcRow[i+1], srcRow[i+2], srcRow[i+3]}
				blendWeight := cpkg.ColorSimilarityWeight(originalColor, fromRGBA, threshold)
				if blendWeight > 0 {
					c := cpkg.BlendColorRGBA(originalColor, toRGBA, blendWeight)
					dstRow[i], dstRow[i+1], dstRow[i+2], dstRow[i+3] = c.R, c.G, c.B, c.A
					replaced = true
				} else {
					copy(dstRow[i:i+4], srcRow[i:i+4])
				}
			}
		}
		if replaced {
			replacementMade.Store(true)
		}
	})

	if !replacementMade.Load() {
		hex := cpkg.RGBtoHex(from.(color.RGBA))
		return nil, fmt.Errorf("the color : %s was not found in the image, nothing to replace", hex)
	}

	return newImg, nil
}
