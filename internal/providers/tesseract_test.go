package providers

import (
	"context"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"strings"
	"testing"

	gwimage "github.com/Achno/gowall/internal/image"
)

func TestParseSceneTSVFilteringAndDedup(t *testing.T) {
	sampleTSV := strings.Join([]string{
		"level\tpage_num\tblock_num\tpar_num\tline_num\tword_num\tleft\ttop\twidth\theight\tconf\ttext",
		"5\t1\t1\t1\t1\t1\t1303\t500\t25\t5\t86.147926\t—",
		"5\t1\t1\t1\t1\t2\t1359\t469\t168\t71\t56.556267\tÇarşamba",
		"5\t1\t1\t1\t1\t3\t1544\t469\t84\t71\t72.967941\tİndir",
		"5\t1\t1\t1\t1\t4\t1550\t470\t139\t49\t93.089119\tİndirimi",
		"5\t1\t4\t1\t1\t1\t271\t581\t482\t117\t19.846277\tSÖYLEME",
		"5\t1\t5\t1\t1\t1\t1351\t567\t324\t137\t95.577606\tSÖYLE",
		"5\t1\t6\t1\t1\t1\t279\t714\t69\t36\t31.903023\tamme",
		"5\t1\t6\t1\t1\t2\t357\t730\t48\t13\t43.987587\tvr",
		"5\t1\t6\t1\t1\t3\t424\t718\t158\t28\t45.346050\tA",
	}, "\n")

	got := parseSceneTSV(sampleTSV)
	want := "Çarşamba İndirimi\n\nSÖYLEME\n\nSÖYLE"
	if got != want {
		t.Fatalf("parseSceneTSV got:\n%q\nwant:\n%q", got, want)
	}
}

func TestTesseractSceneAndHUDOCR(t *testing.T) {
	p, err := NewTesseractProvider(Config{})
	if err != nil {
		t.Fatalf("NewTesseractProvider: %v", err)
	}

	morpheusPath := "/home/thebanri/.gemini/antigravity-cli/brain/9fe6c207-4e89-4ceb-83b1-a740665b2fe4/.user_uploaded/uploaded_media_1793547965918.jpg"
	if f, err := os.Open(morpheusPath); err == nil {
		defer f.Close()
		img, _, err := image.Decode(f)
		if err != nil {
			t.Fatalf("Decode: %v", err)
		}

		resDirect, err := p.OCR(context.Background(), OCRInput{Image: img})
		if err != nil {
			t.Fatalf("Direct OCR error: %v", err)
		}
		if resDirect.Text != "Çarşamba İndirimi\n\nSÖYLEME\n\nSÖYLE" {
			t.Fatalf("Direct OCR got %q, want exact 3 scene lines", resDirect.Text)
		}

		grayProc := gwimage.GrayScaleProcessor{}
		grayImg, _, err := grayProc.Process(img, "", "")
		if err != nil {
			t.Fatalf("GrayScaleProcessor: %v", err)
		}
		resGray, err := p.OCR(context.Background(), OCRInput{Image: grayImg})
		if err != nil {
			t.Fatalf("Gray OCR error: %v", err)
		}
		if resGray.Text != "Çarşamba İndirimi\n\nSÖYLEME\n\nSÖYLE" {
			t.Fatalf("Gray OCR got %q, want exact 3 scene lines", resGray.Text)
		}
	}
}
