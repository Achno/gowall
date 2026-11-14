// Package pixels has helpers for per pixel loops that read and write Pix directly, instead of going through At() and Set()
// which allocate for every pixel, trust me benchmarks don't lie, At() and Set() are very bad
package pixels

import (
	"image"
	"image/draw"
	"runtime"
	"sync"
)

// ToRGBA returns the image as an *image.RGBA whose bounds start at (0,0), so its Pix can be read directly.
// An *image.RGBA that already starts at (0,0) is returned as is (not copied), so callers must not modify it.
func ToRGBA(img image.Image) *image.RGBA {
	if rgba, ok := img.(*image.RGBA); ok && rgba.Rect.Min == (image.Point{}) {
		return rgba
	}
	b := img.Bounds()
	dst := image.NewRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))
	ParallelRows(b.Dy(), func(y0, y1 int) {
		draw.Draw(dst, image.Rect(0, y0, b.Dx(), y1), img, image.Pt(b.Min.X, b.Min.Y+y0), draw.Src)
	})
	return dst
}

// ParallelRows splits the rows [0,height) into one chunk per CPU and runs fn on every chunk concurrently
func ParallelRows(height int, fn func(y0, y1 int)) {
	workers := min(runtime.GOMAXPROCS(0), height)
	if workers <= 1 {
		fn(0, height)
		return
	}

	chunk := (height + workers - 1) / workers
	var wg sync.WaitGroup
	for y0 := 0; y0 < height; y0 += chunk {
		wg.Add(1)
		go func(y0, y1 int) {
			defer wg.Done()
			fn(y0, y1)
		}(y0, min(y0+chunk, height))
	}
	wg.Wait()
}

// CopyPixel copies the pixel at (sx,sy) of src to (dx,dy) of dst. Pixels outside of src are copied as transparent,
// the same as src.At() returns them.
func CopyPixel(dst *image.RGBA, dx, dy int, src *image.RGBA, sx, sy int) {
	d := dst.PixOffset(dx, dy)
	if !(image.Point{sx, sy}.In(src.Rect)) {
		clear(dst.Pix[d : d+4])
		return
	}
	s := src.PixOffset(sx, sy)
	copy(dst.Pix[d:d+4], src.Pix[s:s+4])
}

// Straight undoes the premultiplication of an *image.RGBA pixel: the color as it looks at full opacity.
// a must not be 0, a fully transparent pixel has no color.
func Straight(r, g, b, a uint8) (uint8, uint8, uint8) {
	if a == 255 {
		return r, g, b
	}
	un := func(c uint8) uint8 { return uint8(min(255, (int(c)*255+int(a)/2)/int(a))) }
	return un(r), un(g), un(b)
}

// Premultiply scales a straight color by a, as an *image.RGBA pixel stores it.
func Premultiply(r, g, b, a uint8) (uint8, uint8, uint8) {
	if a == 255 {
		return r, g, b
	}
	pre := func(c uint8) uint8 { return uint8((int(c)*int(a) + 127) / 255) }
	return pre(r), pre(g), pre(b)
}
