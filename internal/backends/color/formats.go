package color

import (
	"encoding/hex"
	"errors"
	"fmt"
	"image/color"
)

func HexToRGBA(hexStr string) (color.RGBA, error) {
	if len(hexStr) != 7 || hexStr[0] != '#' {
		return color.RGBA{}, errors.New("invalid hex color format")
	}
	bytes, err := hex.DecodeString(hexStr[1:])
	if err != nil {
		return color.RGBA{}, err
	}
	return color.RGBA{R: bytes[0], G: bytes[1], B: bytes[2], A: 255}, nil
}

func HexToRGBASlice(hexColors []string) ([]color.Color, error) {
	var rgbaColors []color.Color
	for _, hex := range hexColors {
		rgba, err := HexToRGBA(hex)
		if err != nil {
			return nil, err
		}
		rgbaColors = append(rgbaColors, rgba)
	}
	return rgbaColors, nil
}

func RGBtoHex(c color.RGBA) string {
	return fmt.Sprintf("#%02X%02X%02X", c.R, c.G, c.B)
}
