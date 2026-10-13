package types

import (
	"image"
	"image/color"
	"io"
)

type EncoderFunc func(w io.Writer, img image.Image) error

type ImageMetadata struct {
	EncoderFunction EncoderFunc
}

// Animation holds every fully composited frame of an animated image (e.g. gif).
// It implements image.Image as its first frame, so processors that don't care about animation can use it as is.
type Animation struct {
	Frames    []image.Image
	Delays    []int // per frame, in 100ths of a second
	LoopCount int   // 0 loops forever, -1 plays once, n plays n+1 times
}

func (a *Animation) ColorModel() color.Model { return a.Frames[0].ColorModel() }
func (a *Animation) Bounds() image.Rectangle { return a.Frames[0].Bounds() }
func (a *Animation) At(x, y int) color.Color { return a.Frames[0].At(x, y) }

// ToAnimation returns the image as an Animation, a still image becomes a single frame Animation
func ToAnimation(img image.Image) *Animation {
	if anim, ok := img.(*Animation); ok {
		return anim
	}
	return &Animation{Frames: []image.Image{img}, Delays: []int{0}}
}
