package image

import (
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"math"
	"os"
	"path/filepath"
	"sync"

	"github.com/Achno/gowall/config"
	cpkg "github.com/Achno/gowall/internal/backends/color"
	haldclut "github.com/Achno/gowall/internal/backends/colorthief/haldClut"
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

	clut, err := haldclut.LoadHaldCLUT(filepath.Join(config.GowallConfig.OutputFolder, "cluts", clutPath))
	if err != nil {
		return nil, types.ImageMetadata{}, fmt.Errorf("while loading CLUT: %w", err)
	}
	if clut == nil {
		return nil, types.ImageMetadata{}, fmt.Errorf("CLUT is nil even though is was loaded")
	}
	bounds := img.Bounds()
	rgba := image.NewRGBA(bounds)
	draw.Draw(rgba, bounds, img, bounds.Min, draw.Src)
	newImg := haldclut.ApplyCLUT(rgba, clut, level)

	return newImg, types.ImageMetadata{}, nil
}

func NearestNeighbour(img image.Image, theme Theme) (image.Image, error) {
	bounds := img.Bounds()
	newImg := image.NewRGBA(bounds)

	// replace each pixel with the selected theme's nearest color
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			originalColor := img.At(x, y)
			newColor := nearestColor(originalColor, theme)
			newImg.Set(x, y, newColor)
		}
	}

	if newImg == nil {
		return nil, errors.New("error processing the Image")
	}

	return newImg, nil
}

func nearestColor(clr color.Color, theme Theme) color.Color {
	r, g, b, _ := clr.RGBA()

	// Convert from 16-bit to 8-bit
	r, g, b = r>>8, g>>8, b>>8

	minDist := math.MaxFloat64

	var nearestClr color.Color

	for _, themeColor := range theme.Colors {
		tr, tg, tb, _ := themeColor.RGBA()
		// Convert from 16-bit to 8-bit
		tr, tg, tb = tr>>8, tg>>8, tb>>8

		distance := cpkg.ColorDistance(tr, tg, tb, r, g, b)

		if distance < minDist {
			minDist = distance
			nearestClr = themeColor
		}

	}

	return nearestClr
}
