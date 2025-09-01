package png

import (
	"bytes"
	"fmt"
	"image"
	"image/png"
	"os/exec"

	"github.com/Achno/gowall/internal/types"
	"github.com/Achno/gowall/utils"
)

// PngquantStrategy implements pngquant compression for PNG images
type PngquantStrategy struct {
	BinaryPath string
	Quality    int
	Speed      int
}

// NewPngquantStrategy creates a new pngquant strategy
func NewPngquantStrategy(quality int, speed int) (*PngquantStrategy, error) {

	_, err := CheckPngquantInstalled()
	if err != nil {

		text := `◈ It seems that pngquant is not setup yet, would you like for gowall to set it up.`

		ok := utils.Confirm(utils.BlueColor + text + utils.ResetColor)
		if !ok {
			return nil, fmt.Errorf("pngquant has not been setup, you could always use another backend for png compression via --method, or install pngquant via your package manager")
		}

		err := SetupPngquant()
		if err != nil {
			return nil, fmt.Errorf("while setting up pngquant: %w", err)
		}
	}

	binaryPath, err := CheckPngquantInstalled()
	if err != nil {
	}

	return &PngquantStrategy{
		BinaryPath: binaryPath,
		Quality:    quality,
		Speed:      speed,
	}, nil
}

// Compress implements the CompressionStrategy interface
func (p *PngquantStrategy) Compress(img image.Image) (image.Image, types.ImageMetadata, error) {
	if err := p.ValidateParams(); err != nil {
		return nil, types.ImageMetadata{}, fmt.Errorf("while validating parameters: %w", err)
	}

	// Encode input image to PNG format in memory
	var inputBuffer bytes.Buffer
	err := png.Encode(&inputBuffer, img)
	if err != nil {
		return nil, types.ImageMetadata{}, fmt.Errorf("failed to encode input image to PNG: %w", err)
	}

	qualityRange := fmt.Sprintf("--quality=%d-%d", p.Quality-10, p.Quality)
	if p.Quality <= 10 {
		qualityRange = fmt.Sprintf("--quality=0-%d", p.Quality)
	}

	numColors := "256"
	args := []string{numColors, qualityRange}
	if p.Speed > 0 {
		args = append(args, "--speed", fmt.Sprintf("%d", p.Speed))
	}
	args = append(args, "--force", "-", "--output", "-")

	cmd := exec.Command(p.BinaryPath, args...)
	fmt.Println("THE COMMAND: ", cmd.String()) //!

	// Set up pipes
	cmd.Stdin = &inputBuffer
	var outputBuffer bytes.Buffer
	cmd.Stdout = &outputBuffer
	var errorBuffer bytes.Buffer
	cmd.Stderr = &errorBuffer

	err = cmd.Run()
	if err != nil {
		return nil, types.ImageMetadata{}, fmt.Errorf("pngquant failed: %w, stderr: %s", err, errorBuffer.String())
	}

	compressedImg, err := png.Decode(&outputBuffer)
	if err != nil {
		return nil, types.ImageMetadata{}, fmt.Errorf("failed to decode compressed PNG: %w", err)
	}

	return compressedImg, types.ImageMetadata{}, nil
}

// GetFormat returns the format this strategy handles
func (p *PngquantStrategy) GetFormat() string {
	return "png"
}

func (p *PngquantStrategy) ValidateParams() error {
	if p.Quality < 0 || p.Quality > 100 {
		return fmt.Errorf("quality must be between 0 and 100")
	}
	if p.Speed < 0 || p.Speed > 11 {
		return fmt.Errorf("speed must be between 0 and 11")
	}
	return nil
}
