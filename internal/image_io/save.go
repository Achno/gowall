package imageio

import (
	"fmt"
	"image"
	"image/gif"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Achno/gowall/config"
	"github.com/Achno/gowall/internal/logger"
	types "github.com/Achno/gowall/internal/types"
	"github.com/Achno/gowall/utils"
)

func SaveImage(img image.Image, output ImageWriter, format string, metadata types.ImageMetadata) error {
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

	if metadata.EncoderFunction != nil {
		return metadata.EncoderFunction(file, img)
	}

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

func SaveText(text string, output ImageWriter) error {
	if text == "" {
		return nil
	}

	file, err := output.Create()
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer file.Close()

	_, err = file.WriteString(text)
	if err != nil {
		return fmt.Errorf("failed to write text to file: %w", err)
	}

	return nil
}
