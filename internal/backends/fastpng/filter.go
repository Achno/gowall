package fastpng

// The PNG filter types
const (
	ftNone    = 0
	ftSub     = 1
	ftUp      = 2
	ftAverage = 3
	ftPaeth   = 4
)

// filter is a copy of image/png's: it tries every filter on the row cr[0] (pr is the row above) and returns the one
// with the smallest sum of absolute differences, its filtered row is in cr[filter].
func filter(cr *[5][]byte, pr []byte, bpp int) int {
	cdat0 := cr[0][1:]
	cdat1 := cr[1][1:]
	cdat2 := cr[2][1:]
	cdat3 := cr[3][1:]
	cdat4 := cr[4][1:]
	pdat := pr[1:]
	n := len(cdat0)

	// The up filter.
	sum := 0
	for i := range n {
		cdat2[i] = cdat0[i] - pdat[i]
		sum += abs8(cdat2[i])
	}
	best := sum
	filter := ftUp

	// The Paeth filter.
	sum = 0
	for i := range bpp {
		cdat4[i] = cdat0[i] - pdat[i]
		sum += abs8(cdat4[i])
	}
	for i := bpp; i < n; i++ {
		cdat4[i] = cdat0[i] - paeth(cdat0[i-bpp], pdat[i], pdat[i-bpp])
		sum += abs8(cdat4[i])
		if sum >= best {
			break
		}
	}
	if sum < best {
		best = sum
		filter = ftPaeth
	}

	// The none filter.
	sum = 0
	for i := range n {
		sum += abs8(cdat0[i])
		if sum >= best {
			break
		}
	}
	if sum < best {
		best = sum
		filter = ftNone
	}

	// The sub filter.
	sum = 0
	for i := range bpp {
		cdat1[i] = cdat0[i]
		sum += abs8(cdat1[i])
	}
	for i := bpp; i < n; i++ {
		cdat1[i] = cdat0[i] - cdat0[i-bpp]
		sum += abs8(cdat1[i])
		if sum >= best {
			break
		}
	}
	if sum < best {
		best = sum
		filter = ftSub
	}

	// The average filter.
	sum = 0
	for i := range bpp {
		cdat3[i] = cdat0[i] - pdat[i]/2
		sum += abs8(cdat3[i])
	}
	for i := bpp; i < n; i++ {
		cdat3[i] = cdat0[i] - uint8((int(cdat0[i-bpp])+int(pdat[i]))/2)
		sum += abs8(cdat3[i])
		if sum >= best {
			break
		}
	}
	if sum < best {
		filter = ftAverage
	}

	return filter
}

// paeth implements the Paeth predictor of the PNG spec
func paeth(a, b, c uint8) uint8 {
	pc := int(c)
	pa := int(b) - pc
	pb := int(a) - pc
	pc = abs(pa + pb)
	pa = abs(pa)
	pb = abs(pb)
	if pa <= pb && pa <= pc {
		return a
	} else if pb <= pc {
		return b
	}
	return c
}

// abs8 is the absolute value of a byte interpreted as a signed int8
func abs8(d uint8) int {
	if d < 128 {
		return int(d)
	}
	return 256 - int(d)
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
