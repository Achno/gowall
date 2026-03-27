package bgremoval

import (
	"fmt"
	"image"
	"image/color"

	"github.com/Achno/gowall/internal/backends/onnx"
	"github.com/Achno/gowall/internal/backends/onnx/sessions"
)

type U2NetStrategy struct {
	session *onnx.Session
}

func NewU2NetStrategy() (*U2NetStrategy, error) {
	model := &sessions.U2Net{}
	session, err := onnx.NewSession(model)
	if err != nil {
		return nil, fmt.Errorf("create onnx session: %w", err)
	}

	return &U2NetStrategy{session: session}, nil
}

func (s *U2NetStrategy) Remove(img image.Image) (image.Image, error) {
	masks, err := s.session.Predict(img)
	if err != nil {
		return nil, fmt.Errorf("predict mask: %w", err)
	}

	if len(masks) == 0 {
		return nil, fmt.Errorf("no mask generated")
	}

	return applyMask(img, masks[0]), nil
}

func (s *U2NetStrategy) Close() error {
	if s.session != nil {
		return s.session.Close()
	}
	return nil
}

// applyMask applies a grayscale mask to an image, setting alpha based on mask values
// The mask should have high values (white) for foreground and low values (black) for background
func applyMask(img image.Image, mask image.Image) image.Image {
	bounds := img.Bounds()
	result := image.NewNRGBA(bounds)

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, _ := img.At(x, y).RGBA()

			// Get mask value (grayscale = alpha)
			// For grayscale/NRGBA images, R channel contains the gray value
			maskR, _, _, _ := mask.At(x, y).RGBA()
			alpha := uint8(maskR >> 8)

			// For transparent pixels (alpha=0), set RGB to 0 for cleaner output
			if alpha == 0 {
				result.Set(x, y, color.NRGBA{R: 0, G: 0, B: 0, A: 0})
			} else {
				result.Set(x, y, color.NRGBA{
					R: uint8(r >> 8),
					G: uint8(g >> 8),
					B: uint8(b >> 8),
					A: alpha,
				})
			}
		}
	}

	return result
}
