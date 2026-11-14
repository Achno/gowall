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

const (
	ModeNormal  = "normal"  // move every color toward the theme's colors
	ModeShading = "shading" // the theme's colors, but every pixel keeps its own lightness
	ModeMono    = "mono"    // one color in all its shades, every pixel keeps its own lightness
)

type ThemeConverter struct {
	Mode  string // ModeNormal when empty
	Color string // the hex color ModeMono uses

	// the CLUT of ModeShading and ModeMono, built once in memory and shared by every image of the conversion
	once sync.Once
	clut *image.RGBA
	err  error
}

func (themeConv *ThemeConverter) Process(img image.Image, theme string, format string) (image.Image, types.ImageMetadata, error) {
	level := 8

	if themeConv.Mode == ModeMono {
		clut, err := themeConv.buildOnce(func() (*image.RGBA, error) {
			seed, err := cpkg.HexToRGBA(themeConv.Color)
			if err != nil {
				return nil, fmt.Errorf("while parsing the mono color %s: %w", themeConv.Color, err)
			}
			return haldclut.MonoCLUT(seed, level), nil
		})
		if err != nil {
			return nil, types.ImageMetadata{}, err
		}
		return haldclut.ApplyCLUT(pixels.ToRGBA(img), clut, level), types.ImageMetadata{}, nil
	}

	selectedTheme, err := SelectTheme(theme)
	if err != nil {
		return nil, types.ImageMetadata{}, fmt.Errorf("%w %s", err, theme)
	}

	if config.GowallConfig.ColorCorrectionBackend == "nn" && themeConv.Mode != ModeShading {
		newimg, err := NearestNeighbour(img, selectedTheme)
		if err != nil {
			return nil, types.ImageMetadata{}, err
		}
		return newimg, types.ImageMetadata{}, nil
	}

	clut, err := GetOrCreateCLUT(theme, selectedTheme, level)
	if err != nil {
		return nil, types.ImageMetadata{}, err
	}
	if themeConv.Mode == ModeShading {
		clut, _ = themeConv.buildOnce(func() (*image.RGBA, error) { return haldclut.KeepShading(clut, level), nil })
	}
	rgba := pixels.ToRGBA(img)
	newImg := haldclut.ApplyCLUT(rgba, clut, level)

	return newImg, types.ImageMetadata{}, nil
}

// buildOnce runs build on the first call and returns its result on every call, so the CLUT of ModeShading or
// ModeMono is built once in memory and shared by every image of the conversion
func (themeConv *ThemeConverter) buildOnce(build func() (*image.RGBA, error)) (*image.RGBA, error) {
	themeConv.once.Do(func() { themeConv.clut, themeConv.err = build() })
	return themeConv.clut, themeConv.err
}

// GetOrCreateCLUT loads the theme's CLUT from the cluts folder, generating and saving it first if it isn't there
// The returned CLUT is shared, callers must not modify it
func GetOrCreateCLUT(theme string, selectedTheme Theme, level int) (*image.RGBA, error) {
	// hash colors to know if anything in the custom themes have changed
	clrs, err := GetThemeColors(theme)
	if err != nil {
		return nil, err
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
			return nil, fmt.Errorf("could not generate Identity CLUT")
		}
		mapper := &haldclut.RBFMapper{}
		palette, err := cpkg.ToRGBA(selectedTheme.Colors)
		if err != nil {
			clutMutex.Unlock()
			return nil, fmt.Errorf("could not parse colors to RGBA")
		}
		modifiedClut := haldclut.InterpolateCLUT(identityClut, palette, level, mapper)
		err = haldclut.SaveHaldCLUT(modifiedClut, filepath.Join(config.GowallConfig.OutputFolder, "cluts", clutPath))
		if err != nil {
			clutMutex.Unlock()
			return nil, fmt.Errorf("while saving the CLUT: %w", err)
		}

	}
	clutMutex.Unlock()

	clut, err := haldclut.LoadCachedHaldCLUT(filepath.Join(config.GowallConfig.OutputFolder, "cluts", clutPath))
	if err != nil {
		return nil, fmt.Errorf("while loading CLUT: %w", err)
	}
	if clut == nil {
		return nil, fmt.Errorf("CLUT is nil even though is was loaded")
	}
	return clut, nil
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
		straight := make([]byte, width*4) // the row's colors at full opacity
		for y := y0; y < y1; y++ {
			pixels.Unpremultiply(straight, src.Pix[y*src.Stride:y*src.Stride+width*4])
			dstRow := newImg.Pix[y*newImg.Stride : y*newImg.Stride+width*4]
			for i := 0; i < len(straight); i += 4 {
				// matched at full opacity, and the pixel keeps its own alpha
				a := straight[i+3]
				if a == 0 {
					continue
				}
				c := palette[nearestColor(straight[i], straight[i+1], straight[i+2], palette)]
				dstRow[i], dstRow[i+1], dstRow[i+2] = pixels.Premultiply(c.R, c.G, c.B, a)
				dstRow[i+3] = a
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
