package mediancut

import (
	"image/color"
)

type LongestColor int

const (
	LongRed LongestColor = iota + 1
	LongGreen
	LongBlue
)

type ColorCube struct {
	Count   int          // number of pixels
	Level   int          // cutting depth
	Longest LongestColor // RGB cube's longest edge

	RMin, RMax, GMin, GMax, BMin, BMax uint8 // range of RGB value

	Hist []uint16
}

// Shrink will shrink the range of RGB value
func (cube *ColorCube) Shrink() {
	if cube.Count == 0 {
		return
	}
	cube.RMin = 255
	cube.RMax = 0
	cube.GMin = 255
	cube.GMax = 0
	cube.BMin = 255
	cube.BMax = 0

	var r, g, b uint8
	var cl uint16
	for i := range cube.Hist {
		cl = cube.Hist[i]
		r, g, b = GetRGB(cl)

		if r > cube.RMax {
			cube.RMax = r
		}
		if r < cube.RMin {
			cube.RMin = r
		}
		if g > cube.GMax {
			cube.GMax = g
		}
		if g < cube.GMin {
			cube.GMin = g
		}
		if b > cube.BMax {
			cube.BMax = b
		}
		if b < cube.BMin {
			cube.BMin = b
		}
	}

	lr := cube.RMax - cube.RMin
	lg := cube.GMax - cube.GMin
	lb := cube.BMax - cube.BMin

	if lr >= lg && lr >= lb {
		cube.Longest = LongRed
	} else if lg >= lr && lg >= lb {
		cube.Longest = LongGreen
	} else {
		cube.Longest = LongBlue
	}
}

// GetColor return the cube's average color value
func (cube ColorCube) GetColor(hist []int) color.RGBA {
	if cube.Count == 0 {
		return color.RGBA{}
	}
	var rsum, gsum, bsum int64

	for i := range cube.Hist {
		cl := cube.Hist[i]
		r, g, b := GetRGB(cl)
		rsum += int64(r) * int64(hist[cl])
		gsum += int64(g) * int64(hist[cl])
		bsum += int64(b) * int64(hist[cl])
	}
	count := int64(cube.Count)
	return color.RGBA{
		R: uint8(rsum / count),
		G: uint8(gsum / count),
		B: uint8(bsum / count),
		A: 255,
	}
}

func (cube ColorCube) Clone() ColorCube {
	return ColorCube{
		Count: cube.Count,
		Level: cube.Level,
		Hist:  cube.Hist,
	}
}

func (cube ColorCube) Volume() int {
	return int(cube.RMax-cube.RMin) * int(cube.GMax-cube.GMin) * int(cube.BMax-cube.BMin)
}

func (cube ColorCube) Rank() int {
	return cube.Count * cube.Volume()
}

func (cube *ColorCube) Len() int {
	return len(cube.Hist)
}

func (cube *ColorCube) Less(i, j int) bool {
	colorA := cube.Hist[i]
	colorB := cube.Hist[j]

	switch cube.Longest {
	case LongRed:
		return RedColor(colorA) < RedColor(colorB)
	case LongGreen:
		return GreenColor(colorA) < GreenColor(colorB)
	case LongBlue:
		return BlueColor(colorA) < BlueColor(colorB)
	}
	return true
}

func (cube *ColorCube) Swap(i, j int) {
	cube.Hist[i], cube.Hist[j] = cube.Hist[j], cube.Hist[i]
}
