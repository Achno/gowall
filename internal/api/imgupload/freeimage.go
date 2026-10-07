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
	"strings"

	"github.com/Achno/gowall/config"
)

const freeImageURL = "https://freeimage.host/api/1/upload"

type FreeImage struct {
	BaseURL string
	APIKey  string
	Client  *http.Client
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
	return &FreeImage{BaseURL: freeImageURL, APIKey: apiKey, Client: http.DefaultClient}, nil
}

func (c *FreeImage) Upload(ctx context.Context, img io.Reader, filename string) (FreeImageResponse, error) {
	var body bytes.Buffer
	form := multipart.NewWriter(&body)

	fields := map[string]string{"key": c.APIKey, "action": "upload", "format": "json"}
	for k, v := range fields {
		if err := form.WriteField(k, v); err != nil {
			return FreeImageResponse{}, fmt.Errorf("freeimage: %w", err)
		}
	}

	part, err := form.CreateFormFile("source", filename)
	if err != nil {
		return FreeImageResponse{}, fmt.Errorf("freeimage: %w", err)
	}
	if _, err := io.Copy(part, img); err != nil {
		return FreeImageResponse{}, fmt.Errorf("freeimage: reading image: %w", err)
	}
	if err := form.Close(); err != nil {
		return FreeImageResponse{}, fmt.Errorf("freeimage: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.BaseURL, &body)
	if err != nil {
		return FreeImageResponse{}, fmt.Errorf("freeimage: %w", err)
	}
	req.Header.Set("Content-Type", form.FormDataContentType())
	req.Header.Set("User-Agent", "gowall/"+config.Version)

	res, err := c.Client.Do(req)
	if err != nil {
		return FreeImageResponse{}, fmt.Errorf("freeimage: %w", err)
	}
	defer res.Body.Close()

	data, err := io.ReadAll(res.Body)
	if err != nil {
		return FreeImageResponse{}, fmt.Errorf("freeimage: reading response: %w", err)
	}

	var parsed freeImageAPIResponse
	jsonErr := json.Unmarshal(data, &parsed)

	if res.StatusCode != http.StatusOK {
		msg := parsed.Error.Message
		if jsonErr != nil || msg == "" {
			msg = strings.TrimSpace(string(data))
		}
		return FreeImageResponse{}, fmt.Errorf("freeimage: %s: %s", res.Status, msg)
	}
	if jsonErr != nil {
		return FreeImageResponse{}, fmt.Errorf("freeimage: reading response: %w", jsonErr)
	}
	if parsed.Image.URL == "" {
		return FreeImageResponse{}, fmt.Errorf("freeimage: response has no link: %s", strings.TrimSpace(string(data)))
	}

	return FreeImageResponse{URL: parsed.Image.URL, ViewerURL: parsed.Image.ViewerURL}, nil
}
