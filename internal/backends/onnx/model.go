package onnx

import (
	"image"

	ort "github.com/yalue/onnxruntime_go"
)

type DownloadOptions struct {
	DestPath string
}

type Model interface {
	Name() string
	DownloadURL() string
	Download(DownloadOptions) error
	Normalize(img image.Image) ([]ort.Value, error)
	Predict(img image.Image, outputs []ort.Value) ([]image.Image, error)
}
