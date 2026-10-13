package image

import (
	"image"
	"io"
	"strings"

	imgbackend "github.com/Achno/gowall/internal/backends/image"
	types "github.com/Achno/gowall/internal/types"
)

// AsciiProcessor implements ImageProcessor, gifs stay animated and other formats get the first frame
type AsciiProcessor struct {
	options imgbackend.AsciiOptions
}

type AsciiOption func(*imgbackend.AsciiOptions)

func WithAsciiDetail(detail int) AsciiOption {
	return func(o *imgbackend.AsciiOptions) {
		o.Detail = detail
	}
}

func WithAsciiMode(mode imgbackend.AsciiMode) AsciiOption {
	return func(o *imgbackend.AsciiOptions) {
		o.Mode = mode
	}
}

func NewAsciiProcessor(opts ...AsciiOption) *AsciiProcessor {
	options := imgbackend.DefaultAsciiOptions()
	for _, opt := range opts {
		opt(&options)
	}
	return &AsciiProcessor{options: options}
}

// Process implements the ImageProcessor interface, animated gifs arrive as a *types.Animation with all their frames
func (p *AsciiProcessor) Process(img image.Image, theme string, format string) (image.Image, types.ImageMetadata, error) {
	anim := types.ToAnimation(img)

	converter := imgbackend.NewAsciiConverter(p.options)
	frames, err := converter.Convert(anim.Frames)
	if err != nil {
		return nil, types.ImageMetadata{}, err
	}

	if !strings.EqualFold(format, "gif") {
		return converter.Render(frames[0]), types.ImageMetadata{}, nil
	}

	metadata := types.ImageMetadata{
		EncoderFunction: func(w io.Writer, _ image.Image) error {
			return converter.EncodeGif(w, frames, anim.Delays, anim.LoopCount)
		},
	}
	return nil, metadata, nil
}
