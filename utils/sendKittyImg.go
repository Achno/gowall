package utils

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/png"
	"io"
	"os"

	"golang.org/x/sys/unix"
)

// SendKittyImg sends a file as a Kitty protocol image with aspect-ratio scaling,
// writing directly to /dev/tty to avoid interference from terminal input/output.
func SendKittyImg(filePath string) error {
	// Open the image file.
	f, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("error opening image file: %w", err)
	}
	defer f.Close()
	// Turns out the kitty graphics protocol prefers PNG
	// https://sw.kovidgoyal.net/kitty/graphics-protocol/#png-data
	// so we Decode the image (this works for JPEG, PNG, GIF, etc.) from the filePath
	img, _, err := image.Decode(f)
	if err != nil {
		return fmt.Errorf("error decoding image: %w", err)
	}

	// Create an in-memory buffer to hold the PNG-encoded image.
	// This will be a temporary place in memory simply to easily show an inline preview
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return fmt.Errorf("error encoding PNG: %w", err)
	}

	intRows, intCols := computeDesiredTextSize(img)

	// Open /dev/tty for writing to avoid interference.
	ttyWriter, err := os.OpenFile("/dev/tty", int(unix.O_WRONLY|unix.O_NOCTTY), 0)
	if err != nil {
		return fmt.Errorf("error opening /dev/tty for writing: %w", err)
	}
	defer ttyWriter.Close()

	// Begin the Kitty image escape sequence.
	// a=T is the documented mode (non-scrolling) and we use text-cell sizing (r and c).
	_, err = fmt.Fprintf(ttyWriter, "\x1b_Gf=100,a=T,X=100,r=%d,c=%d;", intRows, intCols)
	if err != nil {
		return fmt.Errorf("error writing escape sequence: %w", err)
	}

	// Base64-encode the PNG data directly to ttyWriter.
	encoder := base64.NewEncoder(base64.StdEncoding, ttyWriter)
	if _, err := io.Copy(encoder, &buf); err != nil {
		return fmt.Errorf("error base64 encoding image data: %w", err)
	}
	encoder.Close()

	// Terminate the escape sequence including newline.
	_, err = fmt.Fprintf(ttyWriter, "\x1b\\\n")
	if err != nil {
		return fmt.Errorf("error writing final escape sequence: %w", err)
	}

	return nil
}

// getTerminalDimensions returns the terminal's text cell dimensions (rows, cols)
// and its pixel dimensions (width, height) by querying /dev/tty.
func getTerminalDimensions() (rows, cols, pxWidth, pxHeight int) {
	tty, err := os.OpenFile("/dev/tty", int(unix.O_RDWR|unix.O_NOCTTY), 0)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error opening /dev/tty: %v\n", err)
		os.Exit(1)
	}
	defer tty.Close()

	sz, err := unix.IoctlGetWinsize(int(tty.Fd()), unix.TIOCGWINSZ)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error getting window size: %v\n", err)
		os.Exit(1)
	}

	return int(sz.Row), int(sz.Col), int(sz.Xpixel), int(sz.Ypixel)
}

// getImageAspect returns the images og dimensions/aspect ratio
func getImageAspect(img image.Image) float64 {

	// Get original image dimensions (in pixels)
	imgWidth := img.Bounds().Dx()
	imgHeight := img.Bounds().Dy()

	// image pixel aspect ratio.
	return float64(imgWidth) / float64(imgHeight)
}

// computeDesiredTextSize calculates and returns the desired number of text cells (r and c)
// that should be used for the image, preserving its aspect ratio.
// It uses both the image dimensions and the terminal's text and pixel dimensions.
func computeDesiredTextSize(img image.Image) (int, int) {
	// Get terminal size: text cells and pixel dimensions.
	termRows, termCols, termPxWidth, termPxHeight := getTerminalDimensions()

	// Get image aspect ratio.
	imgAspect := getImageAspect(img)

	// Compute the size of one cell (in pixels).
	cellWidth := float64(termPxWidth) / float64(termCols)
	cellHeight := float64(termPxHeight) / float64(termRows)
	cellAspect := cellWidth / cellHeight

	// Adjust image aspect ratio to account for the non-square text cells.
	effectiveAspect := imgAspect / cellAspect

	// Use 90% of available text cells as the maximum.
	maxCols := float64(termCols) * 0.9
	maxRows := float64(termRows) * 0.9

	// Start with the maximum width and compute the height from the effective aspect ratio.
	desiredCols := maxCols
	desiredRows := desiredCols / effectiveAspect

	// If the computed rows exceed the maximum available, recalc based on height.
	if desiredRows > maxRows {
		desiredRows = maxRows
		desiredCols = desiredRows * effectiveAspect
	}

	// Return as integers.
	return int(desiredRows), int(desiredCols)
}
