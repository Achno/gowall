// Package inpaint erases rectangles from an image and fills them with what the surroundings suggest.
package inpaint

import (
	"fmt"
	"image"
	"math"

	"github.com/Achno/gowall/internal/backends/pixels"
	"github.com/disintegration/imaging"
)

// InpaintStrategy fills the pixels of img where hole is true, looking only at the Context around it.
type InpaintStrategy interface {
	Context(hole, bounds image.Rectangle) image.Rectangle
	Fill(img *image.NRGBA, hole []bool) (*image.NRGBA, error)
}

const feather = 4

// Erase fills the rectangles one by one, the ones not filled yet are never used as a source.
func Erase(src image.Image, rects []image.Rectangle, s InpaintStrategy) (*image.NRGBA, error) {
	img := imaging.Clone(src)
	b := img.Bounds()

	var pending []image.Rectangle
	for _, r := range rects {
		r = r.Sub(src.Bounds().Min).Intersect(b)
		if !r.Empty() {
			pending = append(pending, r)
		}
	}
	if len(pending) == 0 {
		return nil, fmt.Errorf("none of the areas to erase are inside the image %v", src.Bounds())
	}

	for i, r := range pending {
		grown := r.Inset(-feather).Intersect(b)
		ctx := s.Context(grown, b)

		crop := imaging.Crop(img, ctx)
		w, h := ctx.Dx(), ctx.Dy()
		hole := make([]bool, w*h)
		for _, p := range pending[i:] {
			m := p.Inset(-feather).Intersect(ctx).Sub(ctx.Min)
			for y := m.Min.Y; y < m.Max.Y; y++ {
				for x := m.Min.X; x < m.Max.X; x++ {
					hole[y*w+x] = true
				}
			}
		}

		filled, err := s.Fill(crop, hole)
		if err != nil {
			return nil, fmt.Errorf("while filling %v: %w", r, err)
		}
		blend(img, filled, r, grown, ctx.Min)
	}
	return img, nil
}

// blend fades the fill into img over the feather band around r.
func blend(img, filled *image.NRGBA, r, grown image.Rectangle, offset image.Point) {
	pixels.ParallelRows(grown.Dy(), func(y0, y1 int) {
		for y := grown.Min.Y + y0; y < grown.Min.Y+y1; y++ {
			for x := grown.Min.X; x < grown.Max.X; x++ {
				dx := max(r.Min.X-x, 0, x-(r.Max.X-1))
				dy := max(r.Min.Y-y, 0, y-(r.Max.Y-1))
				a := 1 - math.Hypot(float64(dx), float64(dy))/(feather+1)
				if a <= 0 {
					continue
				}
				i := img.PixOffset(x, y)
				j := filled.PixOffset(x-offset.X, y-offset.Y)
				for c := range 3 {
					v := a*float64(filled.Pix[j+c]) + (1-a)*float64(img.Pix[i+c])
					img.Pix[i+c] = uint8(math.Round(v))
				}
			}
		}
	})
}
