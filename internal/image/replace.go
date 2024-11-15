package image

import (
	"fmt"
	"image"
	"image/color"
)

type ReplaceProcessor struct {
	FromColor string
	ToColor   string
}

func (r *ReplaceProcessor) Process(img image.Image, theme string) (image.Image, error) {

	from, err := HexToRGBA(r.FromColor)

	if err != nil {
		return nil, err
	}

	to, err := HexToRGBA(r.ToColor)

	if err != nil {
		return nil, err
	}

	newimage, err := replaceColor(img, from, to)

	if err != nil {
		return nil, fmt.Errorf("replacing color failed : %w", err)
	}

	return newimage, nil
}

// replaces every pixel from the "from" color over to the "to" color in the image
func replaceColor(img image.Image, from, to color.Color) (image.Image, error) {
	bounds := img.Bounds()
	newImg := image.NewRGBA(bounds)

	// checks if the "from" color exists anywhere in the image
	replacementMade := false

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			originalColor := img.At(x, y)
			if originalColor == from {
				newImg.Set(x, y, to)
				replacementMade = true
			} else {
				newImg.Set(x, y, originalColor)
			}
		}
	}

	if !replacementMade {
		hex := RGBtoHex(from.(color.RGBA))
		return nil, fmt.Errorf("the color : %s was not found in the image,nothing to replace", hex)
	}

	return newImg, nil
}
