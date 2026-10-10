package imgupload

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/Achno/gowall/config"
	request "github.com/Achno/gowall/pkg/requests"
)

const (
	freeImageHost = "freeimage.host"
	freeImagePath = "/api/1/upload"
)

type FreeImage struct {
	APIKey string
	client *request.Client
}

type FreeImageResponse struct {
	URL       string
	ViewerURL string
}

type freeImageAPIResponse struct {
	Image struct {
		URL       string `json:"url"`
		ViewerURL string `json:"url_viewer"`
	} `json:"image"`
	Error struct {
		Message string `json:"message"`
	} `json:"error"`
}

func NewFreeImageClient(apiKey string) (*FreeImage, error) {
	if apiKey == "" {
		return nil, errors.New("freeimage needs FREEIMAGE_API_KEY in your .env, copy the global shared key shown on https://freeimage.host/api")
	}
	return &FreeImage{
		APIKey: apiKey,
		client: request.NewClient("https", freeImageHost, config.UploadImageTimeout),
	}, nil
}

func (c *FreeImage) Upload(ctx context.Context, img io.Reader, filename string) (FreeImageResponse, error) {
	fields := map[string]string{"key": c.APIKey, "action": "upload", "format": "json"}

	res, err := request.Post[freeImageAPIResponse](c.client, freeImagePath, fields,
		request.WithContext(ctx),
		request.WithHeader(request.Header{"User-Agent": "gowall/" + config.Version}),
		request.WithFile("source", filename, img),
	)
	if err != nil {
		var statusErr *request.StatusError
		if errors.As(err, &statusErr) {
			var body freeImageAPIResponse
			if json.Unmarshal(statusErr.Body, &body) == nil && body.Error.Message != "" {
				return FreeImageResponse{}, fmt.Errorf("freeimage: %d %s: %s", statusErr.Code, http.StatusText(statusErr.Code), body.Error.Message)
			}
		}
		return FreeImageResponse{}, fmt.Errorf("freeimage: %w", err)
	}
	if res.Image.URL == "" {
		return FreeImageResponse{}, errors.New("freeimage: response has no link")
	}

	return FreeImageResponse{URL: res.Image.URL, ViewerURL: res.Image.ViewerURL}, nil
}
