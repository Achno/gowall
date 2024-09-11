package image

import (
	"image"
	"image/color"
	"image/draw"
)

type DrawProcessor struct {
	Color           color.RGBA
	BorderThickness int
}

func (b *DrawProcessor) Process(img image.Image, theme string) (image.Image, error) {

	newImg := drawBorder(img, b.BorderThickness, b.Color)

	return newImg, nil

}

func drawBorder(img image.Image, borderThickness int, borderColor color.Color) image.Image {
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// draw on new image
	newImg := image.NewRGBA(bounds)
	draw.Draw(newImg, bounds, img, image.Point{0, 0}, draw.Src)

	// top and bottom borders
	for x := 0; x < width; x++ {
		for t := 0; t < borderThickness; t++ {
			newImg.Set(x, t, borderColor)
			newImg.Set(x, height-borderThickness+t, borderColor)
		}
	}

	// left and right borders
	for y := 0; y < height; y++ {
		for t := 0; t < borderThickness; t++ {
			newImg.Set(t, y, borderColor)
			newImg.Set(width-borderThickness+t, y, borderColor)
		}
	}

	return newImg
}
