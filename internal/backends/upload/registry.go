package upload

import (
	"fmt"
	"maps"
	"slices"
	"strings"
)

type ImageUploadOptions struct {
	FreeImageAPIKey string
	ImgBBAPIKey     string
}

const DefaultUploader = "freeimage"

func getUploaderFactories() map[string]func(ImageUploadOptions) (ImageUploader, error) {
	return map[string]func(ImageUploadOptions) (ImageUploader, error){
		"freeimage": newFreeImage,
		"imgbb":     newImgBB,
	}
}

func New(name string, opts ImageUploadOptions) (ImageUploader, error) {
	newUploader, ok := getUploaderFactories()[name]
	if !ok {
		return nil, fmt.Errorf("unknown uploader %q, use one of [%s]", name, strings.Join(Names(), ", "))
	}
	return newUploader(opts)
}

func Names() []string {
	return slices.Sorted(maps.Keys(getUploaderFactories()))
}
