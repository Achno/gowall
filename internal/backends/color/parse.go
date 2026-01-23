package color

import (
	"fmt"
	"image/color"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// clrParseMap is the package-level map of color format parsers
var clrParseMap = map[string]func(string) (string, error){
	"rgb": ParseRGBToHex,
	"hsl": ParseHSLToHex,
	"hex": ParseHexToHex,
	"lab": ParseLabToHex,
}

// GetClrFormat detects the format of the input color string
func GetClrFormat(colorStr string) string {
	colorStr = strings.TrimSpace(colorStr)

	if strings.HasPrefix(colorStr, "#") {
		return "hex"
	}
	if strings.HasPrefix(colorStr, "rgb") {
		return "rgb"
	}
	if strings.HasPrefix(colorStr, "hsl") {
		return "hsl"
	}
	if strings.HasPrefix(colorStr, "lab") {
		return "lab"
	}

	return "unknown"
}

// GetClrParseMap returns the map of color format parsers
func GetClrParseMap() map[string]func(string) (string, error) {
	return clrParseMap
}

// ValidFormats returns the list of valid color formats derived from the parser map
func ValidFormats() []string {
	formats := make([]string, 0, len(clrParseMap))
	for format := range clrParseMap {
		formats = append(formats, format)
	}
	sort.Strings(formats)
	return formats
}

// ParseColorToHex parses any color format and converts it to hex
func ParseColorToHex(colorStr string) (string, error) {
	inputFormat := GetClrFormat(colorStr)
	if inputFormat == "unknown" {
		return "", fmt.Errorf("unrecognized color format: %s", colorStr)
	}

	parser, exists := clrParseMap[inputFormat]
	if !exists {
		return "", fmt.Errorf("no parser available for format: %s", inputFormat)
	}

	return parser(colorStr)
}

// ParseHexToHex validates and returns hex color
func ParseHexToHex(colorStr string) (string, error) {
	colorStr = strings.TrimSpace(colorStr)
	_, err := HexToRGBA(colorStr)
	if err != nil {
		return "", fmt.Errorf("invalid hex color: %v", err)
	}
	return colorStr, nil
}

// ParseRGBToHex parses RGB format and converts to hex
func ParseRGBToHex(colorStr string) (string, error) {
	colorStr = strings.TrimSpace(colorStr)
	rgbRegex := regexp.MustCompile(`^rgb\s*\(\s*(\d+)\s*,\s*(\d+)\s*,\s*(\d+)\s*\)$`)
	matches := rgbRegex.FindStringSubmatch(colorStr)
	if matches == nil {
		return "", fmt.Errorf("invalid RGB format: %s", colorStr)
	}

	r, _ := strconv.Atoi(matches[1])
	g, _ := strconv.Atoi(matches[2])
	b, _ := strconv.Atoi(matches[3])

	if r > 255 || g > 255 || b > 255 {
		return "", fmt.Errorf("RGB values must be in range 0-255")
	}

	return RGBtoHex(color.RGBA{R: uint8(r), G: uint8(g), B: uint8(b), A: 255}), nil
}

// ParseHSLToHex parses HSL format and converts to hex
func ParseHSLToHex(colorStr string) (string, error) {
	colorStr = strings.TrimSpace(colorStr)
	hslRegex := regexp.MustCompile(`^hsl\s*\(\s*([\d.]+)\s*,\s*([\d.]+)\s*,\s*([\d.]+)\s*\)$`)
	matches := hslRegex.FindStringSubmatch(colorStr)
	if matches == nil {
		return "", fmt.Errorf("invalid HSL format: %s", colorStr)
	}

	h, _ := strconv.ParseFloat(matches[1], 64)
	s, _ := strconv.ParseFloat(matches[2], 64)
	l, _ := strconv.ParseFloat(matches[3], 64)

	return HslToHex(HSL{H: h, S: s, L: l}), nil
}

// ParseLabToHex parses LAB format and converts to hex
func ParseLabToHex(colorStr string) (string, error) {
	colorStr = strings.TrimSpace(colorStr)
	labRegex := regexp.MustCompile(`^lab\s*\(\s*([\d.-]+)\s*,\s*([\d.-]+)\s*,\s*([\d.-]+)\s*\)$`)
	matches := labRegex.FindStringSubmatch(colorStr)
	if matches == nil {
		return "", fmt.Errorf("invalid LAB format: %s", colorStr)
	}

	l, _ := strconv.ParseFloat(matches[1], 64)
	a, _ := strconv.ParseFloat(matches[2], 64)
	b, _ := strconv.ParseFloat(matches[3], 64)

	return LabToHex(LAB{L: l, A: a, B: b}), nil
}

// ConvertHexToFormat converts hex to the specified target format and returns (outputString, outputFormat, error)
func ConvertHexToFormat(hexColor, format string) (string, string, error) {
	normalizedFormat := strings.ToLower(format)

	switch normalizedFormat {
	case "hex":
		return hexColor, "hex", nil

	case "rgb":
		rgba, err := HexToRGBA(hexColor)
		if err != nil {
			return "", "", err
		}
		return fmt.Sprintf("rgb(%d,%d,%d)", rgba.R, rgba.G, rgba.B), "rgb", nil

	case "hsl":
		hsl, err := HexToHsl(hexColor)
		if err != nil {
			return "", "", err
		}
		return fmt.Sprintf("hsl(%.0f,%.0f,%.0f)", hsl.H, hsl.S, hsl.L), "hsl", nil

	case "lab":
		lab, err := HexToLAB(hexColor)
		if err != nil {
			return "", "", err
		}
		return fmt.Sprintf("lab(%.2f,%.2f,%.2f)", lab.L, lab.A, lab.B), "lab", nil

	default:
		return "", "", fmt.Errorf("unsupported format: %s", format)
	}
}
