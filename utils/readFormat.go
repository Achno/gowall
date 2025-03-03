package utils

import (
	"fmt"
	"io"
)

// We could use a library for this purpose

func checkPng(magicBytes []byte) bool {
	png := []byte{137, 80, 78, 71, 13, 10, 26, 10}
	for index, currentByte := range png {
		if magicBytes[index] != currentByte {
			return false
		}
	}
	return true
}

func checkJpeg(magicBytes []byte) bool {
	jpeg := []byte{0xFF, 0xD8, 0xFF}
	for index, currentByte := range jpeg {
		if magicBytes[index] != currentByte {
			return false
		}
	}
	return true
}

func checkWebp(magicBytes []byte) bool {
	webp := []byte{0x52, 0x49, 0x46, 0x46}
	for index, currentByte := range webp {
		if magicBytes[index] != currentByte {
			return false
		}
	}
	return true
}

const (
	pngExt  = "png"
	webpExt = "webp"
	jpgExt  = "jpg"
)

func ReadImageFormat(imgSrc ImageSource) (string, error) {
	magicBytes := make([]byte, 12)

	reader, err := imgSrc.Open()
	if err != nil {
		return "", err
	}

	if _, err := io.ReadFull(reader, magicBytes); err != nil {
		return "", err
	}
	if checkPng(magicBytes) {
		return pngExt, nil
	}
	if checkWebp(magicBytes) {
		return webpExt, nil
	}
	if checkJpeg(magicBytes) {
		return jpgExt, nil
	}
	return "", fmt.Errorf("uknown format")
}
