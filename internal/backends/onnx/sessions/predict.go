package sessions

import (
	"fmt"
	"image"
	"image/color"
	"math"

	"github.com/disintegration/imaging"
	ort "github.com/yalue/onnxruntime_go"
)

// ExtractTensorData extracts float32 data from the first output tensor
func ExtractTensorData(outputs []ort.Value) ([]float32, ort.Shape, error) {
	if len(outputs) == 0 {
		return nil, nil, fmt.Errorf("no outputs from model")
	}

	tensor, ok := outputs[0].(*ort.Tensor[float32])
	if !ok {
		return nil, nil, fmt.Errorf("unexpected output tensor type")
	}

	return tensor.GetData(), tensor.GetShape(), nil
}

// GetMaskDimensions extracts height and width from output tensor shape
// Supports both [1, 1, H, W] and [1, H, W] formats
func GetMaskDimensions(shape ort.Shape) (h, w int64, err error) {
	if len(shape) < 3 {
		return 0, 0, fmt.Errorf("unexpected output shape: %v", shape)
	}

	if len(shape) == 4 {
		return shape[2], shape[3], nil
	}
	return shape[1], shape[2], nil
}

// NormalizeMaskData applies min-max normalization to prediction data
// Returns values in [0, 1] range
func NormalizeMaskData(data []float32) []float32 {
	minVal := float32(math.MaxFloat32)
	maxVal := float32(-math.MaxFloat32)

	for _, v := range data {
		if v < minVal {
			minVal = v
		}
		if v > maxVal {
			maxVal = v
		}
	}

	rangeVal := maxVal - minVal
	if rangeVal < 1e-6 {
		rangeVal = 1e-6
	}

	normalized := make([]float32, len(data))
	for i, v := range data {
		normalized[i] = (v - minVal) / rangeVal
	}

	return normalized
}

// CreateGrayscaleMask creates a grayscale image from normalized float data
func CreateGrayscaleMask(data []float32, w, h int64) *image.Gray {
	mask := image.NewGray(image.Rect(0, 0, int(w), int(h)))

	for y := int64(0); y < h; y++ {
		for x := int64(0); x < w; x++ {
			idx := y*w + x
			val := data[idx]

			// Clamp to [0, 1]
			if val < 0 {
				val = 0
			} else if val > 1 {
				val = 1
			}

			mask.SetGray(int(x), int(y), color.Gray{Y: uint8(val * 255)})
		}
	}

	return mask
}

// ResizeMask resizes a mask to match the original image dimensions
func ResizeMask(mask image.Image, origImg image.Image) image.Image {
	bounds := origImg.Bounds()
	return imaging.Resize(mask, bounds.Dx(), bounds.Dy(), imaging.Lanczos)
}

// PredictSingleMask is a convenience function that handles the common case of
// extracting a single grayscale mask from model output and resizing it
func PredictSingleMask(origImg image.Image, outputs []ort.Value) ([]image.Image, error) {
	data, shape, err := ExtractTensorData(outputs)
	if err != nil {
		return nil, err
	}

	h, w, err := GetMaskDimensions(shape)
	if err != nil {
		return nil, err
	}

	// Extract first channel only
	channelData := make([]float32, h*w)
	copy(channelData, data[:h*w])

	normalized := NormalizeMaskData(channelData)
	mask := CreateGrayscaleMask(normalized, w, h)
	resized := ResizeMask(mask, origImg)

	return []image.Image{resized}, nil
}
