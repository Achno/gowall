package color

import (
	"encoding/hex"
	"errors"
	"fmt"
	"image/color"
	"math"

	"github.com/lucasb-eyer/go-colorful"
	"github.com/muesli/gamut"
)

// HSL represents a color in Hue, Saturation, Lightness color space
type HSL struct {
	H float64 // Hue: 0-360
	S float64 // Saturation: 0-100
	L float64 // Lightness: 0-100
}

// LAB represents a color in CIE L*a*b* color space
type LAB struct {
	L float64 // Lightness: 0-100
	A float64 // Green-Red: -128 to 127
	B float64 // Blue-Yellow: -128 to 127
}

//-------------------            HexTo<Method> methods              ---------------------//

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

func HexToLAB(hex string) (LAB, error) {
	c, err := colorful.Hex(hex)
	if err != nil {
		return LAB{}, err
	}

	l, a, b := c.Lab()

	return LAB{
		L: l * 100,
		A: a * 100,
		B: b * 100,
	}, nil
}

func HexToHsl(hexStr string) (HSL, error) {
	c, err := colorful.Hex(hexStr)
	if err != nil {
		return HSL{}, err
	}

	h, s, l := c.Hsl()

	// Scale S and L to percentage range (0-100)
	return HSL{
		H: math.Round(h),
		S: math.Round(s * 100),
		L: math.Round(l * 100),
	}, nil
}

//-------------------            <Method>ToHex methods              ---------------------//

func RGBtoHex(c color.RGBA) string {
	return fmt.Sprintf("#%02X%02X%02X", c.R, c.G, c.B)
}

func ColorsToHex(colors []color.Color) []string {
	hexColors := make([]string, len(colors))
	for i, c := range colors {
		hexColors[i] = gamut.ToHex(c)
	}
	return hexColors
}

func LabToHex(lab LAB) string {
	c := colorful.Lab(lab.L/100, lab.A/100, lab.B/100)
	return c.Hex()
}

func HslToHex(hsl HSL) string {
	c := colorful.Hsl(hsl.H, hsl.S/100, hsl.L/100)
	return c.Hex()
}

// -------------------            Other methods              ---------------------//
func ToRGBA(clrs []color.Color) ([]color.RGBA, error) {
	rgbaColors := make([]color.RGBA, len(clrs))

	for i, c := range clrs {
		if rgba, ok := c.(color.RGBA); ok {
			rgbaColors[i] = rgba
		} else {
			return nil, fmt.Errorf("while converting theme color at index %d is not color.RGBA: %T", i, c)
		}
	}

	return rgbaColors, nil
}
