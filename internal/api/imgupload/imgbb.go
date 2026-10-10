package imgupload

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/Achno/gowall/config"
	request "github.com/Achno/gowall/pkg/requests"
)

const (
	imgbbHost = "api.imgbb.com"
	imgbbPath = "/1/upload"
)

type ImgBB struct {
	APIKey string
	client *request.Client
}

type ImgBBResponse struct {
	URL       string
	ViewerURL string
	DeleteURL string
}

type imgbbAPIResponse struct {
	Data struct {
		URL       string `json:"url"`
		ViewerURL string `json:"url_viewer"`
		DeleteURL string `json:"delete_url"`
	} `json:"data"`
	Error struct {
		Message string `json:"message"`
	} `json:"error"`
}

func NewImgBBClient(apiKey string) (*ImgBB, error) {
	if apiKey == "" {
		return nil, errors.New("imgbb needs IMGBB_API_KEY in your .env, get one at https://api.imgbb.com")
	}
	return &ImgBB{
		APIKey: apiKey,
		client: request.NewClient("https", imgbbHost, config.UploadImageTimeout),
	}, nil
}

func (c *ImgBB) Upload(ctx context.Context, img io.Reader, filename string) (ImgBBResponse, error) {
	res, err := request.Post[imgbbAPIResponse](c.client, imgbbPath, nil,
		request.WithContext(ctx),
		request.WithQuery(request.Query{"key": c.APIKey}),
		request.WithHeader(request.Header{"User-Agent": "gowall/" + config.Version}),
		request.WithFile("image", filename, img),
	)
	if err != nil {
		// *url.Error prints the full url, which has the api key in it
		var urlErr *url.Error
		if errors.As(err, &urlErr) {
			err = urlErr.Err
		}

		var statusErr *request.StatusError
		if errors.As(err, &statusErr) {
			var body imgbbAPIResponse
			if json.Unmarshal(statusErr.Body, &body) == nil && body.Error.Message != "" {
				return ImgBBResponse{}, fmt.Errorf("imgbb: %d %s: %s", statusErr.Code, http.StatusText(statusErr.Code), body.Error.Message)
			}
		}
		return ImgBBResponse{}, fmt.Errorf("imgbb: %w", err)
	}
	if res.Data.URL == "" {
		return ImgBBResponse{}, errors.New("imgbb: response has no link")
	}

	return ImgBBResponse{
		URL:       res.Data.URL,
		ViewerURL: res.Data.ViewerURL,
		DeleteURL: res.Data.DeleteURL,
	}, nil
}
