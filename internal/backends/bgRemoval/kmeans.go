package bgremoval

import (
	"image"
	"math"
	"math/rand"
	"sync"

	"github.com/Achno/gowall/internal/backends/pixels"
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
	Count    int // points assigned to the cluster in the last iteration
}

// clusterSums accumulates the points assigned to a cluster, so the new centroid is computed without keeping every point
type clusterSums struct {
	sum   point
	count int
}

// toPoint maps 8-bit channels to [0,1] the same way as the 16-bit values of RGBA() (v*0x101) divided by 65535
func toPoint(r, g, b uint8) point {
	return point{
		R: float64(uint32(r)*0x101) / 65535.0,
		G: float64(uint32(g)*0x101) / 65535.0,
		B: float64(uint32(b)*0x101) / 65535.0,
	}
}

func removeBackground(config KMeansOptions, img image.Image) (image.Image, error) {
	src := pixels.ToRGBA(img)
	width, height := src.Rect.Dx(), src.Rect.Dy()

	var points []point
	for y := 0; y < height; y++ {
		row := src.Pix[y*src.Stride : y*src.Stride+width*4]
		for x := 0; x < width; x++ {
			if rand.Float64() > config.SampleRate {
				continue
			}
			points = append(points, toPoint(row[x*4], row[x*4+1], row[x*4+2]))
		}
	}

	clusters := initializeClusters(points)

	chunks := splitPoints(points, config.NumRoutines)
	for iter := 0; iter < config.MaxIter; iter++ {
		var wg sync.WaitGroup
		results := make([][]clusterSums, config.NumRoutines)

		for i := 0; i < config.NumRoutines; i++ {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				sums := make([]clusterSums, len(clusters))
				for _, p := range chunks[i] {
					j := nearestCluster(p, clusters)
					sums[j].sum.R += p.R
					sums[j].sum.G += p.G
					sums[j].sum.B += p.B
					sums[j].count++
				}
				results[i] = sums
			}(i)
		}
		wg.Wait()

		maxChange := 0.0
		for i := range clusters {
			var total clusterSums
			for j := 0; j < config.NumRoutines; j++ {
				total.sum.R += results[j][i].sum.R
				total.sum.G += results[j][i].sum.G
				total.sum.B += results[j][i].sum.B
				total.count += results[j][i].count
			}

			newCentroid := point{}
			if total.count > 0 {
				n := float64(total.count)
				newCentroid = point{total.sum.R / n, total.sum.G / n, total.sum.B / n}
			}
			change := distBetweenPoints(clusters[i].Centroid, newCentroid)
			maxChange = math.Max(maxChange, change)
			clusters[i].Centroid = newCentroid
			clusters[i].Count = total.count
		}

		if maxChange < config.Convergence {
			break
		}
	}

	output := image.NewNRGBA(image.Rect(0, 0, width, height))
	backgroundCluster := 0
	if clusters[1].Count > clusters[0].Count {
		backgroundCluster = 1
	}

	pixels.ParallelRows(height, func(y0, y1 int) {
		for y := y0; y < y1; y++ {
			srcRow := src.Pix[y*src.Stride : y*src.Stride+width*4]
			dstRow := output.Pix[y*output.Stride : y*output.Stride+width*4]
			for i := 0; i < len(srcRow); i += 4 {
				r, g, b := srcRow[i], srcRow[i+1], srcRow[i+2]
				// background pixels stay transparent (all zeros) in the new image
				if nearestCluster(toPoint(r, g, b), clusters) != backgroundCluster {
					dstRow[i], dstRow[i+1], dstRow[i+2], dstRow[i+3] = r, g, b, 255
				}
			}
		}
	})

	return output, nil
}

// nearestCluster returns the index of the cluster whose centroid is closest to p
func nearestCluster(p point, clusters []cluster) int {
	minDist := math.MaxFloat64
	closest := 0
	for i, cluster := range clusters {
		// squared distance, the square root doesn't change which cluster is the closest
		if dist := sqDistBetweenPoints(p, cluster.Centroid); dist < minDist {
			minDist = dist
			closest = i
		}
	}
	return closest
}

func initializeClusters(points []point) []cluster {
	clusters := make([]cluster, 2)

	firstIdx := rand.Intn(len(points))
	clusters[0].Centroid = points[firstIdx]

	distances := make([]float64, len(points))
	sumDist := 0.0
	for i, p := range points {
		distances[i] = sqDistBetweenPoints(p, clusters[0].Centroid)
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
	return math.Sqrt(sqDistBetweenPoints(p1, p2))
}

func sqDistBetweenPoints(p1, p2 point) float64 {
	dr, dg, db := p1.R-p2.R, p1.G-p2.G, p1.B-p2.B
	return dr*dr + dg*dg + db*db
}
