package color

import (
	"crypto/md5"
	"encoding/hex"
	"image/color"
	"math"
)

// gammaToLinear converts sRGB gamma-corrected value to linear
func gammaToLinear(v float64) float64 {
	if v <= 0.04045 {
		return v / 12.92
	}
	return math.Pow((v+0.055)/1.055, 2.4)
}

// linearToGamma converts linear RGB value to sRGB gamma-corrected
func linearToGamma(v float64) float64 {
	if v <= 0.0031308 {
		return v * 12.92
	}
	return 1.055*math.Pow(v, 1.0/2.4) - 0.055
}

// xyzToLab applies the LAB conversion function
func xyzToLab(t float64) float64 {
	if t > 0.008856 {
		return math.Pow(t, 1.0/3.0)
	}
	return 7.787*t + 16.0/116.0
}

// hueToRGB converts hue to RGB component
func hueToRGB(p, q, t float64) float64 {
	if t < 0 {
		t += 360
	}
	if t > 360 {
		t -= 360
	}
	if t < 60 {
		return p + (q-p)*t/60
	}
	if t < 180 {
		return q
	}
	if t < 240 {
		return p + (q-p)*(240-t)/60
	}
	return p
}

// Helper function to check if two colors are similar within a threshold
func ColorsAreSimilar(c1, c2 color.Color, threshold float64) bool {
	r1, g1, b1, _ := c1.RGBA()
	r2, g2, b2, _ := c2.RGBA()

	// Normalize to 8-bit values
	r1, g1, b1 = r1>>8, g1>>8, b1>>8
	r2, g2, b2 = r2>>8, g2>>8, b2>>8

	// Euclidean distance
	distance := math.Sqrt(
		math.Pow(float64(r1)-float64(r2), 2) +
			math.Pow(float64(g1)-float64(g2), 2) +
			math.Pow(float64(b1)-float64(b2), 2),
	)

	return distance <= threshold
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
