package terminal

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"

	"github.com/disintegration/imaging"
	"golang.org/x/term"
)

// Image ids for the kitty protocol. Each image is uploaded once, after that
// moving the slider only re-sends tiny placement commands.
const (
	originalID = 1
	editedID   = 2
	barID      = 3
	barWidth   = 3 // px
)

const (
	// alt screen, hide cursor, clear, report mouse presses/drags in SGR format
	enterSliderScreen = "\x1b[?1049h\x1b[?25l\x1b[2J\x1b[?1002h\x1b[?1006h"
	// undoes enterSliderScreen in reverse order
	leaveSliderScreen = "\x1b[?1006l\x1b[?1002l\x1b[?25h\x1b[?1049l"
	moveCursor        = "\x1b[%d;%dH" // row, col starting from 1

	// kitty image commands, q=2 stops the terminal from replying on stdin
	transmitFirstChunk = "\x1b_Ga=t,f=100,i=%d,q=2,m=%d;%s\x1b\\"
	transmitNextChunk  = "\x1b_Gm=%d;%s\x1b\\"
	placeImage         = "\x1b_Ga=p,i=%d,p=1,z=0,C=1,q=2\x1b\\"
	placeCropped       = "\x1b_Ga=p,i=%d,p=1,w=%d,z=1,C=1,q=2\x1b\\" // only the first w px
	placeShifted       = "\x1b_Ga=p,i=%d,p=1,X=%d,z=2,C=1,q=2\x1b\\" // X px to the right inside the cell
	hideImage          = "\x1b_Ga=d,d=i,i=%d,q=2\x1b\\"              // removes the placements, keeps the image
	deleteImage        = "\x1b_Ga=d,d=I,i=%d,q=2\x1b\\"              // removes the placements and the image

	keyLeft    = "\x1b[D"
	keyRight   = "\x1b[C"
	mouseEvent = "\x1b[<" // ESC [ < btn ; col ; row M|m
)

// RenderKittySlider shows an interactive before/after slider, moved with the arrow keys or by dragging with the mouse.
// The original is placed full size and the edited image is placed on top, cropped to the bar position.
func RenderKittySlider(original, edited image.Image) error {
	s, closeScreen, err := openCompareScreen(original, edited, 1, "←/→ h/l or drag with the mouse to move the slider, q to quit")
	if err != nil {
		return err
	}
	defer closeScreen()

	s.barX = s.width / 2
	s.draw()
	return readInput(func(in []byte) bool {
		if s.handle(in) {
			return true
		}
		s.draw()
		return false
	})
}

// RenderKittySideBySide places the original and the edited image next to each other, split by the bar.
func RenderKittySideBySide(original, edited image.Image) error {
	s, closeScreen, err := openCompareScreen(original, edited, 2, "q to quit")
	if err != nil {
		return err
	}
	defer closeScreen()

	var b bytes.Buffer
	fmt.Fprintf(&b, moveCursor+placeShifted, 1, s.width/s.cellWidth+1, barID, s.width%s.cellWidth)
	x := s.width + barWidth
	fmt.Fprintf(&b, moveCursor+placeShifted, 1, x/s.cellWidth+1, editedID, x%s.cellWidth)
	os.Stdout.Write(b.Bytes())

	return readInput(func(in []byte) bool {
		return bytes.ContainsAny(in, "q\x03")
	})
}

// openCompareScreen fits the images so that `panels` of them fit next to each other, uploads them,
// switches to the alt screen and places the original in the top left. closeScreen undoes all of it.
func openCompareScreen(original, edited image.Image, panels int, help string) (s *slider, closeScreen func(), err error) {
	supportsKitty := IsKittyTerminalRunning() || IsWeztermTerminalRunning() || IsKonsoleTerminalRunning() || IsGhosttyTerminalRunning()
	if !supportsKitty {
		return nil, nil, fmt.Errorf("gowall compare needs a terminal with the kitty image protocol: kitty, wezterm, konsole or ghostty. ")
	}
	fd := int(os.Stdin.Fd())
	if !term.IsTerminal(fd) {
		return nil, nil, fmt.Errorf("comparing images needs an interactive terminal")
	}
	rows, cols, pxWidth, pxHeight, err := getTerminalDimensions()
	if err != nil {
		return nil, nil, fmt.Errorf("while getting terminal dimensions: %w", err)
	}
	if rows == 0 || cols == 0 || pxWidth == 0 || pxHeight == 0 {
		return nil, nil, fmt.Errorf("the terminal doesn't report its size in pixels")
	}
	cellWidth, cellHeight := pxWidth/cols, pxHeight/rows

	// the last row is for the help line
	maxWidth := (pxWidth - barWidth*(panels-1)) / panels
	orig := imaging.Fit(original, maxWidth, pxHeight-cellHeight, imaging.CatmullRom)
	b := orig.Bounds()
	edit := imaging.Resize(edited, b.Dx(), b.Dy(), imaging.CatmullRom)
	bar := image.NewRGBA(image.Rect(0, 0, barWidth, b.Dy()))
	draw.Draw(bar, bar.Bounds(), image.NewUniform(color.White), image.Point{}, draw.Src)

	var out bytes.Buffer
	for id, img := range map[int]image.Image{originalID: orig, editedID: edit, barID: bar} {
		if err := transmit(&out, id, img); err != nil {
			return nil, nil, err
		}
	}

	oldState, err := term.MakeRaw(fd)
	if err != nil {
		return nil, nil, err
	}
	os.Stdout.WriteString(enterSliderScreen)
	closeScreen = func() {
		fmt.Fprintf(os.Stdout, deleteImage+deleteImage+deleteImage+leaveSliderScreen, originalID, editedID, barID)
		term.Restore(fd, oldState)
	}

	imgRows := (b.Dy() + cellHeight - 1) / cellHeight
	fmt.Fprintf(&out, moveCursor+placeImage, 1, 1, originalID)
	fmt.Fprintf(&out, moveCursor+help, imgRows+1, 1)
	os.Stdout.Write(out.Bytes())

	return &slider{width: b.Dx(), cellWidth: cellWidth}, closeScreen, nil
}

// readInput passes everything typed to onInput until it returns true.
func readInput(onInput func(in []byte) bool) error {
	buf := make([]byte, 256)
	for {
		n, err := os.Stdin.Read(buf)
		if err != nil {
			return err
		}
		if onInput(buf[:n]) {
			return nil
		}
	}
}

type slider struct {
	width, cellWidth int
	barX             int
}

// handle updates barX from the input, it returns true when the user quits.
func (s *slider) handle(in []byte) bool {
	step := max(1, s.width/50)
	for len(in) > 0 {
		n := 1
		switch {
		case in[0] == 'q' || in[0] == 3: // 3 is ctrl+c, raw mode doesn't send SIGINT
			return true
		case in[0] == 'h':
			s.barX -= step
		case in[0] == 'l':
			s.barX += step
		case bytes.HasPrefix(in, []byte(keyLeft)):
			s.barX -= step
			n = len(keyLeft)
		case bytes.HasPrefix(in, []byte(keyRight)):
			s.barX += step
			n = len(keyRight)
		case bytes.HasPrefix(in, []byte(mouseEvent)):
			end := bytes.IndexAny(in, "Mm")
			if end < 0 {
				return false
			}
			var btn, col, row int
			fmt.Sscanf(string(in[len(mouseEvent):end]), "%d;%d;%d", &btn, &col, &row)
			if in[end] == 'M' && btn&^32 == 0 {
				s.barX = (col - 1) * s.cellWidth
			}
			n = end + 1
		}
		in = in[n:]
	}
	s.barX = min(max(s.barX, 0), s.width)
	return false
}

func (s *slider) draw() {
	var b bytes.Buffer
	fmt.Fprintf(&b, moveCursor, 1, 1)
	if s.barX > 0 {
		fmt.Fprintf(&b, placeCropped, editedID, s.barX)
	} else {
		// w=0 means the full width, so hide the edited image instead
		fmt.Fprintf(&b, hideImage, editedID)
	}
	x := min(max(s.barX-barWidth/2, 0), s.width-barWidth)
	fmt.Fprintf(&b, moveCursor+placeShifted, 1, x/s.cellWidth+1, barID, x%s.cellWidth)
	os.Stdout.Write(b.Bytes())
}

// transmit uploads img as a png in 4096 byte chunks, the protocol's max chunk size.
func transmit(w *bytes.Buffer, id int, img image.Image) error {
	var p bytes.Buffer
	if err := (&png.Encoder{CompressionLevel: png.BestSpeed}).Encode(&p, img); err != nil {
		return fmt.Errorf("encoding PNG: %w", err)
	}
	data := base64.StdEncoding.EncodeToString(p.Bytes())
	for first := true; len(data) > 0; first = false {
		chunk := data[:min(4096, len(data))]
		data = data[len(chunk):]
		more := 0
		if len(data) > 0 {
			more = 1
		}
		if first {
			fmt.Fprintf(w, transmitFirstChunk, id, more, chunk)
		} else {
			fmt.Fprintf(w, transmitNextChunk, more, chunk)
		}
	}
	return nil
}
