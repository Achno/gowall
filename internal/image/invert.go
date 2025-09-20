package image

import (
	"errors"
	"image"
	"image/color"

	types "github.com/Achno/gowall/internal/types"
)

type Inverter struct {
}

func (Invrt *Inverter) Process(img image.Image, theme string, format string) (image.Image, types.ImageMetadata, error) {

	newImg, err := invertImage(img)

	if err != nil {
		return nil, types.ImageMetadata{}, err
	}

	return newImg, types.ImageMetadata{}, nil
}

func invertImage(img image.Image) (image.Image, error) {
	bounds := img.Bounds()
	newImg := image.NewRGBA(bounds)

	// replace each pixel with the inverted ones
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			originalColor := img.At(x, y)
			newColor := invertColor(originalColor)
			newImg.Set(x, y, newColor)
		}
	}

	if newImg == nil {
		return nil, errors.New("error processing the Image")
	}

	return newImg, nil

}

// You can invert a color
func invertColor(clr color.Color) color.Color {
	r, g, b, a := clr.RGBA()

	return color.RGBA{
		R: uint8(255 - r/257),
		G: uint8(255 - g/257),
		B: uint8(255 - b/257),
		A: uint8(a / 257),
	}

}
