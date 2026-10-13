package image

import "image"

// CloneRGBA returns a copy of the image that doesn't share pixels with the original
func CloneRGBA(src *image.RGBA) *image.RGBA {
	dst := image.NewRGBA(src.Bounds())
	copy(dst.Pix, src.Pix)
	return dst
}

// HasTransparency reports whether any pixel of the image is mostly transparent, like the background of a sprite
func HasTransparency(img *image.NRGBA) bool {
	for i := 3; i < len(img.Pix); i += 4 {
		if img.Pix[i] < 128 {
			return true
		}
	}
	return false
}

// HasOpaque reports whether any pixel of the image is mostly visible
func HasOpaque(img *image.NRGBA) bool {
	for i := 3; i < len(img.Pix); i += 4 {
		if img.Pix[i] >= 128 {
			return true
		}
	}
	return false
}
