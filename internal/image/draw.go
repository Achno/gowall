package image

import (
	"image"
	"image/color"
	"image/draw"
	"math"

	clr "github.com/Achno/gowall/internal/backends/color"
	types "github.com/Achno/gowall/internal/types"
	"github.com/Achno/gowall/utils"
)

type BorderProcessor struct {
	Color           color.RGBA
	BorderThickness int
	CornerRadius    float64
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

	if cornerRadius > 0 {
		newImg = roundImageCorners(newImg, cornerRadius, borderThickness, borderColor)
	}

	return newImg
}

// roundImageCorners rounds the corners of an image with optional border support and anti-aliasing.
// If borderThickness is 0, no border is applied (simple rounding only).
// If borderThickness > 0, a border is drawn around the rounded corners.
func roundImageCorners(img image.Image, radius float64, borderThickness int, borderColor color.Color) *image.RGBA {
	bounds := img.Bounds()
	dst := image.NewRGBA(bounds)
	draw.Draw(dst, bounds, img, image.Point{}, draw.Src)

	w, h := float64(bounds.Dx()), float64(bounds.Dy())

	for y := 0; y < bounds.Dy(); y++ {
		for x := 0; x < bounds.Dx(); x++ {
			fx, fy := float64(x), float64(y)
			var dx, dy float64
			inCorner := false

			// Check all 4 corners
			if fx < radius && fy < radius { // Top Left
				dx, dy = fx-radius, fy-radius
				inCorner = true
			} else if fx > w-radius && fy < radius { // Top Right
				dx, dy = fx-(w-radius), fy-radius
				inCorner = true
			} else if fx < radius && fy > h-radius { // Bottom Left
				dx, dy = fx-radius, fy-(h-radius)
				inCorner = true
			} else if fx > w-radius && fy > h-radius { // Bottom Right
				dx, dy = fx-(w-radius), fy-(h-radius)
				inCorner = true
			}

			if inCorner {
				dist := math.Sqrt(dx*dx + dy*dy)

				if dist > radius {
					dst.Set(x, y, color.Transparent)
				} else if borderThickness > 0 && dist > radius-float64(borderThickness) {
					dst.Set(x, y, borderColor)
				} else if dist > radius-4 {
					samples := 0
					totalSamples := 8
					subPixelOffsets := []float64{-0.375, -0.125, 0.125, 0.375}

					for _, offsetX := range subPixelOffsets {
						for _, offsetY := range subPixelOffsets {
							subDx := dx + offsetX
							subDy := dy + offsetY
							subDist := math.Sqrt(subDx*subDx + subDy*subDy)

							if subDist <= radius && (borderThickness == 0 || subDist < radius-float64(borderThickness)) {
								samples++
							}
						}
					}

					if samples > 0 && samples < totalSamples {
						// Partially covered - apply anti-aliasing
						coverage := float64(samples) / float64(totalSamples)
						c := dst.RGBAAt(x, y)
						c.A = uint8(float64(c.A) * coverage)
						dst.SetRGBA(x, y, c)
					} else if samples == 0 {
						// Fully outside
						dst.Set(x, y, color.Transparent)
					}
					// If samples == totalSamples, pixel is fully inside, keep as-is
				} else if borderThickness > 0 {
					// Handle border color cleanup for non-corner border pixels
					currentPixel := dst.RGBAAt(x, y)
					bcR, bcG, bcB, bcA := borderColor.RGBA()
					cpR, cpG, cpB, cpA := currentPixel.RGBA()
					if cpR == bcR && cpG == bcG && cpB == bcB && cpA == bcA {
						dst.Set(x, y, color.Transparent)
					}
				}
			}
		}
	}

	return dst
}

type RoundProcessor struct {
	CornerRadius float64
}

func (r *RoundProcessor) Process(img image.Image, theme string, format string) (image.Image, types.ImageMetadata, error) {
	newImg := roundImageCorners(img, r.CornerRadius, 0, color.Transparent)
	return newImg, types.ImageMetadata{}, nil
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
		c, err := clr.HexToRGBA(gridColor)
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
