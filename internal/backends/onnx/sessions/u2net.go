package sessions

import (
	"image"

	"github.com/Achno/gowall/internal/backends/onnx"
	"github.com/Achno/gowall/utils"
	ort "github.com/yalue/onnxruntime_go"
)

const u2netModelURL = "https://github.com/danielgatis/rembg/releases/download/v0.0.0/u2net.onnx"

var u2netParams = NormalizeParams{
	Mean: ImageNetParams.Mean,
	Std:  ImageNetParams.Std,
	Size: 320,
}

// U2Net implements the Model interface for U2Net segmentation
type U2Net struct{}

func (m *U2Net) Name() string {
	return "u2net"
}

func (m *U2Net) DownloadURL() string {
	return u2netModelURL
}

func (m *U2Net) Download(opts onnx.DownloadOptions) error {
	return utils.DownloadUrl(u2netModelURL, opts.DestPath)
}

func (m *U2Net) Normalize(img image.Image) ([]ort.Value, error) {
	return NormalizeImage(img, u2netParams)
}

func (m *U2Net) Predict(img image.Image, outputs []ort.Value) ([]image.Image, error) {
	return PredictSingleMask(img, outputs)
}
