package gif

import (
	"fmt"
	"image"
	"image/draw"
	"image/gif"
	"io"

	imgbackend "github.com/Achno/gowall/internal/backends/image"
	types "github.com/Achno/gowall/internal/types"
)

// Decoder decodes every frame of a gif. Gif frames are often only the patch that changed,
// so each one is painted over the previous result (honoring its disposal method) to get full frames.
type Decoder struct {
	gif    *gif.GIF
	canvas *image.RGBA
}

func NewDecoder() *Decoder {
	return &Decoder{}
}

// Decode reads a gif and returns all of its frames fully composited
func (d *Decoder) Decode(r io.Reader) (*types.Animation, error) {
	g, err := gif.DecodeAll(r)
	if err != nil {
		return nil, fmt.Errorf("while decoding gif: %w", err)
	}
	if len(g.Image) == 0 {
		return nil, fmt.Errorf("gif has no frames")
	}

	d.gif = g
	d.canvas = image.NewRGBA(d.bounds())

	return &types.Animation{
		Frames:    d.composite(),
		Delays:    g.Delay,
		LoopCount: g.LoopCount,
	}, nil
}

// composite paints every frame over the canvas and snapshots the result
func (d *Decoder) composite() []image.Image {
	frames := make([]image.Image, len(d.gif.Image))

	for i, frame := range d.gif.Image {
		disposal := d.disposalAt(i)

		var previous *image.RGBA
		if disposal == gif.DisposalPrevious {
			previous = imgbackend.CloneRGBA(d.canvas)
		}

		d.paint(frame)
		frames[i] = imgbackend.CloneRGBA(d.canvas)
		d.dispose(frame, disposal, previous)
	}

	return frames
}

// bounds is the logical screen of the gif, falling back to the first frame
func (d *Decoder) bounds() image.Rectangle {
	bounds := image.Rect(0, 0, d.gif.Config.Width, d.gif.Config.Height)
	if bounds.Empty() {
		return d.gif.Image[0].Bounds()
	}
	return bounds
}

func (d *Decoder) disposalAt(i int) byte {
	if i < len(d.gif.Disposal) {
		return d.gif.Disposal[i]
	}
	return gif.DisposalNone
}

// paint draws the frame over the canvas. Gif colors are either opaque or fully transparent, so drawing over is
// copying the opaque ones and skipping the transparent ones, done directly instead of through draw.Draw's slow generic path.
func (d *Decoder) paint(frame *image.Paletted) {
	colors := make([][4]uint8, len(frame.Palette))
	opaque := make([]bool, len(frame.Palette))
	for i, c := range frame.Palette {
		r, g, b, a := c.RGBA()
		switch a {
		case 0xffff:
			colors[i], opaque[i] = [4]uint8{uint8(r >> 8), uint8(g >> 8), uint8(b >> 8), 0xff}, true
		case 0:
		default:
			draw.Draw(d.canvas, frame.Bounds(), frame, frame.Bounds().Min, draw.Over)
			return
		}
	}

	r := frame.Bounds().Intersect(d.canvas.Bounds())
	for y := r.Min.Y; y < r.Max.Y; y++ {
		src := frame.Pix[frame.PixOffset(r.Min.X, y):frame.PixOffset(r.Max.X, y)]
		dst := d.canvas.Pix[d.canvas.PixOffset(r.Min.X, y):]
		for x, idx := range src {
			if int(idx) >= len(colors) {
				// not a color of the palette, let draw.Draw handle it the way it always has
				draw.Draw(d.canvas, frame.Bounds(), frame, frame.Bounds().Min, draw.Over)
				return
			}
			if opaque[idx] {
				copy(dst[x*4:x*4+4], colors[idx][:])
			}
		}
	}
}

// dispose prepares the canvas for the next frame, as the frame's disposal method asks
func (d *Decoder) dispose(frame *image.Paletted, disposal byte, previous *image.RGBA) {
	switch disposal {
	case gif.DisposalBackground:
		draw.Draw(d.canvas, frame.Bounds(), image.Transparent, image.Point{}, draw.Src)
	case gif.DisposalPrevious:
		d.canvas = previous
	}
}
