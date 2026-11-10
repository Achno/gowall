package image

import (
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"io"
	"strings"

	cpkg "github.com/Achno/gowall/internal/backends/color"
	"github.com/Achno/gowall/internal/backends/colorthief"
	types "github.com/Achno/gowall/internal/types"
)

type ExtractProcessor struct {
	NumOfColors int
	JsonName    string
}

func (e *ExtractProcessor) Process(img image.Image, theme string, format string) (image.Image, types.ImageMetadata, error) {
	clr, err := colorthief.GetPalette(img, e.NumOfColors)
	if err != nil {
		return nil, types.ImageMetadata{}, err
	}

	hexes := make([]string, len(clr))
	for i, c := range clr {
		rgba, ok := c.(color.RGBA)
		if !ok {
			return nil, types.ImageMetadata{}, fmt.Errorf("while RGB casting")
		}
		hexes[i] = cpkg.RGBtoHex(rgba)
	}

	if e.JsonName == "" {
		fmt.Println(strings.Join(hexes, "\n"))
		return nil, types.ImageMetadata{}, nil
	}

	out := JsonTheme{Name: e.JsonName, Colors: hexes}
	return nil, types.ImageMetadata{EncoderFunction: func(w io.Writer, _ image.Image) error {
		enc := json.NewEncoder(w)
		enc.SetIndent("", "    ")
		return enc.Encode(out)
	}}, nil
}
