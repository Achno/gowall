package image

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/Achno/gowall/config"
	imageio "github.com/Achno/gowall/internal/image_io"
	"github.com/Achno/gowall/internal/logger"
	"github.com/Achno/gowall/terminal"
	"github.com/Achno/gowall/utils"
	webp "github.com/HugoSmits86/nativewebp"
	_ "golang.org/x/image/webp"
)

// Available formats to Encode an image in
var encoders = map[string]func(file *os.File, img image.Image) error{
	"png": func(file *os.File, img image.Image) error {
		png := &png.Encoder{
			CompressionLevel: png.BestSpeed,
		}
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

// NoOpImageProcessor  implements ImageProcessor but does nothing.
// Its used to just to convert images from one format to another without altering them.
//
//	Example from "img.webp" --> "img.png"
type NoOpImageProcessor struct{}

// Implement the Process method
func (p *NoOpImageProcessor) Process(img image.Image, options string) (image.Image, error) {
	// Simply return the image without any modifications
	return img, nil
}

func LoadImage(imgSrc imageio.ImageReader) (image.Image, error) {
	reader, err := imgSrc.Open()
	if err != nil {
		return nil, err
	}
	defer reader.Close()
	imgData, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}
	img, _, err := image.Decode(bytes.NewReader(imgData))
	if err != nil {
		return nil, err
	}
	return img, nil
}

func SaveImage(img image.Image, output imageio.ImageWriter, format string) error {
	encoder, ok := encoders[strings.ToLower(format)]

	if !ok {
		return fmt.Errorf("unsupported format: %s", format)
	}

	if img == nil {
		return nil
	}
	file, err := output.Create()
	if err != nil {
		return err
	}
	defer file.Close()
	return encoder(file, img)
}

func SaveGif(gifData gif.GIF, output string) error {
	var file *os.File
	if output == "/dev/stdout" || output == "-" || output == "CON" {
		file = os.Stdout
	} else {
		var err error
		file, err = os.Create(output)
		if err != nil {
			return fmt.Errorf("failed to create output file: %w", err)
		}
		defer file.Close() // Ensure the file gets closed properly
	}
	err := gif.EncodeAll(file, &gifData)
	if err != nil {
		return fmt.Errorf("while Encoding gif : %w", err)
	}

	logger.Printf("Gif processed and saved as %s\n\n", output)
	return nil
}

func SaveUrlAsImg(url string) (string, error) {
	extension, err := utils.GetFileExtensionFromURL(url)
	if err != nil {
		return "", err
	}

	timestamp := time.Now().Format("20060102-150405")
	fileName := fmt.Sprintf("wall-%s%s", timestamp, extension)

	path := filepath.Join(config.GowallConfig.OutputFolder, fileName)

	file, err := os.Create(path)
	if err != nil {
		return "", fmt.Errorf("could not create file: %w", err)
	}
	defer file.Close()

	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("could not fetch the URL: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to fetch image: status code %d", resp.StatusCode)
	}

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return "", fmt.Errorf("could not write to file: %w", err)
	}

	return path, nil
}

// Opens the image on the default viewing application of every operating system.
// or in the terminal for kitty,wezterm,ghostty and konsole
func OpenImageInViewer(filePath string) error {
	if !config.GowallConfig.EnableImagePreviewing {
		return nil
	}

	var cmd *exec.Cmd

	if terminal.IsKittyTerminalRunning() {
		cmd = exec.Command("kitty", "icat", filePath)
		cmd.Stdout = os.Stdout

		return cmd.Run()
	}

	isKonsoleOrGhostty := terminal.IsKonsoleTerminalRunning() || terminal.IsGhosttyTerminalRunning()

	if isKonsoleOrGhostty && terminal.HasIcat() && !config.GowallConfig.InlineImagePreview {
		cmd = exec.Command("kitty", "icat", filePath)
		cmd.Stdout = os.Stdout

		return cmd.Run()
	}

	if isKonsoleOrGhostty && config.GowallConfig.InlineImagePreview {
		return terminal.RenderKittyImg(filePath)
	}

	if terminal.IsWeztermTerminalRunning() {
		cmd = exec.Command("wezterm", "imgcat", filePath)
		cmd.Stdout = os.Stdout

		return cmd.Run()
	}

	switch runtime.GOOS {

	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", filePath)
	case "darwin":
		cmd = exec.Command("open", filePath)
	case "linux", "freebsd", "openbsd":
		cmd = exec.Command("xdg-open", filePath)

	default:
		return fmt.Errorf("unsupported platform")
	}

	return cmd.Start()
}

// Processes the image depending on a processor that impliments the "ImageProcessor" interface.
func ProcessImgs(processor ImageProcessor, imageOps []imageio.ImageIO, theme string) ([]string, error) {
	var wg sync.WaitGroup
	remaining := int32(len(imageOps))
	errChan := make(chan error, len(imageOps))
	var processedImagesFilePaths []string

	// Load the image
	for index, imageOp := range imageOps {
		wg.Add(1)
		go func(i int, imgProcessor ImageProcessor, currentImgOp imageio.ImageIO) {
			defer wg.Done()
			theme := theme
			img, err := LoadImage(currentImgOp.ImageInput)
			if err != nil {
				errChan <- fmt.Errorf("while loading image: %w", err)
				return
			}
			// optionally specify a temporary theme via json file in runtime
			if strings.HasSuffix(theme, ".json") {
				theme, err = loadThemeFromJson(theme)
				if err != nil {
					errChan <- fmt.Errorf("file %s : %w", currentImgOp.ImageInput, err)
					return
				}
			}
			// Process the image
			newImg, err := imgProcessor.Process(img, theme)
			if err != nil {
				errChan <- fmt.Errorf("while processing image: %w", err)
				return
			}

			// Save the image
			err = SaveImage(newImg, currentImgOp.ImageOutput, currentImgOp.Format)
			if err != nil {
				errChan <- fmt.Errorf("while saving image: %w in %s", err, currentImgOp.ImageOutput)
				return
			}
			remainingCount := atomic.AddInt32(&remaining, -1)
			logger.Printf(" ::: Image %d Completed , %d Images left ::: \n", i+1, remainingCount)
			processedImagesFilePaths = append(processedImagesFilePaths, currentImgOp.ImageOutput.String())
		}(index, processor, imageOp)
	}
	wg.Wait()
	close(errChan)

	if len(errChan) > 0 {
		// return <-errChan
		var errs []error

		for err := range errChan {
			errs = append(errs, err)
		}

		return processedImagesFilePaths, errors.New(utils.FormatErrors(errs))
	}
	return processedImagesFilePaths, nil
}

// returns themeName that was inserted to the theme map
func loadThemeFromJson(jsonTheme string) (string, error) {
	reader, err := os.Open(jsonTheme)
	if err != nil {
		return "", fmt.Errorf("error opening image file")
	}
	defer reader.Close()
	data, err := io.ReadAll(reader)
	if err != nil {
		return "", fmt.Errorf("while reading the json file")
	}
	var tm struct {
		Name   string   `json:"name"`
		Colors []string `json:"colors"`
	}

	if err := json.Unmarshal(data, &tm); err != nil {
		return "", fmt.Errorf("while parsing json theme file, ensure your .json is written correctly")
	}
	if len(tm.Name) <= 0 || len(tm.Colors) < 1 {
		return "", fmt.Errorf("json file does not contain a name or colors field(s)")
	}
	clrs, err := HexToRGBASlice(tm.Colors)
	if err != nil {
		return "", err
	}
	themes[strings.ToLower(tm.Name)] = Theme{
		Name:   tm.Name,
		Colors: clrs,
	}

	return tm.Name, nil
}
