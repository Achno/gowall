package types

import (
	"image"
	"io"
)

type EncoderFunc func(w io.Writer, img image.Image) error

type ImageMetadata struct {
	EncoderFunction EncoderFunc
}
