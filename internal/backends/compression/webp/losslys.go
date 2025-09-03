package webp

import (
	"fmt"
	"image"
	"io"

	"github.com/chai2010/webp"
	_ "golang.org/x/image/webp"

	"github.com/Achno/gowall/internal/types"
)

type LosslyWebpStrategy struct {
	Quality int
}

func NewLossllyWebpStrategy(quality int) (*LosslyWebpStrategy, error) {
	return &LosslyWebpStrategy{
		Quality: quality,
	}, nil
}

func (l *LosslyWebpStrategy) Compress(img image.Image) (image.Image, types.ImageMetadata, error) {

	if err := l.ValidateParams(); err != nil {
		return nil, types.ImageMetadata{}, fmt.Errorf("while validating parameters: %w", err)
	}

	webpOptions := &webp.Options{
		Quality:  float32(l.Quality),
		Lossless: false,
	}

	encoderFunc := func(w io.Writer, img image.Image) error {
		return webp.Encode(w, img, webpOptions)
	}

	metadata := types.ImageMetadata{
		EncoderFunction: encoderFunc,
	}

	return img, metadata, nil
}

func (l *LosslyWebpStrategy) GetFormat() string {
	return "webp"
}

func (p *LosslyWebpStrategy) ValidateParams() error {
	if p.Quality < 0 || p.Quality > 100 {
		return fmt.Errorf("quality must be between 0 and 100")
	}
	return nil
}
