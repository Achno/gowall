package terminal

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"os"

	"github.com/Achno/gowall/internal/backends/fastpng"
	"golang.org/x/term"
)

const (
	// alt screen, hide cursor, clear, report mouse presses/drags in SGR format
	enterKittyScreen = "\x1b[?1049h\x1b[?25l\x1b[2J\x1b[?1002h\x1b[?1006h"
	// undoes enterKittyScreen and requestPixelMouse in reverse order
	leaveKittyScreen = "\x1b[?1016l\x1b[?1006l\x1b[?1002l\x1b[?25h\x1b[?1049l"
	// asks for mouse positions in px instead of cells, then asks if the terminal took it.
	// Every terminal answers the device attributes query at the end, so we know when to stop reading.
	requestPixelMouse = "\x1b[?1016h\x1b[?1016$p\x1b[c"
	moveCursor        = "\x1b[%d;%dH" // row, col starting from 1
	clearLine         = "\x1b[2K"

	// kitty image commands, q=2 stops the terminal from replying on stdin
	transmitFirstChunk = "\x1b_Ga=t,f=100,i=%d,q=2,m=%d;%s\x1b\\"
	transmitNextChunk  = "\x1b_Gm=%d;%s\x1b\\"
	placeImage         = "\x1b_Ga=p,i=%d,p=1,z=0,C=1,q=2\x1b\\"
	placeCropped       = "\x1b_Ga=p,i=%d,p=1,w=%d,z=1,C=1,q=2\x1b\\"                  // only the first w px
	placeShifted       = "\x1b_Ga=p,i=%d,p=1,X=%d,z=2,C=1,q=2\x1b\\"                  // X px to the right inside the cell
	placeRegion        = "\x1b_Ga=p,i=%d,p=%d,w=%d,h=%d,X=%d,Y=%d,z=%d,C=1,q=2\x1b\\" // the top left w x h px, X,Y px inside the cell
	hideImage          = "\x1b_Ga=d,d=i,i=%d,q=2\x1b\\"                               // removes the placements, keeps the image
	hidePlacement      = "\x1b_Ga=d,d=i,i=%d,p=%d,q=2\x1b\\"                          // removes one placement, keeps the image
	deleteImage        = "\x1b_Ga=d,d=I,i=%d,q=2\x1b\\"                               // removes the placements and the image

	keyLeft    = "\x1b[D"
	keyRight   = "\x1b[C"
	mouseEvent = "\x1b[<" // ESC [ < btn ; x ; y M|m
)

// kittyScreen is the terminal switched to the alt screen in raw mode, ready for kitty image placements.
type kittyScreen struct {
	fd                    int
	oldState              *term.State
	imageIDs              []int
	pxWidth, pxHeight     int
	cellWidth, cellHeight int
	pixelMouse            bool
}

// openKittyScreen switches to the alt screen in raw mode. close deletes imageIDs and restores the terminal.
func openKittyScreen(imageIDs ...int) (*kittyScreen, error) {
	supportsKitty := IsKittyTerminalRunning() || IsWeztermTerminalRunning() || IsKonsoleTerminalRunning() || IsGhosttyTerminalRunning()
	if !supportsKitty {
		return nil, fmt.Errorf("this needs a terminal with the kitty image protocol: kitty, wezterm, konsole or ghostty")
	}
	fd := int(os.Stdin.Fd())
	if !term.IsTerminal(fd) {
		return nil, fmt.Errorf("this needs an interactive terminal")
	}
	rows, cols, pxWidth, pxHeight, err := getTerminalDimensions()
	if err != nil {
		return nil, fmt.Errorf("while getting terminal dimensions: %w", err)
	}
	if rows == 0 || cols == 0 || pxWidth == 0 || pxHeight == 0 {
		return nil, fmt.Errorf("the terminal doesn't report its size in pixels")
	}

	oldState, err := term.MakeRaw(fd)
	if err != nil {
		return nil, err
	}
	os.Stdout.WriteString(enterKittyScreen)

	return &kittyScreen{
		fd:         fd,
		oldState:   oldState,
		imageIDs:   imageIDs,
		pxWidth:    pxWidth,
		pxHeight:   pxHeight,
		cellWidth:  pxWidth / cols,
		cellHeight: pxHeight / rows,
		pixelMouse: enablePixelMouse(),
	}, nil
}

func (k *kittyScreen) close() {
	var b bytes.Buffer
	for _, id := range k.imageIDs {
		fmt.Fprintf(&b, deleteImage, id)
	}
	b.WriteString(leaveKittyScreen)
	os.Stdout.Write(b.Bytes())
	term.Restore(k.fd, k.oldState)
}

// enablePixelMouse asks for mouse positions in px and returns whether the terminal took it.
func enablePixelMouse() bool {
	os.Stdout.WriteString(requestPixelMouse)
	var reply []byte
	buf := make([]byte, 64)
	for !bytes.HasSuffix(reply, []byte("c")) {
		n, err := os.Stdin.Read(buf)
		if err != nil {
			return false
		}
		reply = append(reply, buf[:n]...)
	}
	return bytes.Contains(reply, []byte("1016;1$y")) || bytes.Contains(reply, []byte("1016;3$y"))
}

// parseMouse reads the mouse report at the start of in, x and y are px from the top left.
// pressed is false when the button was released, n is how many bytes it used and 0 when the report is incomplete.
func (k *kittyScreen) parseMouse(in []byte) (btn, x, y int, pressed bool, n int) {
	end := bytes.IndexAny(in, "Mm")
	if end < 0 {
		return 0, 0, 0, false, 0
	}
	fmt.Sscanf(string(in[len(mouseEvent):end]), "%d;%d;%d", &btn, &x, &y)
	if k.pixelMouse {
		x, y = x-1, y-1
	} else {
		x, y = (x-1)*k.cellWidth, (y-1)*k.cellHeight
	}
	return btn, x, y, in[end] == 'M', end + 1
}

// placeArea places the top left of image id over r (in px), or hides that placement when r is empty.
func (k *kittyScreen) placeArea(b *bytes.Buffer, id, placement, z int, r image.Rectangle) {
	if r.Empty() {
		fmt.Fprintf(b, hidePlacement, id, placement)
		return
	}
	fmt.Fprintf(b, moveCursor+placeRegion, r.Min.Y/k.cellHeight+1, r.Min.X/k.cellWidth+1,
		id, placement, r.Dx(), r.Dy(), r.Min.X%k.cellWidth, r.Min.Y%k.cellHeight, z)
}

// readInput passes everything typed to onInput until it returns true.
func readInput(onInput func(in []byte) bool) error {
	buf := make([]byte, 256)
	for {
		n, err := os.Stdin.Read(buf)
		if err != nil {
			return err
		}
		if onInput(buf[:n]) {
			return nil
		}
	}
}

// transmit uploads img as a png in 4096 byte chunks, the protocol's max chunk size.
func transmit(w *bytes.Buffer, id int, img image.Image) error {
	var p bytes.Buffer
	if err := fastpng.Encode(&p, img); err != nil {
		return fmt.Errorf("encoding PNG: %w", err)
	}
	data := base64.StdEncoding.EncodeToString(p.Bytes())
	for first := true; len(data) > 0; first = false {
		chunk := data[:min(4096, len(data))]
		data = data[len(chunk):]
		more := 0
		if len(data) > 0 {
			more = 1
		}
		if first {
			fmt.Fprintf(w, transmitFirstChunk, id, more, chunk)
		} else {
			fmt.Fprintf(w, transmitNextChunk, more, chunk)
		}
	}
	return nil
}
