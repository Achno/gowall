package haldclut

import (
	"image/color"
	"math"
)

type RBFMapper struct {
	options RBFMapperOptions
}

type RBFMapperOptions struct {
	Sigma float64 // std makes the gaussian wider
}

func NewRBFMapper() RBFMapperOptions {
	return RBFMapperOptions{
		Sigma: 50.0,
	}
}

func (m *RBFMapper) Map(original color.RGBA, palette []color.RGBA) color.RGBA {

	m.options = NewRBFMapper()
	return rbfInterpolation(original, palette, m.options.Sigma)
}

func rbfInterpolation(target color.RGBA, palette []color.RGBA, sigma float64) color.RGBA {
	var numeratorR, numeratorG, numeratorB, denominator float64

	for _, pColor := range palette {
		// Euclidean distance between target and palette color
		distance := math.Sqrt(math.Pow(float64(target.R)-float64(pColor.R), 2) +
			math.Pow(float64(target.G)-float64(pColor.G), 2) +
			math.Pow(float64(target.B)-float64(pColor.B), 2))

		// Gaussian RBF weight
		weight := math.Exp(-distance * distance / (2 * sigma * sigma))

		// Weighted sum
		numeratorR += float64(pColor.R) * weight
		numeratorG += float64(pColor.G) * weight
		numeratorB += float64(pColor.B) * weight
		denominator += weight
	}

	if denominator > 0 {
		return color.RGBA{
			R: uint8(numeratorR / denominator),
			G: uint8(numeratorG / denominator),
			B: uint8(numeratorB / denominator),
			A: 255,
		}
	}

	// Fallback if no weight is applied
	return color.RGBA{R: 0, G: 0, B: 0, A: 255}
}
