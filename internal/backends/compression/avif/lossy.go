package avif

import (
	"image"
	"io"

	avifpkg "github.com/gen2brain/avif"

	"github.com/Achno/gowall/internal/types"
)

type LossyAvifStrategy struct {
	Quality int
	Speed   int
}

func NewLossyAvifStrategy(quality int, speed int) (*LossyAvifStrategy, error) {
	return &LossyAvifStrategy{
		Quality: quality,
		Speed:   speed,
	}, nil
}

func (l *LossyAvifStrategy) Compress(img image.Image) (image.Image, types.ImageMetadata, error) {
	avifOptions := avifpkg.Options{
		Quality: l.Quality,
		Speed:   l.Speed,
	}

	encoderFunc := func(w io.Writer, img image.Image) error {
		return avifpkg.Encode(w, img, avifOptions)
	}

	metadata := types.ImageMetadata{
		EncoderFunction: encoderFunc,
	}

	return img, metadata, nil
}

func (l *LossyAvifStrategy) GetFormat() string {
	return "avif"
}
