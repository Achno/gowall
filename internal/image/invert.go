package image

import (
	"errors"
	"image"

	types "github.com/Achno/gowall/internal/types"
	"github.com/disintegration/imaging"
)

type Inverter struct {
}

func (Invrt *Inverter) Process(img image.Image, theme string, format string) (image.Image, types.ImageMetadata, error) {

	newImg, err := invertImage(img)

	if err != nil {
		return nil, types.ImageMetadata{}, err
	}

	return newImg, types.ImageMetadata{}, nil
}

func invertImage(img image.Image) (image.Image, error) {
	newImg := imaging.Invert(img)
	if newImg == nil {
		return nil, errors.New("error while inverting the image")
	}

	return newImg, nil
}
