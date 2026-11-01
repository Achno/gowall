package tui

import (
	"image/color"

	cpkg "github.com/Achno/gowall/internal/backends/color"
	"github.com/thebanri/limoni/core/accessibility"
	"github.com/thebanri/limoni/core/buffer"
	"github.com/thebanri/limoni/core/cell"
	"github.com/thebanri/limoni/widgets"
)

// SwatchEntry holds precomputed color and hex label data for a single palette swatch
// so drawing a PaletteBar performs zero heap allocations.
type SwatchEntry struct {
	Hex   string
	Color cell.Color
}

// BuildSwatches converts hex strings (e.g. "#1E1E2E") into precomputed SwatchEntry items.
func BuildSwatches(hexColors []string) []SwatchEntry {
	out := make([]SwatchEntry, 0, len(hexColors))
	for _, h := range hexColors {
		rgba, err := cpkg.HexToRGBA(h)
		if err != nil {
			continue
		}
		out = append(out, SwatchEntry{
			Hex:   cpkg.RGBtoHex(rgba),
			Color: cell.NewColorRGB(rgba.R, rgba.G, rgba.B),
		})
	}
	return out
}

// BuildSwatchesFromColors converts image/color.Color items into precomputed SwatchEntry items.
func BuildSwatchesFromColors(clrs []color.Color) []SwatchEntry {
	out := make([]SwatchEntry, 0, len(clrs))
	for _, c := range clrs {
		rgba := cpkg.RGBA8(c)
		out = append(out, SwatchEntry{
			Hex:   cpkg.RGBtoHex(rgba),
			Color: cell.NewColorRGB(rgba.R, rgba.G, rgba.B),
		})
	}
	return out
}

// PaletteBar renders a grid of truecolor color blocks with their hex codes.
// Its Draw method performs zero allocations.
type PaletteBar struct {
	ID       string
	Label    string
	Summary  string
	Swatches []SwatchEntry
	Compact  bool
	Style    cell.Style
}

// Draw renders the color swatches into buf without allocating.
func (p *PaletteBar) Draw(ctx cell.Context, buf *buffer.Buffer) {
	area := ctx.Area
	if area.Width == 0 || area.Height == 0 {
		return
	}
	base := ctx.Style.Merge(p.Style)
	for y := area.Y; y < area.Y+area.Height; y++ {
		for x := area.X; x < area.X+area.Width; x++ {
			buf.SetCell(x, y, cell.Cell{Content: ' ', Style: base})
		}
	}
	if len(p.Swatches) == 0 {
		buf.SetStringWithin(area.X, area.Y, "(no palette colors)", base.Dim(), area.Width)
		return
	}

	curX := area.X
	curY := area.Y
	endX := area.X + area.Width
	endY := area.Y + area.Height

	itemWidth := uint16(11) // "██ #RRGGBB "
	if p.Compact {
		itemWidth = 3 // "██ "
	}

	for i := range p.Swatches {
		sw := &p.Swatches[i]
		if curX+itemWidth > endX {
			curX = area.X
			curY++
			if curY >= endY {
				break
			}
		}
		blockStyle := base.WithFg(sw.Color)
		n := buf.SetStringWithin(curX, curY, "██", blockStyle, endX-curX)
		curX += n
		if !p.Compact && curX < endX {
			curX += buf.SetStringWithin(curX, curY, " ", base, endX-curX)
			curX += buf.SetStringWithin(curX, curY, sw.Hex, base, endX-curX)
			if curX < endX {
				curX += buf.SetStringWithin(curX, curY, "  ", base, endX-curX)
			}
		} else if p.Compact && curX < endX {
			curX += buf.SetStringWithin(curX, curY, " ", base, endX-curX)
		}
	}
}

// SizeHint returns the available area.
func (p *PaletteBar) SizeHint(maxArea cell.Rect) (uint16, uint16) {
	return maxArea.Width, maxArea.Height
}

// AccessibilityNode exposes the palette summary in the semantic tree.
func (p *PaletteBar) AccessibilityNode(bounds cell.Rect, focused bool) accessibility.Node {
	label := p.Label
	if label == "" {
		label = "Palette"
	}
	state := accessibility.NodeState(0)
	if focused {
		state |= accessibility.StateFocused
	}
	return accessibility.Node{
		ID:     p.ID,
		Role:   accessibility.RoleGeneric,
		Label:  label,
		Value:  p.Summary,
		State:  state,
		Bounds: bounds,
	}
}

// ImageView wraps widgets.Image so it also registers a semantic node in the accessibility tree.
type ImageView struct {
	widgets.Image
	Label string
	Value string
}

// AccessibilityNode exposes the image viewport in the Limoni semantic tree.
func (iv *ImageView) AccessibilityNode(area cell.Rect, _ bool) accessibility.Node {
	return accessibility.Node{
		ID:     iv.ID,
		Role:   accessibility.RoleGeneric,
		Label:  iv.Label,
		Value:  iv.Value,
		Bounds: area,
	}
}

// StudioTabs wraps widgets.Tabs to register allocation-free ClickActions per tab.
type StudioTabs struct {
	widgets.Tabs
	Pending *int
}

// Draw renders the tab bar and registers zero-allocation click regions for each tab.
func (st *StudioTabs) Draw(ctx cell.Context, buf *buffer.Buffer) {
	st.Tabs.Draw(ctx, buf)
	if ctx.RegisterClickAction == nil || st.Pending == nil {
		return
	}
	area := ctx.Area
	x := area.X
	limit := area.X + area.Width
	for i, title := range st.Titles {
		if i > 0 {
			if x+1 > limit {
				return
			}
			x++
		}
		w := uint16(len(title)) + 2
		if x >= limit {
			return
		}
		if x+w > limit {
			w = limit - x
		}
		ctx.RegisterClickAction(cell.NewRect(x, area.Y, w, 1), cell.ClickAction{
			Focus:  st.ID,
			Select: st.Pending,
			Index:  i,
		})
		x += w
	}
}


