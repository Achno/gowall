package bgremoval

import (
	"fmt"
	"image"

	"github.com/Achno/gowall/internal/backends/onnx"
	"github.com/Achno/gowall/internal/backends/onnx/sessions"
)

type BriaRmBgStrategy struct {
	session *onnx.Session
}

func NewBriaRmBgStrategy() (*BriaRmBgStrategy, error) {
	model := &sessions.BriaRmBg{}
	session, err := onnx.NewSession(model)
	if err != nil {
		return nil, fmt.Errorf("create onnx session: %w", err)
	}

	return &BriaRmBgStrategy{session: session}, nil
}

func (s *BriaRmBgStrategy) Remove(img image.Image) (image.Image, error) {
	masks, err := s.session.Predict(img)
	if err != nil {
		return nil, fmt.Errorf("predict mask: %w", err)
	}

	if len(masks) == 0 {
		return nil, fmt.Errorf("no mask generated")
	}

	return applyMask(img, masks[0]), nil
}

func (s *BriaRmBgStrategy) Close() error {
	if s.session != nil {
		return s.session.Close()
	}
	return nil
}
