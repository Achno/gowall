package image

import (
	"fmt"
	"image"
	"image/color/palette"
	"image/draw"
	"image/gif"
	"math"
	"time"
)

type GifOptions struct {
	Loop       int    // 0 loops forever, -1 shows the frames only once, anything else loop+1
	Delay      int    // Delay in 100ths of a second between frames
	outputName string // outputName of the gif
}

type GifOption func(*GifOptions)

func WithLoop(LoopForever int) GifOption {
	return func(g *GifOptions) { g.Loop = LoopForever }
}

func WithDelay(delay int) GifOption {
	return func(g *GifOptions) { g.Delay = delay }
}

func WithOutputName(name string) GifOption {
	return func(g *GifOptions) { g.outputName = name }
}

func defaultGifOptions(options []GifOption) GifOptions {
	opts := GifOptions{
		Loop:       0,
		Delay:      200,
		outputName: "",
	}

	if len(options) > 0 {
		for _, option := range options {
			option(&opts)
		}
	}
	return opts
}

func CreateGif(files []string, opts ...GifOption) error {
	options := defaultGifOptions(opts)

	var maxWidth, maxHeight int
	images := []image.Image{}

	for _, pngFile := range files {
		img, err := LoadImage(pngFile)
		if err != nil {
			return fmt.Errorf("while loading image: %w", err)
		}
		images = append(images, img)

		// Update max dimensions
		bounds := img.Bounds()
		if bounds.Dx() > maxWidth {
			maxWidth = bounds.Dx()
		}
		if bounds.Dy() > maxHeight {
			maxHeight = bounds.Dy()
		}
	}

	newGif := &gif.GIF{
		LoopCount: options.Loop,
	}

	for _, img := range images {
		normalized := resizeAspectRatio(img, maxWidth, maxHeight)

		paletted := image.NewPaletted(normalized.Bounds(), palette.Plan9)
		draw.FloydSteinberg.Draw(paletted, normalized.Bounds(), normalized, image.Point{})

		newGif.Image = append(newGif.Image, paletted)
		newGif.Delay = append(newGif.Delay, options.Delay)
	}

	timestamp := time.Now().Format(time.DateTime)
	fileName := fmt.Sprintf("gif-%s", timestamp)

	if options.outputName != "" {
		fileName = options.outputName
	}

	err := SaveGif(*newGif, fileName)
	if err != nil {
		return fmt.Errorf("while saving gif: %w", err)
	}
	return nil
}

func resizeAspectRatio(img image.Image, targetWidth, targetHeight int) image.Image {
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	aspectRatio := float64(width) / float64(height)

	var newWidth, newHeight int
	if float64(targetWidth)/float64(targetHeight) > aspectRatio {
		newHeight = targetHeight
		newWidth = int(math.Round(float64(newHeight) * aspectRatio))
	} else {
		newWidth = targetWidth
		newHeight = int(math.Round(float64(newWidth) / aspectRatio))
	}

	newImg := image.NewRGBA(image.Rect(0, 0, newWidth, newHeight))
	for y := 0; y < newHeight; y++ {
		for x := 0; x < newWidth; x++ {
			srcX := int(math.Round(float64(x) * float64(width) / float64(newWidth)))
			srcY := int(math.Round(float64(y) * float64(height) / float64(newHeight)))
			newImg.Set(x, y, img.At(srcX, srcY))
		}
	}
	return newImg
}
