package image

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"math/rand"
	"sync"
)

// impliments the ImageProcessor interface
type BackgroundProcessor struct {
	options BgOptions
}

// Options with the functional options pattern so you can pick options and set defaults
type BgOptions struct {
	MaxIter     int
	Convergence float64
	SampleRate  float64
	NumRoutines int
}

type BgOption func(*BgOptions)

func WithMaxIter(maxIter int) BgOption {
	return func(bo *BgOptions) {
		bo.MaxIter = maxIter
	}
}
func WithConvergence(conv float64) BgOption {
	return func(bo *BgOptions) {
		bo.Convergence = conv
	}
}
func WithSampleRate(sampleRate float64) BgOption {
	return func(bo *BgOptions) {
		bo.SampleRate = sampleRate
	}
}
func WithNumRoutines(numRoutines int) BgOption {
	return func(bo *BgOptions) {
		bo.NumRoutines = numRoutines
	}
}

// Available options : WithMaxIter,WithConvergence,WithSampleRate,WithNumRoutines
func (p *BackgroundProcessor) SetOptions(options ...BgOption) {
	opts := BgOptions{
		MaxIter:     100,
		Convergence: 0.001,
		SampleRate:  0.5,
		NumRoutines: 4,
	}

	for _, option := range options {
		option(&opts)
	}

	p.options = opts
}

func (p *BackgroundProcessor) Process(img image.Image, theme string) (image.Image, error) {

	// check if options have not been set
	if p.options.Convergence == 0 || p.options.MaxIter == 0 || p.options.SampleRate == 0 || p.options.NumRoutines == 0 {
		p.SetOptions()
	}

	err, newImg := removeBackground(&p.options, img)

	if err != nil {
		return nil, fmt.Errorf("while removing background", err)
	}

	return newImg, nil
}

type Point struct {
	R, G, B float64
}

type Cluster struct {
	Centroid Point
	Points   []Point
}

func removeBackground(config *BgOptions, img image.Image) (error, image.Image) {

	bounds := img.Bounds()
	width, height := bounds.Max.X, bounds.Max.Y

	// Convert image to points
	var points []Point
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {

			// Sample pixels from the image using the SampleRate to speed up the algo
			if rand.Float64() > config.SampleRate {
				continue
			}

			r, g, b, _ := img.At(x, y).RGBA()
			points = append(points, Point{
				R: float64(r) / 65535.0,
				G: float64(g) / 65535.0,
				B: float64(b) / 65535.0,
			})
		}
	}

	// Initialize clusters for k-means
	clusters := initializeClusters(points)

	// Run k-means
	for iter := 0; iter < config.MaxIter; iter++ {

		// Clear previous points
		for i := range clusters {
			clusters[i].Points = clusters[i].Points[:0]
		}

		// Assign points to clusters. First split the points to small chunks and do it in parallel for speed
		chunks := splitPoints(points, config.NumRoutines)
		var wg sync.WaitGroup
		results := make([][][]Point, config.NumRoutines)

		for i := 0; i < config.NumRoutines; i++ {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				results[i] = make([][]Point, len(clusters))
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

		// Merge chunks back to the clusters
		for i := range clusters {
			for j := 0; j < config.NumRoutines; j++ {
				clusters[i].Points = append(clusters[i].Points, results[j][i]...)
			}
		}

		// Update centroids color and check convergence
		maxChange := 0.0
		for i := range clusters {
			newCentroid := averagePoint(clusters[i].Points)
			change := distBetweenPoints(clusters[i].Centroid, newCentroid)
			maxChange = math.Max(maxChange, change)
			clusters[i].Centroid = newCentroid
		}

		// Convergence should be met when the colors stabalized
		if maxChange < config.Convergence {
			break
		}
	}

	// check the length, the biggest cluster is the background
	output := image.NewNRGBA(bounds)
	backgroundCluster := 0
	if len(clusters[1].Points) > len(clusters[0].Points) {
		backgroundCluster = 1
	}

	// set the Background clusters pixels to transparent to remove the bg
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			r, g, b, _ := img.At(x, y).RGBA()
			point := Point{
				R: float64(r) / 65535.0,
				G: float64(g) / 65535.0,
				B: float64(b) / 65535.0,
			}

			// Find closest cluster
			minDist := math.MaxFloat64
			closestCluster := 0
			for i, cluster := range clusters {
				dist := distBetweenPoints(point, cluster.Centroid)
				if dist < minDist {
					minDist = dist
					closestCluster = i
				}
			}

			if closestCluster == backgroundCluster {
				output.Set(x, y, color.NRGBA{0, 0, 0, 0}) // Transparent
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

	return nil, output
}

func initializeClusters(points []Point) []Cluster {
	// Create 2 clusters, 1 for foreground and 1 for background
	clusters := make([]Cluster, 2)

	// Choose first centroid randomly
	firstIdx := rand.Intn(len(points))
	clusters[0].Centroid = points[firstIdx]

	// Choose second centroid and make sure its different than the 1st one
	distances := make([]float64, len(points))
	sumDist := 0.0
	for i, p := range points {
		dist := distBetweenPoints(p, clusters[0].Centroid)
		distances[i] = dist * dist
		sumDist += distances[i]
	}

	// Choose point with probability proportional to square distance
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

// splits the points into chunks of points so you can modify a lot of chunks in parallel
func splitPoints(points []Point, n int) [][]Point {
	chunks := make([][]Point, n)
	chunkSize := len(points) / n

	// divide the points slice to n equal chunks
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

func distBetweenPoints(p1, p2 Point) float64 {
	return math.Sqrt(
		math.Pow(p1.R-p2.R, 2) +
			math.Pow(p1.G-p2.G, 2) +
			math.Pow(p1.B-p2.B, 2),
	)
}

// get the average for R,G,B in all points
func averagePoint(points []Point) Point {
	if len(points) == 0 {
		return Point{}
	}

	var sum Point
	for _, p := range points {
		sum.R += p.R
		sum.G += p.G
		sum.B += p.B
	}

	n := float64(len(points))
	return Point{sum.R / n, sum.G / n, sum.B / n}
}
