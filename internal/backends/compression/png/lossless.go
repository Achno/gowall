package png

import (
	"image"
	"image/png"

	"github.com/Achno/gowall/internal/types"
)

type LosslessPngStrategy struct {
}

func NewLosslessPngStrategy() (*LosslessPngStrategy, error) {
	return &LosslessPngStrategy{}, nil
}

func (l *LosslessPngStrategy) Compress(img image.Image) (image.Image, types.ImageMetadata, error) {

	png := &png.Encoder{
		CompressionLevel: png.BestCompression,
	}

	metadata := types.ImageMetadata{
		EncoderFunction: png.Encode,
	}

	return img, metadata, nil
}

func (l *LosslessPngStrategy) GetFormat() string {
	return "png"
}
