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

// RenderKittyImg renders an image with the Kitty protocol and aspect-ratio scaling,
func RenderKittyImg(filePath string) error {

	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("opening image file: %w", err)
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		return fmt.Errorf("decoding image: %w", err)
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return fmt.Errorf("encoding PNG: %w", err)
	}

	intRows, intCols, err := textCells(img)
	if err != nil {
		return err
	}

	ttyWriter, err := os.OpenFile("/dev/tty", os.O_WRONLY, 0)
	if err != nil {
		return fmt.Errorf("opening /dev/tty: %w", err)
	}
	defer ttyWriter.Close()

	// Start Kitty image protocol sequence
	header := fmt.Sprintf("\x1b_Gf=100,a=T,r=%d,c=%d;", intRows, intCols)
	if _, err := io.WriteString(ttyWriter, header); err != nil {
		return fmt.Errorf("writing header: %w", err)
	}

	encoder := base64.NewEncoder(base64.StdEncoding, ttyWriter)
	if _, err := io.Copy(encoder, &buf); err != nil {
		encoder.Close()
		return fmt.Errorf("writing image data: %w", err)
	}
	if err := encoder.Close(); err != nil {
		return fmt.Errorf("closing encoder: %w", err)
	}

	// Terminate the image sequence
	if _, err := io.WriteString(ttyWriter, "\x1b\\\n"); err != nil {
		return fmt.Errorf("writing footer: %w", err)
	}

	return nil
}

// getTerminalDimensions retrieves terminal size in text cells needed for the kitty protocol and pixels
func getTerminalDimensions() (rows, cols, pxWidth, pxHeight int, err error) {
	// Open /dev/tty for Linux & MacOS , CONOUT$ on Windows
	ttyWrite, err := os.OpenFile("CONOUT$", os.O_WRONLY, 0)
	if err != nil {
		ttyWrite, err = os.OpenFile("/dev/tty", os.O_WRONLY, 0)
		if err != nil {
			return 0, 0, 0, 0, fmt.Errorf("failed to open tty for writing: %w", err)
		}
	}
	defer ttyWrite.Close()

	// Open terminal for reading responses
	ttyRead, err := os.OpenFile("CONOUT$", os.O_RDONLY, 0)
	if err != nil {
		ttyRead, err = os.OpenFile("/dev/tty", os.O_RDONLY, 0)
		if err != nil {
			return 0, 0, 0, 0, fmt.Errorf("failed to open tty for reading: %w", err)
		}
	}
	defer ttyRead.Close()

	// Switch to raw mode and ensure it gets restored
	oldState, err := term.MakeRaw(int(ttyRead.Fd()))
	if err != nil {
		return 0, 0, 0, 0, fmt.Errorf("failed to set raw mode: %w", err)
	}
	defer term.Restore(int(ttyRead.Fd()), oldState)

	// query the terminal for its dimensions via ANSI escape codes
	_, err = ttyWrite.Write([]byte("\033[18t\033[14t"))
	if err != nil {
		return 0, 0, 0, 0, fmt.Errorf("failed to set raw mode: %w", err)
	}

	// Read response
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

	reText := regexp.MustCompile(`\033\[8;(\d+);(\d+)t`)
	matchesText := reText.FindStringSubmatch(string(response))
	if len(matchesText) == 3 {
		rows, _ = strconv.Atoi(matchesText[1])
		cols, _ = strconv.Atoi(matchesText[2])
	}

	rePixel := regexp.MustCompile(`\033\[4;(\d+);(\d+)t`)
	matchesPixel := rePixel.FindStringSubmatch(string(response))
	if len(matchesPixel) == 3 {
		pxHeight, _ = strconv.Atoi(matchesPixel[1])
		pxWidth, _ = strconv.Atoi(matchesPixel[2])
	}

	return rows, cols, pxWidth, pxHeight, nil
}

// aspectRatio returns the aspect ratio of the image
func aspectRatio(img image.Image) float64 {
	imgWidth := img.Bounds().Dx()
	imgHeight := img.Bounds().Dy()

	return float64(imgWidth) / float64(imgHeight)
}

// textCells calculates and returns the desired number of text cells (r and c)
// that should be used for the image, preserving its aspect ratio.
func textCells(img image.Image) (int, int, error) {
	termRows, termCols, termPxWidth, termPxHeight, err := getTerminalDimensions()
	if err != nil {
		return 0, 0, fmt.Errorf("while getting terminal dimensions: %w", err)
	}

	imgAspect := aspectRatio(img)

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

	return int(desiredRows), int(desiredCols), nil
}
