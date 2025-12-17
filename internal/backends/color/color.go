package color

import "github.com/muesli/gamut"

func LightenColor(hex string, amount float64) (string, error) {

	c := gamut.Hex(hex)
	lc := gamut.Lighter(c, amount)

	return gamut.ToHex(lc), nil
}

func DarkenColor(hex string, amount float64) (string, error) {
	c := gamut.Hex(hex)
	dc := gamut.Darker(c, amount)
	return gamut.ToHex(dc), nil
}

func BlendColors(hex1 string, hex2 string, count int) ([]string, error) {
	c1 := gamut.Hex(hex1)
	c2 := gamut.Hex(hex2)
	blended := gamut.Blends(c1, c2, count)

	return ColorsToHex(blended), nil
}

func GenerateShades(hex string, count int) ([]string, error) {
	c := gamut.Hex(hex)
	shades := gamut.Shades(c, count)
	return ColorsToHex(shades), nil
}

func GenerateTints(hex string, count int) ([]string, error) {
	c := gamut.Hex(hex)
	tints := gamut.Tints(c, count)
	return ColorsToHex(tints), nil
}

func GenerateTones(hex string, count int) ([]string, error) {
	c := gamut.Hex(hex)
	tones := gamut.Tones(c, count)
	return ColorsToHex(tones), nil
}

func GenerateMonochromatic(hex string, count int) ([]string, error) {
	c := gamut.Hex(hex)
	monochromatic := gamut.Monochromatic(c, count)
	return ColorsToHex(monochromatic), nil

}
