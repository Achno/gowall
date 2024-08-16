package mediancut

import (
	"image"
	"image/color"
	"sort"
)

const (
	// HistogramSize is the maximum number of 16 bit colors
	HistogramSize = 32768
)

// returns a cluster of similar colors by the median cut algorithm
func GetPalette(img image.Image, maxCubes int) ([]color.Color, error) {
	hist := getHistogram(img)

	cubes, _ := cutCubes(hist, maxCubes)

	colors := make([]color.Color, 0, len(cubes))
	for _, cube := range cubes {
		colors = append(colors, cube.GetColor(hist))
	}

	return colors, nil
}

func getHistogram(img image.Image) []int {
	bounds := img.Bounds()

	hist := make([]int, HistogramSize)

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, a := img.At(x, y).RGBA()
			r = r >> 8
			g = g >> 8
			b = b >> 8
			a = a >> 8

			if a < 125 {
				// skip transparent pixels
				continue
			}
			cl := RGB(uint8(r), uint8(g), uint8(b))
			hist[cl]++

		}
	}
	return hist
}

func cutCubes(hist []int, maxCubes int) ([]ColorCube, int) {
	lenHist := uint16(len(hist))
	histShort := make([]uint16, 0, lenHist)
	firstCube := ColorCube{}

	for clr := uint16(0); clr < lenHist; clr++ {
		if hist[clr] != 0 {
			histShort = append(histShort, clr)
			firstCube.Count += hist[clr]
		}
	}

	firstCube.Hist = histShort[:]
	firstCube.Shrink()
	cubeQueue := NewPriorityQueue(maxCubes)
	cubeQueue.Push(firstCube, firstCube.Rank())

	for nCubes := 1; nCubes < maxCubes; nCubes++ {

		cube, pv, err := cubeQueue.Pop()
		if err != nil {
			break
		}
		if cube.Level > 255 || len(cube.Hist) == 0 {
			cubeQueue.Push(cube, pv)
			break
		}
		sort.Sort(&cube)

		median := -1
		count := 0
		for i, colorIdx := range cube.Hist {
			if count >= cube.Count/2 {
				median = i
				break
			}
			count += hist[colorIdx]
		}
		if median == -1 {
			median = len(cube.Hist) - 1
			count = cube.Count - hist[cube.Hist[median]]
		}

		cubeA := cube.Clone()
		cubeA.Count = count
		cubeA.Hist = cube.Hist[:median]
		cubeA.Level++
		cubeA.Shrink()
		cubeQueue.Push(cubeA, cubeA.Rank())

		cubeB := cube.Clone()
		cubeB.Count -= count
		cubeB.Hist = cube.Hist[median:]
		cubeB.Level++
		cubeB.Shrink()
		cubeQueue.Push(cubeB, cubeB.Rank())
	}

	cubeList := make([]ColorCube, maxCubes)

	sort.Sort(cubeQueue)
	nCubes := 0
	for _, item := range cubeQueue.queue {
		cubeList[nCubes] = item.Value
		nCubes++
	}

	return cubeList, nCubes
}
