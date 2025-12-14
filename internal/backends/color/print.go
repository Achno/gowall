package color

import (
	"fmt"
)

// 1.The client creates a transformation with the input hex clrs and whatever text he wants
// 2. Does the same for the output and boxes with the specified get printed along with the clrs.
//
// t, _ := colorbox.NewTransformation(
//     []string{"#FF0000", "lab(50 0 0)"},
//     []string{"#808080"},
// )
// t.Print()

// ColorBox represents a colored terminal box with its color value
type ColorBox struct {
	ColorStr string // The original color string (can be any format)
	Box      string // The colored box ANSI code
}

// CreateColorBox creates a colored box with ANSI codes from any color format
func CreateColorBox(colorStr string) (ColorBox, error) {
	// Parse the color to hex internally to get RGB values
	hexColor, err := ParseColorToHex(colorStr)
	if err != nil {
		return ColorBox{}, err
	}

	c, err := HexToRGBA(hexColor)
	if err != nil {
		return ColorBox{}, err
	}

	r, g, b := c.R, c.G, c.B
	box := fmt.Sprintf("\033[48;2;%d;%d;%dm    \033[0m", r, g, b)
	return ColorBox{ColorStr: colorStr, Box: box}, nil
}

// Transformation represents input colors -> output colors
type Transformation struct {
	Inputs  []ColorBox
	Outputs []ColorBox
}

// NewTransformation creates a transformation from input hex colors to output hex colors
func NewTransformation(inputs []string, outputs []string) (*Transformation, error) {
	t := &Transformation{
		Inputs:  make([]ColorBox, 0, len(inputs)),
		Outputs: make([]ColorBox, 0, len(outputs)),
	}

	for _, hex := range inputs {
		cb, err := CreateColorBox(hex)
		if err != nil {
			return nil, fmt.Errorf("invalid input color %s: %v", hex, err)
		}
		t.Inputs = append(t.Inputs, cb)
	}

	for _, hex := range outputs {
		cb, err := CreateColorBox(hex)
		if err != nil {
			return nil, fmt.Errorf("invalid output color %s: %v", hex, err)
		}
		t.Outputs = append(t.Outputs, cb)
	}

	return t, nil
}

// Print displays the transformation as: c1 box, c2 box, c3 box -> c4 box, c5 box
func (t *Transformation) Print() {
	for i, cb := range t.Inputs {
		if i > 0 {
			fmt.Print(", ")
		}
		fmt.Printf("%s %s", cb.ColorStr, cb.Box)
	}

	fmt.Print("  ->  ")

	for i, cb := range t.Outputs {
		if i > 0 {
			fmt.Print(", ")
		}
		fmt.Printf("%s %s", cb.ColorStr, cb.Box)
	}

	fmt.Println()
}

// PrintCustom allows custom formatting with separator and arrow strings
func (t *Transformation) PrintCustom(separator, arrow string) {
	for i, cb := range t.Inputs {
		if i > 0 {
			fmt.Print(separator)
		}
		fmt.Printf("%s %s", cb.ColorStr, cb.Box)
	}

	fmt.Print(arrow)

	for i, cb := range t.Outputs {
		if i > 0 {
			fmt.Print(separator)
		}
		fmt.Printf("%s %s", cb.ColorStr, cb.Box)
	}

	fmt.Println()
}

// PrintCompact displays transformation in compact format: inputs -> outputs (no hex labels)
func (t *Transformation) PrintCompact() {
	for i, cb := range t.Inputs {
		if i > 0 {
			fmt.Print(" ")
		}
		fmt.Print(cb.Box)
	}

	fmt.Print(" -> ")

	for i, cb := range t.Outputs {
		if i > 0 {
			fmt.Print(" ")
		}
		fmt.Print(cb.Box)
	}

	fmt.Println()
}
