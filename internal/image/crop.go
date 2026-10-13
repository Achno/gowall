package image

import (
	"fmt"
	"image"

	"github.com/disintegration/imaging"

	types "github.com/Achno/gowall/internal/types"
)

// CropProcessor cuts Rect out of the image, Rect is in the image's pixels.
type CropProcessor struct {
	Rect image.Rectangle
}

func (p *CropProcessor) Process(img image.Image, theme string, format string) (image.Image, types.ImageMetadata, error) {
	cropped := imaging.Crop(img, p.Rect)
	if cropped.Bounds().Empty() {
		return nil, types.ImageMetadata{}, fmt.Errorf("the crop region %v is outside the image %v", p.Rect, img.Bounds())
	}
	return cropped, types.ImageMetadata{}, nil
}
