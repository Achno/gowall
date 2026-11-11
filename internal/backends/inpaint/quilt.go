package inpaint

import (
	"cmp"
	"fmt"
	"image"
	"math"
	"math/rand/v2"
	"slices"

	"github.com/Achno/gowall/internal/backends/pixels"
	"github.com/disintegration/imaging"
)

// Quilt fills the hole's shading smoothly from its edges and covers its texture with blocks
// copied from around it (Efros & Freeman image quilting).
type Quilt struct{}

const (
	qBlock   = 32
	qOverlap = 8
	qTries   = 600
)

func (Quilt) Context(hole, bounds image.Rectangle) image.Rectangle {
	return hole.Inset(-max(hole.Dx(), hole.Dy(), 4*qBlock)).Intersect(bounds)
}

func (Quilt) Fill(img *image.NRGBA, hole []bool) (*image.NRGBA, error) {
	if !slices.Contains(hole, false) {
		return nil, fmt.Errorf("the area covers everything around it, there's nothing left to fill it from")
	}
	w, h := img.Bounds().Dx(), img.Bounds().Dy()
	out := imaging.Clone(img)
	for i, isHole := range hole {
		if out.Pix[4*i+3] = 255; isHole {
			out.Pix[4*i+3] = 0
		}
	}
	small := imaging.Blur(imaging.Resize(out, max(w/4, 1), max(h/4, 1), imaging.Box), 6)
	sv, shole := make([]float32, len(small.Pix)), make([]bool, len(small.Pix)/4)
	for i := range small.Pix {
		sv[i], shole[i/4] = float32(small.Pix[i]), small.Pix[i|3] == 0
	}
	membrane(sv, small.Rect.Dx(), small.Rect.Dy(), shole)
	for i := range small.Pix {
		small.Pix[i] = uint8(min(max(math.Round(float64(sv[i])), 0), 255))
	}
	shading := imaging.Resize(small, w, h, imaging.Linear)

	q := &quilter{w: w, h: h, hole: hole, filled: make([]bool, w*h), tex: make([]float32, 4*w*h), shade: make([]float32, 4*w*h)}
	pixels.ParallelRows(h, func(y0, y1 int) {
		for i := 4 * y0 * w; i < 4*y1*w; i++ {
			q.shade[i] = float32(shading.Pix[i])
			if q.filled[i/4] = !hole[i/4]; q.filled[i/4] {
				q.tex[i] = float32(img.Pix[i]) - q.shade[i]
			}
		}
	})
	// every 2nd pixel of the texture and shading packed together, sources are only looked up in here
	q.hw, q.hh = w/2, h/2
	q.half, q.sat = make([]float32, 6*q.hw*q.hh), make([]int32, (q.hw+1)*(q.hh+1))
	w1 := q.hw + 1
	pixels.ParallelRows(q.hh, func(y0, y1 int) {
		for y := y0; y < y1; y++ {
			for x := range q.hw {
				i := 2*y*w + 2*x
				if hole[i] || hole[i+1] || hole[i+w] || hole[i+w+1] {
					q.sat[(y+1)*w1+x+1] = 1
				}
				copy(q.half[6*(y*q.hw+x):], q.tex[4*i:4*i+3])
				copy(q.half[6*(y*q.hw+x)+3:], q.shade[4*i:4*i+3])
			}
		}
	})
	for y := range q.hh {
		for x := range q.hw {
			q.sat[(y+1)*w1+x+1] += q.sat[y*w1+x+1] + q.sat[(y+1)*w1+x] - q.sat[y*w1+x]
		}
	}
	q.run()

	pixels.ParallelRows(h, func(y0, y1 int) {
		for i := 4 * y0 * w; i < 4*y1*w; i++ {
			out.Pix[i] = img.Pix[i]
			if i%4 != 3 && hole[i/4] {
				out.Pix[i] = uint8(min(max(math.Round(float64(q.shade[i]+q.tex[i])), 0), 255))
			}
		}
	})
	return out, nil
}

// membrane fills the hole smoothly from its edges, solving on a half size copy first.
func membrane(v []float32, w, h int, hole []bool) {
	if !slices.Contains(hole, true) || !slices.Contains(hole, false) {
		return
	}
	if w > 1 && h > 1 {
		cw, ch := (w+1)/2, (h+1)/2
		cv, n, chole := make([]float32, 4*cw*ch), make([]float32, cw*ch), make([]bool, cw*ch)
		for i := range hole {
			if ci := i/w/2*cw + i%w/2; !hole[i] {
				for c := range 4 {
					cv[4*ci+c] += v[4*i+c]
				}
				n[ci]++
			}
		}
		for i := range cv {
			cv[i], chole[i/4] = cv[i]/max(n[i/4], 1), n[i/4] == 0
		}
		membrane(cv, cw, ch, chole)
		for i := range hole {
			if ci := i/w/2*cw + i%w/2; hole[i] {
				copy(v[4*i:4*i+4], cv[4*ci:])
			}
		}
	}
	bb := image.Rectangle{Min: image.Pt(w, h)}
	for i, isHole := range hole {
		if isHole {
			bb = bb.Union(image.Rect(i%w, i/w, i%w+1, i/w+1))
		}
	}
	for r := range 100 { // checkerboard, so neighbours are always the other colour and rows can run at once
		pixels.ParallelRows(bb.Dy(), func(y0, y1 int) {
			for y := bb.Min.Y + y0; y < bb.Min.Y+y1; y++ {
				for x := bb.Min.X + (bb.Min.X+y+r)%2; x < bb.Max.X; x += 2 {
					if !hole[y*w+x] {
						continue
					}
					var sum [4]float32
					var n float32
					for _, p := range [4]image.Point{{x - 1, y}, {x + 1, y}, {x, y - 1}, {x, y + 1}} {
						if p.X >= 0 && p.Y >= 0 && p.X < w && p.Y < h {
							for c := range 4 {
								sum[c] += v[4*(p.Y*w+p.X)+c]
							}
							n++
						}
					}
					for c := range 4 {
						v[4*(y*w+x)+c] = sum[c] / n
					}
				}
			}
		})
	}
}

type quilter struct {
	w, h, hw, hh int
	hole, filled []bool
	tex, shade   []float32
	half         []float32
	sat          []int32
}

type target struct {
	tex, shade [qBlock * qBlock / 4 * 3]float32
	filled     [qBlock * qBlock / 4]bool
}

func (q *quilter) run() {
	if q.w < qBlock || q.h < qBlock {
		return
	}
	hb := image.Rectangle{Min: image.Pt(q.w, q.h)}
	for i, isHole := range q.hole {
		if isHole {
			hb = hb.Union(image.Rect(i%q.w, i/q.w, i%q.w+1, i/q.w+1))
		}
	}
	var blocks []image.Point
	for y := hb.Min.Y - qOverlap; y < hb.Max.Y; y += qBlock - qOverlap {
		for x := hb.Min.X - qOverlap; x < hb.Max.X; x += qBlock - qOverlap {
			blocks = append(blocks, image.Pt(min(max(x, 0), q.w-qBlock), min(max(y, 0), q.h-qBlock)))
		}
	}
	dist := func(p image.Point) int {
		c := p.Add(image.Pt(qBlock/2, qBlock/2))
		return min(c.X-hb.Min.X, hb.Max.X-c.X, c.Y-hb.Min.Y, hb.Max.Y-c.Y)
	}
	slices.SortStableFunc(blocks, func(a, b image.Point) int { return dist(a) - dist(b) })

	// a block waits for the earlier ones it overlaps, the ones in the same wave don't overlap so
	// they're filled at the same time
	var waves [][]int
	wave := make([]int, len(blocks))
	for j, b := range blocks {
		for i := range j {
			if c := blocks[i].Sub(b); max(c.X, -c.X, c.Y, -c.Y) < qBlock {
				wave[j] = max(wave[j], wave[i]+1)
			}
		}
		if wave[j] == len(waves) {
			waves = append(waves, nil)
		}
		waves[wave[j]] = append(waves[wave[j]], j)
	}
	for _, batch := range waves {
		pixels.ParallelRows(len(batch), func(k0, k1 int) {
			for _, j := range batch[k0:k1] {
				b, todo := blocks[j], false
				for y := b.Y; y < b.Y+qBlock && !todo; y++ {
					todo = slices.Contains(q.filled[y*q.w+b.X:y*q.w+b.X+qBlock], false)
				}
				if !todo {
					continue
				}
				if src, ok := q.pick(b, rand.New(rand.NewPCG(3, uint64(j)))); ok {
					q.paste(b, src)
				}
			}
		})
	}
}

func (q *quilter) pick(p image.Point, rng *rand.Rand) (image.Point, bool) {
	var t target
	for y := range qBlock / 2 {
		for x := range qBlock / 2 {
			i, k := (p.Y+2*y)*q.w+p.X+2*x, y*qBlock/2+x
			copy(t.tex[3*k:3*k+3], q.tex[4*i:])
			copy(t.shade[3*k:3*k+3], q.shade[4*i:])
			t.filled[k] = q.filled[i]
		}
	}
	srcs, costs := make([]image.Point, qTries), make([]float32, qTries)
	best := float32(math.MaxFloat32)
	for k := range srcs {
		srcs[k] = image.Pt(rng.IntN(q.hw-qBlock/2+1), rng.IntN(q.hh-qBlock/2+1))
		if k%2 == 0 {
			srcs[k] = image.Pt(min(max(p.X/2+rng.IntN(4*qBlock)-2*qBlock, 0), q.hw-qBlock/2), min(max(p.Y/2+rng.IntN(4*qBlock)-2*qBlock, 0), q.hh-qBlock/2))
		}
		costs[k] = q.cost(&t, srcs[k], best*1.15)
		best = min(best, costs[k])
	}
	if best == math.MaxFloat32 {
		return image.Point{}, false
	}
	var good []image.Point
	for k, d := range costs {
		if d <= best*1.15 {
			good = append(good, srcs[k].Mul(2))
		}
	}
	return good[rng.IntN(len(good))], true
}

func (q *quilter) cost(t *target, s image.Point, limit float32) (d float32) {
	const n = qBlock / 2
	if w1 := q.hw + 1; q.sat[(s.Y+n)*w1+s.X+n]-q.sat[s.Y*w1+s.X+n]-q.sat[(s.Y+n)*w1+s.X]+q.sat[s.Y*w1+s.X] > 0 {
		return math.MaxFloat32
	}
	for y := 0; y < n && d < limit; y++ {
		row := q.half[6*((s.Y+y)*q.hw+s.X):]
		for x := range n {
			k, o := y*n+x, 6*x
			if x%2 == 0 && y%2 == 0 {
				for c := range 3 {
					d += 2 * (t.shade[3*k+c] - row[o+3+c]) * (t.shade[3*k+c] - row[o+3+c])
				}
			}
			if t.filled[k] {
				for c := range 3 {
					d += (t.tex[3*k+c] - row[o+c]) * (t.tex[3*k+c] - row[o+c])
				}
			}
		}
	}
	return d
}

// paste copies src into the block at p, keeping earlier blocks up to the cheapest seam on each side.
func (q *quilter) paste(p, src image.Point) {
	var keep [qBlock][qBlock]bool
	for _, at := range [4]func(i, j int) (int, int){
		func(i, j int) (int, int) { return j, i },
		func(i, j int) (int, int) { return qBlock - 1 - j, i },
		func(i, j int) (int, int) { return i, j },
		func(i, j int) (int, int) { return i, qBlock - 1 - j },
	} {
		var acc [qBlock][qOverlap]float32
		for i := range qBlock {
			for j := range qOverlap {
				x, y := at(i, j)
				if t := (p.Y+y)*q.w + p.X + x; q.filled[t] && q.hole[t] {
					for c, o := 0, (src.Y+y)*q.w+src.X+x; c < 3; c++ {
						acc[i][j] += (q.tex[4*t+c] - q.tex[4*o+c]) * (q.tex[4*t+c] - q.tex[4*o+c])
					}
				}
				if i > 0 {
					acc[i][j] += min(acc[i-1][j], acc[i-1][max(j-1, 0)], acc[i-1][min(j+1, qOverlap-1)])
				}
			}
		}
		j := slices.Index(acc[qBlock-1][:], slices.Min(acc[qBlock-1][:]))
		for i := qBlock - 1; i >= 0; i-- {
			for k := range j {
				x, y := at(i, k)
				keep[y][x] = true
			}
			if i > 0 {
				j = slices.MinFunc([]int{j, max(j-1, 0), min(j+1, qOverlap-1)}, func(a, b int) int { return cmp.Compare(acc[i-1][a], acc[i-1][b]) })
			}
		}
	}
	for y := range qBlock {
		for x := range qBlock {
			if t := (p.Y+y)*q.w + p.X + x; q.hole[t] && !(q.filled[t] && keep[y][x]) {
				copy(q.tex[4*t:4*t+3], q.tex[4*((src.Y+y)*q.w+src.X+x):])
				q.filled[t] = true
			}
		}
	}
}
