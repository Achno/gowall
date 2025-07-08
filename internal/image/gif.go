package image

import (
	"fmt"
	"image"
	"image/color/palette"
	"image/draw"
	"image/gif"
	"path/filepath"
	"sync"
	"time"

	"github.com/Achno/gowall/config"
	imageio "github.com/Achno/gowall/internal/image_io"

	drawx "golang.org/x/image/draw"
)

const (
	NoResize = iota
	Resize
)

type GifOptions struct {
	Loop       int    // 0 loops forever, -1 shows the frames only once, anything else loop+1
	Delay      int    // Delay in 100ths of a second between frames
	Mode       int    // Resize (0) NoResize (1) for resizing all images to same dimensions
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

func WithMode(mode int) GifOption {
	return func(g *GifOptions) { g.Mode = mode }
}

func defaultGifOptions(options []GifOption) GifOptions {
	opts := GifOptions{
		Loop:       0,
		Delay:      200,
		Mode:       Resize,
		outputName: "",
	}

	if len(options) > 0 {
		for _, option := range options {
			option(&opts)
		}
	}
	return opts
}

func CreateGif(files []imageio.ImageIO, opts ...GifOption) error {
	options := defaultGifOptions(opts)

	frames, maxWidth, maxHeight, err := scanImages(files)
	if err != nil {
		return err
	}

	processedImages, err := paletteFrames(frames, maxWidth, maxHeight, options.Mode)
	if err != nil {
		return err
	}

	if len(processedImages) == 0 {
		return fmt.Errorf("no images were processed successfully")
	}

	newGif := &gif.GIF{
		LoopCount: options.Loop,
		Disposal:  make([]byte, len(processedImages)),
	}

	for i, img := range processedImages {
		newGif.Image = append(newGif.Image, img.paletted)
		newGif.Delay = append(newGif.Delay, options.Delay)
		newGif.Disposal[i] = gif.DisposalNone
	}

	fileName := options.outputName
	if options.outputName == "" {
		timestamp := time.Now().Format(time.DateTime)
		fileName = fmt.Sprintf("gif-%s", timestamp)
		fileName = filepath.Join(config.GowallConfig.OutputFolder, fileName)
	}

	err = imageio.SaveGif(*newGif, fileName)
	if err != nil {
		return fmt.Errorf("while saving gif: %w", err)
	}
	return nil
}

// resizeImage resizes an image to the specified width and height while preserving aspect ratio
func resizeImage(img image.Image, width, height int) image.Image {
	srcBounds := img.Bounds()
	srcWidth := srcBounds.Dx()
	srcHeight := srcBounds.Dy()

	widthRatio := float64(width) / float64(srcWidth)
	heightRatio := float64(height) / float64(srcHeight)

	// Use the smaller ratio to ensure the image fits within the target dimensions
	ratio := widthRatio
	if heightRatio < widthRatio {
		ratio = heightRatio
	}
	newWidth := int(float64(srcWidth) * ratio)
	newHeight := int(float64(srcHeight) * ratio)
	dstRect := image.Rect(0, 0, newWidth, newHeight)

	dst := image.NewRGBA(dstRect)
	drawx.CatmullRom.Scale(dst, dstRect, img, img.Bounds(), draw.Over, nil)

	// Center the image in the target dimensions if needed
	if newWidth < width || newHeight < height {
		centered := image.NewRGBA(image.Rect(0, 0, width, height))
		offsetX := (width - newWidth) / 2
		offsetY := (height - newHeight) / 2
		draw.Draw(centered, centered.Bounds(), dst, image.Point{-offsetX, -offsetY}, draw.Over)
		return centered
	}

	return dst
}

// palleted holds the result of image processing
type palleted struct {
	paletted *image.Paletted
	bounds   image.Rectangle
}

// frames holds both the loaded image and its dimensions
type frames struct {
	image         image.Image
	width, height int
}

// scanImages loads images and returns their data plus their max dimensions (Frame)
func scanImages(files []imageio.ImageIO) ([]*frames, int, int, error) {
	var maxWidth, maxHeight int
	var mu sync.Mutex
	var wg sync.WaitGroup

	imageChan := make(chan *frames, len(files))
	errChan := make(chan error, len(files))

	for _, file := range files {
		wg.Add(1)
		go func(f *imageio.ImageIO) {
			defer wg.Done()

			img, err := imageio.LoadImage(f.ImageInput)
			if err != nil {
				errChan <- fmt.Errorf("while loading image: %w", err)
				return
			}

			bounds := img.Bounds()
			width, height := bounds.Dx(), bounds.Dy()

			// Update max dimensions thread safe to be used in resizing of the images
			mu.Lock()
			if width > maxWidth {
				maxWidth = width
			}
			if height > maxHeight {
				maxHeight = height
			}
			mu.Unlock()

			imageChan <- &frames{
				image:  img,
				width:  width,
				height: height,
			}
		}(&file)
	}

	wg.Wait()
	close(imageChan)
	close(errChan)

	select {
	case err := <-errChan:
		if err != nil {
			return nil, 0, 0, err
		}
	default:
	}

	var images []*frames
	for img := range imageChan {
		images = append(images, img)
	}

	return images, maxWidth, maxHeight, nil
}

// paletteFrames converts frames to paletted with a 216-bit palette and optionally resizes all frames to the same dimensions
func paletteFrames(images []*frames, maxWidth, maxHeight int, mode int) ([]*palleted, error) {
	const maxWorkers = 5
	results := make(chan *palleted, len(images))
	errChan := make(chan error, len(images))

	// Process in batches to control concurrency
	var processedImages []*palleted

	for i := 0; i < len(images); i += maxWorkers {
		end := i + maxWorkers
		if end > len(images) {
			end = len(images)
		}

		var wg sync.WaitGroup
		for j := i; j < end; j++ {
			wg.Add(1)
			go func(imgData *frames) {
				defer wg.Done()
				img := imgData.image

				if mode == Resize {
					img = resizeImage(img, maxWidth, maxHeight)
				}
				// Free the original image for garbage collection
				imgData.image = nil

				// Convert to paletted
				bounds := img.Bounds()
				paletted := image.NewPaletted(bounds, palette.WebSafe)
				draw.FloydSteinberg.Draw(paletted, bounds, img, bounds.Min)

				results <- &palleted{
					paletted: paletted,
					bounds:   bounds,
				}
			}(images[j])
		}
		wg.Wait()
	}

	close(results)
	close(errChan)

	select {
	case err := <-errChan:
		if err != nil {
			return nil, err
		}
	default:
	}

	for result := range results {
		processedImages = append(processedImages, result)
	}
	return processedImages, nil
}
