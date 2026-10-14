package image

import (
	"fmt"
	"image"
	"image/color"
	"image/gif"
	"io"
	"math"

	cpkg "github.com/Achno/gowall/internal/backends/color"
	"github.com/Achno/gowall/internal/backends/colorthief/mediancut"
	"github.com/disintegration/imaging"
	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/gomono"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"
)

const (
	// asciiRamp goes from no ink to the most ink
	asciiRamp = " .:-=+*oO0@"
	// inkLevels is how many opacities characters are drawn with, so faint areas also look fainter
	inkLevels = 4
	// minOutputWidth enlarges small images (e.g. sprites) so their characters are readable
	minOutputWidth = 720
	// minColumns is the columns at detail 0, the chunkiest look
	minColumns = 30
	// minCellSize is the size in px of a character at detail 100, the densest look
	minCellSize = 3
	// maxColors is how many colors color mode reduces the image to, every one gets its own shades in the 256 color gif palette
	maxColors = 32
	// maxShades is how many steps from the background to a character color the palette has
	maxShades = 32
	// defaultDelay is used for frames without a delay (in 100ths of a second), browsers do the same
	defaultDelay = 10
)

var (
	AsciiDark  = color.RGBA{R: 20, G: 20, B: 20, A: 255}
	AsciiLight = color.RGBA{R: 244, G: 244, B: 242, A: 255}
)

// AsciiMode decides the colors of the characters and the background
type AsciiMode string

const (
	AsciiColor AsciiMode = "color" // characters keep the colors of the image, on a dark background
	AsciiWhite AsciiMode = "white" // white characters on a dark background
	AsciiBlack AsciiMode = "black" // black characters on a light background, the inverse of white
)

func AsciiModes() []string {
	return []string{string(AsciiColor), string(AsciiWhite), string(AsciiBlack)}
}

type AsciiOptions struct {
	Detail     int // 0-100, how many characters fit in the output width: 0 is chunky, 100 is the densest
	Mode       AsciiMode
	Background color.RGBA
}

func DefaultAsciiOptions() AsciiOptions {
	return AsciiOptions{
		Detail:     50,
		Mode:       AsciiColor,
		Background: AsciiDark,
	}
}

// AsciiCell is one character of an AsciiFrame
type AsciiCell struct {
	Char  int     // index into the ramp, 0 is blank
	Ink   float64 // 0-1, how much ink the character stands for
	Color int     // index into the converter's colors
}

// AsciiFrame is a grid of characters, rows of columns
type AsciiFrame [][]AsciiCell

// AsciiConverter converts images into AsciiFrames and renders them back into images or gifs.
// The output keeps the size of the input (at least minOutputWidth wide), the detail only decides how many characters fill it.
type AsciiConverter struct {
	AsciiOptions

	ramp          []rune
	colors        []color.RGBA // the character colors, a single one unless the mode is color
	palette       color.Palette
	width, height int // of the output in px
	cols, rows    int
	cell          int                       // size in px of the square every character is drawn in
	atlas         [inkLevels][]*image.Alpha // a pre-rendered tile per ink level and character
}

func NewAsciiConverter(opts AsciiOptions) *AsciiConverter {
	return &AsciiConverter{
		AsciiOptions: opts,
		ramp:         []rune(asciiRamp),
	}
}

// Convert turns the frames (all of the same size) into characters.
func (c *AsciiConverter) Convert(imgs []image.Image) ([]AsciiFrame, error) {
	if len(imgs) == 0 {
		return nil, fmt.Errorf("there are no frames to convert")
	}

	c.layout(imgs[0].Bounds())
	if err := c.buildAtlas(); err != nil {
		return nil, err
	}

	smalls := make([]*image.NRGBA, len(imgs))
	for i, img := range imgs {
		smalls[i] = imaging.Resize(img, c.cols, c.rows, imaging.Box)
	}

	c.pickColors(smalls)
	c.buildPalette()

	lo, hi := cpkg.LuminanceRange(smalls)
	solid := HasTransparency(smalls[0])

	frames := make([]AsciiFrame, len(smalls))
	for i, small := range smalls {
		frames[i] = c.toFrame(small, lo, hi, solid)
	}
	return frames, nil
}

// Render draws a frame, every pixel is a shade between the background and a character color so it is paletted without any loss.
// The cells are spread over the exact output size, so some are 1px further apart than others.
func (c *AsciiConverter) Render(frame AsciiFrame) *image.Paletted {
	out := image.NewPaletted(image.Rect(0, 0, c.width, c.height), c.palette)

	for y, row := range frame {
		for x, cell := range row {
			if cell.Char == 0 {
				continue
			}
			tile := c.atlas[c.levelFor(cell.Ink)][cell.Char]
			c.drawTile(out, tile, cell.Color, x*c.width/c.cols, y*c.height/c.rows)
		}
	}

	return out
}

// EncodeGif writes the frames as an animated gif, delays are in 100ths of a second
func (c *AsciiConverter) EncodeGif(w io.Writer, frames []AsciiFrame, delays []int, loopCount int) error {
	g := &gif.GIF{LoopCount: loopCount}

	for i, frame := range frames {
		g.Image = append(g.Image, c.Render(frame))
		g.Delay = append(g.Delay, c.delayAt(delays, i))
	}

	return gif.EncodeAll(w, g)
}

// layout sets the output size and how many characters fill it, the detail goes from minColumns to the columns of minCellSize
func (c *AsciiConverter) layout(bounds image.Rectangle) {
	scale := math.Max(1, float64(minOutputWidth)/float64(bounds.Dx()))
	c.width = int(math.Round(float64(bounds.Dx()) * scale))
	c.height = max(1, int(math.Round(float64(bounds.Dy())*scale)))

	maxCols := c.width / minCellSize
	c.cols = minColumns + (maxCols-minColumns)*cpkg.Clamp(c.Detail, 0, 100)/100
	c.rows = max(1, int(math.Round(float64(c.cols)*float64(c.height)/float64(c.width))))
	c.cell = max(minCellSize, c.width/c.cols)
}

// buildAtlas renders every character at every ink level once, rendering a frame then only copies tiles.
// The glyphs are a bit bigger than the cell and centered in it, the font leaves room around them anyway.
func (c *AsciiConverter) buildAtlas() error {
	parsed, err := opentype.Parse(gomono.TTF)
	if err != nil {
		return fmt.Errorf("while parsing the font: %w", err)
	}

	face, err := opentype.NewFace(parsed, &opentype.FaceOptions{
		Size:    float64(c.cell) * 1.15,
		DPI:     72,
		Hinting: font.HintingNone,
	})
	if err != nil {
		return fmt.Errorf("while creating the font face: %w", err)
	}
	defer face.Close()

	advance, _ := face.GlyphAdvance('0')
	size := fixed.I(c.cell)
	dot := fixed.Point26_6{X: (size - advance) / 2, Y: (size + face.Metrics().CapHeight) / 2}

	for level := range inkLevels {
		c.atlas[level] = make([]*image.Alpha, len(c.ramp))
		opacity := 0.35 + 0.65*float64(level)/float64(inkLevels-1)
		ink := image.NewUniform(color.Alpha{A: uint8(255 * opacity)})

		for i, r := range c.ramp {
			tile := image.NewAlpha(image.Rect(0, 0, c.cell, c.cell))
			drawer := font.Drawer{Dst: tile, Src: ink, Face: face, Dot: dot}
			drawer.DrawString(string(r))
			c.thicken(tile)
			c.atlas[level][i] = tile
		}
	}

	return nil
}

// thicken strengthens the anti aliased edges of a glyph, small glyphs are mostly edges and would look faint otherwise
func (c *AsciiConverter) thicken(tile *image.Alpha) {
	for i, coverage := range tile.Pix {
		tile.Pix[i] = uint8(255 * math.Sqrt(float64(coverage)/255))
	}
}

// pickColors sets the character colors: the one of the mode, or for color mode the image reduced to maxColors colors.
// The colors are picked from all frames at once so they are shared, a fully transparent image falls back to the light color.
func (c *AsciiConverter) pickColors(smalls []*image.NRGBA) {
	switch c.Mode {
	case AsciiBlack:
		c.colors = []color.RGBA{AsciiDark}
		return
	case AsciiWhite:
		c.colors = []color.RGBA{AsciiLight}
		return
	}

	c.colors = []color.RGBA{AsciiLight}
	if !HasOpaque(smalls[0]) {
		return
	}

	all := imaging.New(c.cols, c.rows*len(smalls), color.Transparent)
	for i, small := range smalls {
		copy(all.Pix[i*len(small.Pix):], small.Pix)
	}

	colors, _ := mediancut.GetPalette(all, maxColors)
	if len(colors) == 0 {
		return
	}
	c.colors = make([]color.RGBA, len(colors))
	for i, clr := range colors {
		c.colors[i] = color.RGBAModel.Convert(clr).(color.RGBA)
	}
}

// buildPalette makes the shades from the background to every character color, a pixel's index is color*shades + shade
func (c *AsciiConverter) buildPalette() {
	shades := c.shades()
	c.palette = make(color.Palette, 0, len(c.colors)*shades)
	for _, clr := range c.colors {
		for shade := range shades {
			c.palette = append(c.palette, cpkg.BlendColor(c.Background, clr, float64(shade)/float64(shades-1)))
		}
	}
}

// shades is how many steps every color gets so that all of them fit the 256 colors of a gif
func (c *AsciiConverter) shades() int {
	return min(maxShades, 256/len(c.colors))
}

// toFrame picks a character and a color for every pixel of an already downscaled image.
// solid keeps every visible pixel of a subject on a transparent background at least a ".", so its shape never gets holes.
func (c *AsciiConverter) toFrame(small *image.NRGBA, lo, hi float64, solid bool) AsciiFrame {
	frame := make(AsciiFrame, c.rows)

	for y := range frame {
		frame[y] = make([]AsciiCell, c.cols)
		for x := range frame[y] {
			i := small.PixOffset(x, y)
			px := small.Pix[i : i+4]

			ink := c.inkFor(px, lo, hi)
			char := cpkg.Clamp(int(ink*float64(len(c.ramp))), 0, len(c.ramp)-1)
			if solid && char == 0 && px[3] >= 128 {
				char = 1
			}

			frame[y][x] = AsciiCell{Char: char, Ink: ink, Color: c.colorFor(px)}
		}
	}

	return frame
}

// inkFor maps a pixel to how much ink (0-1) its character gets, stretched to the luminance range.
// Light characters on a dark background give bright pixels the most ink, dark ones on a light background the dark pixels,
// and transparent pixels get no ink.
func (c *AsciiConverter) inkFor(px []uint8, lo, hi float64) float64 {
	lum := cpkg.ClampFloat((cpkg.Luminance(px[0], px[1], px[2])-lo)/math.Max(0.05, hi-lo), 0, 1)

	ink := lum
	if c.Mode == AsciiBlack {
		ink = 1 - lum
	}

	return ink * float64(px[3]) / 255
}

// colorFor returns the index of the character color closest to the pixel
func (c *AsciiConverter) colorFor(px []uint8) int {
	best, bestDist := 0, math.Inf(1)
	for i, clr := range c.colors {
		dist := cpkg.ColorDistance(uint32(px[0]), uint32(px[1]), uint32(px[2]), uint32(clr.R), uint32(clr.G), uint32(clr.B))
		if dist < bestDist {
			best, bestDist = i, dist
		}
	}
	return best
}

// levelFor picks the opacity level a character is drawn with, in color mode the color already shows how bright it is
func (c *AsciiConverter) levelFor(ink float64) int {
	if c.Mode == AsciiColor {
		return inkLevels - 1
	}
	return min(inkLevels-1, int(ink*inkLevels*1.6))
}

// drawTile copies a pre-rendered character into the output as shades of its color, untouched pixels stay index 0 (the background)
func (c *AsciiConverter) drawTile(out *image.Paletted, tile *image.Alpha, colorIndex, px, py int) {
	shades := c.shades()
	size := min(c.cell, c.width-px)

	for row := 0; row < c.cell && py+row < c.height; row++ {
		dst := out.PixOffset(px, py+row)
		src := tile.PixOffset(0, row)
		for i, coverage := range tile.Pix[src : src+size] {
			if coverage > 0 {
				out.Pix[dst+i] = uint8(colorIndex*shades + (int(coverage)*(shades-1)+127)/255)
			}
		}
	}
}

// delayAt returns the delay of frame i in 100ths of a second
func (c *AsciiConverter) delayAt(delays []int, i int) int {
	if i < len(delays) && delays[i] > 1 {
		return delays[i]
	}
	return defaultDelay
}
