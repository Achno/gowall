package color

import (
	"crypto/md5"
	"encoding/hex"
	"image"
	"image/color"
	"math"
)

// ColorSimilarityWeight takes color.RGBA instead of color.Color, so calling it per pixel doesn't allocate
func ColorSimilarityWeight(c1, c2 color.RGBA, threshold float64) float64 {
	dr, dg, db := float64(c1.R)-float64(c2.R), float64(c1.G)-float64(c2.G), float64(c1.B)-float64(c2.B)
	distance := math.Sqrt(dr*dr + dg*dg + db*db)

	if threshold < 0 {
		threshold = 0
	}
	if threshold > 255 {
		threshold = 255
	}

	maxDistance := threshold * math.Sqrt(3)
	if maxDistance <= 0 || distance > maxDistance {
		return 0
	}

	return 1 - (distance / maxDistance)
}

func BlendColor(c1, c2 color.Color, weight float64) color.Color {
	return BlendColorRGBA(RGBA8(c1), RGBA8(c2), weight)
}

// BlendColorRGBA is BlendColor for color.RGBA, it doesn't box the colors in an interface so it doesn't allocate in per pixel loops
func BlendColorRGBA(c1, c2 color.RGBA, weight float64) color.RGBA {
	weight = math.Max(0, math.Min(1, weight))

	blend := func(v1, v2 uint8) uint8 {
		mixed := (1-weight)*float64(v1) + weight*float64(v2)
		return uint8(math.Round(mixed))
	}

	return color.RGBA{
		R: blend(c1.R, c2.R),
		G: blend(c1.G, c2.G),
		B: blend(c1.B, c2.B),
		A: blend(c1.A, c2.A),
	}
}

func ColorDistance(r1, g1, b1, r2, g2, b2 uint32) float64 {
	return math.Sqrt(float64((r1-r2)*(r1-r2) + (g1-g2)*(g1-g2) + (b1-b2)*(b1-b2)))
}

func HashPalette(colors []string) string {
	hasher := md5.New()
	for _, color := range colors {
		hasher.Write([]byte(color))
	}
	// shorten hash
	r := hex.EncodeToString(hasher.Sum(nil))[:16]
	return r
}

func InvertColor(clr color.Color) color.Color {
	r, g, b, a := clr.RGBA()

	return color.RGBA{
		R: uint8(255 - r/257),
		G: uint8(255 - g/257),
		B: uint8(255 - b/257),
		A: uint8(a / 257),
	}

}
func Clamp(val, min, max int) int {
	if val < min {
		return min
	} else if val > max {
		return max
	}
	return val
}

func ClampFloat(val, min, max float64) float64 {
	return math.Max(min, math.Min(max, val))
}

func Luminance(r, g, b uint8) float64 {
	return (0.299*float64(r) + 0.587*float64(g) + 0.114*float64(b)) / 255
}

// LuminanceRange returns the darkest and brightest luminance of the opaque pixels across all images, 0 and 1 if there are none
func LuminanceRange(imgs []*image.NRGBA) (float64, float64) {
	lo, hi := 1.0, 0.0
	for _, img := range imgs {
		for i := 0; i+3 < len(img.Pix); i += 4 {
			if img.Pix[i+3] < 128 {
				continue
			}
			lum := Luminance(img.Pix[i], img.Pix[i+1], img.Pix[i+2])
			lo, hi = math.Min(lo, lum), math.Max(hi, lum)
		}
	}

	if lo > hi {
		return 0, 1
	}
	return lo, hi
}
