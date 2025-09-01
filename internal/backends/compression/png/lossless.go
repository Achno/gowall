package png

import (
	"image"

	"github.com/Achno/gowall/internal/types"
)

// PngquantStrategy implements pngquant compression for PNG images
type LosslessPngStrategy struct {
}

func NewLosslessPngStrategy() (*LosslessPngStrategy, error) {
	return &LosslessPngStrategy{}, nil
}

func (l *LosslessPngStrategy) Compress(img image.Image) (image.Image, types.ImageMetadata, error) {
	return img, types.ImageMetadata{}, nil
}

func (l *LosslessPngStrategy) GetFormat() string {
	return "png"
}
