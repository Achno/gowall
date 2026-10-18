package mediancut

import (
	"image"
	"image/color"
	"sort"
	"sync"

	"github.com/Achno/gowall/internal/backends/pixels"
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
	src := pixels.ToRGBA(img)
	width, height := src.Rect.Dx(), src.Rect.Dy()

	// every chunk of rows counts into its own histogram, they are summed at the end
	var mu sync.Mutex
	hist := make([]int, HistogramSize)

	pixels.ParallelRows(height, func(y0, y1 int) {
		local := make([]int, HistogramSize)
		for y := y0; y < y1; y++ {
			row := src.Pix[y*src.Stride : y*src.Stride+width*4]
			for i := 0; i < len(row); i += 4 {
				if row[i+3] < 125 {
					// skip transparent pixels
					continue
				}
				local[RGB(row[i], row[i+1], row[i+2])]++
			}
		}
		mu.Lock()
		for i, n := range local {
			hist[i] += n
		}
		mu.Unlock()
	})
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
