package terminal

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"image/color"
	"os"

	"github.com/disintegration/imaging"
)

const (
	shadeID     = 4
	borderID    = 5
	borderWidth = 2
)

var ErrCropCancelled = errors.New("crop cancelled")

// SelectCropRegion shows img and lets the user drag a rectangle over it with the mouse, enter confirms it.
// Outside the selection is dimmed by a see-through black image placed as 4 pieces around it,
// the border is a white image placed as 4 thin strips. It returns the selection in img's pixels.
func SelectCropRegion(img image.Image) (image.Rectangle, error) {
	k, err := openKittyScreen(originalID, shadeID, borderID)
	if err != nil {
		return image.Rectangle{}, err
	}
	defer k.close()

	shown := imaging.Fit(img, k.pxWidth, k.pxHeight-k.cellHeight, imaging.CatmullRom)
	b := shown.Bounds()
	shade := imaging.New(b.Dx(), b.Dy(), color.NRGBA{0, 0, 0, 160})
	border := imaging.New(b.Dx(), b.Dy(), color.White)

	var out bytes.Buffer
	for id, im := range map[int]image.Image{originalID: shown, shadeID: shade, borderID: border} {
		if err := transmit(&out, id, im); err != nil {
			return image.Rectangle{}, err
		}
	}
	fmt.Fprintf(&out, moveCursor+placeImage, 1, 1, originalID)
	os.Stdout.Write(out.Bytes())

	c := &cropper{
		kittyScreen: k,
		shown:       b,
		source:      img.Bounds(),
		helpRow:     (b.Dy()+k.cellHeight-1)/k.cellHeight + 1,
	}
	c.draw()

	var confirmed bool
	err = readInput(func(in []byte) bool {
		var quit bool
		quit, confirmed = c.handle(in)
		if !quit {
			c.draw()
		}
		return quit
	})
	if err != nil {
		return image.Rectangle{}, err
	}
	if !confirmed {
		return image.Rectangle{}, ErrCropCancelled
	}
	return c.sourceSelection(), nil
}

type cropper struct {
	*kittyScreen
	shown, source image.Rectangle
	start, end    image.Point // selection corners in px on screen
	helpRow       int
}

func (c *cropper) selection() image.Rectangle {
	return image.Rectangle{c.start, c.end}.Canon()
}

// sourceSelection scales the selection up from the image on screen to the real one.
func (c *cropper) sourceSelection() image.Rectangle {
	sel := c.selection()
	sx := float64(c.source.Dx()) / float64(c.shown.Dx())
	sy := float64(c.source.Dy()) / float64(c.shown.Dy())
	r := image.Rect(int(float64(sel.Min.X)*sx), int(float64(sel.Min.Y)*sy), int(float64(sel.Max.X)*sx), int(float64(sel.Max.Y)*sy))
	return r.Add(c.source.Min).Intersect(c.source)
}

// handle updates the selection from the input. quit is true when the user is done, confirmed when it was enter.
func (c *cropper) handle(in []byte) (quit, confirmed bool) {
	for len(in) > 0 {
		n := 1
		switch {
		case in[0] == 'q' || in[0] == 3:
			return true, false
		case in[0] == '\r' || in[0] == '\n':
			if !c.selection().Empty() {
				return true, true
			}
		case bytes.HasPrefix(in, []byte(mouseEvent)):
			btn, x, y, pressed, used := c.parseMouse(in)
			p := image.Pt(min(max(x, 0), c.shown.Dx()), min(max(y, 0), c.shown.Dy()))
			switch {
			case pressed && btn == 0:
				c.start, c.end = p, p
			case pressed && btn == 32:
				c.end = p
			}
			n = used
			if used == 0 {
				n = len(in)
			}
		}
		in = in[n:]
	}
	return false, false
}

func (c *cropper) draw() {
	sel := c.selection()
	w, h := c.shown.Dx(), c.shown.Dy()

	var shades, borders [4]image.Rectangle
	if !sel.Empty() {
		shades = [4]image.Rectangle{
			image.Rect(0, 0, w, sel.Min.Y),                 // above
			image.Rect(0, sel.Max.Y, w, h),                 // below
			image.Rect(0, sel.Min.Y, sel.Min.X, sel.Max.Y), // left
			image.Rect(sel.Max.X, sel.Min.Y, w, sel.Max.Y), // right
		}
		borders = [4]image.Rectangle{
			image.Rect(sel.Min.X, sel.Min.Y, sel.Max.X, sel.Min.Y+borderWidth).Intersect(sel),
			image.Rect(sel.Min.X, sel.Max.Y-borderWidth, sel.Max.X, sel.Max.Y).Intersect(sel),
			image.Rect(sel.Min.X, sel.Min.Y, sel.Min.X+borderWidth, sel.Max.Y).Intersect(sel),
			image.Rect(sel.Max.X-borderWidth, sel.Min.Y, sel.Max.X, sel.Max.Y).Intersect(sel),
		}
	}

	var b bytes.Buffer
	for i := range 4 {
		c.placeArea(&b, shadeID, i+1, 1, shades[i])
		c.placeArea(&b, borderID, i+1, 2, borders[i])
	}

	help := "drag with the mouse to select, enter to crop, q to cancel"
	if src := c.sourceSelection(); !src.Empty() {
		help += fmt.Sprintf("   %dx%d", src.Dx(), src.Dy())
	}
	fmt.Fprintf(&b, moveCursor+clearLine+"%s", c.helpRow, 1, help)
	os.Stdout.Write(b.Bytes())
}
