package image

import (
	"image/color"
)

type Theme struct{
	Name string
	Colors []color.Color
}

var Catpuccin = Theme{
	Name: "Catpuccin",
	Colors: []color.Color{
		color.RGBA{R: 245, G: 224, B: 220, A: 255},
		color.RGBA{R: 242, G: 205, B: 205, A: 255},
		color.RGBA{R: 245, G: 194, B: 231, A: 255},
		color.RGBA{R: 203, G: 166, B: 247, A: 255},
		color.RGBA{R: 243, G: 139, B: 168, A: 255},
		color.RGBA{R: 235, G: 160, B: 172, A: 255},
		color.RGBA{R: 250, G: 179, B: 135, A: 255},
		color.RGBA{R: 249, G: 226, B: 175, A: 255},
		color.RGBA{R: 166, G: 227, B: 161, A: 255},
		color.RGBA{R: 148, G: 226, B: 213, A: 255},
		color.RGBA{R: 137, G: 220, B: 235, A: 255},
		color.RGBA{R: 116, G: 199, B: 236, A: 255},
		color.RGBA{R: 137, G: 180, B: 250, A: 255},
		color.RGBA{R: 180, G: 190, B: 254, A: 255},
		color.RGBA{R: 205, G: 214, B: 244, A: 255},
		color.RGBA{R: 186, G: 194, B: 222, A: 255},
		color.RGBA{R: 166, G: 173, B: 200, A: 255},
		color.RGBA{R: 147, G: 153, B: 178, A: 255},
		color.RGBA{R: 127, G: 132, B: 156, A: 255},
		color.RGBA{R: 108, G: 112, B: 134, A: 255},
		color.RGBA{R: 88, G: 91, B: 112, A: 255},
		color.RGBA{R: 69, G: 71, B: 90, A: 255},
		color.RGBA{R: 49, G: 50, B: 68, A: 255},
		color.RGBA{R: 30, G: 30, B: 46, A: 255},
		color.RGBA{R: 24, G: 24, B: 37, A: 255},
		color.RGBA{R: 17, G: 17, B: 27, A: 255},

	},
}