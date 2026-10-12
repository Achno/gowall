package terminal

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"os"

	"github.com/disintegration/imaging"
)

// Image ids for the kitty protocol. Each image is uploaded once, after that
// moving the slider only re-sends tiny placement commands.
const (
	originalID = 1
	editedID   = 2
	barID      = 3
	barWidth   = 3 // px
)

// RenderKittySlider shows an interactive before/after slider, moved with the arrow keys or by dragging with the mouse.
// The original is placed full size and the edited image is placed on top, cropped to the bar position.
func RenderKittySlider(original, edited image.Image) error {
	s, err := openCompareScreen(original, edited, 1, "←/→ h/l or drag with the mouse to move the slider, q to quit")
	if err != nil {
		return err
	}
	defer s.close()

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
	s, err := openCompareScreen(original, edited, 2, "q to quit")
	if err != nil {
		return err
	}
	defer s.close()

	var b bytes.Buffer
	fmt.Fprintf(&b, moveCursor+placeShifted, 1, s.width/s.cellWidth+1, barID, s.width%s.cellWidth)
	x := s.width + barWidth
	fmt.Fprintf(&b, moveCursor+placeShifted, 1, x/s.cellWidth+1, editedID, x%s.cellWidth)
	os.Stdout.Write(b.Bytes())

	return readInput(func(in []byte) bool {
		return bytes.ContainsAny(in, "q\x03")
	})
}

// openCompareScreen fits the images so that `panels` of them fit next to each other, uploads them
// and places the original in the top left. s.close undoes all of it.
func openCompareScreen(original, edited image.Image, panels int, help string) (s *slider, err error) {
	k, err := openKittyScreen(originalID, editedID, barID)
	if err != nil {
		return nil, err
	}

	// the last row is for the help line
	maxWidth := (k.pxWidth - barWidth*(panels-1)) / panels
	orig := imaging.Fit(original, maxWidth, k.pxHeight-k.cellHeight, imaging.CatmullRom)
	b := orig.Bounds()
	edit := imaging.Resize(edited, b.Dx(), b.Dy(), imaging.CatmullRom)
	bar := image.NewRGBA(image.Rect(0, 0, barWidth, b.Dy()))
	draw.Draw(bar, bar.Bounds(), image.NewUniform(color.White), image.Point{}, draw.Src)

	var out bytes.Buffer
	for id, img := range map[int]image.Image{originalID: orig, editedID: edit, barID: bar} {
		if err := transmit(&out, id, img); err != nil {
			k.close()
			return nil, err
		}
	}

	imgRows := (b.Dy() + k.cellHeight - 1) / k.cellHeight
	fmt.Fprintf(&out, moveCursor+placeImage, 1, 1, originalID)
	fmt.Fprintf(&out, moveCursor+"%s", imgRows+1, 1, help)
	os.Stdout.Write(out.Bytes())

	return &slider{kittyScreen: k, width: b.Dx()}, nil
}

type slider struct {
	*kittyScreen
	width int
	barX  int
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
			btn, x, _, pressed, used := s.parseMouse(in)
			if pressed && btn&^32 == 0 { // left press, or left drag (motion bit 32)
				s.barX = x
			}
			n = used
			if used == 0 { // incomplete report, drop it
				n = len(in)
			}
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
