package utils

import (
	"encoding/base64"
	"fmt"
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
		return err
	}
	defer f.Close()

	// Begin the inline image sequence with the correct parameter termination.
	fmt.Printf("\x1b_Gf=100,a=T;")
	
	// Base64-encode the file contents directly to stdout.
	encoder := base64.NewEncoder(base64.StdEncoding, os.Stdout)
	if _, err := io.Copy(encoder, f); err != nil {
		return err
	}
	encoder.Close()

	// Terminate the escape sequence correctly.
	fmt.Printf("\x1b\\\n")
	return nil
}

