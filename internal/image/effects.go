package image

import (
	"fmt"
	"image"
	"image/color"
)

type FlipProcessor struct{}

func (p *FlipProcessor) Process(img image.Image, theme string) (image.Image, error) {

	bounds := img.Bounds()
	width, height := bounds.Dx(), bounds.Dy()
	newImg := image.NewRGBA(bounds)

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			pixel := img.At(width-x-1, y)
			newImg.Set(x, y, pixel)
		}
	}
	return newImg, nil
}

type MirrorProcessor struct{}

func (p *MirrorProcessor) Process(img image.Image, theme string) (image.Image, error) {

	bounds := img.Bounds()
	width, height := bounds.Dx(), bounds.Dy()
	newImg := image.NewRGBA(bounds)

	// Copy the original left half
	for y := 0; y < height; y++ {
		for x := 0; x < width/2; x++ {
			pixel := img.At(x, y)
			newImg.Set(x, y, pixel)
		}
	}

	// Mirror the left half onto the right half
	for y := 0; y < height; y++ {
		for x := 0; x < width/2; x++ {
			pixel := img.At(width/2-1-x, y)
			newImg.Set(width/2+x, y, pixel)
		}
	}
	return newImg, nil
}

type GrayScaleProcessor struct{}

func (p *GrayScaleProcessor) Process(img image.Image, theme string) (image.Image, error) {

	bounds := img.Bounds()
	grayImg := image.NewGray(bounds)

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			originalColor := img.At(x, y)
			r, g, b, _ := originalColor.RGBA()

			// luminosity formula
			grayValue := uint8((0.299*float64(r>>8) + 0.587*float64(g>>8) + 0.114*float64(b>>8)))

			grayImg.SetGray(x, y, color.Gray{Y: grayValue})
		}
	}
	return grayImg, nil
}

type BrightnessProcessor struct {
	Factor float64
}

func (p *BrightnessProcessor) Process(img image.Image, theme string) (image.Image, error) {

	if p.Factor <= 0.0 && p.Factor > 5 {
		return nil, fmt.Errorf("Enter a valid factor : from (0.0,5.0] ")
	}

	bounds := img.Bounds()
	newImg := image.NewRGBA(bounds)

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			origColor := color.RGBAModel.Convert(img.At(x, y)).(color.RGBA)

			newR := uint8(clamp(int(float64(origColor.R)*p.Factor), 0, 255))
			newG := uint8(clamp(int(float64(origColor.G)*p.Factor), 0, 255))
			newB := uint8(clamp(int(float64(origColor.B)*p.Factor), 0, 255))
			newA := origColor.A

			newImg.Set(x, y, color.RGBA{R: newR, G: newG, B: newB, A: newA})
		}
	}

	return newImg, nil
}

func clamp(val, min, max int) int {
	if val < min {
		return min
	} else if val > max {
		return max
	}
	return val
}
