package image

import (
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"io"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/Achno/gowall/config"
	imageio "github.com/Achno/gowall/internal/image_io"
	"github.com/Achno/gowall/internal/logger"
	"github.com/Achno/gowall/terminal"
	"github.com/Achno/gowall/utils"
	_ "golang.org/x/image/webp"
)

// Create a Processor of this interface and call 'ProcessImg'
type ImageProcessor interface {
	Process(image.Image, string, string) (image.Image, error)
}

// NoOpImageProcessor  implements ImageProcessor but does nothing.
// Its used to just to convert images from one format to another without altering them.
//
//	Example from "img.webp" --> "img.png"
type NoOpImageProcessor struct{}

// Implement the Process method
func (p *NoOpImageProcessor) Process(img image.Image, options string, format string) (image.Image, error) {
	// Simply return the image without any modifications
	return img, nil
}

// Opens the image on the default viewing application of every operating system.
// or in the terminal for kitty,wezterm,ghostty and konsole
func OpenImageInViewer(filePath string) error {
	if !config.GowallConfig.EnableImagePreviewing {
		return nil
	}
	var cmd *exec.Cmd

	if config.GowallConfig.ImagePreviewBackend == "chafa" {
		if ok := terminal.HasChafa(); !ok {
			return fmt.Errorf("you specified `chafa` in ImagePreviewBackend but gowall could not find chafa in your $PATH,ensure chafa is installed")
		}

		cmd = exec.Command("chafa", filePath)
		cmd.Stdout = os.Stdout

		return cmd.Run()
	}

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
			img, err := imageio.LoadImage(currentImgOp.ImageInput)
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
			newImg, err := imgProcessor.Process(img, theme, currentImgOp.Format)
			if err != nil {
				errChan <- fmt.Errorf("while processing image: %w", err)
				return
			}

			// Save the image
			err = imageio.SaveImage(newImg, currentImgOp.ImageOutput, currentImgOp.Format)
			if err != nil {
				errChan <- fmt.Errorf("while saving image: %w in %s", err, currentImgOp.ImageOutput)
				return
			}
			remainingCount := atomic.AddInt32(&remaining, -1)
			logger.Printf("::: Image completed & saved in %s, %d Images left :::\n", currentImgOp.ImageOutput.String(), remainingCount)
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
