package request

import (
	"context"
	"io"
	"net/http"
)

type ReqOpt func(c *Client)

type Opt func(ctx *Ctx)

func WithClient(client *http.Client) ReqOpt {
	return func(c *Client) {
		c.client = client
	}
}

func WithTransport(tr *http.Transport) ReqOpt {
	return func(c *Client) {
		c.tr = tr
	}
}

func WithHeader(h Header) Opt {
	return func(ctx *Ctx) {
		ctx.header = h
	}
}

func WithQuery(q Query) Opt {
	return func(ctx *Ctx) {
		ctx.query = q
	}
}

func WithBody(body any) Opt {
	return func(ctx *Ctx) {
		ctx.body = body
	}
}

func WithContentType(contentType string) Opt {
	return func(ctx *Ctx) {
		ctx.contentType = contentType
	}
}

func WithContext(requestContext context.Context) Opt {
	return func(ctx *Ctx) {
		ctx.context = requestContext
	}
}

// WithFile adds a file to the form and makes the request multipart/form-data.
func WithFile(field, filename string, r io.Reader) Opt {
	return func(ctx *Ctx) {
		ctx.files = append(ctx.files, formFile{field: field, filename: filename, reader: r})
		ctx.contentType = "multipart/form-data"
	}
}
