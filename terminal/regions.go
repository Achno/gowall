package terminal

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
)

const fillID = 6

// SelectRegions lets the user drag any number of rectangles over img, u undoes, enter confirms.
// Each one is a see-through red image placed over it with a white border.
func SelectRegions(img image.Image) ([]image.Rectangle, error) {
	r := &regionSelector{}
	s, err := selectOnImage(img, map[int]color.Color{fillID: color.NRGBA{255, 40, 40, 90}, borderID: color.White}, r)
	if err != nil {
		return nil, err
	}
	var rects []image.Rectangle
	for _, rect := range r.rects {
		if src := s.toSource(rect); !src.Empty() {
			rects = append(rects, src)
		}
	}
	return rects, nil
}

type regionSelector struct {
	rects    []image.Rectangle // in px on screen, the last one follows the mouse while dragging
	dragging bool
	drawn    int // how many the last draw placed, so removed ones get hidden
}

func (r *regionSelector) ready() bool { return len(r.rects) > 0 && !r.dragging }

func (r *regionSelector) key(k byte) {
	if (k == 'u' || k == 127) && !r.dragging && len(r.rects) > 0 {
		r.rects = r.rects[:len(r.rects)-1]
	}
}

func (r *regionSelector) mouse(btn int, p image.Point, pressed bool) {
	switch last := len(r.rects) - 1; {
	case pressed && btn == 0:
		r.rects = append(r.rects, image.Rectangle{p, p})
		r.dragging = true
	case pressed && btn == 32 && r.dragging:
		r.rects[last].Max = p
	case !pressed && r.dragging:
		r.dragging = false
		if r.rects[last] = r.rects[last].Canon(); r.rects[last].Empty() { // a click without a drag
			r.rects = r.rects[:last]
		}
	}
}

func (r *regionSelector) draw(s *selectScreen, b *bytes.Buffer) string {
	for i := range max(r.drawn, len(r.rects)) {
		var rect image.Rectangle
		if i < len(r.rects) {
			rect = r.rects[i].Canon()
		}
		s.placeArea(b, fillID, i+1, 1, rect)
		for j, br := range borders(rect) {
			s.placeArea(b, borderID, 4*i+j+1, 2, br)
		}
	}
	r.drawn = len(r.rects)

	help := "drag with the mouse to mark areas to erase, u to undo, enter to erase them, q to cancel"
	if len(r.rects) > 0 {
		help += fmt.Sprintf("   %d marked", len(r.rects))
	}
	return help
}
