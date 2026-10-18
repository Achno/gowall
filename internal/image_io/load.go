package imageio

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	"io"
	"os"

	"github.com/Achno/gowall/internal/backends/fastpng"
	gifbackend "github.com/Achno/gowall/internal/backends/gif"
	webp "github.com/chai2010/webp"
	avif "github.com/gen2brain/avif"
	_ "golang.org/x/image/webp"
)

// Available formats to Encode an image in
var encoders = map[string]func(file *os.File, img image.Image) error{
	"png": func(file *os.File, img image.Image) error {
		return fastpng.Encode(file, img)
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
	"avif": func(file *os.File, img image.Image) error {
		return avif.Encode(file, img)
	},
}

func LoadImage(imgSrc ImageReader) (image.Image, error) {
	// For NoInput, return a placeholder image (won't be used by generators)
	if _, ok := imgSrc.(NoInput); ok {
		return image.NewRGBA(image.Rect(0, 0, 1, 1)), nil
	}

	reader, err := imgSrc.Open()
	if err != nil {
		return nil, err
	}
	defer reader.Close()
	imgData, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}
	// animated gifs are returned as a *types.Animation which holds every frame
	if bytes.HasPrefix(imgData, []byte("GIF8")) {
		anim, err := gifbackend.NewDecoder().Decode(bytes.NewReader(imgData))
		if err != nil {
			return nil, fmt.Errorf("%s: %w", imgSrc.String(), err)
		}
		if len(anim.Frames) == 1 {
			return anim.Frames[0], nil
		}
		return anim, nil
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
