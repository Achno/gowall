package color

import (
	"crypto/md5"
	"encoding/hex"
	"image"
	"image/color"
	"math"
)

func ColorSimilarityWeight(c1, c2 color.Color, threshold float64) float64 {
	r1, g1, b1, _ := c1.RGBA()
	r2, g2, b2, _ := c2.RGBA()

	r1, g1, b1 = r1>>8, g1>>8, b1>>8
	r2, g2, b2 = r2>>8, g2>>8, b2>>8

	distance := math.Sqrt(
		math.Pow(float64(r1)-float64(r2), 2) +
			math.Pow(float64(g1)-float64(g2), 2) +
			math.Pow(float64(b1)-float64(b2), 2),
	)

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
	weight = math.Max(0, math.Min(1, weight))

	r1, g1, b1, a1 := c1.RGBA()
	r2, g2, b2, a2 := c2.RGBA()

	blend := func(v1, v2 uint32) uint8 {
		mixed := (1-weight)*float64(v1>>8) + weight*float64(v2>>8)
		return uint8(math.Round(mixed))
	}

	return color.RGBA{
		R: blend(r1, r2),
		G: blend(g1, g2),
		B: blend(b1, b2),
		A: blend(a1, a2),
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
