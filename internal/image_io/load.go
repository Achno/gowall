package imageio

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"os"

	webp "github.com/HugoSmits86/nativewebp"
	_ "golang.org/x/image/webp"
)

// Available formats to Encode an image in
var encoders = map[string]func(file *os.File, img image.Image) error{
	"png": func(file *os.File, img image.Image) error {
		png := &png.Encoder{
			CompressionLevel: png.BestSpeed,
		}
		return png.Encode(file, img)
	},
	"jpg": func(file *os.File, img image.Image) error {
		return jpeg.Encode(file, img, nil)
	},
	"jpeg": func(file *os.File, img image.Image) error {
		return jpeg.Encode(file, img, nil)
	},
	"webp": func(file *os.File, img image.Image) error {
		return webp.Encode(file, img, nil)
	},
}

func LoadImage(imgSrc ImageReader) (image.Image, error) {
	reader, err := imgSrc.Open()
	if err != nil {
		return nil, err
	}
	defer reader.Close()
	imgData, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}
	img, _, err := image.Decode(bytes.NewReader(imgData))
	if err != nil {
		return nil, fmt.Errorf("unknown format : %s", imgSrc.String())
	}
	return img, nil
}

func LoadFileBytes(src ImageReader) ([]byte, error) {
	reader, err := src.Open()
	if err != nil {
		return nil, err
	}
	defer reader.Close()
	bytes, err := io.ReadAll(reader)

	return bytes, err
}
