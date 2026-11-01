package image

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"slices"

	bgremoval "github.com/Achno/gowall/internal/backends/bgRemoval"
	"github.com/Achno/gowall/internal/backends/onnx"
	"github.com/Achno/gowall/internal/backends/onnx/sessions"
	types "github.com/Achno/gowall/internal/types"
)

// BackgroundProcessor implements the ImageProcessor interface.
type BackgroundProcessor struct {
	strategy      bgremoval.BgRemovalStrategy
	backgroundClr color.Color
}

func NewBackgroundProcessor(strategy bgremoval.BgRemovalStrategy, backgroundClr color.Color) *BackgroundProcessor {
	return &BackgroundProcessor{
		strategy:      strategy,
		backgroundClr: backgroundClr,
	}
}

func (p *BackgroundProcessor) Process(img image.Image, theme string, format string) (image.Image, types.ImageMetadata, error) {
	if p.strategy == nil {
		p.strategy = bgremoval.NewKMeansStrategy(bgremoval.DefaultKMeansOptions())
	}

	newImg, err := p.strategy.Remove(img)
	if err != nil {
		return nil, types.ImageMetadata{}, fmt.Errorf("while removing background: %w", err)
	}

	if p.backgroundClr != nil {
		bounds := newImg.Bounds()
		dst := image.NewNRGBA(bounds)
		draw.Draw(dst, bounds, &image.Uniform{C: p.backgroundClr}, image.Point{}, draw.Src)
		draw.Draw(dst, bounds, newImg, bounds.Min, draw.Over)
		newImg = dst
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

// IsBgStrategyReady checks whether the required runtime and model files are already installed on disk.
func IsBgStrategyReady(method string) bool {
	switch method {
	case "kmeans":
		return true
	case "u2net":
		return onnx.IsReady(&sessions.U2Net{})
	case "bria-rmbg":
		return onnx.IsReady(&sessions.BriaRmBg{})
	default:
		return false
	}
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
