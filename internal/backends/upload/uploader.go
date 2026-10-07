package upload

import (
	"context"

	imageio "github.com/Achno/gowall/internal/image_io"
)

type ImageUploader interface {
	Upload(ctx context.Context, img imageio.ImageReader) (ImgUploadResult, error)
}

type ImgUploadResult struct {
	URL      string
	Metadata any // whatever the uploader wants to give back (like the imgbb delete url)
}
