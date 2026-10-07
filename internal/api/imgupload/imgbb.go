package imgupload

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"

	"github.com/Achno/gowall/config"
)

const imgbbURL = "https://api.imgbb.com/1/upload"

type ImgBB struct {
	BaseURL string
	APIKey  string
	Client  *http.Client
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
	return &ImgBB{BaseURL: imgbbURL, APIKey: apiKey, Client: http.DefaultClient}, nil
}

func (c *ImgBB) Upload(ctx context.Context, img io.Reader, filename string) (ImgBBResponse, error) {
	var body bytes.Buffer
	form := multipart.NewWriter(&body)

	part, err := form.CreateFormFile("image", filename)
	if err != nil {
		return ImgBBResponse{}, fmt.Errorf("imgbb: %w", err)
	}
	if _, err := io.Copy(part, img); err != nil {
		return ImgBBResponse{}, fmt.Errorf("imgbb: reading image: %w", err)
	}
	if err := form.Close(); err != nil {
		return ImgBBResponse{}, fmt.Errorf("imgbb: %w", err)
	}

	endpoint, err := url.Parse(c.BaseURL)
	if err != nil {
		return ImgBBResponse{}, fmt.Errorf("imgbb: %w", err)
	}
	endpoint.RawQuery = url.Values{"key": {c.APIKey}}.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint.String(), &body)
	if err != nil {
		return ImgBBResponse{}, fmt.Errorf("imgbb: %w", err)
	}
	req.Header.Set("Content-Type", form.FormDataContentType())
	req.Header.Set("User-Agent", "gowall/"+config.Version)

	res, err := c.Client.Do(req)
	if err != nil {
		// *url.Error prints the full url, which has the api key in it
		var urlErr *url.Error
		if errors.As(err, &urlErr) {
			err = urlErr.Err
		}
		return ImgBBResponse{}, fmt.Errorf("imgbb: %w", err)
	}
	defer res.Body.Close()

	data, err := io.ReadAll(res.Body)
	if err != nil {
		return ImgBBResponse{}, fmt.Errorf("imgbb: reading response: %w", err)
	}

	var parsed imgbbAPIResponse
	jsonErr := json.Unmarshal(data, &parsed)

	if res.StatusCode != http.StatusOK {
		msg := parsed.Error.Message
		if jsonErr != nil || msg == "" {
			msg = strings.TrimSpace(string(data))
		}
		return ImgBBResponse{}, fmt.Errorf("imgbb: %s: %s", res.Status, msg)
	}
	if jsonErr != nil {
		return ImgBBResponse{}, fmt.Errorf("imgbb: reading response: %w", jsonErr)
	}
	if parsed.Data.URL == "" {
		return ImgBBResponse{}, fmt.Errorf("imgbb: response has no link: %s", strings.TrimSpace(string(data)))
	}

	return ImgBBResponse{
		URL:       parsed.Data.URL,
		ViewerURL: parsed.Data.ViewerURL,
		DeleteURL: parsed.Data.DeleteURL,
	}, nil
}
