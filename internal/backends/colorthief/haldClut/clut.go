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

// ApplyCLUT maps every pixel of img through the clut, img must start at (0,0).
// A color is looked up at full opacity and blended between the 8 cube cells around it, so a gradient
// stays smooth instead of stepping from cell to cell, and every pixel keeps its own alpha.
func ApplyCLUT(img *image.RGBA, clut *image.RGBA, level int) *image.RGBA {
	newImg := image.NewRGBA(img.Bounds())
	width, height := img.Bounds().Dx(), img.Bounds().Dy()

	// For every channel value: where the cube cell at or below it starts in clut.Pix, where the next cell
	// starts, and how far towards it the value is (0-255). A cell's offset is the sum of one entry per
	// channel, so the 8 cells around a color cost 8 additions, computed 256 times instead of per pixel.
	cubeSize := level * level
	var rLo, rHi, gLo, gHi, bLo, bHi [256]int
	var rF, gF, bF [256]int32
	base := clut.PixOffset(0, 0)
	for v := range 256 {
		pos := v * (cubeSize - 1) * 256 / 255
		lo, f := pos>>8, int32(pos&255)
		hi := min(lo+1, cubeSize-1)
		rLo[v], rHi[v], rF[v] = base+lo*4, base+hi*4, f
		gLo[v], gHi[v], gF[v] = (lo%level)*cubeSize*4+(lo/level)*clut.Stride, (hi%level)*cubeSize*4+(hi/level)*clut.Stride, f
		bLo[v], bHi[v], bF[v] = lo*level*clut.Stride, hi*level*clut.Stride, f
	}
	p := clut.Pix
	// x*(256-f) + y*f: f of the way from x to y, at 256x scale
	blend := func(x, y, f int32) int32 { return x*(256-f) + y*f }

	pixels.ParallelRows(height, func(y0, y1 int) {
		for y := y0; y < y1; y++ {
			srcRow := img.Pix[y*img.Stride : y*img.Stride+width*4]
			dstRow := newImg.Pix[y*newImg.Stride : y*newImg.Stride+width*4]
			for i := 0; i < len(srcRow); i += 4 {
				a := srcRow[i+3]
				if a == 0 {
					continue // NewRGBA is already transparent black
				}
				r, g, b := pixels.Straight(srcRow[i], srcRow[i+1], srcRow[i+2], a)
				fr, fg, fb := rF[r], gF[g], bF[b]
				c00, c10 := rLo[r]+gLo[g], rHi[r]+gLo[g]
				c01, c11 := rLo[r]+gHi[g], rHi[r]+gHi[g]
				b0, b1 := bLo[b], bHi[b]

				var out [3]uint8
				for ch := range 3 {
					// blend along r, then g (both at 256x scale), then b
					x0 := (blend(blend(int32(p[c00+b0+ch]), int32(p[c10+b0+ch]), fr), blend(int32(p[c01+b0+ch]), int32(p[c11+b0+ch]), fr), fg) + 128) >> 8
					x1 := (blend(blend(int32(p[c00+b1+ch]), int32(p[c10+b1+ch]), fr), blend(int32(p[c01+b1+ch]), int32(p[c11+b1+ch]), fr), fg) + 128) >> 8
					out[ch] = uint8((blend(x0, x1, fb) + 1<<15) >> 16)
				}
				dstRow[i], dstRow[i+1], dstRow[i+2] = pixels.Premultiply(out[0], out[1], out[2], a)
				dstRow[i+3] = a
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
