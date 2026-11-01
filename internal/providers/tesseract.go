package providers

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"unicode"
)

// Implements the Provider interface
type TesseractProvider struct {
	config Config
	client *TesseractClient
}

func NewTesseractProvider(config Config) (OCRProvider, error) {
	return &TesseractProvider{
		config: config,
		client: &TesseractClient{},
	}, nil
}

func (p *TesseractProvider) OCR(ctx context.Context, input OCRInput) (*OCRResult, error) {
	if !p.client.IsTesseractInstalled() {
		return nil, fmt.Errorf("tesseract is not installed")
	}

	if p.config.OCR.Language == "" {
		p.config.OCR.Language = "eng"
	}

	return p.client.OCRImageCmd(ctx, input.Image, p.config.OCR.Language)
}

func (p *TesseractProvider) GetConfig() Config {
	return p.config
}

type TesseractClient struct{}

func resolveTessdataArgs(lang string) ([]string, string) {
	if home, err := os.UserHomeDir(); err == nil {
		customDir := filepath.Join(home, ".config", "gowall", "tessdata")
		if info, err := os.Stat(customDir); err == nil && info.IsDir() {
			if lang == "eng" {
				if _, err := os.Stat(filepath.Join(customDir, "tur.traineddata")); err == nil {
					lang = "tur+eng"
				}
			}
			return []string{"--tessdata-dir", customDir}, lang
		}
	}
	return nil, lang
}

func (c *TesseractClient) OCRImageCmd(ctx context.Context, img image.Image, lang string) (*OCRResult, error) {
	extraArgs, resolvedLang := resolveTessdataArgs(lang)

	// Pass 1: Standard document/HUD OCR on original image (--psm 3)
	buf := new(bytes.Buffer)
	if err := jpeg.Encode(buf, img, &jpeg.Options{Quality: 95}); err != nil {
		return nil, err
	}

	args := append([]string{}, extraArgs...)
	args = append(args, "-l", resolvedLang)
	cmd := c.ConstructTesseractCommand(ctx, args)
	out1, err := c.RunTesseractCommand(ctx, cmd, buf)
	if err != nil && resolvedLang != "eng" {
		cmd = c.ConstructTesseractCommand(ctx, []string{"-l", "eng"})
		out1, err = c.RunTesseractCommand(ctx, cmd, bytes.NewBuffer(buf.Bytes()))
	}
	if err != nil {
		return nil, fmt.Errorf("while running tesseract: %w", err)
	}
	text1 := strings.TrimSpace(out1)

	// Pass 2: Scene / Meme / Outlined-Typography binarization with Sparse Text (--psm 11).
	// On photos/memes/wallpapers where the background is not white paper, standard --psm 3
	// either returns empty or misses large outlined display text (e.g. meme captions).
	sceneImg, isSceneLike := preprocessSceneText(img)
	if isSceneLike || !isCleanDocumentText(text1) {
		sceneBuf := new(bytes.Buffer)
		if err := png.Encode(sceneBuf, sceneImg); err == nil {
			sceneArgs := append([]string{}, extraArgs...)
			sceneArgs = append(sceneArgs, "-l", resolvedLang, "--psm", "11", "-c", "tessedit_create_tsv=1")
			sceneCmd := c.ConstructTesseractCommand(ctx, sceneArgs)
			if out2, err2 := c.RunTesseractCommand(ctx, sceneCmd, sceneBuf); err2 == nil {
				text2 := parseSceneTSV(out2)
				if scoreOCRText(text2) > scoreOCRText(text1) {
					return &OCRResult{Text: text2}, nil
				}
			}
		}
	}

	return &OCRResult{Text: text1}, nil
}

// isCleanDocumentText checks whether standard --psm 3 output looks like clean
// multi-word text without excessive symbol noise lines.
func isCleanDocumentText(s string) bool {
	if s == "" {
		return false
	}
	lines := strings.Split(strings.ReplaceAll(s, "\r\n", "\n"), "\n")
	goodLines, noisyLines := 0, 0
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		letters, symbols := 0, 0
		for _, r := range line {
			if unicode.IsLetter(r) || unicode.IsDigit(r) {
				letters++
			} else if !unicode.IsSpace(r) {
				symbols++
			}
		}
		if letters >= 4 && symbols <= letters/2 {
			goodLines++
		} else {
			noisyLines++
		}
	}
	return goodLines >= 2 && noisyLines == 0
}

func scoreOCRText(s string) int {
	if s == "" {
		return 0
	}
	lines := strings.Split(strings.ReplaceAll(s, "\r\n", "\n"), "\n")
	score := 0
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		letters, symbols := 0, 0
		for _, r := range line {
			if unicode.IsLetter(r) || unicode.IsDigit(r) {
				letters++
			} else if !unicode.IsSpace(r) {
				symbols++
			}
		}
		if letters >= 4 && symbols == 0 {
			score += letters * 4
		} else if letters >= 4 && symbols <= 2 {
			score += letters * 2
		} else {
			score -= 15
		}
	}
	return score
}

// preprocessSceneText converts bright foreground text (white/yellow/light glyphs
// over a dark or photographic background) into black ink on white background.
func preprocessSceneText(src image.Image) (image.Image, bool) {
	b := src.Bounds()
	w, h := b.Dx(), b.Dy()
	dst := image.NewGray(image.Rect(0, 0, w, h))

	brightCount := 0
	total := max(1, w*h)

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			r16, g16, b16, _ := src.At(b.Min.X+x, b.Min.Y+y).RGBA()
			r, g, bl := int(r16>>8), int(g16>>8), int(b16>>8)

			lum := (299*r + 587*g + 114*bl) / 1000
			if lum >= 200 {
				brightCount++
				dst.SetGray(x, y, color.Gray{Y: 0}) // Black glyph ink
			} else {
				dst.SetGray(x, y, color.Gray{Y: 255}) // White background
			}
		}
	}

	isSceneLike := float64(brightCount)/float64(total) < 0.45
	return dst, isSceneLike
}

type tsvWord struct {
	blockNum int
	parNum   int
	lineNum  int
	left     int
	top      int
	width    int
	height   int
	conf     float64
	text     string
}

func parseSceneTSV(raw string) string {
	lines := strings.Split(strings.ReplaceAll(raw, "\r\n", "\n"), "\n")
	type lineKey struct {
		block, par, line int
	}
	var lineOrder []lineKey
	lineWords := make(map[lineKey][]tsvWord)

	for _, row := range lines {
		cols := strings.Split(row, "\t")
		if len(cols) < 12 || cols[0] != "5" {
			continue
		}
		blockNum, _ := strconv.Atoi(cols[2])
		parNum, _ := strconv.Atoi(cols[3])
		lineNum, _ := strconv.Atoi(cols[4])
		left, _ := strconv.Atoi(cols[6])
		top, _ := strconv.Atoi(cols[7])
		width, _ := strconv.Atoi(cols[8])
		height, _ := strconv.Atoi(cols[9])
		conf, err := strconv.ParseFloat(cols[10], 64)
		if err != nil || conf < 0 {
			continue
		}
		text := strings.Trim(cols[11], " \t—–-~_|.,:;!?'\"`‘’“”()[]{}\\/")
		text = strings.TrimSpace(text)
		if text == "" {
			continue
		}

		letters, symbols := 0, 0
		for _, r := range text {
			if unicode.IsLetter(r) || unicode.IsDigit(r) {
				letters++
			} else if !unicode.IsSpace(r) {
				symbols++
			}
		}
		if letters == 0 || symbols > letters/2 {
			continue
		}

		// Filter out low-confidence small background texture noise (e.g. skin/wall creases),
		// while preserving either high-confidence text OR large display/headline words.
		isHighConf := conf >= 52.0 && letters >= 2
		isLargeHeadline := height >= 70 && conf >= 15.0 && letters >= 3
		if !isHighConf && !isLargeHeadline {
			continue
		}

		key := lineKey{blockNum, parNum, lineNum}
		if _, exists := lineWords[key]; !exists {
			lineOrder = append(lineOrder, key)
		}
		lineWords[key] = append(lineWords[key], tsvWord{
			blockNum: blockNum,
			parNum:   parNum,
			lineNum:  lineNum,
			left:     left,
			top:      top,
			width:    width,
			height:   height,
			conf:     conf,
			text:     text,
		})
	}

	var keptLines []string
	for _, key := range lineOrder {
		words := lineWords[key]
		if len(words) == 0 {
			continue
		}
		var dedup []string
		for i, w := range words {
			// Skip a word if the next word on the same line overlaps horizontally or starts with it (curved banner stutter)
			if i+1 < len(words) {
				next := words[i+1]
				overlapsX := w.left+w.width > next.left+4
				isPrefix := len([]rune(w.text)) >= 3 && strings.HasPrefix(strings.ToLower(next.text), strings.ToLower(w.text))
				if (overlapsX || isPrefix) && next.conf >= w.conf {
					continue
				}
			}
			dedup = append(dedup, w.text)
		}
		joined := strings.TrimSpace(strings.Join(dedup, " "))
		runeCount := 0
		for _, r := range joined {
			if unicode.IsLetter(r) || unicode.IsDigit(r) {
				runeCount++
			}
		}
		if runeCount >= 3 {
			keptLines = append(keptLines, joined)
		}
	}

	return strings.Join(keptLines, "\n\n")
}

// IsTesseractInstalled checks if tesseract is installed in $PATH
func (c *TesseractClient) IsTesseractInstalled() bool {
	_, err := exec.LookPath("tesseract")
	return err == nil
}

func (c *TesseractClient) ConstructTesseractCommand(ctx context.Context, args []string) *exec.Cmd {
	fullArgs := append([]string{"stdin", "stdout"}, args...)
	return exec.CommandContext(ctx, "tesseract", fullArgs...)
}

func (c *TesseractClient) RunTesseractCommand(ctx context.Context, cmd *exec.Cmd, input *bytes.Buffer) (string, error) {
	cmd.Stdin = input

	var out bytes.Buffer
	cmd.Stdout = &out
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("%v, stderr: %s", err, stderr.String())
	}

	return out.String(), nil
}
