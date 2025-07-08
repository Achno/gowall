package providers

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"sync"

	"github.com/Achno/gowall/internal/image"
	imageio "github.com/Achno/gowall/internal/image_io"
	"github.com/Achno/gowall/utils"

	ext "github.com/reugn/go-streams/extension"
	"github.com/reugn/go-streams/flow"
)

func ProcessOCR(ops []imageio.ImageIO, provider OCRProvider) error {
	data, err := buildOCRInputs(ops)
	if err != nil {
		return err
	}

	// (1) Pre-Processing Pipeline
	out := make(chan any, len(ops))
	source := ext.NewChanSource(toAnyChan(data))

	gp := image.GrayScaleProcessor{}
	grayScaleFlow := flow.NewMap(func(item any) any {
		dataItem := item.(*OCRInput)
		// If it is a pdf, skip it
		if dataItem.Image == nil {
			return dataItem
		}
		img, err := gp.Process(dataItem.Image, "")
		if err != nil {
			return nil
		}
		dataItem.Image = img
		return dataItem
	}, 5)

	sink := ext.NewChanSink(out)

	source.
		Via(grayScaleFlow).
		// Via(flow.NewMap(func(item any) any {
		// 	return item
		// }, 10)).
		To(sink)

	// (2) OCR
	res, err := provider.OCRBatch(context.Background(), sinkToOCRInputs(out))
	if err != nil {
		return err
	}

	for _, item := range res {
		fmt.Println(item.Text)
		fmt.Println("###################")
	}

	return nil
}

func buildOCRInputs(ops []imageio.ImageIO) (chan *OCRInput, error) {
	inputsChan := make(chan *OCRInput, len(ops))
	var wg sync.WaitGroup

	for _, op := range ops {
		wg.Add(1)
		go func(op imageio.ImageIO) {
			defer wg.Done()

			path := op.ImageInput.String()
			ext := strings.ToLower(filepath.Ext(path))
			var input *OCRInput

			switch ext {
			case ".pdf":
				pdf, err := imageio.LoadFileBytes(op.ImageInput)
				if err != nil {
					utils.HandleError(err)
					return
				}
				input = &OCRInput{
					Type:     InputTypePDF,
					PDFData:  pdf,
					Filename: path,
				}
			case ".png", ".jpg", ".jpeg", ".webp":
				img, err := imageio.LoadImage(op.ImageInput)
				if err != nil {
					utils.HandleError(err)
					return
				}
				input = &OCRInput{
					Type:     InputTypeImage,
					Image:    img,
					Filename: path,
				}
			}

			if input != nil {
				inputsChan <- input
			}
		}(op)
	}

	go func() {
		wg.Wait()
		close(inputsChan)
	}()

	return inputsChan, nil
}

func toAnyChan(inputChan chan *OCRInput) chan any {
	out := make(chan any, cap(inputChan))
	for v := range inputChan {
		out <- v
	}
	close(out)
	return out
}

func sinkToOCRInputs(sinkChan chan any) []OCRInput {
	var results []OCRInput
	for item := range sinkChan {
		if ocrInput, ok := item.(*OCRInput); ok && ocrInput != nil {
			results = append(results, *ocrInput)
		}
	}
	return results
}
