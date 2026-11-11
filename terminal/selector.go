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

const borderWidth = 2

// selector is what a mouse selection (crop, regions) does differently, selectOnImage does the rest.
type selector interface {
	ready() bool                                  // whether enter confirms
	key(k byte)                                   // any key besides q, ctrl+c and enter
	mouse(btn int, p image.Point, pressed bool)   // p is in px on the shown image, btn 32 is a drag
	draw(s *selectScreen, b *bytes.Buffer) string // places the overlays, returns the help line
}

type selectScreen struct {
	*kittyScreen
	shown, source image.Rectangle
	helpRow       int
}

// selectOnImage shows img, uploads a see-through image the size of it for every overlay colour and
// passes the input to sel until enter confirms or q cancels.
func selectOnImage(img image.Image, overlays map[int]color.Color, sel selector) (*selectScreen, error) {
	ids := []int{originalID}
	for id := range overlays {
		ids = append(ids, id)
	}
	k, err := openKittyScreen(ids...)
	if err != nil {
		return nil, err
	}
	defer k.close()

	shown := imaging.Fit(img, k.pxWidth, k.pxHeight-k.cellHeight, imaging.CatmullRom)
	b := shown.Bounds()
	var out bytes.Buffer
	if err := transmit(&out, originalID, shown); err != nil {
		return nil, err
	}
	for id, c := range overlays {
		if err := transmit(&out, id, imaging.New(b.Dx(), b.Dy(), c)); err != nil {
			return nil, err
		}
	}
	fmt.Fprintf(&out, moveCursor+placeImage, 1, 1, originalID)
	os.Stdout.Write(out.Bytes())

	s := &selectScreen{kittyScreen: k, shown: b, source: img.Bounds(), helpRow: (b.Dy()+k.cellHeight-1)/k.cellHeight + 1}
	s.redraw(sel)

	var quit, confirmed bool
	err = readInput(func(in []byte) bool {
		for len(in) > 0 && !quit {
			n := 1
			switch {
			case in[0] == 'q' || in[0] == 3:
				quit = true
			case in[0] == '\r' || in[0] == '\n':
				quit, confirmed = sel.ready(), sel.ready()
			case bytes.HasPrefix(in, []byte(mouseEvent)):
				btn, x, y, pressed, used := s.parseMouse(in)
				if n = used; used == 0 { // cut off, drop it
					n = len(in)
					break
				}
				sel.mouse(btn, image.Pt(min(max(x, 0), b.Dx()), min(max(y, 0), b.Dy())), pressed)
			default:
				sel.key(in[0])
			}
			in = in[n:]
		}
		if !quit {
			s.redraw(sel)
		}
		return quit
	})
	if err == nil && !confirmed {
		err = errors.New("selection cancelled")
	}
	return s, err
}

func (s *selectScreen) redraw(sel selector) {
	var b bytes.Buffer
	help := sel.draw(s, &b)
	fmt.Fprintf(&b, moveCursor+clearLine+"%s", s.helpRow, 1, help)
	os.Stdout.Write(b.Bytes())
}

// toSource scales r up from the image on screen to the real one.
func (s *selectScreen) toSource(r image.Rectangle) image.Rectangle {
	sx := float64(s.source.Dx()) / float64(s.shown.Dx())
	sy := float64(s.source.Dy()) / float64(s.shown.Dy())
	src := image.Rect(int(float64(r.Min.X)*sx), int(float64(r.Min.Y)*sy), int(float64(r.Max.X)*sx), int(float64(r.Max.Y)*sy))
	return src.Add(s.source.Min).Intersect(s.source)
}

// borders returns the 4 strips along the inside of r, or empty ones when r is empty.
func borders(r image.Rectangle) [4]image.Rectangle {
	if r.Empty() {
		return [4]image.Rectangle{}
	}
	return [4]image.Rectangle{
		image.Rect(r.Min.X, r.Min.Y, r.Max.X, r.Min.Y+borderWidth).Intersect(r),
		image.Rect(r.Min.X, r.Max.Y-borderWidth, r.Max.X, r.Max.Y).Intersect(r),
		image.Rect(r.Min.X, r.Min.Y, r.Min.X+borderWidth, r.Max.Y).Intersect(r),
		image.Rect(r.Max.X-borderWidth, r.Min.Y, r.Max.X, r.Max.Y).Intersect(r),
	}
}
