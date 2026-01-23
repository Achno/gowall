package color

import (
	"encoding/hex"
	"errors"
	"fmt"
	"image/color"
	"math"

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
	c, err := HexToRGBA(hex)
	if err != nil {
		return LAB{}, err
	}

	// Convert RGB to 0-1 range
	rf := float64(c.R) / 255.0
	gf := float64(c.G) / 255.0
	bf := float64(c.B) / 255.0

	// Apply gamma correction (sRGB to linear RGB)
	rf = gammaToLinear(rf)
	gf = gammaToLinear(gf)
	bf = gammaToLinear(bf)

	// Convert to XYZ (using D65 illuminant)
	x := rf*0.4124564 + gf*0.3575761 + bf*0.1804375
	y := rf*0.2126729 + gf*0.7151522 + bf*0.0721750
	z := rf*0.0193339 + gf*0.1191920 + bf*0.9503041

	// Normalize for D65 white point
	x /= 0.95047
	y /= 1.00000
	z /= 1.08883

	// Convert XYZ to LAB
	x = xyzToLab(x)
	y = xyzToLab(y)
	z = xyzToLab(z)

	l := 116.0*y - 16.0
	a := 500.0 * (x - y)
	bVal := 200.0 * (y - z)

	return LAB{
		L: l,
		A: a,
		B: bVal,
	}, nil
}

func HexToHsl(hexStr string) (HSL, error) {
	rgba, err := HexToRGBA(hexStr)
	if err != nil {
		return HSL{}, err
	}

	r := float64(rgba.R) / 255.0
	g := float64(rgba.G) / 255.0
	b := float64(rgba.B) / 255.0

	max := math.Max(math.Max(r, g), b)
	min := math.Min(math.Min(r, g), b)
	delta := max - min

	l := (max + min) / 2.0

	var s float64
	if delta == 0 {
		s = 0
	} else {
		if l < 0.5 {
			s = delta / (max + min)
		} else {
			s = delta / (2.0 - max - min)
		}
	}

	var h float64
	if delta == 0 {
		h = 0
	} else {
		switch max {
		case r:
			h = ((g - b) / delta)
			if g < b {
				h += 6.0
			}
		case g:
			h = ((b - r) / delta) + 2.0
		case b:
			h = ((r - g) / delta) + 4.0
		}
		h *= 60.0
	}

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
	// Convert LAB to XYZ
	fy := (lab.L + 16.0) / 116.0
	fx := lab.A/500.0 + fy
	fz := fy - lab.B/200.0

	// Reverse the LAB transformation
	var x, y, z float64
	if math.Pow(fx, 3) > 0.008856 {
		x = math.Pow(fx, 3)
	} else {
		x = (fx - 16.0/116.0) / 7.787
	}
	if math.Pow(fy, 3) > 0.008856 {
		y = math.Pow(fy, 3)
	} else {
		y = (fy - 16.0/116.0) / 7.787
	}
	if math.Pow(fz, 3) > 0.008856 {
		z = math.Pow(fz, 3)
	} else {
		z = (fz - 16.0/116.0) / 7.787
	}

	// Denormalize for D65 white point
	x *= 0.95047
	y *= 1.00000
	z *= 1.08883

	// Convert XYZ to RGB
	r := x*3.2404542 + y*-1.5371385 + z*-0.4985314
	g := x*-0.9692660 + y*1.8760108 + z*0.0415560
	b := x*0.0556434 + y*-0.2040259 + z*1.0572252

	// Apply gamma correction (linear to sRGB)
	r = linearToGamma(r)
	g = linearToGamma(g)
	b = linearToGamma(b)

	// Clamp to [0, 1] and convert to 0-255
	r = math.Max(0, math.Min(1, r))
	g = math.Max(0, math.Min(1, g))
	b = math.Max(0, math.Min(1, b))

	rgba := color.RGBA{
		R: uint8(math.Round(r * 255)),
		G: uint8(math.Round(g * 255)),
		B: uint8(math.Round(b * 255)),
		A: 255,
	}

	return RGBtoHex(rgba)
}

func HslToHex(hsl HSL) string {
	h := hsl.H
	s := hsl.S / 100.0
	l := hsl.L / 100.0

	var r, g, b float64

	if s == 0 {
		// Achromatic (gray)
		r = l
		g = l
		b = l
	} else {
		var q float64
		if l < 0.5 {
			q = l * (1 + s)
		} else {
			q = l + s - l*s
		}
		p := 2*l - q

		r = hueToRGB(p, q, h+120)
		g = hueToRGB(p, q, h)
		b = hueToRGB(p, q, h-120)
	}

	rgba := color.RGBA{
		R: uint8(math.Round(r * 255)),
		G: uint8(math.Round(g * 255)),
		B: uint8(math.Round(b * 255)),
		A: 255,
	}

	return RGBtoHex(rgba)
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
