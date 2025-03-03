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
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/Achno/gowall/config"
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

func LoadImage(imgSrc utils.ImageSource) (image.Image, error) {
	reader, err := imgSrc.Open()
	if err != nil {
		return nil, err
	}

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

func SaveImage(img image.Image, filePath string, format string) error {
	encoder, ok := encoders[strings.ToLower(format)]

	if !ok {
		return fmt.Errorf("unsupported format: %s", format)
	}

	file, err := os.Create(filePath)
	if err != nil {
		return err
	}

	defer file.Close()
	return encoder(file, img)
}

func SaveGif(gifData gif.GIF, fileName string) error {
	dirFolder, err := utils.CreateDirectory()
	if err != nil {
		return err
	}

	outFile, err := os.Create(filepath.Join(dirFolder, "gifs", fileName+".gif"))
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer outFile.Close()

	err = gif.EncodeAll(outFile, &gifData)
	if err != nil {
		return fmt.Errorf("while Encoding gif : %w", err)
	}

	fmt.Printf("Gif processed and saved as %s\n\n", outFile.Name())
	return nil
}

func SaveUrlAsImg(url string) (string, error) {
	extension, err := utils.GetFileExtensionFromURL(url)
	if err != nil {
		return "", err
	}

	dirFolder, err := utils.CreateDirectory()
	if err != nil {
		return "", fmt.Errorf("while creating Directory or getting path")
	}

	timestamp := time.Now().Format("20060102-150405")
	fileName := fmt.Sprintf("wall-%s%s", timestamp, extension)

	path := filepath.Join(dirFolder, fileName)

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
func OpenImage(filePath string) error {
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

type ProcessOptions struct {
	SaveToFile bool   // Whether to save the processed image to file
	OutputExt  string // Optional output extension to override the original
	OutputName string // Optional outputName
	OutputDir  string // Optional Output directory
}

func DefaultProcessOptions() ProcessOptions {
	return ProcessOptions{
		SaveToFile: true,
	}
}

// Processes the image depending on a processor that impliments the "ImageProcessor" interface.
// You can pass an optional  "ProcessOptions" struct with extra options.
func ProcessImg(imgSrc utils.ImageSource, processor ImageProcessor, theme string, opts ...ProcessOptions) (string, *image.Image, error) {
	// Use default options if none provided
	options := DefaultProcessOptions()
	if len(opts) > 0 {
		options = opts[0]
	}

	// Handle directory creation
	dirPath, err := utils.CreateDirectory()
	if err != nil {
		return "", nil, fmt.Errorf("while creating directory: %w", err)
	}

	// Load the image
	img, err := LoadImage(imgSrc)
	if err != nil {
		return "", nil, fmt.Errorf("while loading image: %w", err)
	}

	// optionally specify a temporary theme via json file in runtime
	if strings.HasSuffix(theme, ".json") {
		theme, err = loadThemeFromJson(theme)
		if err != nil {
			return "", nil, err
		}
	}

	// Process the image
	newImg, err := processor.Process(img, theme)
	if err != nil {
		return "", nil, fmt.Errorf("while processing image: %w", err)
	}

	// If we don't need to save, return early with the processed image
	if !options.SaveToFile {
		return "", &newImg, nil
	}

	finalExt, err := determineFinalFormatExt(imgSrc, options)
	if err != nil {
		return "", nil, fmt.Errorf("error while determining file extension")
	}

	outputFilePath, err := buildOutputPath(options, dirPath)
	if err != nil {
		return "", nil, err
	}
	// Save the image
	err = SaveImage(newImg, outputFilePath, finalExt)
	if err != nil {
		return "", nil, fmt.Errorf("while saving image: %w in %s", err, outputFilePath)
	}

	fmt.Printf("Image processed and saved as %s\n\n", outputFilePath)
	return outputFilePath, &newImg, nil
}

// returns themeName that was inserted to the theme map
func loadThemeFromJson(jsonTheme string) (string, error) {
	expandFile := utils.ExpandHomeDirectory([]string{jsonTheme})
	reader, err := expandFile[0].Open()
	if err != nil {
		return "", fmt.Errorf("error opening image file")
	}
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

func determineFinalFormatExt(imgSrc utils.ImageSource, opts ProcessOptions) (string, error) {
	if opts.OutputExt != "" {
		return opts.OutputExt, nil
	}
	if ext := path.Ext(opts.OutputName); ext != "" {
		return ext, nil
	}
	ext, err := utils.ReadImageFormat(imgSrc)
	if err != nil {
		return "", err
	}
	return ext, nil
}

func buildOutputPath(options ProcessOptions,ext string) (string, error) {
  if options.OutputDir
	ext := path.Ext(options.OutputName)
	if ext == "" {
		return filepath.Join(options.OutputDir, options.OutputName), nil
	}
  if 

	return "", nil
}

// Process images concurrently and return the first error if there was one
func ProcessBatchImgs(fileSources []utils.FileSource, theme string, processor ImageProcessor, opts *ProcessOptions) error {
	var wg sync.WaitGroup
	var remaining int32 = int32(len(fileSources))
	errChan := make(chan error, len(fileSources))

	for index, file := range fileSources {

		wg.Add(1)

		go func(fileSrc utils.FileSource, index int) {
			defer wg.Done()
			opts.OutputName = file.Path
			_, _, err := ProcessImg(file, processor, theme, *opts)
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
