package upload

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/Achno/gowall/internal/api/imgupload"
	imageio "github.com/Achno/gowall/internal/image_io"
)

type freeImageUploader struct {
	client *imgupload.FreeImage
}

func newFreeImage(opts ImageUploadOptions) (ImageUploader, error) {
	client, err := imgupload.NewFreeImageClient(opts.FreeImageAPIKey)
	if err != nil {
		return nil, err
	}
	return freeImageUploader{client: client}, nil
}

func (u freeImageUploader) Upload(ctx context.Context, img imageio.ImageReader) (ImgUploadResult, error) {
	f, err := img.Open()
	if err != nil {
		return ImgUploadResult{}, fmt.Errorf("while opening image: %w", err)
	}
	defer f.Close()

	res, err := u.client.Upload(ctx, f, filepath.Base(img.String()))
	if err != nil {
		return ImgUploadResult{}, err
	}
	return ImgUploadResult{URL: res.URL, Metadata: res}, nil
}

type imgBBUploader struct {
	client *imgupload.ImgBB
}

func newImgBB(opts ImageUploadOptions) (ImageUploader, error) {
	client, err := imgupload.NewImgBBClient(opts.ImgBBAPIKey)
	if err != nil {
		return nil, err
	}
	return imgBBUploader{client: client}, nil
}

func (u imgBBUploader) Upload(ctx context.Context, img imageio.ImageReader) (ImgUploadResult, error) {
	f, err := img.Open()
	if err != nil {
		return ImgUploadResult{}, fmt.Errorf("while opening image: %w", err)
	}
	defer f.Close()

	res, err := u.client.Upload(ctx, f, filepath.Base(img.String()))
	if err != nil {
		return ImgUploadResult{}, err
	}
	return ImgUploadResult{URL: res.URL, Metadata: res}, nil
}
