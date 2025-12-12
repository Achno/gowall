package image

import (
	"image"
	"image/color"
	"image/draw"
	"math"

	types "github.com/Achno/gowall/internal/types"
	"github.com/Achno/gowall/utils"
)

type BorderProcessor struct {
	Color           color.RGBA
	BorderThickness int
	CornerRadius    float64 // 0 means no rounding
}

func (b *BorderProcessor) Process(img image.Image, theme string, format string) (image.Image, types.ImageMetadata, error) {
	newImg := drawBorder(img, b.BorderThickness, b.Color, b.CornerRadius)
	return newImg, types.ImageMetadata{}, nil
}

func drawBorder(img image.Image, borderThickness int, borderColor color.Color, cornerRadius float64) image.Image {
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	newImg := image.NewRGBA(bounds)
	draw.Draw(newImg, bounds, img, image.Point{0, 0}, draw.Src)

	// Top and bottom borders
	for x := range width {
		for t := range borderThickness {
			newImg.Set(x, t, borderColor)
			newImg.Set(x, height-borderThickness+t, borderColor)
		}
	}

	// Left and right borders
	for y := range height {
		for t := range borderThickness {
			newImg.Set(t, y, borderColor)
			newImg.Set(width-borderThickness+t, y, borderColor)
		}
	}

	// If corner radius is specified, apply rounding to corners
	if cornerRadius > 0 {
		applyCornerRounding(newImg, cornerRadius, borderThickness, borderColor)
	}

	return newImg
}

func applyCornerRounding(img *image.RGBA, radius float64, borderThickness int, borderColor color.Color) {
	bounds := img.Bounds()
	width := float64(bounds.Dx())
	height := float64(bounds.Dy())

	for y := 0; y < bounds.Dy(); y++ {
		for x := 0; x < bounds.Dx(); x++ {
			fx, fy := float64(x), float64(y)

			var distanceFromCenter float64
			isInCorner := false

			// Top-left corner
			if fx < radius && fy < radius {
				dx := fx - radius
				dy := fy - radius
				distanceFromCenter = math.Sqrt(dx*dx + dy*dy)
				isInCorner = true
			} else if fx > width-radius && fy < radius {
				// Top-right corner
				dx := fx - (width - radius)
				dy := fy - radius
				distanceFromCenter = math.Sqrt(dx*dx + dy*dy)
				isInCorner = true
			} else if fx < radius && fy > height-radius {
				// Bottom-left corner
				dx := fx - radius
				dy := fy - (height - radius)
				distanceFromCenter = math.Sqrt(dx*dx + dy*dy)
				isInCorner = true
			} else if fx > width-radius && fy > height-radius {
				// Bottom-right corner
				dx := fx - (width - radius)
				dy := fy - (height - radius)
				distanceFromCenter = math.Sqrt(dx*dx + dy*dy)
				isInCorner = true
			}

			if isInCorner {
				// Check if pixel is outside the outer radius (cut off completely)
				if distanceFromCenter > radius {
					img.Set(x, y, color.Transparent)
				} else if distanceFromCenter > radius-float64(borderThickness) {
					// Pixel is in the border ring of the rounded corner
					img.Set(x, y, borderColor)
				} else {
					currentPixel := img.RGBAAt(x, y)
					if currentPixel == borderColor.(color.RGBA) {
						img.Set(x, y, color.Transparent)
					}
				}
			}
		}
	}
}

// func roundImageCorners(img *image.RGBA, radius float64) {
// 	applyCornerRounding(img, radius, 0, color.Transparent)
// }

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

func (g *GridProcessor) Process(img image.Image, theme string, format string) (image.Image, types.ImageMetadata, error) {

	newImg, err := applyGridToImage(img, &g.options)
	if err != nil {
		return nil, types.ImageMetadata{}, err
	}
	return newImg, types.ImageMetadata{}, nil

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
