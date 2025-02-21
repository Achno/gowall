package utils

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/png"
	"io"
	"os"
)

// SendKittyImg sends a file as a Kitty protocol inline image without any extra output.
// The escape sequence is formatted per the spec: begins with ESC _ G, parameters terminated with a semicolon,
// then the base64 data, and finally terminated with ESC \.
func SendKittyImg(filePath string) error {
	// Open the file.
	f, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("Error opening image file")
	}
	defer f.Close()

	// Turns out the kitty graphics protocol prefers PNG
	// https://sw.kovidgoyal.net/kitty/graphics-protocol/#png-data
	// so we Decode the image (this works for JPEG, PNG, GIF, etc.) from the filePath
	img, _, err := image.Decode(f)
	if err != nil {
		return fmt.Errorf("Error decoding image")
	}

	// Create an in-memory buffer to hold the PNG-encoded image.
	// This will be a temporary place in memory simply to easily show an inline preview
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return fmt.Errorf("Error encoding to buffer")
	}

	// Begin the inline image sequence with the correct parameter termination.
	fmt.Printf("\x1b_Gf=100,a=T;")

	// Base64-encode the file contents directly to stdout.
	encoder := base64.NewEncoder(base64.StdEncoding, os.Stdout)
	if _, err := io.Copy(encoder, &buf); err != nil {
		return fmt.Errorf("Error base64 encoding")
	}
	encoder.Close()

	// Terminate the escape sequence including newline.
	fmt.Printf("\x1b\\\n")
	return nil
}
