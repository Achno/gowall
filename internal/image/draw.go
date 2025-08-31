package image

import (
	"image"
	"image/color"
	"image/draw"

	"github.com/Achno/gowall/utils"
)

type BorderProcessor struct {
	Color           color.RGBA
	BorderThickness int
}

func (b *BorderProcessor) Process(img image.Image, theme string, format string) (image.Image, error) {

	newImg := drawBorder(img, b.BorderThickness, b.Color)

	return newImg, nil

}

func drawBorder(img image.Image, borderThickness int, borderColor color.Color) image.Image {
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// draw on new image
	newImg := image.NewRGBA(bounds)
	draw.Draw(newImg, bounds, img, image.Point{0, 0}, draw.Src)

	// top and bottom borders
	for x := 0; x < width; x++ {
		for t := 0; t < borderThickness; t++ {
			newImg.Set(x, t, borderColor)
			newImg.Set(x, height-borderThickness+t, borderColor)
		}
	}

	// left and right borders
	for y := 0; y < height; y++ {
		for t := 0; t < borderThickness; t++ {
			newImg.Set(t, y, borderColor)
			newImg.Set(width-borderThickness+t, y, borderColor)
		}
	}

	return newImg
}

type GridProcessor struct {
	options GridOptions
}

type GridOptions struct {
	GridSize      int
	GridColor     color.RGBA
	GridThickness int
	MaskOnly      bool
}

type GridOption func(*GridOptions)

func WithGridSize(gridsize int) GridOption {
	return func(g *GridOptions) {
		g.GridSize = gridsize
	}
}
func WithGridColor(gridColor string) GridOption {
	return func(g *GridOptions) {
		c, err := HexToRGBA(gridColor)
		utils.HandleError(err)
		g.GridColor = c
	}
}
func WithGridThickness(gridThickness int) GridOption {
	return func(g *GridOptions) {
		g.GridThickness = gridThickness
	}
}
func WithMaskonly(maskOnly bool) GridOption {
	return func(g *GridOptions) {
		g.MaskOnly = maskOnly
	}
}

// Available options : WithGridSize,WithGridColor,WithGridThickness,WithMaskonly
func (g *GridProcessor) SetGridOptions(options ...GridOption) {
	opts := GridOptions{
		GridSize:      80,
		GridColor:     color.RGBA{R: 93, G: 63, B: 211, A: 255},
		GridThickness: 1,
		MaskOnly:      false,
	}

	for _, option := range options {
		option(&opts)
	}

	g.options = opts
}

func (g *GridProcessor) Process(img image.Image, theme string, format string) (image.Image, error) {

	newImg, err := applyGridToImage(img, &g.options)
	if err != nil {
		return nil, err
	}
	return newImg, nil

}

func applyGridToImage(img image.Image, options *GridOptions) (image.Image, error) {
	bounds := img.Bounds()
	newImg := image.NewRGBA(bounds)

	draw.Draw(newImg, bounds, img, bounds.Min, draw.Src)

	// optionally use the input image as a mask, and apply the grid only to the transparent areas.
	if options.MaskOnly {
		gridImg := image.NewRGBA(bounds)
		drawGridOnImage(gridImg, options)

		for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
			for x := bounds.Min.X; x < bounds.Max.X; x++ {
				_, _, _, srcAlpha := img.At(x, y).RGBA()

				// (low alpha)
				if srcAlpha < 0x8000 {
					gridColor := gridImg.At(x, y)
					newImg.Set(x, y, gridColor)
				}
			}
		}
	} else {
		drawGridOnImage(newImg, options)
	}

	return newImg, nil
}

// drawGrid draws a grid on the image with the given thickness,color and grid size
func drawGridOnImage(img *image.RGBA, c *GridOptions) {
	bounds := img.Bounds()

	for x := bounds.Min.X; x < bounds.Max.X; x += c.GridSize {
		for thickness := 0; thickness < c.GridThickness; thickness++ {
			if x+thickness >= bounds.Max.X {
				break
			}

			for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
				img.Set(x+thickness, y, c.GridColor)
			}
		}
	}

	for y := bounds.Min.Y; y < bounds.Max.Y; y += c.GridSize {
		for thickness := 0; thickness < c.GridThickness; thickness++ {
			if y+thickness >= bounds.Max.Y {
				break
			}

			for x := bounds.Min.X; x < bounds.Max.X; x++ {
				img.Set(x, y+thickness, c.GridColor)
			}
		}
	}
}
