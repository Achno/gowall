package image

import (
	"fmt"
	"image"
	"image/draw"
	"strings"

	"github.com/disintegration/imaging"
	drawx "golang.org/x/image/draw"

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

// ResizeWithPadding resizes an image to the specified width and height while preserving aspect ratio and padding the image to the target dimensions.
func ResizeWithPadding(img image.Image, width, height int) image.Image {
	srcBounds := img.Bounds()
	srcWidth := srcBounds.Dx()
	srcHeight := srcBounds.Dy()

	widthRatio := float64(width) / float64(srcWidth)
	heightRatio := float64(height) / float64(srcHeight)

	// Use the smaller ratio to ensure the image fits within the target dimensions
	ratio := min(heightRatio, widthRatio)
	newWidth := int(float64(srcWidth) * ratio)
	newHeight := int(float64(srcHeight) * ratio)
	dstRect := image.Rect(0, 0, newWidth, newHeight)

	dst := image.NewRGBA(dstRect)
	drawx.CatmullRom.Scale(dst, dstRect, img, img.Bounds(), draw.Over, nil)

	// Center the image in the target dimensions if needed
	if newWidth < width || newHeight < height {
		centered := image.NewRGBA(image.Rect(0, 0, width, height))
		offsetX := (width - newWidth) / 2
		offsetY := (height - newHeight) / 2
		draw.Draw(centered, centered.Bounds(), dst, image.Point{-offsetX, -offsetY}, draw.Over)
		return centered
	}

	return dst
}
