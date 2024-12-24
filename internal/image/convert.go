package image

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"math"
	"os"
	"path/filepath"

	"github.com/Achno/gowall/config"
	haldclut "github.com/Achno/gowall/internal/backends/colorthief/haldClut"
	"github.com/Achno/gowall/utils"
)

type ThemeConverter struct {
}

func (themeConv *ThemeConverter) Process(img image.Image, theme string) (image.Image, error) {
	level := 8

	selectedTheme, err := SelectTheme(theme)

	if err != nil {
		return nil, fmt.Errorf("%w %s", err, theme)
	}

	// NearestNeighbour backend if specified in the config
	if config.GowallConfig.ColorCorrectionBackend == "nn" {
		newimg, err := NearestNeighbour(img, selectedTheme)
		if err != nil {
			return nil, err
		}
		return newimg, nil
	}

	dirFolder, err := utils.CreateDirectory()
	if err != nil {
		return nil, fmt.Errorf("while creating Directory or getting path : %w", err)
	}

	// hash colors to know if anything in the custom themes have changed
	clrs, err := GetThemeColors(theme)
	if err != nil {
		return nil, err
	}
	hash := hashPalette(clrs)
	clutPath := fmt.Sprintf("%s_%s.png", theme, hash)

	// if clut exists skip to save time
	_, err = os.Stat(filepath.Join(dirFolder, "cluts", clutPath))
	if os.IsNotExist(err) {

		identityClut, err := haldclut.GenerateIdentityCLUT(level)
		if err != nil {
			return nil, fmt.Errorf("could not generate Identity CLUT")
		}
		mapper := &haldclut.RBFMapper{}
		palette, err := toRGBA(selectedTheme.Colors)

		if err != nil {
			return nil, fmt.Errorf("could not parse colors to RGBA")
		}
		modifiedClut := haldclut.InterpolateCLUT(identityClut, palette, level, mapper)
		haldclut.SaveHaldCLUT(modifiedClut, filepath.Join(dirFolder, "cluts", clutPath))
	}

	clut, _ := haldclut.LoadHaldCLUT(filepath.Join(dirFolder, "cluts", clutPath))
	bounds := img.Bounds()
	rgba := image.NewRGBA(bounds)
	draw.Draw(rgba, bounds, img, bounds.Min, draw.Src)
	newImg := haldclut.ApplyCLUT(rgba, clut, level)

	return newImg, nil
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

		distance := colorDistance(tr, tg, tb, r, g, b)

		if distance < minDist {
			minDist = distance
			nearestClr = themeColor
		}

	}

	return nearestClr

}

func colorDistance(r1, g1, b1, r2, g2, b2 uint32) float64 {
	return math.Sqrt(float64((r1-r2)*(r1-r2) + (g1-g2)*(g1-g2) + (b1-b2)*(b1-b2)))
}

func toRGBA(clrs []color.Color) ([]color.RGBA, error) {

	rgbaColors := make([]color.RGBA, len(clrs))

	for i, c := range clrs {
		if rgba, ok := c.(color.RGBA); ok {
			rgbaColors[i] = rgba
		} else {
			return nil, fmt.Errorf("while converting theme color at index %d is not color.RGBA: %T\n", i, c)
		}
	}

	return rgbaColors, nil
}

func hashPalette(colors []string) string {

	hasher := md5.New()
	for _, color := range colors {
		hasher.Write([]byte(color))
	}
	// shorten hash
	r := hex.EncodeToString(hasher.Sum(nil))[:16]
	return r
}
