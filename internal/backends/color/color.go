package color

import (
	"fmt"
	"image"
	"math"

	"github.com/Achno/gowall/internal/backends/pixels"
	"github.com/lucasb-eyer/go-colorful"
	"github.com/muesli/gamut"
)

func LightenColor(hex string, amount float64) (string, error) {

	c := gamut.Hex(hex)
	lc := gamut.Lighter(c, amount)

	return gamut.ToHex(lc), nil
}

func DarkenColor(hex string, amount float64) (string, error) {
	c := gamut.Hex(hex)
	dc := gamut.Darker(c, amount)
	return gamut.ToHex(dc), nil
}

func GenerateComplementary(hex string) (string, error) {
	c := gamut.Hex(hex)
	complementary := gamut.Complementary(c)
	return gamut.ToHex(complementary), nil
}

func GenerateContrast(hex string) (string, error) {
	c := gamut.Hex(hex)
	contrast := gamut.Contrast(c)
	return gamut.ToHex(contrast), nil
}

func BlendColors(hex1 string, hex2 string, count int) ([]string, error) {
	c1 := gamut.Hex(hex1)
	c2 := gamut.Hex(hex2)
	blended := gamut.Blends(c1, c2, count)

	return ColorsToHex(blended), nil
}

func GenerateShades(hex string, count int) ([]string, error) {
	c := gamut.Hex(hex)
	shades := gamut.Shades(c, count)
	return ColorsToHex(shades), nil
}

func GenerateTints(hex string, count int) ([]string, error) {
	c := gamut.Hex(hex)
	tints := gamut.Tints(c, count)
	return ColorsToHex(tints), nil
}

func GenerateTones(hex string, count int) ([]string, error) {
	c := gamut.Hex(hex)
	tones := gamut.Tones(c, count)
	return ColorsToHex(tones), nil
}

func GenerateMonochromatic(hex string, count int) ([]string, error) {
	c := gamut.Hex(hex)
	monochromatic := gamut.Monochromatic(c, count)
	return ColorsToHex(monochromatic), nil

}

func GenerateTriadic(hex string) ([]string, error) {
	c := gamut.Hex(hex)
	triadic := gamut.Triadic(c)
	return ColorsToHex(triadic), nil
}

func GenerateQuadratic(hex string) ([]string, error) {
	c := gamut.Hex(hex)
	quadratic := gamut.Quadratic(c)
	return ColorsToHex(quadratic), nil
}

func GenerateAnalogous(hex string) ([]string, error) {
	c := gamut.Hex(hex)
	analogous := gamut.Analogous(c)
	return ColorsToHex(analogous), nil
}

func GenerateSplitComplementary(hex string) ([]string, error) {
	c := gamut.Hex(hex)
	splitComplementary := gamut.SplitComplementary(c)
	return ColorsToHex(splitComplementary), nil
}

// GradientKeypoint represents a color at a specific position in the gradient
type GradientKeypoint struct {
	Color    colorful.Color
	Position float64 // Position in range [0,1]
}

// GradientTable contains the keypoints of the gradient
type GradientTable []GradientKeypoint

// blendMethods are the ways two colors of a gradient can be blended
var blendMethods = map[string]func(color1, color2 colorful.Color, blendT float64) colorful.Color{
	"hcl": func(color1, color2 colorful.Color, blendT float64) colorful.Color {
		return color1.BlendHcl(color2, blendT).Clamped()
	},
	"lab": func(color1, color2 colorful.Color, blendT float64) colorful.Color {
		return color1.BlendLab(color2, blendT).Clamped()
	},
	"rgb": func(color1, color2 colorful.Color, blendT float64) colorful.Color {
		return color1.BlendRgb(color2, blendT).Clamped()
	},
	"hsv": func(color1, color2 colorful.Color, blendT float64) colorful.Color {
		return color1.BlendHsv(color2, blendT).Clamped()
	},
	"luv": func(color1, color2 colorful.Color, blendT float64) colorful.Color {
		return color1.BlendLuv(color2, blendT).Clamped()
	},
	"luvlch": func(color1, color2 colorful.Color, blendT float64) colorful.Color {
		return color1.BlendLuvLCh(color2, blendT).Clamped()
	},
}

// GetInterpolatedColorFor returns a HCL-blended color for position t [0,1]
func (gt GradientTable) GetInterpolatedColorFor(t float64, method string) (colorful.Color, error) {
	methodFunc, ok := blendMethods[method]
	if !ok {
		return colorful.Color{}, fmt.Errorf("invalid method: %s", method)
	}
	return gt.interpolate(t, methodFunc), nil
}

func (gt GradientTable) interpolate(t float64, methodFunc func(color1, color2 colorful.Color, blendT float64) colorful.Color) colorful.Color {
	for i := 0; i < len(gt)-1; i++ {
		c1 := gt[i]
		c2 := gt[i+1]
		if c1.Position <= t && t <= c2.Position {
			// Blend between c1 and c2
			blendT := (t - c1.Position) / (c2.Position - c1.Position)
			return methodFunc(c1.Color, c2.Color, blendT)
		}
	}
	// Return the last color if past the end
	return gt[len(gt)-1].Color
}

// GenerateGradient creates a gradient image from a list of hex colors.
// angle: gradient direction in degrees (0=left→right, 90=top→bottom, 180=right→left, 270=bottom→top).
// hexColors: list of hex color strings (e.g., "#ff0000").
// width, height: dimensions of the output image.
func GenerateGradient(hexColors []string, width, height int, angle float64, interpolationMethod string) (image.Image, error) {
	if len(hexColors) < 2 {
		return nil, nil
	}

	// Create gradient table with evenly spaced keypoints
	keypoints := make(GradientTable, len(hexColors))
	for i, hexColor := range hexColors {
		c, err := colorful.Hex(hexColor)
		if err != nil {
			return nil, err
		}
		keypoints[i] = GradientKeypoint{
			Color:    c,
			Position: float64(i) / float64(len(hexColors)-1),
		}
	}

	rad := angle * math.Pi / 180.0
	dx := math.Cos(rad)
	dy := math.Sin(rad)

	cx := float64(width) / 2.0
	cy := float64(height) / 2.0

	halfLen := (math.Abs(float64(width)*dx) + math.Abs(float64(height)*dy)) / 2.0

	methodFunc, ok := blendMethods[interpolationMethod]
	if !ok {
		return nil, fmt.Errorf("invalid method: %s", interpolationMethod)
	}

	img := image.NewRGBA(image.Rect(0, 0, width, height))

	pixels.ParallelRows(height, func(y0, y1 int) {
		for y := y0; y < y1; y++ {
			row := img.Pix[y*img.Stride : y*img.Stride+width*4]
			for x := range width {
				proj := (float64(x)-cx)*dx + (float64(y)-cy)*dy
				t := (proj + halfLen) / (2.0 * halfLen)
				if t < 0 {
					t = 0
				} else if t > 1 {
					t = 1
				}

				// the same as img.Set(x, y, c) without boxing c in a color.Color for every pixel
				r, g, b, a := keypoints.interpolate(t, methodFunc).RGBA()
				row[x*4+0], row[x*4+1], row[x*4+2], row[x*4+3] = uint8(r>>8), uint8(g>>8), uint8(b>>8), uint8(a>>8)
			}
		}
	})

	return img, nil
}
