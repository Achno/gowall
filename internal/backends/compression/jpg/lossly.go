package jpg

import (
	"fmt"
	"image"
	"image/jpeg"
	"io"

	"github.com/Achno/gowall/internal/types"
)

type LosslessJpgStrategy struct {
	Quality int
}

func NewLossllyJpgStrategy(quality int) (*LosslessJpgStrategy, error) {
	return &LosslessJpgStrategy{
		Quality: quality,
	}, nil
}

func (l *LosslessJpgStrategy) Compress(img image.Image) (image.Image, types.ImageMetadata, error) {

	if err := l.ValidateParams(); err != nil {
		return nil, types.ImageMetadata{}, fmt.Errorf("while validating parameters: %w", err)
	}

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

func (l *LosslessJpgStrategy) GetFormat() string {
	return "jpeg"
}

func (p *LosslessJpgStrategy) ValidateParams() error {
	if p.Quality < 0 || p.Quality > 100 {
		return fmt.Errorf("quality must be between 0 and 100")
	}
	return nil
}
