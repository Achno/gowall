// Package fastpng encodes PNGs on every CPU core.
//
// image/png filters and deflates the whole image on a single core, which for a 4K wallpaper takes most of the time of a
// gowall command. Here the rows are split into strips that are filtered and deflated concurrently, then joined into one
// zlib stream the way pigz does it: every strip but the last ends with a sync flush, so the deflate streams can simply be
// concatenated, and the adler32 checksums of the strips are combined. The decoded pixels are the same as image/png's.
package fastpng

import (
	"bytes"
	"compress/flate"
	"encoding/binary"
	"hash/adler32"
	"hash/crc32"
	"image"
	"image/png"
	"io"
	"runtime"
	"sync"
)

const (
	ctGrayscale      = 0
	ctTrueColor      = 2
	ctTrueColorAlpha = 6

	// minStripRows keeps strips big enough that restarting the deflate dictionary costs next to nothing in file size
	minStripRows = 16
	// stripsPerWorker gives the workers several strips each, so one slow strip doesn't keep the others waiting
	stripsPerWorker = 4
)

// Encode writes img as a PNG with the same speed/size tradeoff as png.Encoder{CompressionLevel: png.BestSpeed}.
// *image.RGBA, *image.NRGBA and *image.Gray are encoded concurrently, other images are left to image/png.
func Encode(w io.Writer, img image.Image) error {
	e, ok := newEncoder(img)
	if !ok {
		return (&png.Encoder{CompressionLevel: png.BestSpeed}).Encode(w, img)
	}
	return e.encode(w)
}

type encoder struct {
	img       image.Image
	bounds    image.Rectangle
	colorType byte
	bpp       int // bytes per pixel
	rowFn     func(dst []byte, y int)
}

func newEncoder(img image.Image) (*encoder, bool) {
	b := img.Bounds()
	if b.Empty() || int64(b.Dx()) > 1<<31-1 || int64(b.Dy()) > 1<<31-1 {
		// let image/png report the error
		return nil, false
	}

	e := &encoder{img: img, bounds: b}
	switch m := img.(type) {
	case *image.Gray:
		e.colorType, e.bpp = ctGrayscale, 1
		e.rowFn = func(dst []byte, y int) {
			copy(dst, m.Pix[m.PixOffset(b.Min.X, y):m.PixOffset(b.Max.X, y)])
		}
	case *image.NRGBA:
		if m.Opaque() {
			e.colorType, e.bpp = ctTrueColor, 3
			e.rowFn = func(dst []byte, y int) { dropAlpha(dst, m.Pix[m.PixOffset(b.Min.X, y):m.PixOffset(b.Max.X, y)]) }
		} else {
			e.colorType, e.bpp = ctTrueColorAlpha, 4
			e.rowFn = func(dst []byte, y int) {
				copy(dst, m.Pix[m.PixOffset(b.Min.X, y):m.PixOffset(b.Max.X, y)])
			}
		}
	case *image.RGBA:
		if m.Opaque() {
			e.colorType, e.bpp = ctTrueColor, 3
			e.rowFn = func(dst []byte, y int) { dropAlpha(dst, m.Pix[m.PixOffset(b.Min.X, y):m.PixOffset(b.Max.X, y)]) }
		} else {
			e.colorType, e.bpp = ctTrueColorAlpha, 4
			e.rowFn = func(dst []byte, y int) { unpremultiply(dst, m.Pix[m.PixOffset(b.Min.X, y):m.PixOffset(b.Max.X, y)]) }
		}
	default:
		return nil, false
	}
	return e, true
}

// strip is a range of rows deflated on its own
type strip struct {
	y0, y1     int
	compressed bytes.Buffer
	adler      uint32
	rawLen     int64 // filtered bytes, for combining the checksums
	err        error
}

func (e *encoder) encode(w io.Writer) error {
	height := e.bounds.Dy()
	workers := runtime.GOMAXPROCS(0)
	rows := max(minStripRows, (height+workers*stripsPerWorker-1)/(workers*stripsPerWorker))

	var strips []*strip
	for y0 := 0; y0 < height; y0 += rows {
		strips = append(strips, &strip{y0: y0, y1: min(y0+rows, height)})
	}

	next := make(chan int)
	var wg sync.WaitGroup
	for range min(workers, len(strips)) {
		wg.Add(1)
		go func() {
			defer wg.Done()
			fw, _ := flate.NewWriter(nil, flate.BestSpeed)
			rowLen := 1 + e.bpp*e.bounds.Dx()
			var cr [5][]byte
			for i := range cr {
				cr[i] = make([]byte, rowLen)
				cr[i][0] = byte(i)
			}
			pr := make([]byte, rowLen)
			for i := range next {
				e.compressStrip(strips[i], i == len(strips)-1, fw, &cr, pr)
			}
		}()
	}
	for i := range strips {
		next <- i
	}
	close(next)
	wg.Wait()

	adler := uint32(1)
	for _, s := range strips {
		if s.err != nil {
			return s.err
		}
		adler = adler32Combine(adler, s.adler, s.rawLen)
	}

	cw := &chunkWriter{w: w}
	cw.write([]byte("\x89PNG\r\n\x1a\n"))

	ihdr := make([]byte, 13)
	binary.BigEndian.PutUint32(ihdr[0:4], uint32(e.bounds.Dx()))
	binary.BigEndian.PutUint32(ihdr[4:8], uint32(height))
	ihdr[8], ihdr[9] = 8, e.colorType // 8 bit depth, compression, filter and interlace methods stay 0
	cw.chunk("IHDR", ihdr)

	// zlib header for deflate with a 32K window at the fastest level (the same bytes compress/zlib writes for BestSpeed)
	cw.chunk("IDAT", []byte{0x78, 0x01})
	for _, s := range strips {
		cw.chunk("IDAT", s.compressed.Bytes())
	}
	cw.chunk("IDAT", binary.BigEndian.AppendUint32(nil, adler))
	cw.chunk("IEND", nil)
	return cw.err
}

// compressStrip filters and deflates the rows of the strip, cr and pr are the worker's row buffers
func (e *encoder) compressStrip(s *strip, last bool, fw *flate.Writer, cr *[5][]byte, pr []byte) {
	fw.Reset(&s.compressed)
	hash := adler32.New()
	out := io.MultiWriter(fw, hash)

	// the filters of the first row need the unfiltered row above it
	clear(pr)
	if s.y0 > 0 {
		e.rowFn(pr[1:], e.bounds.Min.Y+s.y0-1)
	}

	for y := s.y0; y < s.y1; y++ {
		e.rowFn(cr[0][1:], e.bounds.Min.Y+y)
		f := filter(cr, pr, e.bpp)
		if _, err := out.Write(cr[f]); err != nil {
			s.err = err
			return
		}
		s.rawLen += int64(len(cr[f]))
		// the current row becomes the previous one, unfiltered
		copy(pr, cr[0])
	}

	// a sync flush ends the strip on a byte boundary without ending the stream, so the next strip can follow it
	if last {
		s.err = fw.Close()
	} else {
		s.err = fw.Flush()
	}
	s.adler = hash.Sum32()
}

// dropAlpha copies the RGB of opaque RGBA pixels
func dropAlpha(dst, src []byte) {
	for i, j := 0, 0; j+3 < len(src); i, j = i+3, j+4 {
		dst[i], dst[i+1], dst[i+2] = src[j], src[j+1], src[j+2]
	}
}

// unpremultiply converts alpha-premultiplied RGBA to PNG's non-premultiplied RGBA, the same way image/png does
func unpremultiply(dst, src []byte) {
	for ; len(src) >= 4; dst, src = dst[4:], src[4:] {
		d := (*[4]byte)(dst)
		s := (*[4]byte)(src)
		switch s[3] {
		case 0x00:
			*d = [4]byte{}
		case 0xff:
			*d = *s
		default:
			// 0x101 converts 8-bit color to 16-bit, 0xffff together with the division by a removes the premultiplication
			const m = 0x101 * 0xffff
			a := uint32(s[3]) * 0x101
			d[0] = uint8((uint32(s[0]) * m / a) >> 8)
			d[1] = uint8((uint32(s[1]) * m / a) >> 8)
			d[2] = uint8((uint32(s[2]) * m / a) >> 8)
			d[3] = s[3]
		}
	}
}

// adler32Combine returns the adler32 of A followed by B from the checksums of A and B, as zlib's adler32_combine does
func adler32Combine(adlerA, adlerB uint32, lenB int64) uint32 {
	const base = 65521
	rem := uint32(lenB % base)
	sum1 := adlerA & 0xffff
	sum2 := (rem * sum1) % base
	sum1 += (adlerB & 0xffff) + base - 1
	sum2 += (adlerA >> 16) + (adlerB >> 16) + base - rem
	if sum1 >= base {
		sum1 -= base
	}
	if sum1 >= base {
		sum1 -= base
	}
	if sum2 >= base<<1 {
		sum2 -= base << 1
	}
	if sum2 >= base {
		sum2 -= base
	}
	return sum1 | sum2<<16
}

// chunkWriter writes PNG chunks and keeps the first error
type chunkWriter struct {
	w   io.Writer
	err error
}

func (cw *chunkWriter) write(b []byte) {
	if cw.err == nil {
		_, cw.err = cw.w.Write(b)
	}
}

func (cw *chunkWriter) chunk(name string, data []byte) {
	header := binary.BigEndian.AppendUint32(nil, uint32(len(data)))
	header = append(header, name...)
	crc := crc32.NewIEEE()
	crc.Write(header[4:])
	crc.Write(data)

	cw.write(header)
	cw.write(data)
	cw.write(binary.BigEndian.AppendUint32(nil, crc.Sum32()))
}
