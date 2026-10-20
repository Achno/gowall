package image

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"os"
	"path/filepath"
	"sync"

	"github.com/Achno/gowall/config"
	cpkg "github.com/Achno/gowall/internal/backends/color"
	haldclut "github.com/Achno/gowall/internal/backends/colorthief/haldClut"
	"github.com/Achno/gowall/internal/backends/pixels"
	types "github.com/Achno/gowall/internal/types"
)

var clutMutex sync.Mutex

type ThemeConverter struct{}

func (themeConv *ThemeConverter) Process(img image.Image, theme string, format string) (image.Image, types.ImageMetadata, error) {
	level := 8

	selectedTheme, err := SelectTheme(theme)
	if err != nil {
		return nil, types.ImageMetadata{}, fmt.Errorf("%w %s", err, theme)
	}

	// NearestNeighbour backend if specified in the config
	if config.GowallConfig.ColorCorrectionBackend == "nn" {
		newimg, err := NearestNeighbour(img, selectedTheme)
		if err != nil {
			return nil, types.ImageMetadata{}, err
		}
		return newimg, types.ImageMetadata{}, nil
	}

	// hash colors to know if anything in the custom themes have changed
	clrs, err := GetThemeColors(theme)
	if err != nil {
		return nil, types.ImageMetadata{}, err
	}
	hash := cpkg.HashPalette(clrs)
	clutPath := fmt.Sprintf("%s_%s.png", theme, hash)

	clutMutex.Lock()
	// if clut exists skip to save time
	_, err = os.Stat(filepath.Join(config.GowallConfig.OutputFolder, "cluts", clutPath))
	if os.IsNotExist(err) {

		identityClut, err := haldclut.GenerateIdentityCLUT(level)
		if err != nil {
			clutMutex.Unlock()
			return nil, types.ImageMetadata{}, fmt.Errorf("could not generate Identity CLUT")
		}
		mapper := &haldclut.RBFMapper{}
		palette, err := cpkg.ToRGBA(selectedTheme.Colors)
		if err != nil {
			clutMutex.Unlock()
			return nil, types.ImageMetadata{}, fmt.Errorf("could not parse colors to RGBA")
		}
		modifiedClut := haldclut.InterpolateCLUT(identityClut, palette, level, mapper)
		err = haldclut.SaveHaldCLUT(modifiedClut, filepath.Join(config.GowallConfig.OutputFolder, "cluts", clutPath))
		if err != nil {
			clutMutex.Unlock()
			return nil, types.ImageMetadata{}, fmt.Errorf("while saving the CLUT: %w", err)
		}

	}
	clutMutex.Unlock()

	clut, err := haldclut.LoadCachedHaldCLUT(filepath.Join(config.GowallConfig.OutputFolder, "cluts", clutPath))
	if err != nil {
		return nil, types.ImageMetadata{}, fmt.Errorf("while loading CLUT: %w", err)
	}
	if clut == nil {
		return nil, types.ImageMetadata{}, fmt.Errorf("CLUT is nil even though is was loaded")
	}
	rgba := pixels.ToRGBA(img)
	newImg := haldclut.ApplyCLUT(rgba, clut, level)

	return newImg, types.ImageMetadata{}, nil
}

func NearestNeighbour(img image.Image, theme Theme) (image.Image, error) {
	src := pixels.ToRGBA(img)
	newImg := image.NewRGBA(src.Rect)
	width, height := src.Rect.Dx(), src.Rect.Dy()

	palette := make([]color.RGBA, len(theme.Colors))
	for i, themeColor := range theme.Colors {
		palette[i] = cpkg.RGBA8(themeColor)
	}

	// replace each pixel with the selected theme's nearest color
	pixels.ParallelRows(height, func(y0, y1 int) {
		for y := y0; y < y1; y++ {
			srcRow := src.Pix[y*src.Stride : y*src.Stride+width*4]
			dstRow := newImg.Pix[y*newImg.Stride : y*newImg.Stride+width*4]
			for i := 0; i < len(srcRow); i += 4 {
				c := palette[nearestColor(srcRow[i], srcRow[i+1], srcRow[i+2], palette)]
				dstRow[i], dstRow[i+1], dstRow[i+2], dstRow[i+3] = c.R, c.G, c.B, c.A
			}
		}
	})

	return newImg, nil
}

// nearestColor returns the index of the palette color closest to r,g,b.
// It compares squared distances, the square root doesn't change which color is the closest.
func nearestColor(r, g, b uint8, palette []color.RGBA) int {
	nearest, minDist := 0, math.MaxInt
	for i, pc := range palette {
		dr, dg, db := int(pc.R)-int(r), int(pc.G)-int(g), int(pc.B)-int(b)
		if dist := dr*dr + dg*dg + db*db; dist < minDist {
			nearest, minDist = i, dist
		}
	}
	return nearest
}
