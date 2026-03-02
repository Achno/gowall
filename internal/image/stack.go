package image

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"strings"

	types "github.com/Achno/gowall/internal/types"
)

const (
	StackLayoutHorizontal = "horizontal"
	StackLayoutVertical   = "vertical"
	StackLayoutGrid       = "grid"

	StackResizeOff     = "off"
	StackResizeBiggest = "biggest"
)

func StackLayoutList() []string {
	return []string{StackLayoutHorizontal, StackLayoutVertical, StackLayoutGrid}
}

func StackResizeList() []string {
	return []string{StackResizeOff, StackResizeBiggest}
}

type StackProcessor struct {
	LayoutMode      string
	Rows            int
	Cols            int
	BorderThickness int
	BorderColor     color.RGBA
	ResizeMode      string
}

func (p *StackProcessor) Composite(images []image.Image, theme string, format string) (image.Image, types.ImageMetadata, error) {
	resizeMode := strings.ToLower(p.ResizeMode)
	rows, cols, err := p.resolveGrid(len(images))
	if err != nil {
		return nil, types.ImageMetadata{}, err
	}

	cellWidth, cellHeight := stackCellDimensions(images, resizeMode)
	if cellWidth <= 0 || cellHeight <= 0 {
		return nil, types.ImageMetadata{}, fmt.Errorf("invalid image dimensions")
	}

	canvasWidth := cols*cellWidth + (cols+1)*p.BorderThickness
	canvasHeight := rows*cellHeight + (rows+1)*p.BorderThickness
	stacked := image.NewRGBA(image.Rect(0, 0, canvasWidth, canvasHeight))

	if p.BorderThickness > 0 {
		draw.Draw(stacked, stacked.Bounds(), &image.Uniform{C: p.BorderColor}, image.Point{}, draw.Src)
	}

	// Clear all cell interiors so empty cells remain transparent.
	for row := range rows {
		for col := range cols {
			x := p.BorderThickness + col*(cellWidth+p.BorderThickness)
			y := p.BorderThickness + row*(cellHeight+p.BorderThickness)
			cellRect := image.Rect(x, y, x+cellWidth, y+cellHeight)
			draw.Draw(stacked, cellRect, image.Transparent, image.Point{}, draw.Src)
		}
	}

	for idx, img := range images {
		if idx >= rows*cols {
			break
		}

		current := img
		if resizeMode == StackResizeBiggest {
			current = ResizeWithPadding(img, cellWidth, cellHeight)
		}

		p.drawImageInCell(stacked, current, idx, cols, cellWidth, cellHeight)
	}

	return stacked, types.ImageMetadata{}, nil
}

func (p *StackProcessor) resolveGrid(imageCount int) (int, int, error) {
	switch strings.ToLower(p.LayoutMode) {
	case StackLayoutHorizontal:
		return 1, imageCount, nil
	case StackLayoutVertical:
		return imageCount, 1, nil
	case StackLayoutGrid:
		if p.Rows <= 0 || p.Cols <= 0 {
			return 0, 0, fmt.Errorf("grid layout requires positive rows and cols")
		}
		if imageCount > p.Rows*p.Cols {
			return 0, 0, fmt.Errorf("layout has %d cells but %d images were provided", p.Rows*p.Cols, imageCount)
		}
		return p.Rows, p.Cols, nil
	default:
		return 0, 0, fmt.Errorf("invalid layout mode: %s", p.LayoutMode)
	}
}

func stackCellDimensions(images []image.Image, resizeMode string) (int, int) {
	first := images[0].Bounds()
	if resizeMode == StackResizeBiggest {
		minWidth := first.Dx()
		minHeight := first.Dy()

		for _, img := range images[1:] {
			bounds := img.Bounds()
			if bounds.Dx() < minWidth {
				minWidth = bounds.Dx()
			}
			if bounds.Dy() < minHeight {
				minHeight = bounds.Dy()
			}
		}

		return minWidth, minHeight
	}

	maxWidth := first.Dx()
	maxHeight := first.Dy()

	for _, img := range images[1:] {
		bounds := img.Bounds()
		if bounds.Dx() > maxWidth {
			maxWidth = bounds.Dx()
		}
		if bounds.Dy() > maxHeight {
			maxHeight = bounds.Dy()
		}
	}

	return maxWidth, maxHeight
}

func (p *StackProcessor) drawImageInCell(dst *image.RGBA, src image.Image, index, cols, cellWidth, cellHeight int) {
	row := index / cols
	col := index % cols

	x := p.BorderThickness + col*(cellWidth+p.BorderThickness)
	y := p.BorderThickness + row*(cellHeight+p.BorderThickness)

	bounds := src.Bounds()
	srcWidth := bounds.Dx()
	srcHeight := bounds.Dy()

	offsetX := x + (cellWidth-srcWidth)/2
	offsetY := y + (cellHeight-srcHeight)/2
	dstRect := image.Rect(offsetX, offsetY, offsetX+srcWidth, offsetY+srcHeight)

	draw.Draw(dst, dstRect, src, bounds.Min, draw.Over)
}
