package image

import (
	"image"
	"image/color"
	"image/draw"
	"math"

	types "github.com/Achno/gowall/internal/types"
)

type FlipProcessor struct{}

func (p *FlipProcessor) Process(img image.Image, theme string, format string) (image.Image, types.ImageMetadata, error) {

	bounds := img.Bounds()
	width, height := bounds.Dx(), bounds.Dy()
	newImg := image.NewRGBA(bounds)

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			pixel := img.At(width-x-1, y)
			newImg.Set(x, y, pixel)
		}
	}
	return newImg, types.ImageMetadata{}, nil
}

type MirrorProcessor struct{}

func (p *MirrorProcessor) Process(img image.Image, theme string, format string) (image.Image, types.ImageMetadata, error) {

	bounds := img.Bounds()
	width, height := bounds.Dx(), bounds.Dy()
	newImg := image.NewRGBA(bounds)

	// Copy the original left half
	for y := 0; y < height; y++ {
		for x := 0; x < width/2; x++ {
			pixel := img.At(x, y)
			newImg.Set(x, y, pixel)
		}
	}

	// Mirror the left half onto the right half
	for y := 0; y < height; y++ {
		for x := 0; x < width/2; x++ {
			pixel := img.At(width/2-1-x, y)
			newImg.Set(width/2+x, y, pixel)
		}
	}
	return newImg, types.ImageMetadata{}, nil
}

type GrayScaleProcessor struct{}

func (p *GrayScaleProcessor) Process(img image.Image, theme string, format string) (image.Image, types.ImageMetadata, error) {

	bounds := img.Bounds()
	grayImg := image.NewGray(bounds)

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			originalColor := img.At(x, y)
			r, g, b, _ := originalColor.RGBA()

			// luminosity formula
			grayValue := uint8((0.299*float64(r>>8) + 0.587*float64(g>>8) + 0.114*float64(b>>8)))

			grayImg.SetGray(x, y, color.Gray{Y: grayValue})
		}
	}
	return grayImg, types.ImageMetadata{}, nil
}

type BrightnessProcessor struct {
	Factor float64
}

func (p *BrightnessProcessor) Process(img image.Image, theme string, format string) (image.Image, types.ImageMetadata, error) {

	bounds := img.Bounds()
	newImg := image.NewRGBA(bounds)

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			origColor := color.RGBAModel.Convert(img.At(x, y)).(color.RGBA)

			newR := uint8(clamp(int(float64(origColor.R)*p.Factor), 0, 255))
			newG := uint8(clamp(int(float64(origColor.G)*p.Factor), 0, 255))
			newB := uint8(clamp(int(float64(origColor.B)*p.Factor), 0, 255))
			newA := origColor.A

			newImg.Set(x, y, color.RGBA{R: newR, G: newG, B: newB, A: newA})
		}
	}

	return newImg, types.ImageMetadata{}, nil
}

func clamp(val, min, max int) int {
	if val < min {
		return min
	} else if val > max {
		return max
	}
	return val
}

type Preset struct {
	BackgroundStart color.RGBA
	BackgroundEnd   color.RGBA
	TiltX           float64
	TiltY           float64
	TiltZ           float64 // Z-axis rotation in degrees (positive = clockwise)
	Scale           float64
	CornerRadius    float64
}

var TiltPresets = map[string]Preset{
	"p1": {
		BackgroundStart: color.RGBA{0, 0, 0, 255},    // Black
		BackgroundEnd:   color.RGBA{40, 40, 40, 255}, // Dark Gray
		TiltX:           5.0,
		TiltY:           -8.0,
		TiltZ:           3.0,
		Scale:           0.70,
		CornerRadius:    40.0,
	},
	"p2": {
		BackgroundStart: color.RGBA{0, 0, 0, 255},    // Black
		BackgroundEnd:   color.RGBA{40, 40, 40, 255}, // Dark Gray
		TiltX:           -5.0,
		TiltY:           8.0,
		TiltZ:           -4.0,
		Scale:           0.70,
		CornerRadius:    40.0,
	},
	"p3": {
		BackgroundStart: color.RGBA{240, 240, 245, 255}, // Light Gray
		BackgroundEnd:   color.RGBA{200, 200, 210, 255}, // Medium Gray
		TiltX:           5.0,
		TiltY:           -8.0,
		TiltZ:           0.0,
		Scale:           0.60,
		CornerRadius:    40.0,
	},
	"p4": {
		BackgroundStart: color.RGBA{240, 240, 245, 255}, // Light Gray
		BackgroundEnd:   color.RGBA{200, 200, 210, 255}, // Medium Gray
		TiltX:           10.0,
		TiltY:           5.0,
		TiltZ:           0.0,
		Scale:           0.70,
		CornerRadius:    30.0,
	},
}

func GetTiltPresetNames() []string {
	names := make([]string, 0, len(TiltPresets))
	for k := range TiltPresets {
		names = append(names, k)
	}
	return names
}

type TiltProcessor struct {
	Preset Preset
}

func (p *TiltProcessor) Process(img image.Image, theme string, format string) (image.Image, types.ImageMetadata, error) {

	rounded := roundImageCorners(img, p.Preset.CornerRadius, 0, color.Transparent)

	var rotated image.Image = rounded
	if p.Preset.TiltZ != 0 {
		rotated = rotate2D(rounded, p.Preset.TiltZ)
	}

	tilted := apply3DTransform(rotated, p.Preset.TiltX, p.Preset.TiltY, p.Preset.Scale)
	trimmed, _ := trimContent(tilted)

	outW, outH := 1920, 1080
	final := image.NewRGBA(image.Rect(0, 0, outW, outH))

	// Draw Gradient
	for y := range outH {
		for x := range outW {
			progress := float64(x+y) / float64(outW+outH)
			r := uint8(float64(p.Preset.BackgroundStart.R)*(1-progress) + float64(p.Preset.BackgroundEnd.R)*progress)
			g := uint8(float64(p.Preset.BackgroundStart.G)*(1-progress) + float64(p.Preset.BackgroundEnd.G)*progress)
			b := uint8(float64(p.Preset.BackgroundStart.B)*(1-progress) + float64(p.Preset.BackgroundEnd.B)*progress)
			final.Set(x, y, color.RGBA{r, g, b, 255})
		}
	}

	contentBounds := trimmed.Bounds()
	targetX := (outW - contentBounds.Dx()) / 2
	targetY := (outH - contentBounds.Dy()) / 2

	draw.Draw(final, image.Rect(targetX, targetY, targetX+contentBounds.Dx(), targetY+contentBounds.Dy()), trimmed, image.Point{}, draw.Over)

	return final, types.ImageMetadata{}, nil
}

func rotate2D(src image.Image, angleDegrees float64) *image.RGBA {
	bounds := src.Bounds()
	w, h := float64(bounds.Dx()), float64(bounds.Dy())

	// Convert angle to radians
	angleRad := angleDegrees * math.Pi / 180
	cos := math.Cos(angleRad)
	sin := math.Sin(angleRad)

	// Calculate bounding box for rotated image
	corners := [][2]float64{
		{0, 0}, {w, 0}, {0, h}, {w, h},
	}

	minX, minY := math.MaxFloat64, math.MaxFloat64
	maxX, maxY := -math.MaxFloat64, -math.MaxFloat64

	for _, corner := range corners {
		x := corner[0] - w/2
		y := corner[1] - h/2
		newX := x*cos - y*sin
		newY := x*sin + y*cos

		if newX < minX {
			minX = newX
		}
		if newX > maxX {
			maxX = newX
		}
		if newY < minY {
			minY = newY
		}
		if newY > maxY {
			maxY = newY
		}
	}

	outW := int(maxX - minX + 1)
	outH := int(maxY - minY + 1)
	dst := image.NewRGBA(image.Rect(0, 0, outW, outH))

	centerSrcX := w / 2
	centerSrcY := h / 2
	centerDstX := float64(outW) / 2
	centerDstY := float64(outH) / 2

	for dstY := range outH {
		for dstX := range outW {
			x := float64(dstX) - centerDstX
			y := float64(dstY) - centerDstY

			srcX := x*cos + y*sin + centerSrcX
			srcY := -x*sin + y*cos + centerSrcY

			if srcX >= 0 && srcX < w-1 && srcY >= 0 && srcY < h-1 {
				dst.Set(dstX, dstY, getPixelBilinear(src, srcX, srcY))
			}
		}
	}

	return dst
}

func apply3DTransform(src image.Image, tiltX, tiltY, scale float64) *image.RGBA {
	bounds := src.Bounds()
	w, h := float64(bounds.Dx()), float64(bounds.Dy())

	// Large canvas to catch rotation
	outW, outH := int(w*2.5), int(h*2.5)
	dst := image.NewRGBA(image.Rect(0, 0, outW, outH))

	centerX := float64(outW) / 2
	centerY := float64(outH) / 2

	radX := tiltX * math.Pi / 180
	radY := tiltY * math.Pi / 180
	cosX, sinX := math.Cos(radX), math.Sin(radX)
	cosY, sinY := math.Cos(radY), math.Sin(radY)

	f := 1500.0 // Focal Length

	for y := range outH {
		for x := range outW {
			dx := float64(x) - centerX
			dy := float64(y) - centerY

			z := dx*sinY + dy*(-sinX)*cosY + f*cosX*cosY
			if z <= 0 {
				continue
			}

			scaleFactor := f / z * scale

			srcX := (dx*cosY + dy*sinX*sinY + f*(-sinY)) / scaleFactor
			srcY := (dy*cosX + f*sinX) / scaleFactor

			srcX += w / 2
			srcY += h / 2

			if srcX >= 0 && srcX < w-1 && srcY >= 0 && srcY < h-1 {
				dst.Set(x, y, getPixelBilinear(src, srcX, srcY))
			}
		}
	}
	return dst
}

// Sub-pixel sampling for smooth edges
func getPixelBilinear(img image.Image, x, y float64) color.RGBA {
	x0 := int(math.Floor(x))
	y0 := int(math.Floor(y))
	x1, y1 := x0+1, y0+1
	wx, wy := x-float64(x0), y-float64(y0)

	get := func(ix, iy int) (uint32, uint32, uint32, uint32) {
		return img.At(ix, iy).RGBA()
	}

	interp := func(v00, v10, v01, v11 uint32) uint8 {
		top := float64(v00)*(1-wx) + float64(v10)*wx
		bot := float64(v01)*(1-wx) + float64(v11)*wx
		return uint8((top*(1-wy) + bot*wy) / 256)
	}

	r00, g00, b00, a00 := get(x0, y0)
	r10, g10, b10, a10 := get(x1, y0)
	r01, g01, b01, a01 := get(x0, y1)
	r11, g11, b11, a11 := get(x1, y1)

	return color.RGBA{
		interp(r00, r10, r01, r11),
		interp(g00, g10, g01, g11),
		interp(b00, b10, b01, b11),
		interp(a00, a10, a01, a11),
	}
}

// TrimContent trims the content of the image to remove the empty space around the image
func trimContent(img image.Image) (*image.RGBA, image.Rectangle) {
	bounds := img.Bounds()
	minX, minY, maxX, maxY := bounds.Max.X, bounds.Max.Y, bounds.Min.X, bounds.Min.Y
	found := false

	// Scan for visible pixels find the box the image is in
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			_, _, _, a := img.At(x, y).RGBA()
			if a > 0 {
				if x < minX {
					minX = x
				}
				if x > maxX {
					maxX = x
				}
				if y < minY {
					minY = y
				}
				if y > maxY {
					maxY = y
				}
				found = true
			}
		}
	}

	if !found {
		return image.NewRGBA(image.Rect(0, 0, 1, 1)), image.Rect(0, 0, 1, 1)
	}

	trimmedRect := image.Rect(minX, minY, maxX+1, maxY+1)
	trimmed := image.NewRGBA(image.Rect(0, 0, trimmedRect.Dx(), trimmedRect.Dy()))
	draw.Draw(trimmed, trimmed.Bounds(), img, trimmedRect.Min, draw.Src)

	return trimmed, trimmedRect
}
