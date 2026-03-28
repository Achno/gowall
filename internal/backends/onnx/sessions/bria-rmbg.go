package sessions

import (
	"image"

	"github.com/Achno/gowall/internal/backends/onnx"
	"github.com/Achno/gowall/utils"
	ort "github.com/yalue/onnxruntime_go"
)

const briaRmBgModelURL = "https://github.com/danielgatis/rembg/releases/download/v0.0.0/bria-rmbg-2.0.onnx"

var briaRmBgParams = NormalizeParams{
	Mean: ImageNetParams.Mean,
	Std:  ImageNetParams.Std,
	Size: 1024,
}

// BriaRmBg implements the Model interface for BRIA RMBG segmentation.
type BriaRmBg struct{}

func (m *BriaRmBg) Name() string {
	return "bria-rmbg"
}

func (m *BriaRmBg) DownloadURL() string {
	return briaRmBgModelURL
}

func (m *BriaRmBg) Download(opts onnx.DownloadOptions) error {
	return utils.DownloadUrl(briaRmBgModelURL, opts.DestPath)
}

func (m *BriaRmBg) Normalize(img image.Image) ([]ort.Value, error) {
	return NormalizeImage(img, briaRmBgParams)
}

func (m *BriaRmBg) Predict(img image.Image, outputs []ort.Value) ([]image.Image, error) {
	return PredictSingleMask(img, outputs)
}
