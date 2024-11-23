package image

import (
	"fmt"
	"image"
	"os"
	"os/exec"

	"github.com/Achno/gowall/internal/upscaler"
)

type UpscaleProcessor struct {
}

func (p *UpscaleProcessor) Process(img image.Image, theme string) (image.Image, error) {

	upscaler.SetupUpscaler()

	cmd := exec.Command("/home/achno/Pictures/gowall/upscaler/realesrgan-ncnn-vulkan-v0.2.0-ubuntu/realesrgan-ncnn-vulkan", "-h")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr // Capture error output as well

	err := cmd.Run()
	if err != nil {
		// Check if the error is related to the specific exit code
		exitError, ok := err.(*exec.ExitError)
		if ok && exitError.ExitCode() == 255 {
			// Treat exit code 255 as a "success" for the help menu
			fmt.Println("Help menu displayed successfully (exit code 255).")
			return nil, nil
		}
		// Log other errors as real failures
		fmt.Printf("Command failed: %v\n", err)
	}

	return nil, nil
}
