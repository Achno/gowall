package haldclut

import (
	"image/color"
	"math"
	"sort"
)

type ShepardMapper struct {
	options ShepardMapperOptions
}

type ShepardMapperOptions struct {
	Nearest int
	Power   float64
}

func NewShepardMapper(opts ShepardMapperOptions) *ShepardMapper {
	return &ShepardMapper{options: opts}
}

func (m *ShepardMapper) Map(original color.RGBA, palette []color.RGBA) color.RGBA {
	return shepardInterpolation(original, palette, m.options)
}

// Core Shepard's Method implementation
func shepardInterpolation(originalRGBA color.RGBA, paletteRGBAs []color.RGBA, opts ShepardMapperOptions) color.RGBA {
	if len(paletteRGBAs) == 0 {
		return originalRGBA
	}

	// Find N closest colors based on original color
	closest := findNClosestColors(originalRGBA, paletteRGBAs, opts.Nearest)
	if len(closest) == 0 {
		return originalRGBA
	}

	// If exact match or only one neighbor, return it
	if len(closest) == 1 || closest[0].dist == 0 {
		return closest[0].color
	}

	// Calculate inverse distance weights
	weights := make([]float64, len(closest))
	var totalWeight float64

	for i, c := range closest {
		if c.dist == 0 {
			return c.color // Exact match found
		}
		weight := 1.0 / math.Pow(math.Sqrt(c.dist), opts.Power)
		weights[i] = weight
		totalWeight += weight
	}

	// Blend colors using inverse distance weights
	blended := blendColors(extractColors(closest), weights)
	return blended
}

func colorDistanceSquared(c1, c2 color.RGBA) float64 {
	dr := float64(c1.R) - float64(c2.R)
	dg := float64(c1.G) - float64(c2.G)
	db := float64(c1.B) - float64(c2.B)
	return dr*dr + dg*dg + db*db
}

func findNClosestColors(originalRGBA color.RGBA, paletteRGBAs []color.RGBA, n int) []struct {
	dist  float64
	color color.RGBA
} {
	// Early termination if exact match found
	for _, pRGBA := range paletteRGBAs {
		if originalRGBA == pRGBA {
			return []struct {
				dist  float64
				color color.RGBA
			}{{dist: 0, color: pRGBA}}
		}
	}

	distances := make([]struct {
		dist  float64
		color color.RGBA
	}, 0, len(paletteRGBAs))

	for _, pRGBA := range paletteRGBAs {
		distances = append(distances, struct {
			dist  float64
			color color.RGBA
		}{dist: colorDistanceSquared(originalRGBA, pRGBA), color: pRGBA})
	}

	sort.Slice(distances, func(i, j int) bool {
		return distances[i].dist < distances[j].dist
	})

	if n > len(distances) {
		n = len(distances)
	}
	return distances[:n]
}

func blendColors(colors []color.RGBA, weights []float64) color.RGBA {
	if len(colors) == 0 || len(colors) != len(weights) {
		return color.RGBA{}
	}

	var sumR, sumG, sumB float64
	var totalWeight float64

	for i := range colors {
		sumR += float64(colors[i].R) * weights[i]
		sumG += float64(colors[i].G) * weights[i]
		sumB += float64(colors[i].B) * weights[i]
		totalWeight += weights[i]
	}

	if totalWeight == 0 {
		return colors[0]
	}

	return color.RGBA{
		R: uint8(math.Round(sumR / totalWeight)),
		G: uint8(math.Round(sumG / totalWeight)),
		B: uint8(math.Round(sumB / totalWeight)),
		A: 255,
	}
}

func extractColors(sortedColors []struct {
	dist  float64
	color color.RGBA
}) []color.RGBA {
	colors := make([]color.RGBA, len(sortedColors))
	for i, item := range sortedColors {
		colors[i] = item.color
	}
	return colors
}
