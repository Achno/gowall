package image

import (
	"fmt"
	"image"
	"strings"

	"github.com/Achno/gowall/internal/backends/compression/jpg"
	png "github.com/Achno/gowall/internal/backends/compression/png"
	types "github.com/Achno/gowall/internal/types"
)

type CompressionStrategy interface {
	Compress(img image.Image) (image.Image, types.ImageMetadata, error)
	GetFormat() string
}

type CompressionProcessor struct {
	options CompressionOptions
}

// Options with the functional options pattern so you can pick options and set defaults
type CompressionOptions struct {
	Quality  int
	Speed    int
	Strategy string // Name of the backend to use
}
type CompressionOption func(*CompressionOptions)

func WithQuality(quality int) CompressionOption {
	return func(co *CompressionOptions) {
		co.Quality = quality
	}
}

func WithSpeed(speed int) CompressionOption {
	return func(co *CompressionOptions) {
		co.Speed = speed
	}
}

func WithStrategy(strategy string) CompressionOption {
	return func(co *CompressionOptions) {
		co.Strategy = strategy
	}
}

// NewCompressionProcessor creates a new compression processor with default strategies
func NewCompressionProcessor(opts ...CompressionOption) *CompressionProcessor {
	// Default options
	options := CompressionOptions{
		Quality:  80,
		Speed:    4,
		Strategy: "",
	}

	// Apply functional options
	for _, opt := range opts {
		opt(&options)
	}

	processor := &CompressionProcessor{
		options: options,
	}

	return processor
}

// Process implements the ImageProcessor interface (updated to include format)
func (p *CompressionProcessor) Process(img image.Image, theme string, format string) (image.Image, types.ImageMetadata, error) {

	format = strings.ToLower(format)

	strategy, err := p.GetStrategy(format)
	if err != nil {
		return nil, types.ImageMetadata{}, fmt.Errorf("failed to get compression strategy: %w", err)
	}

	return strategy.Compress(img)
}

// GetStrategies returns all available strategies
func (p *CompressionProcessor) GetStrategies() map[string]func(quality int, speed int) (CompressionStrategy, error) {

	//? Here is where strategies are registered.
	// Keys: <strategy_name>-<format> and they map to a function that returns a CompressionStrategy
	// the -<format> part is used to filter strategies by format and is required.
	var strategies = map[string]func(quality int, speed int) (CompressionStrategy, error){
		"pngquant-png": func(quality int, speed int) (CompressionStrategy, error) {
			return png.NewPngquantStrategy(quality, speed)
		},
		"losslesspng-png": func(quality int, speed int) (CompressionStrategy, error) {
			return png.NewLosslessPngStrategy()
		},
		"lossllyjpeg-jpeg": func(quality int, speed int) (CompressionStrategy, error) {
			return jpg.NewLossllyJpgStrategy(quality)
		},
		"lossllyjpg-jpg": func(quality int, speed int) (CompressionStrategy, error) {
			return jpg.NewLossllyJpgStrategy(quality)
		},
	}

	return strategies
}

// GetDefaultStrategyNameForFormat returns the default strategy name for a format
func (p *CompressionProcessor) GetDefaultStrategyNameForFormat(format string) (string, error) {

	//? Here is where default strategies are registered.
	var defaultStrategyName string
	switch format {
	case "png":
		fmt.Println("using default strategy for png")
		defaultStrategyName = "pngquant"
	case "jpeg":
		fmt.Println("using default strategy for jpeg")
		defaultStrategyName = "lossllyjpeg"
	case "jpg":
		fmt.Println("using default strategy for jpg")
		defaultStrategyName = "lossllyjpg"
	}

	return defaultStrategyName, nil
}

func (p *CompressionProcessor) GetStrategy(format string) (CompressionStrategy, error) {

	strategies := p.GetStrategies()
	strategyName := p.options.Strategy

	// Filter strategies that contain the format
	var availableStrategies []string
	for strategyName := range strategies {
		if strings.Contains(strategyName, format) {
			availableStrategies = append(availableStrategies, strategyName)
		}
	}

	if len(availableStrategies) == 0 {
		return nil, fmt.Errorf("no compression strategies available for format: %s", format)
	}

	// If a strategy is not specified, fallback to the default strategy of the format
	if strategyName == "" {
		defaultStrategyName, err := p.GetDefaultStrategyNameForFormat(format)
		if err != nil {
			return nil, fmt.Errorf("failed to get default strategy name for format: %w", err)
		}
		strategyName = defaultStrategyName
	}

	key := strategyName + "-" + format

	strategyFunc, ok := strategies[key]
	if !ok {
		return nil, fmt.Errorf("strategy %q not available for format %q", p.options.Strategy, format)
	}
	return strategyFunc(p.options.Quality, p.options.Speed)
}

// GetAllStrategies returns all available strategy names (without format suffix)
func (p *CompressionProcessor) GetAllStrategiesNames() []string {
	strategies := p.GetStrategies()
	var strategyNames []string

	for key := range strategies {
		strategyName, _, found := strings.Cut(key, "-")
		if found {
			strategyNames = append(strategyNames, strategyName)
		} else {
			strategyNames = append(strategyNames, key)
		}
	}

	return strategyNames
}
