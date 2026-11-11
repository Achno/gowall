package terminal

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
)

const (
	shadeID  = 4
	borderID = 5
)

// SelectCropRegion shows img and lets the user drag a rectangle over it with the mouse, enter confirms it.
// Outside the selection is dimmed by a see-through black image placed as 4 pieces around it,
// the border is a white image placed as 4 thin strips. It returns the selection in img's pixels.
func SelectCropRegion(img image.Image) (image.Rectangle, error) {
	c := &cropper{}
	s, err := selectOnImage(img, map[int]color.Color{shadeID: color.NRGBA{0, 0, 0, 160}, borderID: color.White}, c)
	if err != nil {
		return image.Rectangle{}, err
	}
	return s.toSource(c.selection()), nil
}

type cropper struct {
	start, end image.Point // selection corners in px on screen
}

func (c *cropper) selection() image.Rectangle {
	return image.Rectangle{c.start, c.end}.Canon()
}

func (c *cropper) ready() bool { return !c.selection().Empty() }

func (c *cropper) key(byte) {}

func (c *cropper) mouse(btn int, p image.Point, pressed bool) {
	switch {
	case pressed && btn == 0:
		c.start, c.end = p, p
	case pressed && btn == 32:
		c.end = p
	}
}

func (c *cropper) draw(s *selectScreen, b *bytes.Buffer) string {
	sel := c.selection()
	w, h := s.shown.Dx(), s.shown.Dy()

	var shades [4]image.Rectangle
	if !sel.Empty() {
		shades = [4]image.Rectangle{
			image.Rect(0, 0, w, sel.Min.Y),                 // above
			image.Rect(0, sel.Max.Y, w, h),                 // below
			image.Rect(0, sel.Min.Y, sel.Min.X, sel.Max.Y), // left
			image.Rect(sel.Max.X, sel.Min.Y, w, sel.Max.Y), // right
		}
	}
	for i, br := range borders(sel) {
		s.placeArea(b, shadeID, i+1, 1, shades[i])
		s.placeArea(b, borderID, i+1, 2, br)
	}

	help := "drag with the mouse to select, enter to crop, q to cancel"
	if src := s.toSource(sel); !src.Empty() {
		help += fmt.Sprintf("   %dx%d", src.Dx(), src.Dy())
	}
	return help
}
