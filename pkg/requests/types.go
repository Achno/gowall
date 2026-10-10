package request

import (
	"context"
	"fmt"
	"io"
)

type Ctx struct {
	body        any
	header      Header
	query       Query
	contentType string
	context     context.Context
	files       []formFile
}

type formFile struct {
	field    string
	filename string
	reader   io.Reader
}

type Response[T any] struct {
	Code    int    `json:"code"`
	Data    T      `json:"data"`
	Message string `json:"message"`
}

type StatusError struct {
	Code int
	Body []byte
}

func (e *StatusError) Error() string {
	return fmt.Sprintf("status code: %d", e.Code)
}

type Query map[string]string
type Header map[string]string
