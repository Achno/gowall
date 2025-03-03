package utils

import (
	"io"
	"os"
)

// Input image abstraction
type ImageSource interface {
	Open() (io.Reader, error)
}

type FileSource struct {
	Path string
}

func (fs FileSource) Open() (io.Reader, error) {
	f, err := os.Open(fs.Path)
	if err != nil {
		return nil, err
	}
	return f, nil
}

type StdinSource struct{}

func (ss StdinSource) Open() (io.Reader, error) {
	return os.Stdin, nil
}
