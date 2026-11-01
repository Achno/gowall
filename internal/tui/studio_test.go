package tui

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Achno/gowall/config"
	"github.com/thebanri/limoni/core/buffer"
	"github.com/thebanri/limoni/core/cell"
	"github.com/thebanri/limoni/core/driver"
	"github.com/thebanri/limoni/core/terminal"
	"github.com/thebanri/limoni/uitest"
)

func TestStudioSemanticSurfaceAndNavigation(t *testing.T) {
	s := NewStudio(Options{SplitPreview: true})
	s.wakeup = nil

	page := uitest.Run(t, 120, 40, s.Draw)

	// Themes tab (default)
	page.Expect(page.GetByID("theme-list")).ToBeVisible()
	page.Expect(page.GetByID("theme-swatches")).ToBeVisible()
	page.Expect(page.GetByID("active-swatches")).ToBeVisible()
	page.Expect(page.GetByID("btn-save")).ToBeVisible()
	page.Expect(page.GetByID("btn-backend")).ToBeVisible()
	page.Expect(page.GetByID("img-original")).ToBeVisible()
	page.Expect(page.GetByID("img-preview")).ToBeVisible()
	page.Expect(page.GetByID("pipeline-info")).ToContainValue("Split (Before/After)")

	// Toggle backend (HaldCLUT -> Nearest-Neighbour)
	page.Press("b")
	page.Expect(page.GetByID("status-message")).ToContainValue("Nearest-Neighbour")

	// Toggle split view (Split -> Single)
	page.Press("v")
	page.Expect(page.GetByID("img-original")).Not().ToBeVisible()
	page.Expect(page.GetByID("img-preview")).ToBeVisible()
	page.Expect(page.GetByID("pipeline-info")).ToContainValue("Single (Preview)")

	// Switch to Effects tab (2)
	page.Press("2")
	page.Expect(page.GetByID("effect-list")).ToBeVisible()
	page.Expect(page.GetByID("effect-param")).ToBeVisible()
	page.Expect(page.GetByID("btn-chain")).ToBeVisible()
	page.Expect(page.GetByID("btn-reset")).ToBeVisible()

	// Switch to BgRm tab (3)
	page.Press("3")
	page.Expect(page.GetByID("bg-list")).ToBeVisible()
	page.Expect(page.GetByID("bg-iterations")).ToBeVisible()
	page.Expect(page.GetByID("btn-chain")).ToBeVisible()
	page.Expect(page.GetByID("btn-save")).ToBeVisible()

	// Switch to OCR tab (4)
	page.Press("4")
	page.Expect(page.GetByID("ocr-list")).ToBeVisible()
	page.Expect(page.GetByID("ocr-meta")).ToBeVisible()
	page.Expect(page.GetByID("ocr-output")).ToBeVisible()
	page.Expect(page.GetByID("btn-run-ocr")).ToBeVisible()
	page.Expect(page.GetByID("btn-save-ocr")).ToBeVisible()

	// Switch to Palette tab (5)
	page.Press("5")
	page.Expect(page.GetByID("palette-list")).ToBeVisible()
	page.Expect(page.GetByID("palette-count")).ToBeVisible()
	page.Expect(page.GetByID("btn-extract-src")).ToBeVisible()
	page.Expect(page.GetByID("btn-lighten")).ToBeVisible()

	// Switch to Files tab (6)
	page.Press("6")
	page.Expect(page.GetByID("file-picker")).ToBeVisible()
	page.Expect(page.GetByID("btn-drop")).ToBeVisible()

	// Open Drag & Drop Image Zone ('d') and close with Esc
	page.Press("d")
	page.Expect(page.GetByID("drop-input")).ToBeVisible()
	page.Expect(page.GetByID("drop-zone-hint")).ToBeVisible()
	page.Expect(page.GetByID("btn-drop-load")).ToBeVisible()
	page.Expect(page.GetByID("btn-drop-cancel")).ToBeVisible()
	page.Press("esc")
}

func TestStudioDragAndDrop(t *testing.T) {
	tmpOut := t.TempDir()
	s := NewStudio(Options{SplitPreview: true})
	s.wakeup = nil
	config.GowallConfig.OutputFolder = tmpOut

	// Save a real PNG to disk with spaces in the filename to test drag-and-drop parsing
	s.SaveCurrentOutput()
	entries, err := os.ReadDir(tmpOut)
	if err != nil || len(entries) == 0 {
		t.Fatalf("expected saved sample image in %s", tmpOut)
	}
	srcFile := filepath.Join(tmpOut, entries[0].Name())
	spacedFile := filepath.Join(tmpOut, "my dropped wallpaper.png")
	data, err := os.ReadFile(srcFile)
	if err != nil {
		t.Fatalf("read saved sample: %v", err)
	}
	if err := os.WriteFile(spacedFile, data, 0644); err != nil {
		t.Fatalf("write spaced sample: %v", err)
	}

	// 1. Test ParseDroppedPaths with single-quoted path, file:// URI, and backslash-escaped path
	if got := ParseDroppedPaths("'" + spacedFile + "' "); len(got) != 1 || got[0] != spacedFile {
		t.Fatalf("single-quoted drop parse failed: got %v, want [%s]", got, spacedFile)
	}
	uriPayload := "file://" + filepath.ToSlash(tmpOut) + "/my%20dropped%20wallpaper.png\r\n"
	if got := ParseDroppedPaths(uriPayload); len(got) != 1 || got[0] != spacedFile {
		t.Fatalf("file:// URI drop parse failed: got %v, want [%s]", got, spacedFile)
	}
	escapedPayload := filepath.Join(tmpOut, `my\ dropped\ wallpaper.png`)
	if got := ParseDroppedPaths(escapedPayload); len(got) != 1 || got[0] != spacedFile {
		t.Fatalf("backslash-escaped drop parse failed: got %v, want [%s]", got, spacedFile)
	}

	// 2. Test terminal Bracketed Paste drag-and-drop event (driver.EventPaste)
	ev := driver.Event{
		Type:  driver.EventPaste,
		Paste: driver.PasteEvent{Text: "'" + spacedFile + "'"},
	}
	s.HandleEvent(nil, &ev)
	if s.fileName != "my dropped wallpaper.png" {
		t.Fatalf("expected Studio to load dropped file via EventPaste, got fileName=%q", s.fileName)
	}
}

func TestStudioPipelineAndSave(t *testing.T) {
	tmpOut := t.TempDir()
	s := NewStudio(Options{SplitPreview: true})
	s.wakeup = nil
	config.GowallConfig.OutputFolder = tmpOut

	// Select an effect (Invert) and save output
	s.onSelectEffect(1) // Invert Colors
	s.SaveCurrentOutput()

	entries, err := os.ReadDir(tmpOut)
	if err != nil {
		t.Fatalf("failed reading temp output dir: %v", err)
	}
	if len(entries) == 0 {
		t.Fatalf("expected saved output file in %s, got none", tmpOut)
	}
	savedPath := filepath.Join(tmpOut, entries[0].Name())
	info, err := os.Stat(savedPath)
	if err != nil || info.Size() == 0 {
		t.Fatalf("expected non-empty saved image at %s", savedPath)
	}
}

func TestStudioBgRmOCRAndPalette(t *testing.T) {
	tmpOut := t.TempDir()
	s := NewStudio(Options{SplitPreview: true})
	s.wakeup = nil
	config.GowallConfig.OutputFolder = tmpOut

	// 1. Background Removal (K-Means + Checkerboard preview)
	s.onSelectBgMode(1) // K-Means Cutout (Alpha Grid)
	if s.previewImg == nil {
		t.Fatalf("expected preview image after K-Means background removal")
	}

	// Verify U2-Net and Bria-RMBG AI Cutout selections cleanly fallback without prompting or hanging
	s.onSelectBgMode(7) // U2-Net AI Cutout (ONNX)
	if s.previewImg == nil {
		t.Fatalf("expected preview image after U2-Net fallback")
	}
	s.onSelectBgMode(8) // Bria-RMBG AI Cutout (ONNX)
	if s.previewImg == nil {
		t.Fatalf("expected preview image after Bria-RMBG fallback")
	}

	// 2. Palette Lighten + HSL formatting
	if len(s.extractedItems) == 0 {
		t.Fatalf("expected non-empty extracted palette items with Hex/RGB/HSL")
	}
	initialLine := s.extractedItems[0]
	s.LightenSelectedColor()
	afterLighten := s.extractedItems[0]
	if afterLighten == initialLine {
		t.Fatalf("expected LightenSelectedColor to transform extracted colors")
	}

	// 3. OCR Output Save
	s.ocrText = "# Extracted OCR Sample\nGowall Studio TUI"
	s.SaveOCROutput()
	entries, err := os.ReadDir(tmpOut)
	if err != nil || len(entries) == 0 {
		t.Fatalf("expected saved OCR markdown file in %s", tmpOut)
	}
}

func BenchmarkStudioDraw(b *testing.B) {
	s := NewStudio(Options{SplitPreview: true})
	s.wakeup = nil

	buf := buffer.NewBuffer(cell.NewRect(0, 0, 120, 40))
	focus := terminal.NewFocusManager()
	f := terminal.NewFrame(buf, focus)

	// Warm up one pass per tab so internal widget buffers are sized
	for tab := TabThemes; tab <= TabFiles; tab++ {
		s.activeTab = tab
		s.tabs.Selected = tab
		s.syncStrings()
		f.Reset()
		s.Draw(f, nil)
	}
	s.activeTab = TabThemes
	s.tabs.Selected = TabThemes
	s.syncStrings()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		f.Reset()
		s.Draw(f, nil)
	}
}
