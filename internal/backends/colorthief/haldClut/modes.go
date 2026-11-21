package haldclut

import (
	"image"
	"image/color"

	cpkg "github.com/Achno/gowall/internal/backends/color"
	"github.com/Achno/gowall/internal/backends/pixels"
	"github.com/lucasb-eyer/go-colorful"
)

// KeepShading returns a new CLUT where every cell keeps the color clut gives it (CIELAB a*, b*) but takes back the
// lightness (L*) of the color that goes into that cell. Converting through it changes only the colors of an image,
// every light and dark area stays as light and as dark. clut is not modified.
func KeepShading(clut *image.RGBA, level int) *image.RGBA {
	return mapCells(level, func(in colorful.Color, cell int) color.RGBA {
		l, _, _ := in.Lab()
		themed, _ := colorful.MakeColor(color.RGBA{clut.Pix[cell], clut.Pix[cell+1], clut.Pix[cell+2], 255})
		_, a, b := themed.Lab()
		return cpkg.LabInGamut(l, a, b)
	})
}

// MonoCLUT returns a CLUT that gives every color seed's hue and chroma (its CIELAB a*, b*) at the color's own
// lightness (L*): one color in all its shades from black to white, every light and dark area stays as light and as dark.
func MonoCLUT(seed color.RGBA, level int) *image.RGBA {
	seedColor, _ := colorful.MakeColor(seed)
	_, a, b := seedColor.Lab()
	return mapCells(level, func(in colorful.Color, _ int) color.RGBA {
		l, _, _ := in.Lab()
		return cpkg.LabInGamut(l, a, b)
	})
}

// mapCells returns a CLUT of level where every cell is fn(the color that goes into that cell, the cell's offset in Pix)
func mapCells(level int, fn func(in colorful.Color, cell int) color.RGBA) *image.RGBA {
	identity, _ := GenerateIdentityCLUT(level)
	out := image.NewRGBA(identity.Bounds())
	pixels.ParallelRows(identity.Bounds().Dy(), func(y0, y1 int) {
		for cell := y0 * identity.Stride; cell < y1*identity.Stride; cell += 4 {
			in, _ := colorful.MakeColor(color.RGBA{identity.Pix[cell], identity.Pix[cell+1], identity.Pix[cell+2], 255})
			c := fn(in, cell)
			out.Pix[cell], out.Pix[cell+1], out.Pix[cell+2], out.Pix[cell+3] = c.R, c.G, c.B, 255
		}
	})
	return out
}
