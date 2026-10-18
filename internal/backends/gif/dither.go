package gif

import "image"

// webSafeStep is the distance between the levels of a channel in palette.WebSafe, in the 16-bit values of RGBA() (0x33 * 0x101)
const webSafeStep = 0x3333

// DitherWebSafe gives the same result as draw.FloydSteinberg.Draw(dst, dst.Bounds(), src, dst.Bounds().Min) when dst has the
// palette.WebSafe palette, but much faster.
//
// draw.FloydSteinberg compares every pixel against all 216 colors of the palette. palette.WebSafe is a 6x6x6 cube of evenly
// spaced levels (r*36 + g*6 + b), so the closest color is found per channel instead: the squared distance is a sum of one term
// per channel and alpha is the same for every color, so the closest level of each channel gives the closest color. Ties go to
// the lower level, which is the lower index draw picks too. The error diffusion is a copy of draw's.
func DitherWebSafe(dst *image.Paletted, src image.Image) {
	r := dst.Bounds()
	sp := r.Min

	pxRGBA := func(x, y int) (r, g, b, a uint32) { return src.At(x, y).RGBA() }
	switch src0 := src.(type) {
	case *image.RGBA:
		pxRGBA = func(x, y int) (r, g, b, a uint32) { return src0.RGBAAt(x, y).RGBA() }
	case *image.NRGBA:
		pxRGBA = func(x, y int) (r, g, b, a uint32) { return src0.NRGBAAt(x, y).RGBA() }
	case *image.YCbCr:
		pxRGBA = func(x, y int) (r, g, b, a uint32) { return src0.YCbCrAt(x, y).RGBA() }
	}

	pix, stride := dst.Pix[dst.PixOffset(r.Min.X, r.Min.Y):], dst.Stride

	// the Floyd-Steinberg errors propagated to the current and next row, the +2 simplifies the edges.
	// Alpha is left out as it never changes which color is picked.
	quantErrorCurr := make([][3]int32, r.Dx()+2)
	quantErrorNext := make([][3]int32, r.Dx()+2)

	for y := 0; y != r.Dy(); y++ {
		for x := 0; x != r.Dx(); x++ {
			sr, sg, sb, _ := pxRGBA(sp.X+x, sp.Y+y)
			er := clamp(int32(sr) + quantErrorCurr[x+1][0]/16)
			eg := clamp(int32(sg) + quantErrorCurr[x+1][1]/16)
			eb := clamp(int32(sb) + quantErrorCurr[x+1][2]/16)

			lr, lg, lb := webSafeLevel(er), webSafeLevel(eg), webSafeLevel(eb)
			pix[y*stride+x] = byte(lr*36 + lg*6 + lb)

			er -= int32(lr) * webSafeStep
			eg -= int32(lg) * webSafeStep
			eb -= int32(lb) * webSafeStep

			// Propagate the Floyd-Steinberg quantization error.
			quantErrorNext[x+0][0] += er * 3
			quantErrorNext[x+0][1] += eg * 3
			quantErrorNext[x+0][2] += eb * 3
			quantErrorNext[x+1][0] += er * 5
			quantErrorNext[x+1][1] += eg * 5
			quantErrorNext[x+1][2] += eb * 5
			quantErrorNext[x+2][0] += er * 1
			quantErrorNext[x+2][1] += eg * 1
			quantErrorNext[x+2][2] += eb * 1
			quantErrorCurr[x+2][0] += er * 7
			quantErrorCurr[x+2][1] += eg * 7
			quantErrorCurr[x+2][2] += eb * 7
		}

		// Recycle the quantization error buffers.
		quantErrorCurr, quantErrorNext = quantErrorNext, quantErrorCurr
		clear(quantErrorNext)
	}
}

// webSafeLevel returns the closest of the 6 levels (0-5) to a 16-bit channel value, measured like draw does
func webSafeLevel(v int32) int {
	level := v / webSafeStep
	if level >= 5 {
		return 5
	}
	if sqDiff(v, (level+1)*webSafeStep) < sqDiff(v, level*webSafeStep) {
		return int(level + 1)
	}
	return int(level)
}

// clamp and sqDiff are copies of image/draw's
func clamp(i int32) int32 {
	if i < 0 {
		return 0
	}
	if i > 0xffff {
		return 0xffff
	}
	return i
}

func sqDiff(x, y int32) uint32 {
	d := uint32(x - y)
	return (d * d) >> 2
}
