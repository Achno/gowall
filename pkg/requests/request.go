package request

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Client struct {
	scheme string
	host   string
	client *http.Client
	tr     *http.Transport
}

func NewClient(scheme string, host string, timeout time.Duration, opts ...ReqOpt) *Client {
	req := &Client{
		scheme: scheme,
		host:   host,
		client: &http.Client{
			Timeout: timeout,
		},
	}

	for _, opt := range opts {
		opt(req)
	}

	if req.tr != nil {
		req.client.Transport = req.tr
	}

	return req
}

func NewURLClient(timeout time.Duration, opts ...ReqOpt) *Client {
	return NewClient("", "", timeout, opts...)
}

func (c *Client) SetTransport(tr *http.Transport) {
	c.client.Transport = tr
}

func sendRequest[T any](c *Client, method, path string, opts ...Opt) (*T, error) {
	u, err := url.Parse(path)
	if err != nil {
		return nil, err
	}
	if !u.IsAbs() {
		baseURL := &url.URL{Scheme: c.scheme, Host: c.host}
		u = baseURL.ResolveReference(u)
	}

	ctx := &Ctx{}
	for _, opt := range opts {
		opt(ctx)
	}

	if len(ctx.query) > 0 {
		values := u.Query()
		for k, v := range ctx.query {
			values.Add(k, v)
		}
		u.RawQuery = values.Encode()
	}

	body, contentType, err := encodeBody(ctx)
	if err != nil {
		return nil, err
	}

	var req *http.Request
	if ctx.context == nil {
		req, err = http.NewRequest(method, u.String(), body)
	} else {
		req, err = http.NewRequestWithContext(ctx.context, method, u.String(), body)
	}
	if err != nil {
		return nil, err
	}
	for k, v := range ctx.header {
		req.Header.Add(k, v)
	}
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, &StatusError{Code: resp.StatusCode, Body: b}
	}
	if len(b) == 0 {
		return new(T), nil
	}

	var rr T
	// T = []byte gives back the raw body, for responses that aren't json like HTML
	if raw, ok := any(&rr).(*[]byte); ok {
		*raw = b
		return &rr, nil
	}
	if err := json.Unmarshal(b, &rr); err != nil {
		return nil, err
	}
	return &rr, nil
}

// encodeBody returns the request body and its Content-Type, no body means no Content-Type.
func encodeBody(ctx *Ctx) (io.Reader, string, error) {
	if ctx.contentType == "multipart/form-data" {
		return encodeMultipart(ctx) // can be just a file, without fields
	}
	if ctx.body == nil {
		return nil, "", nil
	}

	switch ctx.contentType {
	case "application/x-www-form-urlencoded":
		fields, err := toStringMap(ctx.body)
		if err != nil {
			return nil, "", err
		}
		data := url.Values{}
		for k, v := range fields {
			data.Add(k, v)
		}
		return strings.NewReader(data.Encode()), ctx.contentType, nil
	case "", "application/json":
		bs, err := json.Marshal(ctx.body)
		if err != nil {
			return nil, "", err
		}
		return bytes.NewReader(bs), "application/json", nil
	default:
		return nil, "", fmt.Errorf("unsupported content type %q", ctx.contentType)
	}
}

func encodeMultipart(ctx *Ctx) (io.Reader, string, error) {
	if ctx.body == nil && len(ctx.files) == 0 {
		return nil, "", errors.New("multipart/form-data requires a body or a file")
	}

	fields, err := toStringMap(ctx.body)
	if err != nil {
		return nil, "", err
	}

	buf := &bytes.Buffer{}
	writer := multipart.NewWriter(buf)
	for k, v := range fields {
		if err := writer.WriteField(k, v); err != nil {
			return nil, "", err
		}
	}
	for _, f := range ctx.files {
		part, err := writer.CreateFormFile(f.field, f.filename)
		if err != nil {
			return nil, "", err
		}
		if _, err := io.Copy(part, f.reader); err != nil {
			return nil, "", err
		}
	}
	if err := writer.Close(); err != nil {
		return nil, "", err
	}

	return buf, writer.FormDataContentType(), nil
}

// toStringMap turns a body like map[string]string or a struct with string fields into form fields.
func toStringMap(body any) (map[string]string, error) {
	m := make(map[string]string)
	if body == nil {
		return m, nil
	}

	bs, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(bs, &m); err != nil {
		return nil, err
	}
	return m, nil
}

func Get[T any](c *Client, path string, opts ...Opt) (*T, error) {
	return sendRequest[T](c, http.MethodGet, path, opts...)
}

func Post[T any](c *Client, path string, body any, opts ...Opt) (*T, error) {
	opts = append(opts, WithBody(body))
	return sendRequest[T](c, http.MethodPost, path, opts...)
}

func Put[T any](c *Client, path string, body any, opts ...Opt) (*T, error) {
	opts = append(opts, WithBody(body))
	return sendRequest[T](c, http.MethodPut, path, opts...)
}

func Delete[T any](c *Client, path string, opts ...Opt) (*T, error) {
	return sendRequest[T](c, http.MethodDelete, path, opts...)
}

func GetHeaderMap(header string) map[string]string {
	headerMap := make(map[string]string)
	for _, h := range strings.Split(header, "\n") {
		if key, value, ok := strings.Cut(h, "="); ok {
			headerMap[key] = value
		}
	}
	return headerMap
}
