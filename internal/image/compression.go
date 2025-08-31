package image

import "image"

type CompressionStrategy interface {
	Compress(img image.Image, quality int) (image.Image, error)
	GetFormat() string
}

type CompressionProcessor struct {
	options CompressionOptions
}

// Options with the functional options pattern so you can pick options and set defaults
type CompressionOptions struct {
	Quality int
	Speed   int
	Method  string
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

func WithMethod(method string) CompressionOption {
	return func(co *CompressionOptions) {
		co.Method = method
	}
}

func (p *CompressionProcessor) Process(img image.Image, theme string) (image.Image, error) {
	return img, nil
}
