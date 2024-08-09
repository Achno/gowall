package image

import (
	"errors"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/Achno/gowall/config"
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

// Opens the image on the default viewing application of every operating system.
//
//	If the terminal emulator "kitty" is running --> it will print the image on the terminal
func OpenImage(filePath string) error {

	if !config.GowallConfig.EnableImagePreviewing {
		return nil
	}

	var cmd *exec.Cmd

	// 50ms
	if utils.IsKittyTerminalRunning() {
		cmd = exec.Command("kitty", "icat", filePath)
		cmd.Stdout = os.Stdout

		return cmd.Run()
	}

	// 300ms for gwen
	switch runtime.GOOS {

	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", filePath)
	case "darwin":
		cmd = exec.Command("open", filePath)
	case "linux":
		cmd = exec.Command("xdg-open", filePath)

	default:
		return fmt.Errorf("unsupported platform")
	}

	return cmd.Start()

}

// 1. Loads the img, 2. Processes it depending on the type of Processor you put which impliments
// the 'ImageProcessor' interface, 3. Creates the necessary directories ,4. Saves the image there
func ProcessImg(imgPath string, processor ImageProcessor, theme string) (string, error) {

	img, err := LoadImage(imgPath)

	if err != nil {
		return "", fmt.Errorf("while loading image : %w", err)
	}

	newImg, err := processor.Process(img, theme)

	if err != nil {
		return "", fmt.Errorf("while processing image : %w", err)
	}

	//Extract file extension from imgPath
	extension := strings.ToLower(filepath.Ext(imgPath))

	if extension == "" {
		return "", fmt.Errorf("error: Could not determine file extension")
	}

	// remove '.' from the extension
	extension = extension[1:]

	dirPath, err := utils.CreateDirectory()
	nameOfFile := filepath.Base(imgPath)

	outputFilePath := filepath.Join(dirPath, nameOfFile)

	if err != nil {
		return "", fmt.Errorf("while creating Directory or getting path")
	}

	err = SaveImage(newImg, outputFilePath, extension)

	if err != nil {
		return "", fmt.Errorf("while saving image: %w in %s ", err, outputFilePath)
	}

	fmt.Printf("Image processed and saved as %s\n\n", outputFilePath)

	return outputFilePath, nil

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

			_, err := ProcessImg(file, processor, theme)

			if err != nil {
				errChan <- fmt.Errorf("file %s : %w", file, err)
				return
			}
			remainingCount := atomic.AddInt32(&remaining, -1)
			fmt.Printf(" ::: Image %d Completed , %d Images left ::: \n", index, remainingCount)
		}(file, index)

	}

	wg.Wait()
	close(errChan)

	if len(errChan) > 0 {
		// return <-errChan
		var errs []error

		for err := range errChan {
			errs = append(errs, err)
		}

		return errors.New(utils.FormatErrors(errs))
	}

	return nil
}
