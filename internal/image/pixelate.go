package image

import (
	"image"
	"math"

	"github.com/Achno/gowall/internal/backends/pixels"
	types "github.com/Achno/gowall/internal/types"
)

type PixelateProcessor struct {
	Scale float64
}

func (p *PixelateProcessor) Process(img image.Image, theme string, format string) (image.Image, types.ImageMetadata, error) {

	bounds := img.Bounds()
	originalWidth := bounds.Dx()
	originalHeight := bounds.Dy()

	scaleFactor := float64(p.Scale) * 0.01

	downscaled := downscale(img, scaleFactor)

	// Upscale back to the original dimensions
	upscaled := upscale(downscaled, originalWidth, originalHeight)

	return upscaled, types.ImageMetadata{}, nil

}

func downscale(img image.Image, scale float64) *image.RGBA {

	src := pixels.ToRGBA(img)
	bounds := src.Bounds()
	width := int(math.Round(float64(bounds.Dx()) * scale))
	height := int(math.Round(float64(bounds.Dy()) * scale))

	newImage := image.NewRGBA(image.Rect(0, 0, width, height))

	// pick the nearest pixel for the new scaled image
	pixels.ParallelRows(height, func(y0, y1 int) {
		for y := y0; y < y1; y++ {
			origY := int(float64(y) / scale)
			for x := range width {
				origX := int(float64(x) / scale)
				pixels.CopyPixel(newImage, x, y, src, origX, origY)
			}
		}
	})

	return newImage
}

func upscale(img *image.RGBA, originalWidth, originalHeight int) image.Image {

	newImage := image.NewRGBA(image.Rect(0, 0, originalWidth, originalHeight))

	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	pixels.ParallelRows(originalHeight, func(y0, y1 int) {
		for y := y0; y < y1; y++ {
			// Get the pixel from the downscaled image
			origY := y * height / originalHeight
			for x := range originalWidth {
				origX := x * width / originalWidth
				pixels.CopyPixel(newImage, x, y, img, origX, origY)
			}
		}
	})

	return newImage
}
