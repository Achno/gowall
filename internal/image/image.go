package image

import (
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/Achno/gowall/utils"
	"github.com/chai2010/webp"
)

// Available formats to encode an image in
var encoders = map[string]func(file *os.File, img image.Image) error{
	"png": func(file *os.File, img image.Image) error {
		return png.Encode(file, img)
	},
	"jpg": func(file *os.File, img image.Image) error {
		return jpeg.Encode(file, img, nil)
	},
	"jpeg": func(file *os.File, img image.Image) error {
		return jpeg.Encode(file, img, nil)
	},
	"webp": func(file *os.File, img image.Image) error {
		return webp.Encode(file, img, nil)
	},
}

// ImageProcessor defines an interface for processing images
type ImageProcessor interface {
	Process(image.Image, string) (image.Image, error)
}

// LoadImage opens and decodes an image from the given file path
func LoadImage(filePath string) (image.Image, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	return img, err
}

// SaveImage encodes and saves an image to the given file path
func SaveImage(img image.Image, filePath string, format string) error {
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder, ok := encoders[strings.ToLower(format)]
	if !ok {
		return fmt.Errorf("unsupported format: %s", format)
	}

	return encoder(file, img)
}

// ProcessImg handles the workflow of loading, processing, and saving an image
func ProcessImg(imgPath string, processor ImageProcessor, theme string) error {
	img, err := LoadImage(imgPath)
	if err != nil {
		fmt.Println("Error loading image:", err)
		return err
	}

	newImg, err := processor.Process(img, theme)
	if err != nil {
		fmt.Println("Error processing image:", err)
		return err
	}

	extension := strings.ToLower(filepath.Ext(imgPath))
	if extension == "" {
		return fmt.Errorf("could not determine file extension")
	}

	// Remove '.' from the extension
	extension = extension[1:]

	dirPath, err := utils.CreateDirectory()
	if err != nil {
		fmt.Println("Error creating directory:", err)
		return err
	}

	nameOfFile := filepath.Base(imgPath)
	outputFilePath := filepath.Join(dirPath, nameOfFile)

	err = SaveImage(newImg, outputFilePath, extension)
	if err != nil {
		fmt.Println("Error saving image:", err)
		return err
	}

	fmt.Printf("Image processed and saved as %s\n", outputFilePath)
	return nil
}

// ProcessBatchImgs processes a batch of images concurrently
func ProcessBatchImgs(files []string, theme string, processor ImageProcessor) {
	var wg sync.WaitGroup
	var remaining int32 = int32(len(files))
	var hasError bool

	for index, file := range files {
		wg.Add(1)

		go func(file string, index int) {
			defer wg.Done()

			err := ProcessImg(file, processor, theme)
			if err != nil {
				hasError = true
				fmt.Printf("Error processing image %d: %v\n", index, err)
			}

			remainingCount := atomic.AddInt32(&remaining, -1)
			fmt.Printf(" ::: Image %d Completed, %d Images left ::: \n", index, remainingCount)
		}(file, index)
	}

	wg.Wait()

	if hasError {
		fmt.Println("Some images were not processed successfully.")
	}
}
