package sessions

import (
	"fmt"
	"image"

	"github.com/disintegration/imaging"
	ort "github.com/yalue/onnxruntime_go"
)

// NormalizeParams holds normalization parameters for image preprocessing
type NormalizeParams struct {
	Mean [3]float32
	Std  [3]float32
	Size int
}

// ImageNet normalization constants (commonly used across models)
var ImageNetParams = NormalizeParams{
	Mean: [3]float32{0.485, 0.456, 0.406},
	Std:  [3]float32{0.229, 0.224, 0.225},
	Size: 320,
}

// NormalizeImage preprocesses an image for model inference.
// Returns a tensor in NCHW format [1, 3, H, W].
func NormalizeImage(img image.Image, params NormalizeParams) ([]ort.Value, error) {
	resized := imaging.Resize(img, params.Size, params.Size, imaging.Lanczos)
	bounds := resized.Bounds()
	w, h := bounds.Dx(), bounds.Dy()

	maxVal := findMaxPixelValue(resized)
	tensorData := buildNCHWTensor(resized, w, h, maxVal, params.Mean, params.Std)

	shape := ort.Shape{1, 3, int64(h), int64(w)}
	tensor, err := ort.NewTensor(shape, tensorData)
	if err != nil {
		return nil, fmt.Errorf("create input tensor: %w", err)
	}

	return []ort.Value{tensor}, nil
}

// findMaxPixelValue scans the image and returns the maximum pixel value (0-1 range)
func findMaxPixelValue(img image.Image) float32 {
	bounds := img.Bounds()
	var maxVal float32 = 1e-6

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, _ := img.At(x, y).RGBA()
			rf := float32(r) / 65535.0
			gf := float32(g) / 65535.0
			bf := float32(b) / 65535.0

			if rf > maxVal {
				maxVal = rf
			}
			if gf > maxVal {
				maxVal = gf
			}
			if bf > maxVal {
				maxVal = bf
			}
		}
	}

	return maxVal
}

// buildNCHWTensor creates a normalized tensor in NCHW format [1, 3, H, W]
func buildNCHWTensor(img image.Image, w, h int, maxVal float32, mean, std [3]float32) []float32 {
	tensorData := make([]float32, 3*h*w)
	bounds := img.Bounds()

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			r, g, b, _ := img.At(bounds.Min.X+x, bounds.Min.Y+y).RGBA()

			// Normalize to [0,1], scale by max, then apply mean/std normalization
			rf := ((float32(r) / 65535.0) / maxVal - mean[0]) / std[0]
			gf := ((float32(g) / 65535.0) / maxVal - mean[1]) / std[1]
			bf := ((float32(b) / 65535.0) / maxVal - mean[2]) / std[2]

			// NCHW layout: [channel][height][width]
			tensorData[0*h*w+y*w+x] = rf // R channel
			tensorData[1*h*w+y*w+x] = gf // G channel
			tensorData[2*h*w+y*w+x] = bf // B channel
		}
	}

	return tensorData
}
