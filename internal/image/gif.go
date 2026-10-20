package image

import (
	"image"
	"image/color/palette"
	"image/gif"
	"io"
	"runtime"
	"sync"

	gifbackend "github.com/Achno/gowall/internal/backends/gif"
	types "github.com/Achno/gowall/internal/types"
)

const (
	NoResize = iota
	Resize
)

// GifProcessor implements MultiImageProcessor for creating animated GIFs
type GifProcessor struct {
	Loop  int // 0 loops forever, -1 shows the frames only once, anything else loop+1
	Delay int // Delay in 100ths of a second between frames
	Mode  int // Resize (1) NoResize (0) for resizing all images to same dimensions
}

// Composite processes multiple images into a single animated GIF
func (g *GifProcessor) Composite(images []image.Image, theme string, format string) (image.Image, types.ImageMetadata, error) {

	var maxWidth, maxHeight int
	for _, img := range images {
		bounds := img.Bounds()
		width, height := bounds.Dx(), bounds.Dy()
		if width > maxWidth {
			maxWidth = width
		}
		if height > maxHeight {
			maxHeight = height
		}
	}

	processedImages, err := paletteFramesFromImages(images, maxWidth, maxHeight, g.Mode)
	if err != nil {
		return nil, types.ImageMetadata{}, err
	}

	newGif := &gif.GIF{
		LoopCount: g.Loop,
		Disposal:  make([]byte, len(processedImages)),
	}

	for i, img := range processedImages {
		newGif.Image = append(newGif.Image, img.paletted)
		newGif.Delay = append(newGif.Delay, g.Delay)
		newGif.Disposal[i] = gif.DisposalNone
	}

	// Return a custom encoder function in metadata for a very hacky solution to allow Composite() to work with gifs.
	metadata := types.ImageMetadata{
		EncoderFunction: func(w io.Writer, img image.Image) error {
			return gif.EncodeAll(w, newGif)
		},
	}

	return nil, metadata, nil
}

// palleted holds the result of image processing
type palleted struct {
	paletted *image.Paletted
	bounds   image.Rectangle
}

// paletteFramesFromImages converts images to paletted with a 216-bit palette and optionally resizes all frames to the same dimensions.
func paletteFramesFromImages(images []image.Image, maxWidth, maxHeight int, mode int) ([]*palleted, error) {
	processedImages := make([]*palleted, len(images))
	next := make(chan int)

	var wg sync.WaitGroup
	for range min(runtime.GOMAXPROCS(0), len(images)) {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := range next {
				img := images[i]
				if mode == Resize {
					img = ResizeWithPadding(img, maxWidth, maxHeight)
				}

				// Convert to paletted
				bounds := img.Bounds()
				paletted := image.NewPaletted(bounds, palette.WebSafe)
				gifbackend.DitherWebSafe(paletted, img)

				processedImages[i] = &palleted{
					paletted: paletted,
					bounds:   bounds,
				}
			}
		}()
	}

	for i := range images {
		next <- i
	}
	close(next)
	wg.Wait()

	return processedImages, nil
}
