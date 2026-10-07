package image

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"

	"github.com/Achno/gowall/config"
	"github.com/Achno/gowall/internal/backends/upload"
	imageio "github.com/Achno/gowall/internal/image_io"
	"github.com/Achno/gowall/internal/logger"
	"github.com/Achno/gowall/utils"
)

// UploadImgs uploads the images one by one and prints each link to stdout.
func UploadImgs(uploader upload.ImageUploader, imageOps []imageio.ImageIO) error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	failed := 0
	for _, op := range imageOps {
		imgName := filepath.Base(op.ImageInput.String())
		utils.Spinner.Start()
		utils.Spinner.Message(fmt.Sprintf("Uploading %s...", imgName))

		uploadCtx, cancel := context.WithTimeout(ctx, config.UploadImageTimeout)
		res, err := uploader.Upload(uploadCtx, op.ImageInput)
		cancel()
		if err != nil {
			utils.Spinner.StopFail()
			logger.Errorf("%s: %v", op.ImageInput, err)
			failed++
			if ctx.Err() != nil {
				break // Ctrl+C, don't start the rest
			}
			continue
		}

		utils.Spinner.StopMessage("Uploaded " + imgName)
		utils.Spinner.Stop()
		fmt.Println(res.URL)
	}

	if failed > 0 {
		return fmt.Errorf("%d of %d uploads failed", failed, len(imageOps))
	}
	return nil
}
