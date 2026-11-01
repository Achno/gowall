package tui

import (
	"context"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/Achno/gowall/config"
	bgremoval "github.com/Achno/gowall/internal/backends/bgRemoval"
	cpkg "github.com/Achno/gowall/internal/backends/color"
	"github.com/Achno/gowall/internal/backends/colorthief"
	imgbackend "github.com/Achno/gowall/internal/backends/image"
	gwimage "github.com/Achno/gowall/internal/image"
	imageio "github.com/Achno/gowall/internal/image_io"
	"github.com/Achno/gowall/internal/logger"
	"github.com/Achno/gowall/internal/providers"
	"github.com/Achno/gowall/internal/types"
	"github.com/disintegration/imaging"
	"github.com/thebanri/limoni"
	"github.com/thebanri/limoni/core/cell"
	"github.com/thebanri/limoni/core/driver"
	"github.com/thebanri/limoni/graphics"
	"github.com/thebanri/limoni/widgets"
)

const (
	TabThemes = iota
	TabEffects
	TabBgRm
	TabOCR
	TabPalette
	TabFiles
	numTabs
)

const (
	EffectNone = iota
	EffectInvert
	EffectAsciiColor
	EffectAsciiMono
	EffectPixelate
	EffectGrayscale
	EffectMirror
	EffectFlip
	EffectBrightness
	EffectContrast
	EffectSaturation
	EffectGrid
	EffectBorder
	EffectTilt
	EffectCropAspect
	EffectGradient
	EffectCompress
)

type effectDef struct {
	Name       string
	Kind       int
	HasParam   bool
	ParamLabel string
	Min        int
	Max        int
	Default    int
}

var studioEffects = []effectDef{
	{Name: "(None / Original)", Kind: EffectNone},
	{Name: "Invert Colors", Kind: EffectInvert},
	{Name: "ASCII Art (Full Color)", Kind: EffectAsciiColor, HasParam: true, ParamLabel: "ASCII Detail (%)", Min: 15, Max: 85, Default: 45},
	{Name: "ASCII Art (Monochrome)", Kind: EffectAsciiMono, HasParam: true, ParamLabel: "ASCII Detail (%)", Min: 15, Max: 85, Default: 45},
	{Name: "Pixelate (Block Art)", Kind: EffectPixelate, HasParam: true, ParamLabel: "Block Scale (%)", Min: 4, Max: 40, Default: 12},
	{Name: "Grayscale", Kind: EffectGrayscale},
	{Name: "Mirror (Symmetry)", Kind: EffectMirror},
	{Name: "Flip Horizontal", Kind: EffectFlip},
	{Name: "Brightness", Kind: EffectBrightness, HasParam: true, ParamLabel: "Brightness (%)", Min: 30, Max: 200, Default: 130},
	{Name: "Contrast", Kind: EffectContrast, HasParam: true, ParamLabel: "Contrast (-50..+80)", Min: -50, Max: 80, Default: 30},
	{Name: "Saturation", Kind: EffectSaturation, HasParam: true, ParamLabel: "Saturation (-80..+80)", Min: -80, Max: 80, Default: 45},
	{Name: "Draw Grid Overlay", Kind: EffectGrid, HasParam: true, ParamLabel: "Grid Size (px)", Min: 24, Max: 160, Default: 64},
	{Name: "Rounded Border", Kind: EffectBorder, HasParam: true, ParamLabel: "Corner Radius (px)", Min: 10, Max: 80, Default: 36},
	{Name: "3D Tilt Showcase", Kind: EffectTilt, HasParam: true, ParamLabel: "Preset (1..4)", Min: 1, Max: 4, Default: 1},
	{Name: "Aspect Ratio Crop", Kind: EffectCropAspect, HasParam: true, ParamLabel: "Preset (1:16x9 2:1x1 3:9x16 4:21x9)", Min: 1, Max: 4, Default: 2},
	{Name: "Palette Gradient Gen", Kind: EffectGradient, HasParam: true, ParamLabel: "Gradient Angle (°)", Min: 0, Max: 315, Default: 45},
	{Name: "Compress (PNGQuant)", Kind: EffectCompress, HasParam: true, ParamLabel: "Quality (10..100)", Min: 10, Max: 100, Default: 75},
}

const (
	BgModeOff = iota
	BgModeKMeansChecker
	BgModeKMeansTransparent
	BgModeKMeansMocha
	BgModeKMeansMauve
	BgModeKMeansTeal
	BgModeKMeansPeach
	BgModeU2Net
	BgModeBria
)

type bgModeDef struct {
	Name   string
	Mode   int
	Method string
	Fill   color.RGBA
	HasClr bool
}

var studioBgModes = []bgModeDef{
	{Name: "(Keep Original Background)", Mode: BgModeOff},
	{Name: "K-Means Cutout (Alpha Grid)", Mode: BgModeKMeansChecker, Method: "kmeans"},
	{Name: "K-Means Cutout (Transparent)", Mode: BgModeKMeansTransparent, Method: "kmeans"},
	{Name: "K-Means + Mocha Fill (#1E1E2E)", Mode: BgModeKMeansMocha, Method: "kmeans", Fill: color.RGBA{R: 30, G: 30, B: 46, A: 255}, HasClr: true},
	{Name: "K-Means + Mauve Fill (#CBA6F7)", Mode: BgModeKMeansMauve, Method: "kmeans", Fill: color.RGBA{R: 203, G: 166, B: 247, A: 255}, HasClr: true},
	{Name: "K-Means + Teal Fill (#94E2D5)", Mode: BgModeKMeansTeal, Method: "kmeans", Fill: color.RGBA{R: 148, G: 226, B: 213, A: 255}, HasClr: true},
	{Name: "K-Means + Peach Fill (#FAB387)", Mode: BgModeKMeansPeach, Method: "kmeans", Fill: color.RGBA{R: 250, G: 179, B: 135, A: 255}, HasClr: true},
	{Name: "U2-Net AI Cutout (ONNX)", Mode: BgModeU2Net, Method: "u2net"},
	{Name: "Bria-RMBG AI Cutout (ONNX)", Mode: BgModeBria, Method: "bria-rmbg"},
}

type ocrPresetDef struct {
	Label    string
	Provider string
	Model    string
	Language string
}

var studioOCRPresets = []ocrPresetDef{
	{Label: "Tesseract OCR (eng — Local)", Provider: "tesseract", Model: "tesseract", Language: "eng"},
	{Label: "Tesseract OCR (tur+eng)", Provider: "tesseract", Model: "tesseract", Language: "tur+eng"},
	{Label: "Ollama Vision (llava:latest)", Provider: "ollama", Model: "llava:latest"},
	{Label: "Ollama Vision (qwen2.5vl)", Provider: "ollama", Model: "qwen2.5vl"},
	{Label: "Gemini Vision (gemini-2.0-flash)", Provider: "gemini", Model: "gemini-2.0-flash"},
	{Label: "OpenAI Vision (gpt-4o-mini)", Provider: "openai", Model: "gpt-4o-mini"},
	{Label: "Mistral OCR (mistral-ocr-latest)", Provider: "mistral", Model: "mistral-ocr-latest"},
	{Label: "Docling Document Parser", Provider: "docling", Model: "docling"},
}

// Options configures the Gowall Studio TUI.
type Options struct {
	InitialPath  string
	SplitPreview bool
	SocketPath   string
}

// Studio is the interactive Gowall TUI application state.
// All widgets and layout structures are kept on the struct so Draw allocates 0 B/op.
type Studio struct {
	mu     sync.Mutex
	wakeup func()

	activeTab    int
	pendingTab   int
	splitPreview bool
	nnBackend    bool
	extractOrig  bool

	// Loaded images
	filePath   string
	fileName   string
	origImg    image.Image // Full resolution original from disk
	baseImg    image.Image // Working base (allows chaining theme + effects + bg removal)
	previewSrc image.Image // Downscaled copy (max 960x540) for real-time interactive preview
	previewImg image.Image // Processed preview image

	// Theme state
	themeNames      []string
	themeItems      []string
	selectedTheme   int
	selectedEffect  int
	selectedBgMode  int
	selectedOCR     int
	themeSwatches   []SwatchEntry
	extractedHex    []string
	extractedItems  []string
	extractSwatches []SwatchEntry

	// Effect & BgRm parameter values
	effectParams []int
	bgIterations int

	// OCR state
	ocrText      string
	ocrRawText   string
	ocrMeta      string
	ocrRunning   atomic.Bool
	ocrLastDurMS int64
	ocrRanFile   string

	// Async worker state
	jobGen    atomic.Uint64
	busy      atomic.Bool
	lastDurMS int64

	// Pre-formatted strings for zero-allocation Draw
	headerTitle    string
	origBlockTitle string
	prevBlockTitle string
	ocrBlockTitle  string
	paramTitle     string
	paletteTitle   string
	bgParamTitle   string
	statusMsg      string
	infoLine       string
	protoLabel     string

	// Persistent widgets and states
	tabsState      widgets.TabsState
	tabs           StudioTabs
	themeListState *widgets.ListState
	themeList      widgets.List
	effectState    *widgets.ListState
	effectList     widgets.List
	bgListState    *widgets.ListState
	bgList         widgets.List
	ocrListState   *widgets.ListState
	ocrList        widgets.List
	paletteState   *widgets.ListState
	paletteList    widgets.List
	fileState      *widgets.FilePickerState
	filePicker     widgets.FilePicker
	paramSliderSt  *widgets.SliderState
	paramSlider    widgets.Slider
	bgSliderSt     *widgets.SliderState
	bgSlider       widgets.Slider
	palSliderSt    *widgets.SliderState
	palSlider      widgets.Slider
	paletteStateW  PaletteBar
	bottomSwatches PaletteBar

	headerBlock   widgets.Block
	headerPara    widgets.Paragraph
	sidebarBlock  widgets.Block
	swatchBlock   widgets.Block
	paramBlock    widgets.Block
	ocrMetaBlock  widgets.Block
	ocrMetaPara   widgets.Paragraph
	actionsBlock  widgets.Block
	origBlock     widgets.Block
	prevBlock     widgets.Block
	ocrBlock      widgets.Block
	ocrTextPara   widgets.Paragraph
	infoBlock     widgets.Block
	infoPara      widgets.Paragraph
	origImageW    ImageView
	prevImageW    ImageView
	btnSave       widgets.Button
	btnSplit      widgets.Button
	btnBackend    widgets.Button
	btnChain      widgets.Button
	btnReset      widgets.Button
	btnExtractSrc widgets.Button
	btnRunOCR     widgets.Button
	btnSaveOCR    widgets.Button
	btnLighten    widgets.Button
	btnDropOpen   widgets.Button
	statusBar     widgets.StatusBar
	statusNode    widgets.Paragraph

	// Drag & Drop Image Zone state and widgets
	dropZoneOpen   bool
	lastDropInfo   string
	dropInputSt    *widgets.TextInputState
	dropInput      widgets.TextInput
	dropModalBlock widgets.Block
	dropZoneBlock  widgets.Block
	dropHintPara   widgets.Paragraph
	dropInputBlock widgets.Block
	dropStatusPara widgets.Paragraph
	btnDropLoad    widgets.Button
	btnDropCancel  widgets.Button

	cmdPaletteSt *widgets.CommandPaletteState
	cmdPalette   widgets.CommandPalette
}

// NewStudio initializes a ready-to-render Studio instance.
func NewStudio(opts Options) *Studio {
	config.LoadConfig()
	gwimage.LoadCustomThemes()

	rawThemes := gwimage.ListThemes()
	slices.Sort(rawThemes)

	themeNames := make([]string, 0, len(rawThemes)+1)
	themeItems := make([]string, 0, len(rawThemes)+1)
	themeNames = append(themeNames, "")
	themeItems = append(themeItems, "(Original / None)")
	for _, t := range rawThemes {
		themeNames = append(themeNames, t)
		themeItems = append(themeItems, t)
	}

	effItems := make([]string, len(studioEffects))
	effParams := make([]int, len(studioEffects))
	for i, e := range studioEffects {
		effItems[i] = e.Name
		effParams[i] = e.Default
	}

	bgItems := make([]string, len(studioBgModes))
	for i, b := range studioBgModes {
		name := b.Name
		if (b.Method == "u2net" || b.Method == "bria-rmbg") && !gwimage.IsBgStrategyReady(b.Method) {
			name = b.Name + " (not installed)"
		}
		bgItems[i] = name
	}

	ocrItems := make([]string, len(studioOCRPresets))
	for i, o := range studioOCRPresets {
		ocrItems[i] = o.Label
	}

	s := &Studio{
		activeTab:    TabThemes,
		pendingTab:   -1,
		splitPreview: opts.SplitPreview,
		nnBackend:    config.GowallConfig.ColorCorrectionBackend == "nn",
		themeNames:   themeNames,
		themeItems:   themeItems,
		effectParams: effParams,
		bgIterations: 80,
		ocrText:      "Press Enter or 'o' to extract text from the current image using Gowall OCR.\n\nSupported engines:\n • Local Tesseract CLI (eng, tur+eng)\n • Local Vision LLMs via Ollama (llava, qwen2.5vl)\n • Cloud VLMs (Gemini, OpenAI, Mistral OCR)\n • Docling Document Layout Parser",
		ocrMeta:      "Engine: Tesseract (eng)\nStatus: Idle — press 'o' to run",
		wakeup:       limoni.Wakeup,
		protoLabel:   "Auto",
	}

	s.themeListState = widgets.NewListState()
	s.themeListState.Selected = 0
	s.effectState = widgets.NewListState()
	s.effectState.Selected = 0
	s.bgListState = widgets.NewListState()
	s.bgListState.Selected = 0
	s.ocrListState = widgets.NewListState()
	s.ocrListState.Selected = 0
	s.paletteState = widgets.NewListState()
	s.paletteState.Selected = 0

	startDir := "."
	var initialFile string
	if opts.InitialPath != "" {
		if info, err := os.Stat(opts.InitialPath); err == nil {
			if info.IsDir() {
				startDir = opts.InitialPath
			} else {
				startDir = filepath.Dir(opts.InitialPath)
				initialFile = opts.InitialPath
			}
		}
	} else if info, err := os.Stat("assets"); err == nil && info.IsDir() {
		startDir = "assets"
	} else if config.GowallConfig.OutputFolder != "" {
		if info, err := os.Stat(config.GowallConfig.OutputFolder); err == nil && info.IsDir() {
			startDir = config.GowallConfig.OutputFolder
		}
	}

	s.fileState = widgets.NewFilePickerState(startDir)
	s.fileState.Extensions = []string{".png", ".jpg", ".jpeg", ".webp", ".avif", ".gif"}
	_ = s.fileState.Load(startDir)

	s.paramSliderSt = widgets.NewSliderState(45)
	s.bgSliderSt = widgets.NewSliderState(80)
	s.palSliderSt = widgets.NewSliderState(8)

	// Configure persistent widgets and callbacks once so Draw allocates nothing
	s.tabs = StudioTabs{
		Tabs: widgets.Tabs{
			ID:       "mode-tabs",
			Titles:   []string{"1:Themes", "2:Effects", "3:BgRm", "4:OCR", "5:Palette", "6:Files"},
			Selected: 0,
			State:    &s.tabsState,
			Style:    cell.Style{Fg: cell.NewColorRGB(180, 190, 254)},
			SelectedStyle: cell.Style{
				Fg:       cell.NewColorRGB(17, 17, 27),
				Bg:       cell.NewColorRGB(203, 166, 247),
				Modifier: cell.ModifierBold,
			},
		},
		Pending: &s.pendingTab,
	}

	s.themeList = widgets.List{
		ID:               "theme-list",
		Label:            "Themes",
		Items:            s.themeItems,
		State:            s.themeListState,
		Scrollbar:        true,
		HighlightSymbol:  "▸ ",
		HighlightSpacing: true,
		SelectedStyle: cell.Style{
			Fg:       cell.NewColorRGB(203, 166, 247),
			Modifier: cell.ModifierBold,
		},
	}

	s.effectList = widgets.List{
		ID:               "effect-list",
		Label:            "Effects",
		Items:            effItems,
		State:            s.effectState,
		Scrollbar:        true,
		HighlightSymbol:  "▸ ",
		HighlightSpacing: true,
		SelectedStyle: cell.Style{
			Fg:       cell.NewColorRGB(137, 180, 250),
			Modifier: cell.ModifierBold,
		},
	}

	s.bgList = widgets.List{
		ID:               "bg-list",
		Label:            "Background Removal",
		Items:            bgItems,
		State:            s.bgListState,
		Scrollbar:        true,
		HighlightSymbol:  "▸ ",
		HighlightSpacing: true,
		SelectedStyle: cell.Style{
			Fg:       cell.NewColorRGB(148, 226, 213),
			Modifier: cell.ModifierBold,
		},
	}

	s.ocrList = widgets.List{
		ID:               "ocr-list",
		Label:            "OCR Engines",
		Items:            ocrItems,
		State:            s.ocrListState,
		Scrollbar:        true,
		HighlightSymbol:  "▸ ",
		HighlightSpacing: true,
		SelectedStyle: cell.Style{
			Fg:       cell.NewColorRGB(250, 179, 135),
			Modifier: cell.ModifierBold,
		},
	}

	s.paletteList = widgets.List{
		ID:               "palette-list",
		Label:            "Extracted Colors",
		State:            s.paletteState,
		Scrollbar:        true,
		HighlightSymbol:  "▸ ",
		HighlightSpacing: true,
		SelectedStyle: cell.Style{
			Fg:       cell.NewColorRGB(249, 226, 175),
			Modifier: cell.ModifierBold,
		},
	}

	s.filePicker = widgets.FilePicker{
		ID:    "file-picker",
		State: s.fileState,
		DirStyle: cell.Style{
			Fg:       cell.NewColorRGB(137, 180, 250),
			Modifier: cell.ModifierBold,
		},
		SelectedStyle: cell.Style{
			Fg:       cell.NewColorRGB(17, 17, 27),
			Bg:       cell.NewColorRGB(166, 227, 161),
			Modifier: cell.ModifierBold,
		},
	}

	s.paramSlider = widgets.Slider{
		ID:          "effect-param",
		State:       s.paramSliderSt,
		Min:         1,
		Max:         100,
		FilledStyle: cell.Style{Fg: cell.NewColorRGB(137, 180, 250)},
		ThumbStyle:  cell.Style{Fg: cell.NewColorRGB(245, 194, 231), Modifier: cell.ModifierBold},
		OnChange:    s.onEffectParamChange,
	}

	s.bgSlider = widgets.Slider{
		ID:          "bg-iterations",
		State:       s.bgSliderSt,
		Min:         20,
		Max:         200,
		FilledStyle: cell.Style{Fg: cell.NewColorRGB(148, 226, 213)},
		ThumbStyle:  cell.Style{Fg: cell.NewColorRGB(166, 227, 161), Modifier: cell.ModifierBold},
		OnChange:    s.onBgIterChange,
	}

	s.palSlider = widgets.Slider{
		ID:          "palette-count",
		State:       s.palSliderSt,
		Min:         4,
		Max:         16,
		FilledStyle: cell.Style{Fg: cell.NewColorRGB(249, 226, 175)},
		ThumbStyle:  cell.Style{Fg: cell.NewColorRGB(250, 179, 135), Modifier: cell.ModifierBold},
		OnChange:    s.onPaletteCountChange,
	}

	s.paletteStateW = PaletteBar{
		ID:    "theme-swatches",
		Label: "Theme Swatches",
	}
	s.bottomSwatches = PaletteBar{
		ID:      "active-swatches",
		Label:   "Active Palette",
		Compact: false,
	}

	s.origImageW = ImageView{
		Image: widgets.Image{
			ID:     "img-original",
			ZIndex: -3,
		},
		Label: "Original Image",
	}
	s.prevImageW = ImageView{
		Image: widgets.Image{
			ID:     "img-preview",
			ZIndex: -3,
		},
		Label: "Processed Preview",
	}

	s.btnSave = widgets.Button{
		ID:      "btn-save",
		Label:   "Save (s)",
		OnPress: s.SaveCurrentOutput,
		Style:   cell.Style{Fg: cell.NewColorRGB(166, 227, 161), Modifier: cell.ModifierBold},
	}
	s.btnSplit = widgets.Button{
		ID:      "btn-split",
		Label:   "Split (v)",
		OnPress: s.ToggleSplitView,
		Style:   cell.Style{Fg: cell.NewColorRGB(137, 180, 250)},
	}
	s.btnBackend = widgets.Button{
		ID:      "btn-backend",
		Label:   "CLUT/NN (b)",
		OnPress: s.ToggleBackend,
		Style:   cell.Style{Fg: cell.NewColorRGB(249, 226, 175)},
	}
	s.btnChain = widgets.Button{
		ID:      "btn-chain",
		Label:   "Chain (a)",
		OnPress: s.ChainPreviewToBase,
		Style:   cell.Style{Fg: cell.NewColorRGB(203, 166, 247)},
	}
	s.btnReset = widgets.Button{
		ID:      "btn-reset",
		Label:   "Reset (r)",
		OnPress: s.ResetToOriginal,
		Style:   cell.Style{Fg: cell.NewColorRGB(243, 139, 168)},
	}
	s.btnExtractSrc = widgets.Button{
		ID:      "btn-extract-src",
		Label:   "Src:Prev (e)",
		OnPress: s.ToggleExtractSource,
		Style:   cell.Style{Fg: cell.NewColorRGB(249, 226, 175)},
	}
	s.btnRunOCR = widgets.Button{
		ID:      "btn-run-ocr",
		Label:   "Run OCR (o)",
		OnPress: s.RunOCR,
		Style:   cell.Style{Fg: cell.NewColorRGB(250, 179, 135), Modifier: cell.ModifierBold},
	}
	s.btnSaveOCR = widgets.Button{
		ID:      "btn-save-ocr",
		Label:   "Save .md (s)",
		OnPress: s.SaveOCROutput,
		Style:   cell.Style{Fg: cell.NewColorRGB(166, 227, 161)},
	}
	s.btnLighten = widgets.Button{
		ID:      "btn-lighten",
		Label:   "Lighten (l)",
		OnPress: s.LightenSelectedColor,
		Style:   cell.Style{Fg: cell.NewColorRGB(203, 166, 247)},
	}
	s.btnDropOpen = widgets.Button{
		ID:      "btn-drop",
		Label:   "Drop (d)",
		OnPress: s.ToggleDropZone,
		Style:   cell.Style{Fg: cell.NewColorRGB(148, 226, 213), Modifier: cell.ModifierBold},
	}

	s.dropInputSt = widgets.NewTextInputState()
	s.dropInput = widgets.TextInput{
		ID:               "drop-input",
		State:            s.dropInputSt,
		Placeholder:      "Drop file/folder here or paste path / file:// URI...",
		Style:            cell.Style{Fg: cell.NewColorRGB(205, 214, 244), Bg: cell.NewColorRGB(24, 24, 37)},
		FocusedStyle:     cell.Style{Fg: cell.NewColorRGB(166, 227, 161), Bg: cell.NewColorRGB(24, 24, 37), Modifier: cell.ModifierBold},
		PlaceholderStyle: cell.Style{Fg: cell.NewColorRGB(108, 112, 134), Bg: cell.NewColorRGB(24, 24, 37)},
		Focused:          true,
		SelectionStart:   -1,
		SelectionEnd:     -1,
	}
	s.dropModalBlock = widgets.Block{
		Title:         " 📥 Drag & Drop Image — Gowall Studio ",
		Borders:       limoni.BorderAll,
		BorderSymbols: limoni.SymbolsRounded,
		BorderStyle:   cell.Style{Fg: cell.NewColorRGB(148, 226, 213), Bg: cell.NewColorRGB(17, 17, 27)},
		TitleStyle:    cell.Style{Fg: cell.NewColorRGB(148, 226, 213), Bg: cell.NewColorRGB(17, 17, 27), Modifier: cell.ModifierBold},
		Style:         cell.Style{Fg: cell.NewColorRGB(205, 214, 244), Bg: cell.NewColorRGB(17, 17, 27)},
		Padding:       widgets.Insets{Top: 0, Right: 2, Bottom: 0, Left: 2},
	}
	s.dropZoneBlock = widgets.Block{
		Title:         " Drop Target (Drag from File Manager Anywhere on Terminal) ",
		Borders:       limoni.BorderAll,
		BorderSymbols: limoni.SymbolsRounded,
		BorderStyle:   cell.Style{Fg: cell.NewColorRGB(203, 166, 247), Bg: cell.NewColorRGB(24, 24, 37)},
		TitleStyle:    cell.Style{Fg: cell.NewColorRGB(245, 194, 231), Bg: cell.NewColorRGB(24, 24, 37), Modifier: cell.ModifierBold},
		Style:         cell.Style{Fg: cell.NewColorRGB(205, 214, 244), Bg: cell.NewColorRGB(24, 24, 37)},
		Padding:       widgets.Insets{Top: 0, Right: 1, Bottom: 0, Left: 1},
	}
	s.dropHintPara = widgets.Paragraph{
		ID: "drop-zone-hint",
		Text: "• Drag & drop any image file or folder directly onto the terminal window\n" +
			"• Supports: PNG, JPG, WebP, GIF, BMP, TIFF, SVG, ICO  •  file:// URIs  •  Quoted paths",
		Style: cell.Style{Fg: cell.NewColorRGB(180, 190, 254), Bg: cell.NewColorRGB(24, 24, 37)},
		Wrap:  true,
	}
	s.dropInputBlock = widgets.Block{
		Title:         " Dropped Path / file:// URI ",
		Borders:       limoni.BorderAll,
		BorderSymbols: limoni.SymbolsRounded,
		BorderStyle:   cell.Style{Fg: cell.NewColorRGB(137, 180, 250), Bg: cell.NewColorRGB(24, 24, 37)},
		TitleStyle:    cell.Style{Fg: cell.NewColorRGB(137, 180, 250), Bg: cell.NewColorRGB(24, 24, 37), Modifier: cell.ModifierBold},
		Style:         cell.Style{Fg: cell.NewColorRGB(205, 214, 244), Bg: cell.NewColorRGB(24, 24, 37)},
		Padding:       widgets.Insets{Top: 0, Right: 1, Bottom: 0, Left: 1},
	}
	s.dropStatusPara = widgets.Paragraph{
		ID:    "drop-zone-status",
		Text:  "Ready — drop a file from your file manager or enter a path and press Enter",
		Style: cell.Style{Fg: cell.NewColorRGB(249, 226, 175), Bg: cell.NewColorRGB(17, 17, 27)},
	}
	s.btnDropLoad = widgets.Button{
		ID:      "btn-drop-load",
		Label:   "Load Image (Enter)",
		OnPress: func() { _ = s.SubmitDropInput() },
		Style:   cell.Style{Fg: cell.NewColorRGB(166, 227, 161), Bg: cell.NewColorRGB(17, 17, 27), Modifier: cell.ModifierBold},
	}
	s.btnDropCancel = widgets.Button{
		ID:      "btn-drop-cancel",
		Label:   "Cancel (Esc)",
		OnPress: s.CloseDropZone,
		Style:   cell.Style{Fg: cell.NewColorRGB(243, 139, 168), Bg: cell.NewColorRGB(17, 17, 27)},
	}

	borderStyle := cell.Style{Fg: cell.NewColorRGB(88, 91, 112)}
	titleStyle := cell.Style{Fg: cell.NewColorRGB(205, 214, 244), Modifier: cell.ModifierBold}

	s.headerBlock = widgets.Block{
		Borders:       limoni.BorderAll,
		BorderSymbols: limoni.SymbolsRounded,
		BorderStyle:   cell.Style{Fg: cell.NewColorRGB(203, 166, 247)},
		TitleStyle:    titleStyle,
	}
	s.headerPara = widgets.Paragraph{
		ID:    "studio-header",
		Style: cell.Style{Fg: cell.NewColorRGB(205, 214, 244), Modifier: cell.ModifierBold},
	}
	s.sidebarBlock = widgets.Block{
		Borders:       limoni.BorderAll,
		BorderSymbols: limoni.SymbolsRounded,
		BorderStyle:   borderStyle,
		TitleStyle:    titleStyle,
	}
	s.swatchBlock = widgets.Block{
		Title:         " Theme Palette ",
		Borders:       limoni.BorderAll,
		BorderSymbols: limoni.SymbolsRounded,
		BorderStyle:   borderStyle,
		TitleStyle:    titleStyle,
	}
	s.paramBlock = widgets.Block{
		Borders:       limoni.BorderAll,
		BorderSymbols: limoni.SymbolsRounded,
		BorderStyle:   borderStyle,
		TitleStyle:    titleStyle,
	}
	s.ocrMetaBlock = widgets.Block{
		Title:         " OCR Engine Status ",
		Borders:       limoni.BorderAll,
		BorderSymbols: limoni.SymbolsRounded,
		BorderStyle:   borderStyle,
		TitleStyle:    titleStyle,
	}
	s.ocrMetaPara = widgets.Paragraph{
		ID:    "ocr-meta",
		Style: cell.Style{Fg: cell.NewColorRGB(249, 226, 175)},
		Wrap:  true,
	}
	s.actionsBlock = widgets.Block{
		Title:         " Actions ",
		Borders:       limoni.BorderAll,
		BorderSymbols: limoni.SymbolsRounded,
		BorderStyle:   borderStyle,
		TitleStyle:    titleStyle,
	}
	s.origBlock = widgets.Block{
		Borders:       limoni.BorderAll,
		BorderSymbols: limoni.SymbolsRounded,
		BorderStyle:   borderStyle,
		TitleStyle:    cell.Style{Fg: cell.NewColorRGB(148, 226, 213), Modifier: cell.ModifierBold},
	}
	s.prevBlock = widgets.Block{
		Borders:       limoni.BorderAll,
		BorderSymbols: limoni.SymbolsRounded,
		BorderStyle:   cell.Style{Fg: cell.NewColorRGB(203, 166, 247)},
		TitleStyle:    cell.Style{Fg: cell.NewColorRGB(203, 166, 247), Modifier: cell.ModifierBold},
	}
	s.ocrBlock = widgets.Block{
		Borders:       limoni.BorderAll,
		BorderSymbols: limoni.SymbolsRounded,
		BorderStyle:   cell.Style{Fg: cell.NewColorRGB(250, 179, 135)},
		TitleStyle:    cell.Style{Fg: cell.NewColorRGB(250, 179, 135), Modifier: cell.ModifierBold},
		Padding:       widgets.Insets{Top: 1, Right: 2, Bottom: 1, Left: 2},
	}
	s.ocrTextPara = widgets.Paragraph{
		ID:    "ocr-output",
		Style: cell.Style{Fg: cell.NewColorRGB(205, 214, 244)},
		Wrap:  true,
	}
	s.infoBlock = widgets.Block{
		Borders:       limoni.BorderAll,
		BorderSymbols: limoni.SymbolsRounded,
		BorderStyle:   borderStyle,
		TitleStyle:    titleStyle,
	}
	s.infoPara = widgets.Paragraph{
		ID:    "pipeline-info",
		Style: cell.Style{Fg: cell.NewColorRGB(166, 173, 200)},
	}
	s.statusNode = widgets.Paragraph{
		ID: "status-message",
	}

	s.statusBar = widgets.StatusBar{
		Style:    cell.Style{Fg: cell.NewColorRGB(205, 214, 244), Bg: cell.NewColorRGB(24, 24, 37)},
		KeyStyle: cell.Style{Fg: cell.NewColorRGB(203, 166, 247), Bg: cell.NewColorRGB(24, 24, 37), Modifier: cell.ModifierBold},
		Left: []widgets.StatusItem{
			{Key: "1-6/Tab", Text: "Mode"},
			{Key: "↑/↓", Text: "Select"},
			{Key: "d", Text: "Drop"},
			{Key: "v", Text: "Split"},
			{Key: "b", Text: "CLUT/NN"},
			{Key: "o", Text: "OCR"},
			{Key: "a", Text: "Chain"},
			{Key: "s", Text: "Save"},
			{Key: "/", Text: "Search"},
			{Key: "q", Text: "Quit"},
		},
		Center: []widgets.StatusItem{{Text: "Ready"}},
		Right:  []widgets.StatusItem{{Text: "Gowall Studio"}},
	}

	s.initCommandPalette()

	if initialFile == "" {
		initialFile = s.firstImageInPicker()
	}
	if initialFile != "" {
		_ = s.LoadImageFile(initialFile)
	} else {
		s.loadFallbackSample()
	}
	s.syncStrings()
	return s
}

func (s *Studio) initCommandPalette() {
	s.cmdPaletteSt = widgets.NewCommandPaletteState()
	items := make([]widgets.CommandItem, 0, len(s.themeNames)+len(studioEffects)+len(studioBgModes)+16)
	items = append(items,
		widgets.CommandItem{Label: "Open Drag & Drop Image Zone", Detail: "d", Category: "File", Handler: s.OpenDropZone},
		widgets.CommandItem{Label: "Run OCR Text Extraction", Detail: "o", Category: "OCR", Handler: func() {
			s.onSelectTab(TabOCR)
			s.RunOCR()
		}},
		widgets.CommandItem{Label: "Save Processed Output", Detail: "s", Category: "Action", Handler: s.SaveCurrentOutput},
		widgets.CommandItem{Label: "Toggle Split / Single Preview", Detail: "v", Category: "View", Handler: s.ToggleSplitView},
		widgets.CommandItem{Label: "Toggle HaldCLUT / Nearest-Neighbour Backend", Detail: "b", Category: "Engine", Handler: s.ToggleBackend},
		widgets.CommandItem{Label: "Chain Current Preview to Base Image", Detail: "a", Category: "Pipeline", Handler: s.ChainPreviewToBase},
		widgets.CommandItem{Label: "Reset to Original Image", Detail: "r", Category: "Pipeline", Handler: s.ResetToOriginal},
		widgets.CommandItem{Label: "Switch to 1:Themes Tab", Detail: "1", Category: "Navigation", Handler: func() { s.onSelectTab(TabThemes) }},
		widgets.CommandItem{Label: "Switch to 2:Effects Tab", Detail: "2", Category: "Navigation", Handler: func() { s.onSelectTab(TabEffects) }},
		widgets.CommandItem{Label: "Switch to 3:BgRm Tab (Background Removal)", Detail: "3", Category: "Navigation", Handler: func() { s.onSelectTab(TabBgRm) }},
		widgets.CommandItem{Label: "Switch to 4:OCR Tab (Text Extraction)", Detail: "4", Category: "Navigation", Handler: func() { s.onSelectTab(TabOCR) }},
		widgets.CommandItem{Label: "Switch to 5:Palette Tab", Detail: "5", Category: "Navigation", Handler: func() { s.onSelectTab(TabPalette) }},
		widgets.CommandItem{Label: "Switch to 6:Files Tab", Detail: "6", Category: "Navigation", Handler: func() { s.onSelectTab(TabFiles) }},
	)
	for i, bg := range studioBgModes {
		idx := i
		items = append(items, widgets.CommandItem{
			Label:    "BgRm: " + bg.Name,
			Category: "BgRm",
			Handler: func() {
				s.onSelectTab(TabBgRm)
				s.onSelectBgMode(idx)
			},
		})
	}
	for i, eff := range studioEffects {
		idx := i
		items = append(items, widgets.CommandItem{
			Label:    "Effect: " + eff.Name,
			Category: "Effect",
			Handler: func() {
				s.onSelectTab(TabEffects)
				s.onSelectEffect(idx)
			},
		})
	}
	for i, name := range s.themeItems {
		idx := i
		items = append(items, widgets.CommandItem{
			Label:    "Theme: " + name,
			Category: "Theme",
			Handler: func() {
				s.onSelectTab(TabThemes)
				s.onSelectTheme(idx)
			},
		})
	}
	s.cmdPaletteSt.AllItems = items
	s.cmdPaletteSt.Filtered = items
	s.cmdPalette = widgets.CommandPalette{
		ID:          "command-palette",
		State:       s.cmdPaletteSt,
		Title:       " Gowall Command Palette ",
		Placeholder: "Search themes, effects, background removal, OCR, or studio actions...",
	}
}

func (s *Studio) firstImageInPicker() string {
	for _, e := range s.fileState.Entries {
		if !e.IsDir {
			return filepath.Join(s.fileState.Dir, e.Name)
		}
	}
	return ""
}

func (s *Studio) loadFallbackSample() {
	w, h := 480, 270
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			r := uint8((x * 255) / w)
			g := uint8((y * 220) / h)
			b := uint8(180 - (x*100)/w + (y*60)/h)
			img.SetRGBA(x, y, color.RGBA{R: r, G: g, B: b, A: 255})
		}
	}
	s.filePath = "sample_wallpaper.png"
	s.fileName = "sample_wallpaper.png"
	s.setBaseImages(img)
	s.setStatus("Loaded built-in sample canvas (use 6:Files to open any image)")
}

// LoadImageFile loads an image from disk and updates the preview pipeline.
func (s *Studio) LoadImageFile(path string) error {
	abs, err := filepath.Abs(path)
	if err != nil {
		abs = path
	}
	img, err := imageio.LoadImage(imageio.FileReader{Path: abs})
	if err != nil {
		s.setStatus("Failed to load " + filepath.Base(path) + ": " + err.Error())
		return err
	}
	s.filePath = abs
	s.fileName = filepath.Base(abs)
	s.setBaseImages(img)
	s.setStatus("Loaded " + s.fileName)
	s.triggerPreview()
	if s.activeTab == TabOCR {
		s.RunOCR()
	}
	return nil
}

// OpenDropZone opens the interactive Drag & Drop Image Zone modal.
func (s *Studio) OpenDropZone() {
	s.cmdPaletteSt.Close()
	s.dropZoneOpen = true
	s.dropInputSt.SetValue("")
	s.syncDropStatus()
	s.setStatus("Drag & Drop Zone open — drop any image onto the terminal or enter a path")
	s.syncStrings()
}

// CloseDropZone closes the Drag & Drop Image Zone modal.
func (s *Studio) CloseDropZone() {
	s.dropZoneOpen = false
	s.syncStrings()
}

// ToggleDropZone toggles the Drag & Drop Image Zone modal.
func (s *Studio) ToggleDropZone() {
	if s.dropZoneOpen {
		s.CloseDropZone()
	} else {
		s.OpenDropZone()
	}
}

// HandleDropPayload parses a terminal drag-and-drop payload (from bracketed
// paste, file:// URI, or shell-quoted path) and loads the dropped image or folder.
func (s *Studio) HandleDropPayload(raw string) bool {
	paths := ParseDroppedPaths(raw)
	if len(paths) == 0 {
		return false
	}
	first := paths[0]
	info, err := os.Stat(first)
	if err != nil {
		return false
	}
	s.dropZoneOpen = false
	s.cmdPaletteSt.Close()
	s.dropInputSt.SetValue("")

	if info.IsDir() {
		_ = s.fileState.Load(first)
		if firstImg := s.firstImageInPicker(); firstImg != "" {
			_ = s.LoadImageFile(firstImg)
		}
		s.lastDropInfo = "Dropped folder: " + filepath.Base(first)
		s.setStatus("📥 Dropped folder loaded: " + filepath.Base(first))
		s.syncStrings()
		return true
	}

	dir := filepath.Dir(first)
	base := filepath.Base(first)
	_ = s.fileState.Load(dir)
	for i, e := range s.fileState.Entries {
		if e.Name == base {
			s.fileState.Selected = i
			break
		}
	}
	if err := s.LoadImageFile(first); err != nil {
		return false
	}
	w, h := 0, 0
	if s.origImg != nil {
		b := s.origImg.Bounds()
		w, h = b.Dx(), b.Dy()
	}
	s.lastDropInfo = fmt.Sprintf("Last dropped: %s (%d×%d)", base, w, h)
	s.setStatus(fmt.Sprintf("📥 Dropped & Loaded: %s (%d×%d)", base, w, h))
	s.syncStrings()
	return true
}

// SubmitDropInput attempts to load the path currently entered in the Drop Zone input.
func (s *Studio) SubmitDropInput() bool {
	val := strings.TrimSpace(s.dropInputSt.Value())
	if s.HandleDropPayload(val) {
		return true
	}
	if val == "" {
		s.dropStatusPara.Text = "⚠ Please drop an image file onto the terminal or enter a valid image path"
	} else {
		s.dropStatusPara.Text = "⚠ No valid image file found at: " + val
	}
	s.setStatus("Drop path not found: " + val)
	s.syncStrings()
	return false
}

func (s *Studio) syncDropStatus() {
	val := strings.TrimSpace(s.dropInputSt.Value())
	if val == "" {
		if s.lastDropInfo != "" {
			s.dropStatusPara.Text = "✓ " + s.lastDropInfo + "  │  Drop another file or press Esc"
		} else {
			s.dropStatusPara.Text = "Ready — drop a file from your file manager or enter a path and press Enter"
		}
		return
	}
	if paths := ParseDroppedPaths(val); len(paths) > 0 {
		s.dropStatusPara.Text = "✓ Valid image detected: " + filepath.Base(paths[0]) + " — press Enter to load!"
	} else {
		s.dropStatusPara.Text = "Path: " + val + " (press Enter when complete)"
	}
}

func (s *Studio) setBaseImages(img image.Image) {
	if s.origImg != nil {
		graphics.ForgetImage(s.origImg)
	}
	if s.previewImg != nil {
		graphics.ForgetImage(s.previewImg)
	}
	s.origImg = img
	s.baseImg = img
	s.previewSrc = makePreviewSource(img, 960, 540)
	s.previewImg = s.previewSrc
	s.origImageW.Img = s.previewSrc
	s.prevImageW.Img = s.previewImg
	s.updateExtractedPalette(s.previewImg)
	s.syncStrings()
}

func makePreviewSource(img image.Image, maxW, maxH int) image.Image {
	if img == nil {
		return nil
	}
	b := img.Bounds()
	if b.Dx() <= maxW && b.Dy() <= maxH {
		return img
	}
	return imaging.Fit(img, maxW, maxH, imaging.Linear)
}

func (s *Studio) onSelectTab(index int) {
	if index < 0 || index >= numTabs {
		return
	}
	s.activeTab = index
	s.tabs.Selected = index
	s.syncStrings()
	if index == TabOCR && s.ocrRanFile != s.fileName && !s.ocrRunning.Load() {
		s.RunOCR()
	}
}

func (s *Studio) onSelectTheme(index int) {
	if index < 0 || index >= len(s.themeNames) {
		return
	}
	s.selectedTheme = index
	s.themeListState.Selected = index
	themeName := s.themeNames[index]
	if themeName == "" {
		s.themeSwatches = nil
		s.paletteStateW.Swatches = nil
		s.paletteStateW.Summary = "Original colors"
	} else if hexes, err := gwimage.GetThemeColors(themeName); err == nil {
		s.themeSwatches = BuildSwatches(hexes)
		s.paletteStateW.Swatches = s.themeSwatches
		s.paletteStateW.Summary = strings.Join(hexes, " ")
	}
	s.triggerPreview()
}

func (s *Studio) onSelectEffect(index int) {
	if index < 0 || index >= len(studioEffects) {
		return
	}
	s.selectedEffect = index
	s.effectState.Selected = index
	eff := studioEffects[index]
	if eff.HasParam {
		s.paramSlider.Min = eff.Min
		s.paramSlider.Max = eff.Max
		s.paramSliderSt.Set(s.effectParams[index], eff.Min, eff.Max)
	}
	s.syncStrings()
	s.triggerPreview()
}

func (s *Studio) onSelectBgMode(index int) {
	if index < 0 || index >= len(studioBgModes) {
		return
	}
	s.selectedBgMode = index
	s.bgListState.Selected = index
	bg := studioBgModes[index]
	if (bg.Method == "u2net" || bg.Method == "bria-rmbg") && !gwimage.IsBgStrategyReady(bg.Method) {
		s.setStatus(bg.Name + ": Model not installed (falling back to K-Means). Run 'gowall bg -m " + bg.Method + "' to download.")
	} else {
		s.setStatus("Background removal: " + bg.Name)
	}
	s.syncStrings()
	s.triggerPreview()
}

func (s *Studio) onSelectOCREngine(index int) {
	if index < 0 || index >= len(studioOCRPresets) {
		return
	}
	s.selectedOCR = index
	s.ocrListState.Selected = index
	preset := studioOCRPresets[index]
	s.ocrMeta = fmt.Sprintf("Provider: %s\nModel: %s\nStatus: Ready ('o' to run)", preset.Provider, preset.Model)
	s.syncStrings()
	if preset.Provider == "tesseract" {
		s.RunOCR()
	}
}

func (s *Studio) onEffectParamChange(val int) {
	idx := s.effectState.Selected
	if idx >= 0 && idx < len(studioEffects) {
		s.effectParams[idx] = val
		s.syncStrings()
		s.triggerPreview()
	}
}

func (s *Studio) onBgIterChange(val int) {
	s.bgIterations = val
	s.syncStrings()
	if s.selectedBgMode > 0 {
		s.triggerPreview()
	}
}

func (s *Studio) onPaletteCountChange(val int) {
	if s.extractOrig {
		s.updateExtractedPalette(s.previewSrc)
	} else {
		s.updateExtractedPalette(s.previewImg)
	}
	s.syncStrings()
}

// LightenSelectedColor lightens the currently selected extracted palette color (`gowall color light`).
func (s *Studio) LightenSelectedColor() {
	idx := s.paletteState.Selected
	if idx < 0 || idx >= len(s.extractedHex) {
		return
	}
	hex := s.extractedHex[idx]
	lighter, err := cpkg.LightenColor(hex, 0.20)
	if err != nil {
		return
	}
	rgba, err := cpkg.HexToRGBA(lighter)
	if err != nil {
		return
	}
	hsl, _, _ := cpkg.ConvertHexToFormat(lighter, "hsl")
	hslCompact := strings.ReplaceAll(hsl, " ", "")
	s.extractedHex[idx] = lighter
	s.extractedItems[idx] = fmt.Sprintf("%s %3d,%3d,%3d %s", strings.ToUpper(lighter), rgba.R, rgba.G, rgba.B, hslCompact)
	s.extractSwatches = BuildSwatches(s.extractedHex)
	s.setStatus("Lightened palette color -> " + strings.ToUpper(lighter) + " (" + hslCompact + ")")
	s.syncStrings()
}

// RunOCR executes Gowall's OCR pipeline on the current working image asynchronously.
func (s *Studio) RunOCR() {
	if s.ocrRunning.Load() {
		return
	}
	src := s.origImg
	if src == nil {
		src = s.previewImg
	}
	if src == nil {
		return
	}

	presetIdx := max(0, min(s.ocrListState.Selected, len(studioOCRPresets)-1))
	preset := studioOCRPresets[presetIdx]
	fileName := s.fileName

	s.ocrRunning.Store(true)
	s.ocrMeta = fmt.Sprintf("Provider: %s (%s)\nStatus: Extracting text...", preset.Provider, preset.Model)
	s.setStatus("Running OCR (" + preset.Label + ") on " + fileName + "...")
	s.syncStrings()

	go func() {
		t0 := time.Now()
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		var cfg providers.Config
		cfg.OCR.Provider = preset.Provider
		cfg.OCR.Model = preset.Model
		cfg.OCR.Language = preset.Language

		var extracted string
		var runErr error

		prov, err := providers.NewOCRProvider(cfg)
		if err != nil {
			runErr = err
		} else {
			svc := providers.NewProviderService(prov, cfg)
			res, err := svc.OCR(ctx, providers.OCRInput{
				Type:     providers.InputTypeImage,
				Image:    src,
				Filename: fileName,
			})
			if err != nil {
				runErr = err
			} else if res != nil {
				extracted = strings.TrimSpace(res.Text)
			}
		}

		dur := time.Since(t0).Milliseconds()
		s.mu.Lock()
		s.ocrRunning.Store(false)
		s.ocrLastDurMS = dur
		s.ocrRanFile = fileName
		if runErr != nil {
			s.ocrText = fmt.Sprintf("OCR Provider Notice (%s / %s):\n%v\n\nTip: Select 'Tesseract OCR (eng — Local)' in the left sidebar for offline OCR, or configure your API key / local Ollama server for VLM providers.",
				preset.Provider, preset.Model, runErr)
			s.ocrMeta = fmt.Sprintf("Provider: %s\nStatus: Error (%d ms)", preset.Provider, dur)
			s.statusMsg = "OCR notice: " + runErr.Error()
		} else if extracted == "" {
			s.ocrText = "(No text regions detected in " + fileName + ")\n\nTry loading an image with text/code in 6:Files, or switch OCR language/provider."
			s.ocrMeta = fmt.Sprintf("Provider: %s (%d ms)\nResult: 0 words", preset.Provider, dur)
			s.statusMsg = fmt.Sprintf("OCR finished (%d ms): no text detected", dur)
		} else {
			words := len(strings.Fields(extracted))
			lines := strings.Count(extracted, "\n") + 1
			langInfo := preset.Model
			if preset.Language != "" {
				langInfo = preset.Language
			}
			s.ocrRawText = extracted
			s.ocrText = fmt.Sprintf(
				"# OCR Result — %s\nProvider: %s (%s) • %d ms • %d lines, %d words\n──────────────────────────────────────────\n\n%s\n\n──────────────────────────────────────────\n[o/Enter] Re-run OCR   [c] Copy Clipboard   [s] Save .md   [v] Full/Split",
				fileName, preset.Provider, langInfo, dur, lines, words, extracted,
			)
			s.ocrMeta = fmt.Sprintf("Provider: %s (%d ms)\nExtracted: %d lines, %d words", preset.Provider, dur, lines, words)
			s.statusMsg = fmt.Sprintf("OCR extracted %d words (%d lines) in %d ms", words, lines, dur)
		}
		s.syncStringsLocked()
		s.mu.Unlock()
		if s.wakeup != nil {
			s.wakeup()
		}
	}()
}

// CopyToClipboard copies the extracted OCR text (on 4:OCR) or selected palette Hex color (on 5:Palette) via OSC 52.
func (s *Studio) CopyToClipboard() {
	if s.activeTab == TabOCR {
		text := strings.TrimSpace(s.ocrRawText)
		if text == "" {
			text = strings.TrimSpace(s.ocrText)
		}
		if text == "" {
			return
		}
		_ = driver.WriteClipboard(os.Stdout, text)
		words := len(strings.Fields(text))
		lines := strings.Count(text, "\n") + 1
		s.setStatus(fmt.Sprintf("📋 Copied OCR text to clipboard (%d lines, %d words)", lines, words))
		s.syncStrings()
		return
	}
	idx := s.paletteState.Selected
	if idx >= 0 && idx < len(s.extractedHex) {
		hex := s.extractedHex[idx]
		_ = driver.WriteClipboard(os.Stdout, hex)
		s.setStatus("📋 Copied color to clipboard: " + hex)
		s.syncStrings()
	}
}

// SaveOCROutput saves the extracted OCR text as a Markdown file in the Gowall output folder.
func (s *Studio) SaveOCROutput() {
	outDir := config.GowallConfig.OutputFolder
	if outDir == "" {
		home, _ := os.UserHomeDir()
		outDir = filepath.Join(home, "Pictures", "gowall")
	}
	_ = os.MkdirAll(outDir, 0755)
	base := strings.TrimSuffix(s.fileName, filepath.Ext(s.fileName))
	if base == "" {
		base = "document"
	}
	outPath := filepath.Join(outDir, base+"_ocr.md")
	content := strings.TrimSpace(s.ocrRawText)
	if content == "" {
		content = strings.TrimSpace(s.ocrText)
	}
	if err := os.WriteFile(outPath, []byte(content+"\n"), 0644); err != nil {
		s.setStatus("Failed saving OCR markdown: " + err.Error())
	} else {
		s.setStatus("💾 Saved OCR Markdown -> " + outPath)
	}
	s.syncStrings()
}

// ToggleSplitView switches between side-by-side (Original + Preview) and single Preview mode.
func (s *Studio) ToggleSplitView() {
	s.splitPreview = !s.splitPreview
	if s.splitPreview {
		s.setStatus("View mode: Side-by-Side (Original vs Preview)")
	} else {
		s.setStatus("View mode: Single Full Preview")
	}
	s.syncStrings()
}

// ToggleBackend switches between HaldCLUT (RBF) and Nearest-Neighbour theme conversion.
func (s *Studio) ToggleBackend() {
	s.nnBackend = !s.nnBackend
	if s.nnBackend {
		config.GowallConfig.ColorCorrectionBackend = "nn"
		s.setStatus("Color backend: Nearest-Neighbour (Fast)")
	} else {
		config.GowallConfig.ColorCorrectionBackend = "clut"
		s.setStatus("Color backend: HaldCLUT (Smooth RBF)")
	}
	s.syncStrings()
	if s.selectedTheme > 0 {
		s.triggerPreview()
	}
}

// ChainPreviewToBase promotes the current processed preview to be the new working base image.
func (s *Studio) ChainPreviewToBase() {
	s.mu.Lock()
	cur := s.previewImg
	s.mu.Unlock()
	if cur == nil {
		return
	}
	fullOut, err := s.applyPipeline(s.baseImg, false)
	if err == nil && fullOut != nil {
		s.baseImg = fullOut
	}
	s.previewSrc = cur
	s.selectedTheme = 0
	s.selectedEffect = 0
	s.selectedBgMode = 0
	s.themeListState.Selected = 0
	s.effectState.Selected = 0
	s.bgListState.Selected = 0
	s.setStatus("Chained current output as new base layer!")
	s.syncStrings()
}

// ResetToOriginal discards chained edits and resets theme/effect/bg selection to Original.
func (s *Studio) ResetToOriginal() {
	if s.origImg == nil {
		return
	}
	s.baseImg = s.origImg
	s.previewSrc = makePreviewSource(s.origImg, 960, 540)
	if s.previewImg != nil && s.previewImg != s.previewSrc {
		graphics.ForgetImage(s.previewImg)
	}
	s.previewImg = s.previewSrc
	s.prevImageW.Img = s.previewImg
	s.selectedTheme = 0
	s.selectedEffect = 0
	s.selectedBgMode = 0
	s.themeListState.Selected = 0
	s.effectState.Selected = 0
	s.bgListState.Selected = 0
	s.themeSwatches = nil
	s.paletteStateW.Swatches = nil
	s.updateExtractedPalette(s.previewImg)
	s.setStatus("Reset to original image")
	s.syncStrings()
}

// ToggleExtractSource toggles palette extraction between the processed preview and original image.
func (s *Studio) ToggleExtractSource() {
	s.extractOrig = !s.extractOrig
	if s.extractOrig {
		s.btnExtractSrc.Label = "Src:Orig (e)"
		s.updateExtractedPalette(s.previewSrc)
		s.setStatus("Extracted palette from Original image")
	} else {
		s.btnExtractSrc.Label = "Src:Prev (e)"
		s.updateExtractedPalette(s.previewImg)
		s.setStatus("Extracted palette from Processed image")
	}
	s.syncStrings()
}

// SaveCurrentOutput processes the full-resolution image and saves it to the gowall output folder.
func (s *Studio) SaveCurrentOutput() {
	if s.activeTab == TabOCR {
		s.SaveOCROutput()
		return
	}
	if s.baseImg == nil {
		return
	}
	outDir := config.GowallConfig.OutputFolder
	if outDir == "" {
		home, _ := os.UserHomeDir()
		outDir = filepath.Join(home, "Pictures", "gowall")
	}
	_ = os.MkdirAll(outDir, 0755)

	ext := filepath.Ext(s.fileName)
	if ext == "" || s.selectedBgMode > 0 {
		ext = ".png"
	}
	base := strings.TrimSuffix(s.fileName, filepath.Ext(s.fileName))
	suffix := s.activeOperationSuffix()
	outName := fmt.Sprintf("%s_%s%s", base, suffix, ext)
	outPath := filepath.Join(outDir, outName)

	s.setStatus("Saving full-resolution image to " + outName + "...")
	s.syncStrings()

	imgToSave, err := s.applyPipeline(s.baseImg, true)
	if err != nil || imgToSave == nil {
		imgToSave = s.previewImg
	}
	format := strings.TrimPrefix(strings.ToLower(ext), ".")
	if format == "" {
		format = "png"
	}
	if err := imageio.SaveImage(imgToSave, imageio.FileWriter{Path: outPath}, format, types.ImageMetadata{}); err != nil {
		outPath = filepath.Join(outDir, base+"_"+suffix+".png")
		if err2 := imageio.SaveImage(imgToSave, imageio.FileWriter{Path: outPath}, "png", types.ImageMetadata{}); err2 != nil {
			s.setStatus("Save error: " + err2.Error())
			s.syncStrings()
			return
		}
	}
	s.setStatus("Saved -> " + outPath)
	s.syncStrings()
}

func (s *Studio) activeOperationSuffix() string {
	var parts []string
	if s.selectedTheme > 0 && s.selectedTheme < len(s.themeNames) {
		parts = append(parts, s.themeNames[s.selectedTheme])
	}
	if s.selectedBgMode > 0 && s.selectedBgMode < len(studioBgModes) {
		parts = append(parts, "nobg")
	}
	if s.effectState.Selected > 0 && s.effectState.Selected < len(studioEffects) {
		eff := studioEffects[s.effectState.Selected]
		tag := strings.ToLower(strings.Fields(eff.Name)[0])
		parts = append(parts, tag)
	}
	if len(parts) == 0 {
		return "edited"
	}
	return strings.Join(parts, "_")
}

func (s *Studio) triggerPreview() {
	src := s.previewSrc
	if src == nil {
		return
	}
	gen := s.jobGen.Add(1)
	s.busy.Store(true)
	s.syncStrings()

	themeIdx := s.selectedTheme
	effIdx := s.effectState.Selected
	var paramVal int
	if effIdx >= 0 && effIdx < len(s.effectParams) {
		paramVal = s.effectParams[effIdx]
	}
	bgIdx := s.selectedBgMode
	bgIter := s.bgIterations
	nn := s.nnBackend

	go func() {
		t0 := time.Now()
		out, err := s.computeImage(src, themeIdx, effIdx, paramVal, bgIdx, bgIter, nn, false)
		if s.jobGen.Load() != gen {
			return
		}
		dur := time.Since(t0).Milliseconds()
		s.mu.Lock()
		if err == nil && out != nil {
			if s.previewImg != nil && s.previewImg != s.previewSrc {
				graphics.ForgetImage(s.previewImg)
			}
			s.previewImg = out
			s.prevImageW.Img = out
			s.lastDurMS = dur
			if !s.extractOrig {
				s.updateExtractedPaletteLocked(out)
			}
		} else if err != nil {
			s.statusMsg = "Pipeline notice: " + err.Error()
		}
		s.busy.Store(false)
		s.syncStringsLocked()
		s.mu.Unlock()
		if s.wakeup != nil {
			s.wakeup()
		}
	}()
}

func (s *Studio) applyPipeline(src image.Image, forExport bool) (image.Image, error) {
	effIdx := s.effectState.Selected
	paramVal := 0
	if effIdx >= 0 && effIdx < len(s.effectParams) {
		paramVal = s.effectParams[effIdx]
	}
	return s.computeImage(src, s.selectedTheme, effIdx, paramVal, s.selectedBgMode, s.bgIterations, s.nnBackend, forExport)
}

func (s *Studio) computeImage(src image.Image, themeIdx, effIdx, paramVal, bgIdx, bgIter int, nn bool, forExport bool) (image.Image, error) {
	cur := src

	// 1. Background Removal (`gowall bg`)
	if bgIdx > 0 && bgIdx < len(studioBgModes) {
		bgDef := studioBgModes[bgIdx]
		var strategy bgremoval.BgRemovalStrategy
		var cleanup func() error
		if bgDef.Method == "u2net" || bgDef.Method == "bria-rmbg" {
			if gwimage.IsBgStrategyReady(bgDef.Method) {
				st, cl, err := gwimage.GetBgStrategy(bgDef.Method, bgIter, 0.001, 0.5, 4)
				if err == nil {
					strategy = st
					cleanup = cl
				}
			}
			if strategy == nil {
				// Fallback to K-Means if ONNX runtime/model is not downloaded yet
				strategy = bgremoval.NewKMeansStrategy(bgremoval.KMeansOptions{
					MaxIter:     max(20, bgIter),
					Convergence: 0.001,
					SampleRate:  0.6,
					NumRoutines: 4,
				})
			}
		} else {
			strategy = bgremoval.NewKMeansStrategy(bgremoval.KMeansOptions{
				MaxIter:     max(20, bgIter),
				Convergence: 0.001,
				SampleRate:  0.6,
				NumRoutines: 4,
			})
		}
		if cleanup != nil {
			defer cleanup()
		}
		var fillClr color.Color
		if bgDef.HasClr {
			fillClr = bgDef.Fill
		}
		proc := gwimage.NewBackgroundProcessor(strategy, fillClr)
		cutout, _, err := proc.Process(cur, "", "png")
		if err != nil {
			return nil, err
		}
		if bgDef.Mode == BgModeKMeansChecker && !forExport {
			cur = compositeAlphaCheckerboard(cutout)
		} else {
			cur = cutout
		}
	}

	// 2. Theme Conversion (`gowall convert`)
	if themeIdx > 0 && themeIdx < len(s.themeNames) {
		themeName := s.themeNames[themeIdx]
		if nn {
			tm, err := gwimage.SelectTheme(themeName)
			if err != nil {
				return nil, err
			}
			next, err := gwimage.NearestNeighbour(cur, tm)
			if err != nil {
				return nil, err
			}
			cur = next
		} else {
			conv := &gwimage.ThemeConverter{}
			next, _, err := conv.Process(cur, themeName, "png")
			if err != nil {
				return nil, err
			}
			cur = next
		}
	}

	// 3. Effects, ASCII, Crop, Gradient, Compression
	if effIdx > 0 && effIdx < len(studioEffects) {
		eff := studioEffects[effIdx]
		var proc gwimage.ImageProcessor
		switch eff.Kind {
		case EffectInvert:
			proc = &gwimage.Inverter{}
		case EffectAsciiColor:
			proc = gwimage.NewAsciiProcessor(
				gwimage.WithAsciiDetail(max(15, min(paramVal, 90))),
				gwimage.WithAsciiMode(imgbackend.AsciiColor),
				gwimage.WithAsciiBackground(color.RGBA{R: 24, G: 24, B: 37, A: 255}),
			)
		case EffectAsciiMono:
			proc = gwimage.NewAsciiProcessor(
				gwimage.WithAsciiDetail(max(15, min(paramVal, 90))),
				gwimage.WithAsciiMode(imgbackend.AsciiWhite),
				gwimage.WithAsciiBackground(color.RGBA{R: 17, G: 17, B: 27, A: 255}),
			)
		case EffectPixelate:
			proc = &gwimage.PixelateProcessor{Scale: float64(max(paramVal, 2))}
		case EffectGrayscale:
			proc = &gwimage.GrayScaleProcessor{}
		case EffectMirror:
			proc = &gwimage.MirrorProcessor{}
		case EffectFlip:
			proc = &gwimage.FlipProcessor{}
		case EffectBrightness:
			proc = &gwimage.BrightnessProcessor{Factor: float64(max(paramVal, 10)) / 100.0}
		case EffectContrast:
			proc = &gwimage.ContrastProcessor{Mode: gwimage.ContrastModeNormal, Factor: float64(paramVal)}
		case EffectSaturation:
			proc = &gwimage.SaturationProcessor{Percentage: float64(paramVal)}
		case EffectGrid:
			gp := &gwimage.GridProcessor{}
			gp.SetGridOptions(gwimage.WithGridSize(max(paramVal, 16)), gwimage.WithGridColor("#CBA6F7"), gwimage.WithGridThickness(1))
			proc = gp
		case EffectBorder:
			proc = &gwimage.BorderProcessor{
				Color:           color.RGBA{R: 203, G: 166, B: 247, A: 255},
				BorderThickness: 6,
				CornerRadius:    float64(max(paramVal, 4)),
			}
		case EffectTilt:
			presetName := fmt.Sprintf("p%d", max(1, min(paramVal, 4)))
			preset, _ := gwimage.GetTiltPreset(presetName)
			proc = &gwimage.TiltProcessor{Preset: preset}
		case EffectCropAspect:
			proc = &gwimage.CropProcessor{Rect: aspectCropRect(cur.Bounds(), paramVal)}
		case EffectGradient:
			gradColors := s.activeGradientHexes()
			gp := &gwimage.GradientProcessor{}
			b := cur.Bounds()
			gp.SetOptions(
				gwimage.WithColors(gradColors),
				gwimage.WithGradientWidth(max(320, b.Dx())),
				gwimage.WithGradientHeight(max(180, b.Dy())),
				gwimage.WithAngle(float64(paramVal)),
				gwimage.WithGradientMethod("hcl"),
			)
			proc = gp
		case EffectCompress:
			proc = gwimage.NewCompressionProcessor(
				gwimage.WithQuality(max(10, min(paramVal, 100))),
				gwimage.WithSpeed(4),
			)
		}
		if proc != nil {
			next, _, err := proc.Process(cur, "", "png")
			if err != nil {
				return nil, err
			}
			if next != nil {
				cur = next
			}
		}
	}
	return cur, nil
}

func aspectCropRect(b image.Rectangle, preset int) image.Rectangle {
	w, h := b.Dx(), b.Dy()
	targetW, targetH := w, h
	switch preset {
	case 1: // 16:9
		if w*9 > h*16 {
			targetW = (h * 16) / 9
		} else {
			targetH = (w * 9) / 16
		}
	case 2: // 1:1 Square
		side := min(w, h)
		targetW, targetH = side, side
	case 3: // 9:16 Vertical
		if w*16 > h*9 {
			targetW = (h * 9) / 16
		} else {
			targetH = (w * 16) / 9
		}
	case 4: // 21:9 Ultrawide
		if w*9 > h*21 {
			targetW = (h * 21) / 9
		} else {
			targetH = (w * 9) / 21
		}
	}
	x0 := (w - targetW) / 2
	y0 := (h - targetH) / 2
	return image.Rect(x0, y0, x0+targetW, y0+targetH)
}

func (s *Studio) activeGradientHexes() []string {
	if s.selectedTheme > 0 && s.selectedTheme < len(s.themeNames) {
		if hexes, err := gwimage.GetThemeColors(s.themeNames[s.selectedTheme]); err == nil && len(hexes) >= 2 {
			if len(hexes) > 5 {
				return hexes[:5]
			}
			return hexes
		}
	}
	if len(s.extractedHex) >= 2 {
		if len(s.extractedHex) > 4 {
			return s.extractedHex[:4]
		}
		return s.extractedHex
	}
	return []string{"#1E1E2E", "#89B4FA", "#CBA6F7", "#FAB387"}
}

func compositeAlphaCheckerboard(img image.Image) image.Image {
	b := img.Bounds()
	dst := image.NewRGBA(b)
	c1 := color.RGBA{R: 30, G: 30, B: 46, A: 255}
	c2 := color.RGBA{R: 69, G: 71, B: 90, A: 255}
	tile := 14
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			if ((x/tile)+(y/tile))%2 == 0 {
				dst.SetRGBA(x, y, c1)
			} else {
				dst.SetRGBA(x, y, c2)
			}
		}
	}
	draw.Draw(dst, b, img, b.Min, draw.Over)
	return dst
}

func (s *Studio) updateExtractedPalette(img image.Image) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.updateExtractedPaletteLocked(img)
}

func (s *Studio) updateExtractedPaletteLocked(img image.Image) {
	if img == nil {
		return
	}
	count := max(4, min(s.palSliderSt.Value, 16))
	clrs, err := colorthief.GetPalette(img, count)
	if err != nil {
		return
	}
	s.extractSwatches = BuildSwatchesFromColors(clrs)
	s.extractedHex = s.extractedHex[:0]
	s.extractedItems = s.extractedItems[:0]
	for _, c := range clrs {
		rgba := cpkg.RGBA8(c)
		hex := cpkg.RGBtoHex(rgba)
		hsl, _, _ := cpkg.ConvertHexToFormat(hex, "hsl")
		hslCompact := strings.ReplaceAll(hsl, " ", "")
		s.extractedHex = append(s.extractedHex, hex)
		s.extractedItems = append(s.extractedItems, fmt.Sprintf("%s %3d,%3d,%3d %s", hex, rgba.R, rgba.G, rgba.B, hslCompact))
	}
	s.paletteList.Items = s.extractedItems
}

func (s *Studio) setStatus(msg string) {
	s.statusMsg = msg
}

func (s *Studio) syncStrings() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.syncStringsLocked()
}

func (s *Studio) syncStringsLocked() {
	w, h := 0, 0
	if s.origImg != nil {
		b := s.origImg.Bounds()
		w, h = b.Dx(), b.Dy()
	}
	backendName := "HaldCLUT"
	if s.nnBackend {
		backendName = "Nearest-NB"
	}

	s.headerTitle = fmt.Sprintf(" ❖ GOWALL STUDIO │ %s (%d×%d) │ %s │ %s",
		s.fileName, w, h, backendName, s.protoLabel)
	s.headerPara.Text = s.headerTitle

	s.origBlockTitle = fmt.Sprintf(" Original — %s (%d×%d) ", s.fileName, w, h)
	s.origBlock.Title = s.origBlockTitle
	s.origImageW.Value = s.fileName

	var parts []string
	if s.selectedBgMode > 0 && s.selectedBgMode < len(studioBgModes) {
		parts = append(parts, studioBgModes[s.selectedBgMode].Name)
	}
	if s.selectedTheme > 0 && s.selectedTheme < len(s.themeNames) {
		parts = append(parts, "Theme: "+s.themeNames[s.selectedTheme])
	}
	if s.effectState.Selected > 0 && s.effectState.Selected < len(studioEffects) {
		parts = append(parts, studioEffects[s.effectState.Selected].Name)
	}
	activeDesc := "Original"
	if len(parts) > 0 {
		activeDesc = strings.Join(parts, " + ")
	}
	if s.busy.Load() {
		activeDesc += " [working...]"
	}
	s.prevBlockTitle = fmt.Sprintf(" Preview — %s (%d ms) ", activeDesc, s.lastDurMS)
	s.prevBlock.Title = s.prevBlockTitle
	s.prevImageW.Value = activeDesc

	presetIdx := max(0, min(s.ocrListState.Selected, len(studioOCRPresets)-1))
	preset := studioOCRPresets[presetIdx]
	ocrTag := preset.Provider
	if preset.Language != "" {
		ocrTag = preset.Provider + "/" + preset.Language
	}
	s.ocrBlockTitle = fmt.Sprintf(" OCR Output — %s (%d ms) ", ocrTag, s.ocrLastDurMS)
	s.ocrBlock.Title = s.ocrBlockTitle
	s.ocrTextPara.Text = s.ocrText
	s.ocrMetaPara.Text = s.ocrMeta

	switch s.activeTab {
	case TabThemes:
		s.sidebarBlock.Title = fmt.Sprintf(" Themes (%d) ", len(s.themeItems)-1)
	case TabEffects:
		s.sidebarBlock.Title = fmt.Sprintf(" Effects & Tools (%d) ", len(studioEffects)-1)
		eff := studioEffects[max(0, min(s.effectState.Selected, len(studioEffects)-1))]
		if eff.HasParam {
			s.paramTitle = fmt.Sprintf(" %s: %d ", eff.ParamLabel, s.effectParams[s.effectState.Selected])
		} else {
			s.paramTitle = " Parameter (N/A) "
		}
		s.paramBlock.Title = s.paramTitle
	case TabBgRm:
		s.sidebarBlock.Title = " Background Removal (bg) "
		s.bgParamTitle = fmt.Sprintf(" K-Means Iterations: %d ", s.bgIterations)
		s.paramBlock.Title = s.bgParamTitle
	case TabOCR:
		s.sidebarBlock.Title = " OCR Providers (ocr) "
	case TabPalette:
		s.sidebarBlock.Title = fmt.Sprintf(" Extracted Palette (%d) ", len(s.extractedItems))
		s.paletteTitle = fmt.Sprintf(" Dominant Colors: %d ", s.palSliderSt.Value)
		s.paramBlock.Title = s.paletteTitle
	case TabFiles:
		s.sidebarBlock.Title = " Image Browser "
	}

	if len(s.themeSwatches) > 0 && s.activeTab == TabThemes {
		s.bottomSwatches.Swatches = s.themeSwatches
		s.bottomSwatches.Summary = "Theme palette"
	} else {
		s.bottomSwatches.Swatches = s.extractSwatches
		s.bottomSwatches.Summary = strings.Join(s.extractedHex, " ")
	}

	viewMode := "Split (Before/After)"
	if !s.splitPreview {
		viewMode = "Single (Preview)"
	}
	s.infoLine = fmt.Sprintf("Active: %-28s │ View: %-20s │ Render: %d ms │ Colors: %d",
		activeDesc, viewMode, s.lastDurMS, len(s.bottomSwatches.Swatches))
	s.infoPara.Text = s.infoLine

	if s.statusMsg == "" {
		s.statusMsg = "Ready"
	}
	s.statusNode.Text = s.statusMsg
	s.statusBar.Center[0].Text = s.statusMsg
	s.statusBar.Right[0].Text = fmt.Sprintf("%d×%d", w, h)
}

func (s *Studio) updateProtoLabel(proto graphics.Protocol) {
	want := "Half-Block"
	switch proto {
	case graphics.ProtocolOverlay:
		if os.Getenv("LIMONI_BACKDROP_SHELL") != "" || os.Getenv("TERMIMG_X11") != "" {
			want = "termimg"
		} else {
			want = "X11 Overlay"
		}
	case graphics.ProtocolKitty:
		if os.Getenv("LIMONI_BACKDROP_SHELL") != "" || os.Getenv("TERMIMG_X11") != "" {
			want = "termimg"
		} else {
			want = "Kitty Graphics"
		}
	case graphics.ProtocolSixel:
		want = "Sixel Graphics"
	case graphics.ProtocolIterm2:
		want = "iTerm2 Inline"
	}
	if want != s.protoLabel {
		s.protoLabel = want
		s.syncStrings()
	}
}

// HandleEvent processes keyboard and drag-and-drop paste events and returns false when the application should exit.
func (s *Studio) HandleEvent(f *limoni.Frame, ev *limoni.Event) bool {
	if ev == nil {
		return true
	}

	// Terminal Drag & Drop arrives as a Bracketed Paste event (EventPaste)
	if ev.Type == driver.EventPaste {
		if s.HandleDropPayload(ev.Paste.Text) {
			return true
		}
		if s.dropZoneOpen {
			s.dropInputSt.SetValue(s.dropInputSt.Value() + strings.TrimSpace(ev.Paste.Text))
			if !s.HandleDropPayload(s.dropInputSt.Value()) {
				s.syncDropStatus()
			}
		}
		return true
	}

	if ev.Type != limoni.EventKey {
		return true
	}
	k := ev.Key
	if k.Release {
		return true
	}

	// Drag & Drop Zone Modal routing
	if s.dropZoneOpen {
		switch k.Type {
		case limoni.KeyEsc:
			s.CloseDropZone()
			return true
		case limoni.KeyEnter:
			s.SubmitDropInput()
			return true
		default:
			s.dropInputSt.HandleKey(k)
			val := strings.TrimSpace(s.dropInputSt.Value())
			// Auto-load if an un-bracketed quoted drop ('/path/img.png' or "/path/img.png") just completed
			if len(val) >= 4 && ((val[0] == '\'' && val[len(val)-1] == '\'') || (val[0] == '"' && val[len(val)-1] == '"')) {
				if s.HandleDropPayload(val) {
					return true
				}
			}
			s.syncDropStatus()
			return true
		}
	}

	// Command Palette toggle & routing
	if k.Ctrl && (k.Ch == 'p' || k.Ch == 'P') {
		s.cmdPaletteSt.Toggle()
		return true
	}
	if s.cmdPaletteSt.IsOpen {
		s.cmdPaletteSt.HandleKey(k)
		return true
	}

	// Global shortcuts
	if k.Type == limoni.KeyEsc || (k.Type == limoni.KeyRune && (k.Ch == 'q' || k.Ch == 'Q')) {
		return false
	}
	if k.Type == limoni.KeyTab {
		next := (s.activeTab + 1) % numTabs
		s.onSelectTab(next)
		return true
	}

	if k.Type == limoni.KeyRune && !k.Ctrl && !k.Alt {
		switch k.Ch {
		case '1':
			s.onSelectTab(TabThemes)
			return true
		case '2':
			s.onSelectTab(TabEffects)
			return true
		case '3':
			s.onSelectTab(TabBgRm)
			return true
		case '4':
			s.onSelectTab(TabOCR)
			return true
		case '5':
			s.onSelectTab(TabPalette)
			return true
		case '6':
			s.onSelectTab(TabFiles)
			return true
		case 'd', 'D':
			s.ToggleDropZone()
			return true
		case '\'', '"':
			// Fallback for terminals without bracketed paste that drop quoted file paths
			s.OpenDropZone()
			s.dropInputSt.SetValue(string(k.Ch))
			s.syncDropStatus()
			return true
		case 'v', 'V':
			s.ToggleSplitView()
			return true
		case 'b', 'B':
			s.ToggleBackend()
			return true
		case 'a', 'A':
			s.ChainPreviewToBase()
			return true
		case 'r', 'R':
			s.ResetToOriginal()
			return true
		case 's', 'S':
			s.SaveCurrentOutput()
			return true
		case 'e', 'E':
			s.ToggleExtractSource()
			return true
		case 'o', 'O':
			s.onSelectTab(TabOCR)
			s.RunOCR()
			return true
		case 'c', 'C':
			s.CopyToClipboard()
			return true
		case 'l', 'L':
			s.LightenSelectedColor()
			return true
		case '/':
			s.cmdPaletteSt.Toggle()
			return true
		case '[', '-':
			s.adjustActiveSlider(-4)
			return true
		case ']', '+', '=':
			s.adjustActiveSlider(4)
			return true
		}
	}

	// Tab-specific navigation
	switch s.activeTab {
	case TabThemes:
		switch k.Type {
		case limoni.KeyUp:
			s.themeListState.Previous()
			s.onSelectTheme(s.themeListState.Selected)
		case limoni.KeyDown:
			if s.themeListState.Selected < len(s.themeItems)-1 {
				s.themeListState.Next()
				s.onSelectTheme(s.themeListState.Selected)
			}
		case limoni.KeyHome:
			s.onSelectTheme(0)
		case limoni.KeyEnd:
			s.onSelectTheme(len(s.themeItems) - 1)
		}
	case TabEffects:
		switch k.Type {
		case limoni.KeyUp:
			s.effectState.Previous()
			s.onSelectEffect(s.effectState.Selected)
		case limoni.KeyDown:
			if s.effectState.Selected < len(studioEffects)-1 {
				s.effectState.Next()
				s.onSelectEffect(s.effectState.Selected)
			}
		case limoni.KeyLeft:
			s.adjustActiveSlider(-4)
		case limoni.KeyRight:
			s.adjustActiveSlider(4)
		}
	case TabBgRm:
		switch k.Type {
		case limoni.KeyUp:
			s.bgListState.Previous()
			s.onSelectBgMode(s.bgListState.Selected)
		case limoni.KeyDown:
			if s.bgListState.Selected < len(studioBgModes)-1 {
				s.bgListState.Next()
				s.onSelectBgMode(s.bgListState.Selected)
			}
		case limoni.KeyLeft:
			s.adjustActiveSlider(-10)
		case limoni.KeyRight:
			s.adjustActiveSlider(10)
		}
	case TabOCR:
		switch k.Type {
		case limoni.KeyUp:
			s.ocrListState.Previous()
			s.onSelectOCREngine(s.ocrListState.Selected)
		case limoni.KeyDown:
			if s.ocrListState.Selected < len(studioOCRPresets)-1 {
				s.ocrListState.Next()
				s.onSelectOCREngine(s.ocrListState.Selected)
			}
		case limoni.KeyEnter:
			s.RunOCR()
		}
	case TabPalette:
		switch k.Type {
		case limoni.KeyUp:
			s.paletteState.Previous()
		case limoni.KeyDown:
			if s.paletteState.Selected < len(s.extractedItems)-1 {
				s.paletteState.Next()
			}
		case limoni.KeyLeft:
			s.adjustActiveSlider(-1)
		case limoni.KeyRight:
			s.adjustActiveSlider(1)
		}
	case TabFiles:
		s.fileState.Chosen = ""
		if s.fileState.HandleKey(k) && s.fileState.Chosen != "" {
			_ = s.LoadImageFile(s.fileState.Chosen)
		}
	}
	return true
}

func (s *Studio) adjustActiveSlider(delta int) {
	switch s.activeTab {
	case TabEffects:
		idx := s.effectState.Selected
		if idx >= 0 && idx < len(studioEffects) && studioEffects[idx].HasParam {
			eff := studioEffects[idx]
			s.paramSliderSt.Set(s.paramSliderSt.Value+delta, eff.Min, eff.Max)
			s.onEffectParamChange(s.paramSliderSt.Value)
		}
	case TabBgRm:
		s.bgSliderSt.Set(s.bgSliderSt.Value+delta, 20, 200)
		s.onBgIterChange(s.bgSliderSt.Value)
	case TabPalette:
		s.palSliderSt.Set(s.palSliderSt.Value+delta, 4, 16)
		s.onPaletteCountChange(s.palSliderSt.Value)
	}
}

// Draw renders the entire Studio frame with zero heap allocations.
func (s *Studio) Draw(f *limoni.Frame, ev *limoni.Event) bool {
	if !s.HandleEvent(f, ev) {
		return false
	}
	// Sync tab and list selections updated via mouse ClickAction on previous frame
	if s.pendingTab >= 0 {
		tab := s.pendingTab
		s.pendingTab = -1
		s.onSelectTab(tab)
	}
	if s.themeListState.Selected != s.selectedTheme {
		s.onSelectTheme(s.themeListState.Selected)
	}
	if s.effectState.Selected != s.selectedEffect {
		s.onSelectEffect(s.effectState.Selected)
	}
	if s.bgListState.Selected != s.selectedBgMode {
		s.onSelectBgMode(s.bgListState.Selected)
	}
	if s.ocrListState.Selected != s.selectedOCR {
		s.onSelectOCREngine(s.ocrListState.Selected)
	}
	// Check if a file was chosen via mouse double-click in FilePicker
	if s.fileState.Chosen != "" {
		chosen := s.fileState.Chosen
		s.fileState.Chosen = ""
		_ = s.LoadImageFile(chosen)
	}

	s.updateProtoLabel(graphics.Protocol(f.ImageProtocol))

	area := f.Area()
	if area.Width == 0 || area.Height == 0 {
		return true
	}

	headerArea, bodyArea, statusArea := splitV3(area, 3, 1)

	// 1. Header Bar
	f.RenderWidget(&s.headerBlock, headerArea)
	headerInner := s.headerBlock.Inner(headerArea)
	hLeft, hRight := splitH2FixedRight(headerInner, 62)
	f.RenderWidget(&s.headerPara, hLeft)
	f.RenderWidget(&s.tabs, hRight)

	// 2. Main Body (Sidebar + Viewport)
	leftArea, rightArea := splitH2FixedLeft(bodyArea, 36, 1)

	s.drawSidebar(f, leftArea)
	s.drawViewport(f, rightArea)

	// 3. Status Bar
	f.RenderWidget(&s.statusBar, statusArea)
	f.RenderWidget(&s.statusNode, cell.Rect{X: statusArea.X, Y: statusArea.Y, Width: 0, Height: 0})

	// 4. Command Palette Overlay (if open)
	if s.cmdPaletteSt.IsOpen {
		palArea := s.cmdPalette.DebugArea(area)
		f.RegisterModal("command-palette", palArea, s.cmdPaletteSt.Close)
		f.BeginLayer("command-palette")
		f.RenderWidget(&s.cmdPalette, area)
		f.EndLayer()
	}

	// 5. Drag & Drop Image Zone Overlay (if open)
	if s.dropZoneOpen {
		s.drawDropZoneModal(f, area)
	}

	return true
}

func (s *Studio) drawDropZoneModal(f *limoni.Frame, area cell.Rect) {
	mw := min(area.Width-4, uint16(76))
	mh := min(area.Height-4, uint16(13))
	if mw < 30 || mh < 9 {
		return
	}
	mx := area.X + (area.Width-mw)/2
	my := area.Y + (area.Height-mh)/2
	modalRect := cell.NewRect(mx, my, mw, mh)

	f.RegisterModal("drop-zone", modalRect, s.CloseDropZone)
	f.BeginLayer("drop-zone")
	f.RenderWidget(&s.dropModalBlock, modalRect)
	inner := s.dropModalBlock.Inner(modalRect)

	topBox, midBox, botBox := splitV3(inner, 4, 2)
	f.RenderWidget(&s.dropZoneBlock, topBox)
	f.RenderWidget(&s.dropHintPara, s.dropZoneBlock.Inner(topBox))

	inputBox, _ := splitV2FixedBottom(midBox, 0)
	if inputBox.Height > 3 {
		inputBox.Height = 3
	}
	f.RenderWidget(&s.dropInputBlock, inputBox)
	f.RenderWidget(&s.dropInput, s.dropInputBlock.Inner(inputBox))

	statusRow, btnRow := splitV2FixedBottom(botBox, 1)
	f.RenderWidget(&s.dropStatusPara, statusRow)
	b0, b1 := splitH2Equal(btnRow, 2)
	f.RenderWidget(&s.btnDropLoad, b0)
	f.RenderWidget(&s.btnDropCancel, b1)
	f.EndLayer()
}

func (s *Studio) drawSidebar(f *limoni.Frame, area cell.Rect) {
	switch s.activeTab {
	case TabThemes:
		r0, r1, r2 := splitV3Bottom2(area, 6, 4)
		f.RenderWidget(&s.sidebarBlock, r0)
		f.RenderWidget(&s.themeList, s.sidebarBlock.Inner(r0))

		f.RenderWidget(&s.swatchBlock, r1)
		f.RenderWidget(&s.paletteStateW, s.swatchBlock.Inner(r1))

		f.RenderWidget(&s.actionsBlock, r2)
		actInner := s.actionsBlock.Inner(r2)
		b0, b1 := splitH2Equal(actInner, 1)
		f.RenderWidget(&s.btnSave, b0)
		f.RenderWidget(&s.btnBackend, b1)

	case TabEffects:
		r0, r1, r2 := splitV3Bottom2(area, 4, 4)
		f.RenderWidget(&s.sidebarBlock, r0)
		f.RenderWidget(&s.effectList, s.sidebarBlock.Inner(r0))

		f.RenderWidget(&s.paramBlock, r1)
		paramInner := s.paramBlock.Inner(r1)
		f.RenderWidget(&s.paramSlider, paramInner)

		f.RenderWidget(&s.actionsBlock, r2)
		actInner := s.actionsBlock.Inner(r2)
		b0, b1 := splitH2Equal(actInner, 1)
		f.RenderWidget(&s.btnChain, b0)
		f.RenderWidget(&s.btnReset, b1)

	case TabBgRm:
		r0, r1, r2 := splitV3Bottom2(area, 4, 4)
		f.RenderWidget(&s.sidebarBlock, r0)
		f.RenderWidget(&s.bgList, s.sidebarBlock.Inner(r0))

		f.RenderWidget(&s.paramBlock, r1)
		f.RenderWidget(&s.bgSlider, s.paramBlock.Inner(r1))

		f.RenderWidget(&s.actionsBlock, r2)
		actInner := s.actionsBlock.Inner(r2)
		b0, b1 := splitH2Equal(actInner, 1)
		f.RenderWidget(&s.btnChain, b0)
		f.RenderWidget(&s.btnSave, b1)

	case TabOCR:
		r0, r1, r2 := splitV3Bottom2(area, 5, 4)
		f.RenderWidget(&s.sidebarBlock, r0)
		f.RenderWidget(&s.ocrList, s.sidebarBlock.Inner(r0))

		f.RenderWidget(&s.ocrMetaBlock, r1)
		f.RenderWidget(&s.ocrMetaPara, s.ocrMetaBlock.Inner(r1))

		f.RenderWidget(&s.actionsBlock, r2)
		actInner := s.actionsBlock.Inner(r2)
		b0, b1 := splitH2Equal(actInner, 1)
		f.RenderWidget(&s.btnRunOCR, b0)
		f.RenderWidget(&s.btnSaveOCR, b1)

	case TabPalette:
		r0, r1, r2 := splitV3(area, 4, 4)
		f.RenderWidget(&s.paramBlock, r0)
		f.RenderWidget(&s.palSlider, s.paramBlock.Inner(r0))

		f.RenderWidget(&s.sidebarBlock, r1)
		f.RenderWidget(&s.paletteList, s.sidebarBlock.Inner(r1))

		f.RenderWidget(&s.actionsBlock, r2)
		actInner := s.actionsBlock.Inner(r2)
		b0, b1 := splitH2Equal(actInner, 1)
		f.RenderWidget(&s.btnExtractSrc, b0)
		f.RenderWidget(&s.btnLighten, b1)

	case TabFiles:
		r0, r1, r2 := splitV3Bottom2(area, 4, 4)
		f.RenderWidget(&s.sidebarBlock, r0)
		f.RenderWidget(&s.filePicker, s.sidebarBlock.Inner(r0))

		f.RenderWidget(&s.swatchBlock, r1)
		f.RenderWidget(&s.paletteStateW, s.swatchBlock.Inner(r1))

		f.RenderWidget(&s.actionsBlock, r2)
		actInner := s.actionsBlock.Inner(r2)
		b0, b1 := splitH2Equal(actInner, 1)
		f.RenderWidget(&s.btnDropOpen, b0)
		f.RenderWidget(&s.btnSave, b1)
	}
}

func (s *Studio) drawViewport(f *limoni.Frame, area cell.Rect) {
	imgArea, infoArea := splitV2FixedBottom(area, 4)

	if s.activeTab == TabOCR {
		if s.splitPreview {
			c0, c1 := splitH2Equal(imgArea, 1)
			f.RenderWidget(&s.origBlock, c0)
			f.RenderWidget(&s.origImageW, s.origBlock.Inner(c0))

			f.RenderWidget(&s.ocrBlock, c1)
			f.RenderWidget(&s.ocrTextPara, s.ocrBlock.Inner(c1))
		} else {
			f.RenderWidget(&s.ocrBlock, imgArea)
			f.RenderWidget(&s.ocrTextPara, s.ocrBlock.Inner(imgArea))
		}
	} else if s.splitPreview {
		c0, c1 := splitH2Equal(imgArea, 1)
		f.RenderWidget(&s.origBlock, c0)
		f.RenderWidget(&s.origImageW, s.origBlock.Inner(c0))

		f.RenderWidget(&s.prevBlock, c1)
		f.RenderWidget(&s.prevImageW, s.prevBlock.Inner(c1))
	} else {
		f.RenderWidget(&s.prevBlock, imgArea)
		f.RenderWidget(&s.prevImageW, s.prevBlock.Inner(imgArea))
	}

	f.RenderWidget(&s.infoBlock, infoArea)
	infoInner := s.infoBlock.Inner(infoArea)
	ir0, ir1 := splitV2FixedBottom(infoInner, 1)
	f.RenderWidget(&s.infoPara, ir0)
	f.RenderWidget(&s.bottomSwatches, ir1)
}

func splitV3(area cell.Rect, topH, botH uint16) (top, mid, bot cell.Rect) {
	if area.Height <= topH+botH {
		return area, cell.Rect{}, cell.Rect{}
	}
	midH := area.Height - topH - botH
	top = cell.Rect{X: area.X, Y: area.Y, Width: area.Width, Height: topH}
	mid = cell.Rect{X: area.X, Y: area.Y + topH, Width: area.Width, Height: midH}
	bot = cell.Rect{X: area.X, Y: area.Y + topH + midH, Width: area.Width, Height: botH}
	return top, mid, bot
}

func splitV3Bottom2(area cell.Rect, midH, botH uint16) (top, mid, bot cell.Rect) {
	if area.Height <= midH+botH {
		return area, cell.Rect{}, cell.Rect{}
	}
	topH := area.Height - midH - botH
	top = cell.Rect{X: area.X, Y: area.Y, Width: area.Width, Height: topH}
	mid = cell.Rect{X: area.X, Y: area.Y + topH, Width: area.Width, Height: midH}
	bot = cell.Rect{X: area.X, Y: area.Y + topH + midH, Width: area.Width, Height: botH}
	return top, mid, bot
}

func splitV2FixedBottom(area cell.Rect, botH uint16) (top, bot cell.Rect) {
	if area.Height <= botH {
		return area, cell.Rect{}
	}
	topH := area.Height - botH
	top = cell.Rect{X: area.X, Y: area.Y, Width: area.Width, Height: topH}
	bot = cell.Rect{X: area.X, Y: area.Y + topH, Width: area.Width, Height: botH}
	return top, bot
}

func splitH2FixedRight(area cell.Rect, rightW uint16) (left, right cell.Rect) {
	if area.Width <= rightW {
		return area, cell.Rect{}
	}
	leftW := area.Width - rightW
	left = cell.Rect{X: area.X, Y: area.Y, Width: leftW, Height: area.Height}
	right = cell.Rect{X: area.X + leftW, Y: area.Y, Width: rightW, Height: area.Height}
	return left, right
}

func splitH2FixedLeft(area cell.Rect, leftW, gap uint16) (left, right cell.Rect) {
	if area.Width <= leftW+gap {
		return area, cell.Rect{}
	}
	rightW := area.Width - leftW - gap
	left = cell.Rect{X: area.X, Y: area.Y, Width: leftW, Height: area.Height}
	right = cell.Rect{X: area.X + leftW + gap, Y: area.Y, Width: rightW, Height: area.Height}
	return left, right
}

func splitH2Equal(area cell.Rect, gap uint16) (left, right cell.Rect) {
	if area.Width <= gap {
		return area, cell.Rect{}
	}
	usable := area.Width - gap
	leftW := usable / 2
	rightW := usable - leftW
	left = cell.Rect{X: area.X, Y: area.Y, Width: leftW, Height: area.Height}
	right = cell.Rect{X: area.X + leftW + gap, Y: area.Y, Width: rightW, Height: area.Height}
	return left, right
}

// Run launches the interactive Gowall Studio TUI on the active terminal.
func Run(opts Options) error {
	logger.SetQuiet(true)
	defer logger.SetQuiet(false)

	// When running inside a termimg X11/PTY session, use the X11 overlay
	// protocol so full-resolution pixels render cleanly in Alacritty and
	// modal overlays (such as the Command Palette) automatically occlude images.
	if os.Getenv("LIMONI_GRAPHICS") == "" && (os.Getenv("LIMONI_BACKDROP_SHELL") != "" || os.Getenv("TERMIMG_X11") != "") {
		_ = os.Setenv("LIMONI_GRAPHICS", "overlay")
	}

	studio := NewStudio(opts)
	appOpts := []limoni.AppOption{
		limoni.WithTitle("Gowall Studio"),
		limoni.WithOverlayPictures(),
		limoni.WithSuspend(),
	}
	sock := opts.SocketPath
	if sock == "" {
		sock = os.Getenv("LIMONI_AUTOMATION_SOCKET")
	}
	if sock != "" {
		appOpts = append(appOpts, limoni.WithAutomation(sock, limoni.AutomationPolicy{
			AllowInput:        true,
			ExposeScreen:      true,
			ExposeInputValues: true,
		}))
	}
	return limoni.Run(studio.Draw, appOpts...)
}
