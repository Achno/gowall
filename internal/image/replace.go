package image

import (
	"fmt"
	"image"
	"image/color"
	"math"

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
			if colorsAreSimilar(originalColor, from, threshold) {
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

// Helper function to check if two colors are similar within a threshold
func colorsAreSimilar(c1, c2 color.Color, threshold float64) bool {
	r1, g1, b1, _ := c1.RGBA()
	r2, g2, b2, _ := c2.RGBA()

	// Normalize to 8-bit values
	r1, g1, b1 = r1>>8, g1>>8, b1>>8
	r2, g2, b2 = r2>>8, g2>>8, b2>>8

	// Euclidean distance
	distance := math.Sqrt(
		math.Pow(float64(r1)-float64(r2), 2) +
			math.Pow(float64(g1)-float64(g2), 2) +
			math.Pow(float64(b1)-float64(b2), 2),
	)

	return distance <= threshold
}
