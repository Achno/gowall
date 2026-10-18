package image

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"math"
	"strings"
	"sync"

	cpkg "github.com/Achno/gowall/internal/backends/color"
	"github.com/Achno/gowall/internal/backends/pixels"
	types "github.com/Achno/gowall/internal/types"
	"github.com/disintegration/imaging"
)

type FlipProcessor struct{}

func (p *FlipProcessor) Process(img image.Image, theme string, format string) (image.Image, types.ImageMetadata, error) {

	src := pixels.ToRGBA(img)
	width, height := src.Rect.Dx(), src.Rect.Dy()
	newImg := image.NewRGBA(src.Rect)

	pixels.ParallelRows(height, func(y0, y1 int) {
		for y := y0; y < y1; y++ {
			srcRow := src.Pix[y*src.Stride : y*src.Stride+width*4]
			dstRow := newImg.Pix[y*newImg.Stride : y*newImg.Stride+width*4]
			for x := range width {
				copy(dstRow[x*4:x*4+4], srcRow[(width-x-1)*4:(width-x)*4])
			}
		}
	})
	return newImg, types.ImageMetadata{}, nil
}

type MirrorProcessor struct{}

func (p *MirrorProcessor) Process(img image.Image, theme string, format string) (image.Image, types.ImageMetadata, error) {

	src := pixels.ToRGBA(img)
	width, height := src.Rect.Dx(), src.Rect.Dy()
	half := width / 2
	newImg := image.NewRGBA(src.Rect)

	pixels.ParallelRows(height, func(y0, y1 int) {
		for y := y0; y < y1; y++ {
			srcRow := src.Pix[y*src.Stride : y*src.Stride+width*4]
			dstRow := newImg.Pix[y*newImg.Stride : y*newImg.Stride+width*4]
			// Copy the original left half
			copy(dstRow[:half*4], srcRow[:half*4])
			// Mirror the left half onto the right half
			for x := range half {
				copy(dstRow[(half+x)*4:(half+x+1)*4], srcRow[(half-1-x)*4:(half-x)*4])
			}
		}
	})
	return newImg, types.ImageMetadata{}, nil
}

type GrayScaleProcessor struct{}

func (p *GrayScaleProcessor) Process(img image.Image, theme string, format string) (image.Image, types.ImageMetadata, error) {

	src := pixels.ToRGBA(img)
	width, height := src.Rect.Dx(), src.Rect.Dy()
	grayImg := image.NewGray(src.Rect)

	pixels.ParallelRows(height, func(y0, y1 int) {
		for y := y0; y < y1; y++ {
			srcRow := src.Pix[y*src.Stride : y*src.Stride+width*4]
			dstRow := grayImg.Pix[y*grayImg.Stride : y*grayImg.Stride+width]
			for x := range dstRow {
				r, g, b := srcRow[x*4], srcRow[x*4+1], srcRow[x*4+2]
				// luminosity formula
				dstRow[x] = uint8((0.299*float64(r) + 0.587*float64(g) + 0.114*float64(b)))
			}
		}
	})
	return grayImg, types.ImageMetadata{}, nil
}

type BrightnessProcessor struct {
	Factor float64
}

func (p *BrightnessProcessor) Process(img image.Image, theme string, format string) (image.Image, types.ImageMetadata, error) {

	// every channel value maps to the same new value, so compute them once
	var lut [256]uint8
	for v := range lut {
		lut[v] = uint8(cpkg.Clamp(int(float64(v)*p.Factor), 0, 255))
	}

	src := pixels.ToRGBA(img)
	newImg := image.NewRGBA(src.Rect)
	width, height := src.Rect.Dx(), src.Rect.Dy()

	pixels.ParallelRows(height, func(y0, y1 int) {
		for y := y0; y < y1; y++ {
			srcRow := src.Pix[y*src.Stride : y*src.Stride+width*4]
			dstRow := newImg.Pix[y*newImg.Stride : y*newImg.Stride+width*4]
			for i := 0; i < len(srcRow); i += 4 {
				dstRow[i+0] = lut[srcRow[i+0]]
				dstRow[i+1] = lut[srcRow[i+1]]
				dstRow[i+2] = lut[srcRow[i+2]]
				dstRow[i+3] = srcRow[i+3]
			}
		}
	})

	return newImg, types.ImageMetadata{}, nil
}

type ContrastProcessor struct {
	Mode          string
	Factor        float64
	Midpoint      float64
	SigmoidFactor float64
}

const (
	ContrastModeNormal  = "normal"
	ContrastModeSigmoid = "sigmoid"
)

func (p *ContrastProcessor) Process(img image.Image, theme string, format string) (image.Image, types.ImageMetadata, error) {
	switch strings.ToLower(p.Mode) {
	case ContrastModeNormal:
		return imaging.AdjustContrast(img, p.Factor), types.ImageMetadata{}, nil
	case ContrastModeSigmoid:
		return imaging.AdjustSigmoid(img, p.Midpoint, p.SigmoidFactor), types.ImageMetadata{}, nil
	default:
		return nil, types.ImageMetadata{}, fmt.Errorf("invalid contrast mode: %s", p.Mode)
	}
}

type GammaProcessor struct {
	Gamma float64
}

func (p *GammaProcessor) Process(img image.Image, theme string, format string) (image.Image, types.ImageMetadata, error) {
	return imaging.AdjustGamma(img, p.Gamma), types.ImageMetadata{}, nil
}

type SaturationProcessor struct {
	Percentage float64
}

func (p *SaturationProcessor) Process(img image.Image, theme string, format string) (image.Image, types.ImageMetadata, error) {
	return imaging.AdjustSaturation(img, p.Percentage), types.ImageMetadata{}, nil
}

type Preset struct {
	BackgroundStart color.RGBA
	BackgroundEnd   color.RGBA
	BackgroundImage image.Image // Optional: if set, used instead of gradient
	TiltX           float64
	TiltY           float64
	TiltZ           float64 // Z-axis rotation in degrees (positive = clockwise)
	Scale           float64
	CornerRadius    float64
}

var DefaultTiltPreset = Preset{
	BackgroundStart: color.RGBA{18, 18, 18, 255},
	BackgroundEnd:   color.RGBA{40, 40, 40, 255},
	TiltX:           5.0,
	TiltY:           -8.0,
	TiltZ:           3.0,
	Scale:           0.65,
	CornerRadius:    40.0,
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
		CornerRadius:    40.0,
	},
}

func GetTiltPresetNames() []string {
	names := make([]string, 0, len(TiltPresets))
	for k := range TiltPresets {
		names = append(names, k)
	}
	return names
}

func GetTiltPreset(name string) (Preset, error) {
	if name == "" {
		return DefaultTiltPreset, nil
	}

	preset, exists := TiltPresets[name]
	if !exists {
		return Preset{}, fmt.Errorf("invalid preset '%s'. Valid presets: %v", name, GetTiltPresetNames())
	}

	return preset, nil
}

type TiltProcessor struct {
	Preset Preset
}

func (p *TiltProcessor) Process(img image.Image, theme string, format string) (image.Image, types.ImageMetadata, error) {

	rounded := roundImageCorners(img, p.Preset.CornerRadius, 0, color.Transparent)

	rotated := rounded
	if p.Preset.TiltZ != 0 {
		rotated = rotate2D(rounded, p.Preset.TiltZ)
	}

	tilted := apply3DTransform(rotated, p.Preset.TiltX, p.Preset.TiltY, p.Preset.Scale)
	trimmed, _ := trimContent(tilted)

	outW, outH := 1920, 1080
	final := image.NewRGBA(image.Rect(0, 0, outW, outH))

	if p.Preset.BackgroundImage != nil {
		p.drawBackgroundImage(final, outW, outH)
	} else {
		p.drawBackgroundGradient(final, outW, outH)
	}

	contentBounds := trimmed.Bounds()
	targetX := (outW - contentBounds.Dx()) / 2
	targetY := (outH - contentBounds.Dy()) / 2

	draw.Draw(final, image.Rect(targetX, targetY, targetX+contentBounds.Dx(), targetY+contentBounds.Dy()), trimmed, image.Point{}, draw.Over)

	return final, types.ImageMetadata{}, nil
}

func (p *TiltProcessor) drawBackgroundGradient(final *image.RGBA, outW, outH int) {
	start, end := p.Preset.BackgroundStart, p.Preset.BackgroundEnd
	pixels.ParallelRows(outH, func(y0, y1 int) {
		for y := y0; y < y1; y++ {
			row := final.Pix[y*final.Stride : y*final.Stride+outW*4]
			for x := range outW {
				progress := float64(x+y) / float64(outW+outH)
				row[x*4+0] = uint8(float64(start.R)*(1-progress) + float64(end.R)*progress)
				row[x*4+1] = uint8(float64(start.G)*(1-progress) + float64(end.G)*progress)
				row[x*4+2] = uint8(float64(start.B)*(1-progress) + float64(end.B)*progress)
				row[x*4+3] = 255
			}
		}
	})
}

func (p *TiltProcessor) drawBackgroundImage(final *image.RGBA, outW, outH int) {
	bgImg := p.Preset.BackgroundImage
	// Scale/crop the background image to fill the output canvas (cover mode)
	resized := imaging.Fill(bgImg, outW, outH, imaging.Center, imaging.Lanczos)
	draw.Draw(final, final.Bounds(), resized, image.Point{}, draw.Src)
}

func rotate2D(src *image.RGBA, angleDegrees float64) *image.RGBA {
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

	pixels.ParallelRows(outH, func(y0, y1 int) {
		for dstY := y0; dstY < y1; dstY++ {
			for dstX := range outW {
				x := float64(dstX) - centerDstX
				y := float64(dstY) - centerDstY

				srcX := x*cos + y*sin + centerSrcX
				srcY := -x*sin + y*cos + centerSrcY

				if srcX >= 0 && srcX < w-1 && srcY >= 0 && srcY < h-1 {
					dst.SetRGBA(dstX, dstY, getPixelBilinear(src, srcX, srcY))
				}
			}
		}
	})

	return dst
}

func apply3DTransform(src *image.RGBA, tiltX, tiltY, scale float64) *image.RGBA {
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

	pixels.ParallelRows(outH, func(y0, y1 int) {
		for y := y0; y < y1; y++ {
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
					dst.SetRGBA(x, y, getPixelBilinear(src, srcX, srcY))
				}
			}
		}
	})
	return dst
}

// Sub-pixel sampling for smooth edges, reads the pixels directly since going through At() allocates for every one
func getPixelBilinear(img *image.RGBA, x, y float64) color.RGBA {
	x0 := int(math.Floor(x))
	y0 := int(math.Floor(y))
	wx, wy := x-float64(x0), y-float64(y0)

	// the 4 neighbours, x0+1 and y0+1 are always inside the image as callers keep x < w-1 and y < h-1
	p00 := img.PixOffset(x0, y0)
	p10 := p00 + 4
	p01 := p00 + img.Stride
	p11 := p01 + 4

	// same as interpolating the 16-bit values of RGBA() (v*0x101) and dividing by 256
	interp := func(c int) uint8 {
		v00 := float64(uint32(img.Pix[p00+c]) * 0x101)
		v10 := float64(uint32(img.Pix[p10+c]) * 0x101)
		v01 := float64(uint32(img.Pix[p01+c]) * 0x101)
		v11 := float64(uint32(img.Pix[p11+c]) * 0x101)
		top := v00*(1-wx) + v10*wx
		bot := v01*(1-wx) + v11*wx
		return uint8((top*(1-wy) + bot*wy) / 256)
	}

	return color.RGBA{interp(0), interp(1), interp(2), interp(3)}
}

// TrimContent trims the content of the image to remove the empty space around the image
func trimContent(img *image.RGBA) (*image.RGBA, image.Rectangle) {
	bounds := img.Bounds()
	minX, minY, maxX, maxY := bounds.Max.X, bounds.Max.Y, bounds.Min.X, bounds.Min.Y
	found := false
	var mu sync.Mutex

	// Scan for visible pixels find the box the image is in, every chunk of rows finds its own box and they are merged
	pixels.ParallelRows(bounds.Dy(), func(y0, y1 int) {
		cMinX, cMinY, cMaxX, cMaxY := bounds.Max.X, bounds.Max.Y, bounds.Min.X, bounds.Min.Y
		cFound := false
		for y := bounds.Min.Y + y0; y < bounds.Min.Y+y1; y++ {
			row := img.Pix[img.PixOffset(bounds.Min.X, y):]
			for x := bounds.Min.X; x < bounds.Max.X; x++ {
				if row[(x-bounds.Min.X)*4+3] > 0 {
					cMinX, cMaxX = min(cMinX, x), max(cMaxX, x)
					cMinY, cMaxY = min(cMinY, y), max(cMaxY, y)
					cFound = true
				}
			}
		}
		if !cFound {
			return
		}
		mu.Lock()
		minX, minY, maxX, maxY = min(minX, cMinX), min(minY, cMinY), max(maxX, cMaxX), max(maxY, cMaxY)
		found = true
		mu.Unlock()
	})

	if !found {
		return image.NewRGBA(image.Rect(0, 0, 1, 1)), image.Rect(0, 0, 1, 1)
	}

	trimmedRect := image.Rect(minX, minY, maxX+1, maxY+1)
	trimmed := image.NewRGBA(image.Rect(0, 0, trimmedRect.Dx(), trimmedRect.Dy()))
	draw.Draw(trimmed, trimmed.Bounds(), img, trimmedRect.Min, draw.Src)

	return trimmed, trimmedRect
}
