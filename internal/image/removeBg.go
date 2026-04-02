package image

import (
	"fmt"
	"image"
	"slices"

	bgremoval "github.com/Achno/gowall/internal/backends/bgRemoval"
	types "github.com/Achno/gowall/internal/types"
)

// BackgroundProcessor implements the ImageProcessor interface.
type BackgroundProcessor struct {
	strategy bgremoval.BgRemovalStrategy
}

func NewBackgroundProcessor(strategy bgremoval.BgRemovalStrategy) *BackgroundProcessor {
	return &BackgroundProcessor{strategy: strategy}
}

func (p *BackgroundProcessor) Process(img image.Image, theme string, format string) (image.Image, types.ImageMetadata, error) {
	if p.strategy == nil {
		p.strategy = bgremoval.NewKMeansStrategy(bgremoval.DefaultKMeansOptions())
	}

	newImg, err := p.strategy.Remove(img)
	if err != nil {
		return nil, types.ImageMetadata{}, fmt.Errorf("while removing background: %w", err)
	}

	return newImg, types.ImageMetadata{}, nil
}

func GetBgStrategyNames() []string {
	strategies := []string{
		"kmeans",
		"u2net",
		"bria-rmbg",
	}

	slices.Sort(strategies)

	return strategies
}

func IsValidBgStrategy(name string) bool {
	for _, strategy := range GetBgStrategyNames() {
		if strategy == name {
			return true
		}
	}

	return false
}

func GetBgStrategy(method string, maxIter int, convergence float64, sampleRate float64, numRoutines int) (bgremoval.BgRemovalStrategy, func() error, error) {
	switch method {
	case "kmeans":
		return bgremoval.NewKMeansStrategy(bgremoval.KMeansOptions{
			MaxIter:     maxIter,
			Convergence: convergence,
			SampleRate:  sampleRate,
			NumRoutines: numRoutines,
		}), nil, nil
	case "u2net":
		strategy, err := bgremoval.NewU2NetStrategy()
		if err != nil {
			return nil, nil, fmt.Errorf("initializing U2Net: %w", err)
		}
		return strategy, strategy.Close, nil
	case "bria-rmbg":
		strategy, err := bgremoval.NewBriaRmBgStrategy()
		if err != nil {
			return nil, nil, fmt.Errorf("initializing Bria RMBG: %w", err)
		}
		return strategy, strategy.Close, nil
	default:
		return nil, nil, fmt.Errorf("invalid background removal method %q", method)
	}
}
