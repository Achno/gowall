package image

import (
	"fmt"
	"image"
	"maps"
	"slices"

	"github.com/Achno/gowall/internal/backends/inpaint"
	types "github.com/Achno/gowall/internal/types"
)

type InpaintProcessor struct {
	options InpaintOptions
}

type InpaintOptions struct {
	Rects    []image.Rectangle
	Strategy string // Name of the backend to use
}
type InpaintOption func(*InpaintOptions)

func WithRects(rects []image.Rectangle) InpaintOption {
	return func(o *InpaintOptions) {
		o.Rects = rects
	}
}

func WithInpaintStrategy(strategy string) InpaintOption {
	return func(o *InpaintOptions) {
		o.Strategy = strategy
	}
}

func NewInpaintProcessor(opts ...InpaintOption) *InpaintProcessor {
	options := InpaintOptions{Strategy: "quilt"}
	for _, opt := range opts {
		opt(&options)
	}
	if options.Strategy == "" {
		options.Strategy = "quilt"
	}
	return &InpaintProcessor{options: options}
}

func (p *InpaintProcessor) Process(img image.Image, theme string, format string) (image.Image, types.ImageMetadata, error) {
	strategy, err := p.GetStrategy()
	if err != nil {
		return nil, types.ImageMetadata{}, err
	}
	out, err := inpaint.Erase(img, p.options.Rects, strategy)
	if err != nil {
		return nil, types.ImageMetadata{}, fmt.Errorf("while inpainting: %w", err)
	}
	return out, types.ImageMetadata{}, nil
}

// GetStrategies returns all available strategies
func (p *InpaintProcessor) GetStrategies() map[string]func() (inpaint.InpaintStrategy, error) {
	//? Here is where strategies are registered.
	return map[string]func() (inpaint.InpaintStrategy, error){
		"quilt": func() (inpaint.InpaintStrategy, error) {
			return inpaint.Quilt{}, nil
		},
	}
}

func (p *InpaintProcessor) GetStrategy() (inpaint.InpaintStrategy, error) {
	strategyFunc, ok := p.GetStrategies()[p.options.Strategy]
	if !ok {
		return nil, fmt.Errorf("inpainting strategy %q not available, use one of %v", p.options.Strategy, p.GetAllStrategiesNames())
	}
	return strategyFunc()
}

func (p *InpaintProcessor) GetAllStrategiesNames() []string {
	return slices.Sorted(maps.Keys(p.GetStrategies()))
}
