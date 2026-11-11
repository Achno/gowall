package image

import (
	"fmt"
	"image"
	"strconv"
	"strings"
)

// ParseRects parses every "x,y,width,height" in pixels into a rectangle.
func ParseRects(rects ...string) ([]image.Rectangle, error) {
	var out []image.Rectangle
	for _, s := range rects {
		parts := strings.Split(s, ",")
		if len(parts) != 4 {
			return nil, fmt.Errorf("invalid rect %q, use x,y,width,height (e.g. 100,50,800,600)", s)
		}
		var n [4]int
		for i, p := range parts {
			v, err := strconv.Atoi(strings.TrimSpace(p))
			if err != nil || v < 0 {
				return nil, fmt.Errorf("invalid rect %q, %q is not a positive number", s, p)
			}
			n[i] = v
		}
		if n[2] == 0 || n[3] == 0 {
			return nil, fmt.Errorf("invalid rect %q, width and height can't be 0", s)
		}
		out = append(out, image.Rect(n[0], n[1], n[0]+n[2], n[1]+n[3]))
	}
	return out, nil
}
