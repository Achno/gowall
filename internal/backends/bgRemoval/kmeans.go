package bgremoval

import (
	"image"
	"image/color"
	"math"
	"math/rand"
	"sync"
)

type KMeansOptions struct {
	MaxIter     int
	Convergence float64
	SampleRate  float64
	NumRoutines int
}

type KMeansStrategy struct {
	options KMeansOptions
}

func DefaultKMeansOptions() KMeansOptions {
	return KMeansOptions{
		MaxIter:     100,
		Convergence: 0.001,
		SampleRate:  0.5,
		NumRoutines: 4,
	}
}

func NewKMeansStrategy(opts KMeansOptions) BgRemovalStrategy {
	defaults := DefaultKMeansOptions()

	if opts.MaxIter == 0 {
		opts.MaxIter = defaults.MaxIter
	}
	if opts.Convergence == 0 {
		opts.Convergence = defaults.Convergence
	}
	if opts.SampleRate == 0 {
		opts.SampleRate = defaults.SampleRate
	}
	if opts.NumRoutines == 0 {
		opts.NumRoutines = defaults.NumRoutines
	}

	return &KMeansStrategy{options: opts}
}

func (k *KMeansStrategy) Remove(img image.Image) (image.Image, error) {
	return removeBackground(k.options, img)
}

type point struct {
	R, G, B float64
}

type cluster struct {
	Centroid point
	Points   []point
}

func removeBackground(config KMeansOptions, img image.Image) (image.Image, error) {
	bounds := img.Bounds()
	width, height := bounds.Max.X, bounds.Max.Y

	var points []point
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			if rand.Float64() > config.SampleRate {
				continue
			}

			r, g, b, _ := img.At(x, y).RGBA()
			points = append(points, point{
				R: float64(r) / 65535.0,
				G: float64(g) / 65535.0,
				B: float64(b) / 65535.0,
			})
		}
	}

	clusters := initializeClusters(points)

	for iter := 0; iter < config.MaxIter; iter++ {
		for i := range clusters {
			clusters[i].Points = clusters[i].Points[:0]
		}

		chunks := splitPoints(points, config.NumRoutines)
		var wg sync.WaitGroup
		results := make([][][]point, config.NumRoutines)

		for i := 0; i < config.NumRoutines; i++ {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				results[i] = make([][]point, len(clusters))
				for _, p := range chunks[i] {
					minDist := math.MaxFloat64
					minCluster := 0
					for j, cluster := range clusters {
						dist := distBetweenPoints(p, cluster.Centroid)
						if dist < minDist {
							minDist = dist
							minCluster = j
						}
					}
					results[i][minCluster] = append(results[i][minCluster], p)
				}
			}(i)
		}
		wg.Wait()

		for i := range clusters {
			for j := 0; j < config.NumRoutines; j++ {
				clusters[i].Points = append(clusters[i].Points, results[j][i]...)
			}
		}

		maxChange := 0.0
		for i := range clusters {
			newCentroid := averagePoint(clusters[i].Points)
			change := distBetweenPoints(clusters[i].Centroid, newCentroid)
			maxChange = math.Max(maxChange, change)
			clusters[i].Centroid = newCentroid
		}

		if maxChange < config.Convergence {
			break
		}
	}

	output := image.NewNRGBA(bounds)
	backgroundCluster := 0
	if len(clusters[1].Points) > len(clusters[0].Points) {
		backgroundCluster = 1
	}

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			r, g, b, _ := img.At(x, y).RGBA()
			pixelPoint := point{
				R: float64(r) / 65535.0,
				G: float64(g) / 65535.0,
				B: float64(b) / 65535.0,
			}

			minDist := math.MaxFloat64
			closestCluster := 0
			for i, cluster := range clusters {
				dist := distBetweenPoints(pixelPoint, cluster.Centroid)
				if dist < minDist {
					minDist = dist
					closestCluster = i
				}
			}

			if closestCluster == backgroundCluster {
				output.Set(x, y, color.NRGBA{0, 0, 0, 0})
			} else {
				output.Set(x, y, color.NRGBA{
					R: uint8(r >> 8),
					G: uint8(g >> 8),
					B: uint8(b >> 8),
					A: 255,
				})
			}
		}
	}

	return output, nil
}

func initializeClusters(points []point) []cluster {
	clusters := make([]cluster, 2)

	firstIdx := rand.Intn(len(points))
	clusters[0].Centroid = points[firstIdx]

	distances := make([]float64, len(points))
	sumDist := 0.0
	for i, p := range points {
		dist := distBetweenPoints(p, clusters[0].Centroid)
		distances[i] = dist * dist
		sumDist += distances[i]
	}

	target := rand.Float64() * sumDist
	currentSum := 0.0
	for i, dist := range distances {
		currentSum += dist
		if currentSum >= target {
			clusters[1].Centroid = points[i]
			break
		}
	}

	return clusters
}

func splitPoints(points []point, n int) [][]point {
	chunks := make([][]point, n)
	chunkSize := len(points) / n

	for i := 0; i < n; i++ {
		start := i * chunkSize
		end := start + chunkSize
		if i == n-1 {
			end = len(points)
		}
		chunks[i] = points[start:end]
	}

	return chunks
}

func distBetweenPoints(p1, p2 point) float64 {
	return math.Sqrt(
		math.Pow(p1.R-p2.R, 2) +
			math.Pow(p1.G-p2.G, 2) +
			math.Pow(p1.B-p2.B, 2),
	)
}

func averagePoint(points []point) point {
	if len(points) == 0 {
		return point{}
	}

	var sum point
	for _, p := range points {
		sum.R += p.R
		sum.G += p.G
		sum.B += p.B
	}

	n := float64(len(points))
	return point{sum.R / n, sum.G / n, sum.B / n}
}
