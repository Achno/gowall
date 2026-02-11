package image

import (
	"fmt"
	"image"
	"image/color"

	cpkg "github.com/Achno/gowall/internal/backends/color"
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
	bounds := img.Bounds()
	newImg := image.NewRGBA(bounds)

	replacementMade := false

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			originalColor := img.At(x, y)
			if cpkg.ColorsAreSimilar(originalColor, from, threshold) {
				newImg.Set(x, y, to)
				replacementMade = true
			} else {
				newImg.Set(x, y, originalColor)
			}
		}
	}

	if !replacementMade {
		hex := cpkg.RGBtoHex(from.(color.RGBA))
		return nil, fmt.Errorf("the color : %s was not found in the image, nothing to replace", hex)
	}

	return newImg, nil
}
