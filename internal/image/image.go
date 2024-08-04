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

// Available formats to Encode an image in
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

// Create a Processor of this interface and call 'ProcessImg'
type ImageProcessor interface {
	Process(image.Image, string) (image.Image, error)
}

func LoadImage(filePath string) (image.Image, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	return img, err
}

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

// 1. Loads the img, 2. Processes it depending on the type of Processor you put which implements
// the 'ImageProcessor' interface, 3. Creates the necessary directories, 4. Saves the image there
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

	// Extract file extension from imgPath
	extension := strings.ToLower(filepath.Ext(imgPath))
	if extension == "" {
		fmt.Println("Error: Could not determine file extension.")
		return fmt.Errorf("could not determine file extension")
	}

	// Remove '.' from the extension
	extension = extension[1:]

	dirPath, err := utils.CreateDirectory()
	if err != nil {
		fmt.Println("Error creating Directory or getting path:", err)
		return err
	}

	nameOfFile := filepath.Base(imgPath)
	outputFilePath := filepath.Join(dirPath, nameOfFile)

	err = SaveImage(newImg, outputFilePath, extension)
	if err != nil {
		fmt.Println("Error saving image:", err, outputFilePath)
		return err
	}

	fmt.Printf("Image processed and saved as %s\n", outputFilePath)
	return nil
}

// Process images concurrently and return the first error if there was one
func ProcessBatchImgs(files []string, theme string, processor ImageProcessor) error {
	var wg sync.WaitGroup
	var remaining int32 = int32(len(files))
	errChan := make(chan error, len(files))

	for index, file := range files {
		wg.Add(1)

		go func(file string, index int) {
			defer wg.Done()

			err := ProcessImg(file, processor, theme)
			if err != nil {
				errChan <- fmt.Errorf("file %s: %w", file, err)
				return
			}
			remainingCount := atomic.AddInt32(&remaining, -1)
			fmt.Printf(" ::: Image %d Completed, %d Images left ::: \n", index, remainingCount)
		}(file, index)
	}

	wg.Wait()
	close(errChan)

	if len(errChan) > 0 {
		return <-errChan
	}

	return nil
}
