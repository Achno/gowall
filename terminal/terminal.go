package terminal

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/png"
	"io"
	"os"
	"regexp"
	"strconv"

	"golang.org/x/term"
)

// RenderKittyImg sends a file as a Kitty protocol image with aspect-ratio scaling,
// writing directly to /dev/tty to avoid interference from terminal input/output.
func RenderKittyImg(filePath string) error {

	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("error opening image file: %w", err)
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		return fmt.Errorf("error decoding image: %w", err)
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return fmt.Errorf("error encoding PNG: %w", err)
	}

	// Calculate dimensions in text cells
	intRows, intCols := computeDesiredTextSize(img)

	// Open terminal for direct writing
	ttyWriter, err := os.OpenFile("/dev/tty", os.O_WRONLY, 0)
	if err != nil {
		return fmt.Errorf("error opening /dev/tty: %w", err)
	}
	defer ttyWriter.Close()

	// Start Kitty image protocol sequence
	_, err = fmt.Fprintf(ttyWriter, "\x1b_Gf=100,a=T,r=%d,c=%d;", intRows, intCols)
	if err != nil {
		return fmt.Errorf("error writing header: %w", err)
	}

	// Stream base64-encoded image data
	encoder := base64.NewEncoder(base64.StdEncoding, ttyWriter)
	if _, err := io.Copy(encoder, &buf); err != nil {
		return fmt.Errorf("error writing image data: %w", err)
	}
	encoder.Close()

	// Terminate the image sequence
	_, err = fmt.Fprintf(ttyWriter, "\x1b\\\n")
	if err != nil {
		return fmt.Errorf("error writing footer: %w", err)
	}

	return nil
}

// getTerminalDimensions retrieves terminal size in text cells and pixels
func getTerminalDimensions() (rows, cols, pxWidth, pxHeight int) {
	// Open /dev/tty for Linux & MacOS , CONOUT$ on Windows
	// for writing
	ttyWrite, err := os.OpenFile("CONOUT$", os.O_WRONLY, 0)
	if err != nil {
		ttyWrite, err = os.OpenFile("/dev/tty", os.O_WRONLY, 0)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to open terminal for writing: %v\n", err)
			os.Exit(1)
		}
	}
	defer ttyWrite.Close()

	// Open terminal for reading responses
	ttyRead, err := os.OpenFile("CONOUT$", os.O_RDONLY, 0)
	if err != nil {
		ttyRead, err = os.OpenFile("/dev/tty", os.O_RDONLY, 0)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to open terminal for reading: %v\n", err)
			os.Exit(1)
		}
	}
	defer ttyRead.Close()

	// Switch to raw mode and ensure it gets restored
	oldState, err := term.MakeRaw(int(ttyRead.Fd()))
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to set raw mode: %v\n", err)
		os.Exit(1)
	}
	defer term.Restore(int(ttyRead.Fd()), oldState)

	// query the terminal for its dimensions via ANSI escape codes
	_, err = ttyWrite.Write([]byte("\033[18t\033[14t"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to write to terminal: %v\n", err)
		os.Exit(1)
	}

	// Read responses
	var buf [32]byte
	var response []byte
	for {
		n, err := ttyRead.Read(buf[:])
		if err != nil || n == 0 {
			break
		}
		response = append(response, buf[:n]...)
		if bytes.Count(response, []byte("t")) >= 2 {
			break
		}
	}

	// Parse text dimensions
	reText := regexp.MustCompile(`\033\[8;(\d+);(\d+)t`)
	matchesText := reText.FindStringSubmatch(string(response))
	if len(matchesText) == 3 {
		rows, _ = strconv.Atoi(matchesText[1])
		cols, _ = strconv.Atoi(matchesText[2])
	}

	// Parse pixel dimensions
	rePixel := regexp.MustCompile(`\033\[4;(\d+);(\d+)t`)
	matchesPixel := rePixel.FindStringSubmatch(string(response))
	if len(matchesPixel) == 3 {
		pxHeight, _ = strconv.Atoi(matchesPixel[1])
		pxWidth, _ = strconv.Atoi(matchesPixel[2])
	}

	return
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
