package haldclut

import (
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"
	"sync"

	"github.com/Achno/gowall/internal/backends/pixels"
)

// CLUTs decoded by LoadCachedHaldCLUT, by file path
var (
	cacheMutex sync.Mutex
	cache      = map[string]*image.RGBA{}
)

// Interface for all the possible interpolation and mapping algorithms
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

// LoadCachedHaldCLUT is LoadHaldCLUT but every path is decoded only once, so a batch doesn't decode the same CLUT for every image.
// It doesn't notice if the file changes afterwards, and the returned CLUT is shared so callers must not modify it.
func LoadCachedHaldCLUT(filePath string) (*image.RGBA, error) {
	cacheMutex.Lock()
	defer cacheMutex.Unlock()

	if clut, ok := cache[filePath]; ok {
		return clut, nil
	}
	clut, err := LoadHaldCLUT(filePath)
	if err != nil {
		return nil, err
	}
	cache[filePath] = clut
	return clut, nil
}

// ApplyCLUT maps every pixel of img through the clut, img must start at (0,0)
func ApplyCLUT(img *image.RGBA, clut *image.RGBA, level int) *image.RGBA {
	newImg := image.NewRGBA(img.Bounds())
	width, height := img.Bounds().Dx(), img.Bounds().Dy()

	// the cube coordinate of every channel value, computed 256 times instead of 3 times per pixel
	cubeSize := level * level
	var cube [256]int
	for v := range cube {
		cube[v] = v * (cubeSize - 1) / 255
	}

	pixels.ParallelRows(height, func(y0, y1 int) {
		for y := y0; y < y1; y++ {
			srcRow := img.Pix[y*img.Stride : y*img.Stride+width*4]
			dstRow := newImg.Pix[y*newImg.Stride : y*newImg.Stride+width*4]
			for i := 0; i < len(srcRow); i += 4 {
				r, g, b := cube[srcRow[i]], cube[srcRow[i+1]], cube[srcRow[i+2]]
				clutX := (r % cubeSize) + (g%level)*cubeSize
				clutY := (b * level) + (g / level)

				c := clut.PixOffset(clutX, clutY)
				copy(dstRow[i:i+4], clut.Pix[c:c+4])
				if srcRow[i+3] == 0 {
					dstRow[i+3] = 0
				}
			}
		}
	})
	return newImg
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
