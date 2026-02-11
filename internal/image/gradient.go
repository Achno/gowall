package image

import (
	"image"

	cpkg "github.com/Achno/gowall/internal/backends/color"
	types "github.com/Achno/gowall/internal/types"
)

// implements the ImageProcessor interface
type GradientProcessor struct {
	options GradientOptions
}

// Options with the functional options pattern so you can pick options and set defaults
type GradientOptions struct {
	Colors    []string
	Width     int
	Height    int
	Direction string // "horizontal" or "vertical"
	Method    string // "rgb", "hcl", "lab", "hsv", "luv", "luvlch"
}

type GradientOption func(*GradientOptions)

func WithColors(colors []string) GradientOption {
	return func(go_ *GradientOptions) {
		go_.Colors = colors
	}
}

func WithGradientWidth(width int) GradientOption {
	return func(go_ *GradientOptions) {
		go_.Width = width
	}
}

func WithGradientHeight(height int) GradientOption {
	return func(go_ *GradientOptions) {
		go_.Height = height
	}
}

func WithDirection(direction string) GradientOption {
	return func(go_ *GradientOptions) {
		go_.Direction = direction
	}
}

func WithGradientMethod(method string) GradientOption {
	return func(go_ *GradientOptions) {
		go_.Method = method
	}
}

// Available options: WithColors, WithGradientWidth, WithGradientHeight, WithDirection
func (p *GradientProcessor) SetOptions(options ...GradientOption) {
	opts := GradientOptions{
		Colors:    []string{},
		Width:     1920,
		Height:    1080,
		Direction: "horizontal",
		Method:    "rgb",
	}

	for _, option := range options {
		option(&opts)
	}

	p.options = opts
}

func (p *GradientProcessor) Process(img image.Image, theme string, format string) (image.Image, types.ImageMetadata, error) {
	// Generate gradient (input image is ignored for gradient generation)
	gradientImg, err := cpkg.GenerateGradient(
		p.options.Colors,
		p.options.Width,
		p.options.Height,
		p.options.Direction,
		p.options.Method,
	)
	if err != nil {
		return nil, types.ImageMetadata{}, err
	}

	return gradientImg, types.ImageMetadata{}, nil
}
