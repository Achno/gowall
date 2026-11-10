package jpg

import (
	"image"
	"io"

	"github.com/Achno/gowall/internal/types"
	"github.com/gen2brain/jpegli"
)

type JpegliStrategy struct {
	Quality int
}

func NewJpegliStrategy(quality int) (*JpegliStrategy, error) {
	return &JpegliStrategy{
		Quality: quality,
	}, nil
}

func (j *JpegliStrategy) Compress(img image.Image) (image.Image, types.ImageMetadata, error) {

	jpegOptions := &jpegli.EncodingOptions{
		Quality:              j.Quality,
		ChromaSubsampling:    image.YCbCrSubsampleRatio420,
		ProgressiveLevel:     2,
		OptimizeCoding:       true,
		AdaptiveQuantization: true,
		DCTMethod:            jpegli.DefaultDCTMethod,
	}

	encoderFunc := func(w io.Writer, img image.Image) error {
		return jpegli.Encode(w, img, jpegOptions)
	}

	metadata := types.ImageMetadata{
		EncoderFunction: encoderFunc,
	}

	return img, metadata, nil
}

func (j *JpegliStrategy) GetFormat() string {
	return "jpeg"
}
