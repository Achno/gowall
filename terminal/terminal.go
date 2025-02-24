package terminal

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/png"
	"io"
	"os"
)

func RenderKittyImg(filePath string) error {

	file, err := os.Open(filePath)
	if err != nil {
		return err
	}

	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		return err
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return fmt.Errorf("while encoding to buffer")
	}

	// Begin the inline image sequence with the correct parameter termination.
	fmt.Printf("\x1b_Gf=100,a=T;")

	encoder := base64.NewEncoder(base64.StdEncoding, os.Stdout)
	if _, err := io.Copy(encoder, &buf); err != nil {
		return fmt.Errorf("while base64 encoding")
	}
	encoder.Close()

	// Terminate the escape sequence including newline.
	fmt.Printf("\x1b\\\n")
	return nil
}
