package haldclut

import (
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"
	"sync"
)

// Interface for all the possible interpolation algorithms
type Mapperfunc interface {
	Map(color.RGBA, []color.RGBA) color.RGBA
}

func GenerateIdentityCLUT(level int) (*image.RGBA, error) {
	cubeSize := level * level
	imageSize := cubeSize * level

	clut := image.NewRGBA(image.Rect(0, 0, imageSize, imageSize))

	// Generate the identity CLUT
	index := 0
	for blue := 0; blue < cubeSize; blue++ {
		b := uint8(blue * 255 / (cubeSize - 1))
		for green := 0; green < cubeSize; green++ {
			g := uint8(green * 255 / (cubeSize - 1))
			for red := 0; red < cubeSize; red++ {
				r := uint8(red * 255 / (cubeSize - 1))

				x := index % imageSize
				y := index / imageSize
				clut.SetRGBA(x, y, color.RGBA{R: r, G: g, B: b, A: 255})
				index++
			}
		}
	}

	return clut, nil
}

func SaveHaldCLUT(clut *image.RGBA, filePath string) error {
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()
	return png.Encode(file, clut)
}

func LoadHaldCLUT(filePath string) (*image.RGBA, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	img, err := png.Decode(file)
	if err != nil {
		return nil, err
	}

	bounds := img.Bounds()
	clut := image.NewRGBA(bounds)
	draw.Draw(clut, bounds, img, bounds.Min, draw.Src)
	return clut, nil
}

func ApplyCLUT(img *image.RGBA, clut *image.RGBA, level int) *image.RGBA {
	// cubeSize := level * level
	newImg := image.NewRGBA(img.Bounds())

	for y := 0; y < img.Bounds().Dy(); y++ {
		for x := 0; x < img.Bounds().Dx(); x++ {
			original := img.RGBAAt(x, y)

			clutX, clutY := correctPixel(original, level)

			mappedColor := clut.RGBAAt(clutX, clutY)
			newImg.SetRGBA(x, y, mappedColor)
		}
	}
	return newImg
}

func correctPixel(original color.RGBA, level int) (int, int) {
	cubeSize := level * level

	r := int(original.R) * (cubeSize - 1) / 255
	g := int(original.G) * (cubeSize - 1) / 255
	b := int(original.B) * (cubeSize - 1) / 255

	x := (r % cubeSize) + (g%level)*cubeSize
	y := (b * level) + (g / level)

	return x, y
}

func InterpolateCLUT(identityClut *image.RGBA, palette []color.RGBA, level int, mapper Mapperfunc) *image.RGBA {
	bounds := identityClut.Bounds()
	newClut := image.NewRGBA(bounds)

	wg := sync.WaitGroup{}

	chunkSize := 128 // goroutines on 128x128 chunks
	numChunksY := (bounds.Dy() + chunkSize - 1) / chunkSize
	numChunksX := (bounds.Dx() + chunkSize - 1) / chunkSize

	for i := 0; i < numChunksY; i++ {
		for j := 0; j < numChunksX; j++ {
			startY := i * chunkSize
			endY := min(startY+chunkSize, bounds.Max.Y)
			startX := j * chunkSize
			endX := min(startX+chunkSize, bounds.Max.X)

			wg.Add(1)
			go func(startX, endX, startY, endY int) {
				defer wg.Done()
				for y := startY; y < endY; y++ {
					for x := startX; x < endX; x++ {
						originalColor := identityClut.RGBAAt(x, y)
						interpolatedColor := mapper.Map(originalColor, palette)
						newClut.SetRGBA(x, y, interpolatedColor)
					}
				}
			}(startX, endX, startY, endY)
		}
	}
	wg.Wait()
	return newClut
}
