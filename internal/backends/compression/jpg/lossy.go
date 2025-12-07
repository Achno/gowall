package jpg

import (
	"image"
	"image/jpeg"
	"io"

	"github.com/Achno/gowall/internal/types"
)

type LossyJpgStrategy struct {
	Quality int
}

func NewLossyJpgStrategy(quality int) (*LossyJpgStrategy, error) {
	return &LossyJpgStrategy{
		Quality: quality,
	}, nil
}

func (l *LossyJpgStrategy) Compress(img image.Image) (image.Image, types.ImageMetadata, error) {

	jpegOptions := &jpeg.Options{
		Quality: l.Quality,
	}

	encoderFunc := func(w io.Writer, img image.Image) error {
		return jpeg.Encode(w, img, jpegOptions)
	}

	metadata := types.ImageMetadata{
		EncoderFunction: encoderFunc,
	}

	return img, metadata, nil
}

func (l *LossyJpgStrategy) GetFormat() string {
	return "jpeg"
}
