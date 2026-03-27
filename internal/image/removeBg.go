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

func GetBgStrategyNames() []string {
	strategies := []string{
		"kmeans",
		"u2net",
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
