package image

import (
	"fmt"
	"image"
	"math"

	types "github.com/Achno/gowall/internal/types"
)

type PixelateProcessor struct {
	Scale float64
}

func (p *PixelateProcessor) Process(img image.Image, theme string, format string) (image.Image, types.ImageMetadata, error) {

	// check if scale is valid
	if p.Scale < 1 || p.Scale > 25 {
		return nil, types.ImageMetadata{}, fmt.Errorf("scale must be between 1 and 25")
	}

	bounds := img.Bounds()
	originalWidth := bounds.Dx()
	originalHeight := bounds.Dy()

	scaleFactor := float64(p.Scale) * 0.01

	downscaled := downscale(img, scaleFactor)

	// Upscale back to the original dimensions
	upscaled := upscale(downscaled, originalWidth, originalHeight)

	return upscaled, types.ImageMetadata{}, nil

}

func downscale(img image.Image, scale float64) image.Image {

	bounds := img.Bounds()
	width := int(math.Round(float64(bounds.Dx()) * scale))
	height := int(math.Round(float64(bounds.Dy()) * scale))

	newImage := image.NewRGBA(image.Rect(0, 0, width, height))

	// pick the nearest pixel for the new scaled image
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			origX := int(float64(x) / scale)
			origY := int(float64(y) / scale)
			newImage.Set(x, y, img.At(origX, origY))
		}
	}

	return newImage
}

func upscale(img image.Image, originalWidth, originalHeight int) image.Image {

	newImage := image.NewRGBA(image.Rect(0, 0, originalWidth, originalHeight))

	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	for y := 0; y < originalHeight; y++ {
		for x := 0; x < originalWidth; x++ {
			// Get the pixel from the downscaled image
			origX := x * width / originalWidth
			origY := y * height / originalHeight
			newImage.Set(x, y, img.At(origX, origY))
		}
	}

	return newImage
}
