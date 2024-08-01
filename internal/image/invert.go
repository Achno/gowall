package image

import (
	"image"
	"image/color"
)

type Inverter struct{}

func (inv *Inverter) Process(img image.Image, theme string) (image.Image, error) {
	newImg, err := invertImage(img)
	if err != nil {
		return nil, err
	}
	return newImg, nil
}

func invertImage(img image.Image) (image.Image, error) {
	bounds := img.Bounds()
	newImg := image.NewRGBA(bounds)

	// Replace each pixel with the inverted ones
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			originalColor := img.At(x, y)
			newColor := invertColor(originalColor)
			newImg.Set(x, y, newColor)
		}
	}

	return newImg, nil
}

// invertColor returns a new color with inverted RGB values, keeping the alpha channel the same
func invertColor(clr color.Color) color.Color {
	r, g, b, a := clr.RGBA()

	return color.RGBA{
		R: uint8(255 - r>>8), // Divide by 256 using bitwise shift for performance
		G: uint8(255 - g>>8),
		B: uint8(255 - b>>8),
		A: uint8(a >> 8),
	}
}
