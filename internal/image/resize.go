package image

import (
	"fmt"
	"image"
	"strings"

	"github.com/disintegration/imaging"

	types "github.com/Achno/gowall/internal/types"
)

// implements the ImageProcessor interface
type ResizeProcessor struct {
	options ResizeOptions
}

// Options with the functional options pattern so you can pick options and set defaults
type ResizeOptions struct {
	Width  int
	Height int
	Method string
}

type ResizeOption func(*ResizeOptions)

func WithWidth(width int) ResizeOption {
	return func(ro *ResizeOptions) {
		ro.Width = width
	}
}

func WithHeight(height int) ResizeOption {
	return func(ro *ResizeOptions) {
		ro.Height = height
	}
}

func WithMethod(method string) ResizeOption {
	return func(ro *ResizeOptions) {
		ro.Method = method
	}
}

// Available options: WithWidth, WithHeight, WithMethod
func (p *ResizeProcessor) SetOptions(options ...ResizeOption) {
	opts := ResizeOptions{
		Width:  0,
		Height: 0,
		Method: "lanczos",
	}

	for _, option := range options {
		option(&opts)
	}

	p.options = opts
}

func (p *ResizeProcessor) Process(img image.Image, theme string, format string) (image.Image, types.ImageMetadata, error) {

	newImg, err := resize(&p.options, img)
	if err != nil {
		return nil, types.ImageMetadata{}, err
	}

	return newImg, types.ImageMetadata{}, nil
}

func resize(config *ResizeOptions, img image.Image) (image.Image, error) {

	if config.Width == 0 && config.Height == 0 {
		return img, nil
	}

	filter, err := mapMethodNameToFilter(config.Method)
	if err != nil {
		return nil, err
	}

	return imaging.Resize(img, config.Width, config.Height, filter), nil
}

func mapMethodNameToFilter(method string) (imaging.ResampleFilter, error) {
	method = strings.ToLower(method)

	filterMap := map[string]imaging.ResampleFilter{
		"catmullrom": imaging.CatmullRom,
		"lanczos":    imaging.Lanczos,
	}

	if filter, ok := filterMap[method]; ok {
		return filter, nil
	}

	return imaging.ResampleFilter{}, fmt.Errorf("invalid resampling method: %s", method)
}
