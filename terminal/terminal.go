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

// RenderKittyImg renders an image using the Kitty protocol and aspect-ratio scaling.
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

// getTerminalDimensions retrieves the terminal dimensions (text cells and pixel sizes) required for the Kitty protocol.
func getTerminalDimensions() (rows, cols, pxWidth, pxHeight int, err error) {
	// Open tty for writing; try "CONOUT$" (Windows) then "/dev/tty" (Linux/MacOS)
	ttyWrite, err := os.OpenFile("CONOUT$", os.O_WRONLY, 0)
	if err != nil {
		ttyWrite, err = os.OpenFile("/dev/tty", os.O_WRONLY, 0)
		if err != nil {
			return 0, 0, 0, 0, fmt.Errorf("failed to open tty for writing: %w", err)
		}
	}
	defer ttyWrite.Close()

	// Open tty for reading
	ttyRead, err := os.OpenFile("CONOUT$", os.O_RDONLY, 0)
	if err != nil {
		ttyRead, err = os.OpenFile("/dev/tty", os.O_RDONLY, 0)
		if err != nil {
			return 0, 0, 0, 0, fmt.Errorf("failed to open tty for reading: %w", err)
		}
	}
	defer ttyRead.Close()

	// Set raw mode so we can read terminal responses directly.
	oldState, err := term.MakeRaw(int(ttyRead.Fd()))
	if err != nil {
		return 0, 0, 0, 0, fmt.Errorf("failed to set raw mode: %w", err)
	}
	defer term.Restore(int(ttyRead.Fd()), oldState)

	// Query the terminal for its dimensions via ANSI escape codes.
	// \033[18t requests text cell dimensions; \033[14t requests pixel dimensions.
	if _, err = ttyWrite.Write([]byte("\033[18t\033[14t")); err != nil {
		return 0, 0, 0, 0, fmt.Errorf("failed to send terminal query: %w", err)
	}

	// Read the terminal response.
	var buf [32]byte
	var response []byte
	for {
		n, err := ttyRead.Read(buf[:])
		if err != nil {
			return 0, 0, 0, 0, fmt.Errorf("failed to read terminal response: %w", err)
		}
		if n == 0 {
			break
		}
		response = append(response, buf[:n]...)
		// Break if expected escape sequences for text and pixel dimensions are detected.
		if bytes.Contains(response, []byte("\033[8;")) && bytes.Contains(response, []byte("\033[4;")) {
			break
		}
	}

	// Parse text dimensions: "\033[8;<rows>;<cols>t"
	reText := regexp.MustCompile(`\033\[8;(\d+);(\d+)t`)
	matchesText := reText.FindStringSubmatch(string(response))
	if len(matchesText) == 3 {
		rows, err = strconv.Atoi(matchesText[1])
		if err != nil {
			return 0, 0, 0, 0, fmt.Errorf("converting rows: %w", err)
		}
		cols, err = strconv.Atoi(matchesText[2])
		if err != nil {
			return 0, 0, 0, 0, fmt.Errorf("converting cols: %w", err)
		}
	}

	// Parse pixel dimensions: "\033[4;<pxHeight>;<pxWidth>t"
	rePixel := regexp.MustCompile(`\033\[4;(\d+);(\d+)t`)
	matchesPixel := rePixel.FindStringSubmatch(string(response))
	if len(matchesPixel) == 3 {
		pxHeight, err = strconv.Atoi(matchesPixel[1])
		if err != nil {
			return 0, 0, 0, 0, fmt.Errorf("converting pixel height: %w", err)
		}
		pxWidth, err = strconv.Atoi(matchesPixel[2])
		if err != nil {
			return 0, 0, 0, 0, fmt.Errorf("converting pixel width: %w", err)
		}
	}

	return rows, cols, pxWidth, pxHeight, nil
}

// aspectRatio returns the aspect ratio of the image.
func aspectRatio(img image.Image) float64 {
	return float64(img.Bounds().Dx()) / float64(img.Bounds().Dy())
}

// textCells calculates and returns the desired number of text cells (rows and columns)
// to display the image, preserving its aspect ratio.
func textCells(img image.Image) (int, int, error) {
	termRows, termCols, termPxWidth, termPxHeight, err := getTerminalDimensions()
	if err != nil {
		return 0, 0, fmt.Errorf("while getting terminal dimensions: %w", err)
	}

	imgAspect := aspectRatio(img)

	// Calculate the size of one text cell in pixels.
	cellWidth := float64(termPxWidth) / float64(termCols)
	cellHeight := float64(termPxHeight) / float64(termRows)
	cellAspect := cellWidth / cellHeight

	// Adjust image aspect ratio to account for non-square text cells.
	effectiveAspect := imgAspect / cellAspect

	// Use 90% of available text cells as the maximum area.
	maxCols := float64(termCols) * 0.9
	maxRows := float64(termRows) * 0.9

	// Start with maximum width and compute the height using the effective aspect ratio.
	desiredCols := maxCols
	desiredRows := desiredCols / effectiveAspect

	// If computed rows exceed the maximum available, recalc based on maximum rows.
	if desiredRows > maxRows {
		desiredRows = maxRows
		desiredCols = desiredRows * effectiveAspect
	}

	return int(desiredRows), int(desiredCols), nil
}
