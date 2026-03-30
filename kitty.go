package main

import (
	"encoding/base64"
	"fmt"
	"image"
	"image/color"
	"math"
	"os"
	"strings"

	"github.com/gdamore/tcell/v2"
)

// isKittyTerminal returns true when running inside the Kitty terminal emulator.
func isKittyTerminal() bool {
	return strings.Contains(os.Getenv("TERM"), "kitty") ||
		os.Getenv("TERM_PROGRAM") == "kitty"
}

// makeKittyColorMap returns a richer, 24-bit RGB color map for use inside Kitty.
func makeKittyColorMap() map[byte]tcell.Color {
	return map[byte]tcell.Color{
		// bright tropical palette
		'c': tcell.NewRGBColor(0, 220, 220),   // vivid cyan
		'C': tcell.NewRGBColor(0, 140, 160),   // teal
		'r': tcell.NewRGBColor(255, 60, 60),   // coral red
		'R': tcell.NewRGBColor(180, 20, 20),   // deep red
		'y': tcell.NewRGBColor(255, 230, 0),   // golden yellow
		'Y': tcell.NewRGBColor(200, 160, 0),   // amber
		'b': tcell.NewRGBColor(80, 140, 255),  // sky blue
		'B': tcell.NewRGBColor(20, 60, 200),   // deep blue
		'g': tcell.NewRGBColor(60, 230, 80),   // lime green
		'G': tcell.NewRGBColor(20, 140, 40),   // forest green
		'm': tcell.NewRGBColor(220, 80, 220),  // magenta
		'M': tcell.NewRGBColor(140, 20, 160),  // purple
		'w': tcell.NewRGBColor(255, 255, 255), // white
		'W': tcell.NewRGBColor(180, 180, 200), // silver/grey
		'k': tcell.NewRGBColor(0, 8, 24),      // ocean black
		'K': tcell.NewRGBColor(0, 8, 24),
	}
}

// ── Kitty Graphics Protocol ───────────────────────────────────────────────────

// KittyRenderer manages pixel-art fish overlaid using the Kitty graphics
// protocol.  After each tcell frame (ASCII art) is shown, we place small
// pixel-art images over the fish positions.  Kitty images are NOT erased by
// normal ANSI clear-screen sequences, so we explicitly delete them each frame.
type KittyRenderer struct {
	enabled bool
	// Pre-baked image IDs uploaded once at startup.
	// id layout: variant 0..N go right; N..2N go left (mirrored).
	imageIDs   []int // one ID per fish color variant
	fishColors []fishPalette
	tty        *os.File
}

type fishPalette struct {
	body, tail, fin, eye color.NRGBA
}

func NewKittyRenderer() *KittyRenderer {
	return &KittyRenderer{
		enabled: isKittyTerminal(),
		fishColors: []fishPalette{
			{body: color.NRGBA{255, 140, 0, 255}, tail: color.NRGBA{200, 80, 0, 255}, fin: color.NRGBA{255, 200, 0, 255}, eye: color.NRGBA{0, 0, 0, 255}},
			{body: color.NRGBA{0, 180, 220, 255}, tail: color.NRGBA{0, 100, 180, 255}, fin: color.NRGBA{100, 220, 255, 255}, eye: color.NRGBA{0, 0, 0, 255}},
			{body: color.NRGBA{80, 220, 80, 255}, tail: color.NRGBA{20, 160, 20, 255}, fin: color.NRGBA{180, 255, 100, 255}, eye: color.NRGBA{0, 0, 0, 255}},
			{body: color.NRGBA{220, 60, 220, 255}, tail: color.NRGBA{140, 20, 160, 255}, fin: color.NRGBA{255, 140, 255, 255}, eye: color.NRGBA{0, 0, 0, 255}},
			{body: color.NRGBA{255, 80, 80, 255}, tail: color.NRGBA{180, 20, 20, 255}, fin: color.NRGBA{255, 180, 100, 255}, eye: color.NRGBA{0, 0, 0, 255}},
			{body: color.NRGBA{255, 220, 60, 255}, tail: color.NRGBA{200, 160, 0, 255}, fin: color.NRGBA{255, 255, 120, 255}, eye: color.NRGBA{0, 0, 0, 255}},
		},
	}
}

// Init uploads all fish images to the Kitty terminal at startup.
func (k *KittyRenderer) Init() {
	if !k.enabled {
		return
	}

	// Open /dev/tty so writes are independent of tcell's stdout buffering.
	tty, err := os.OpenFile("/dev/tty", os.O_WRONLY, 0)
	if err != nil {
		// Fall back to stdout
		tty = os.Stdout
	}
	k.tty = tty

	// Upload right-facing and left-facing variants for each palette.
	id := 1
	for _, pal := range k.fishColors {
		img := generateFishImage(64, 20, pal, false)
		k.uploadImage(id, img)
		k.imageIDs = append(k.imageIDs, id)
		id++
		imgL := generateFishImage(64, 20, pal, true)
		k.uploadImage(id, imgL)
		k.imageIDs = append(k.imageIDs, id)
		id++
	}
}

// Render deletes previous frame's images and places new ones over fish.
func (k *KittyRenderer) Render(anim *Animation) {
	if !k.enabled || k.tty == nil {
		return
	}

	// Delete all current image placements (images stay uploaded, placements removed)
	fmt.Fprintf(k.tty, "\x1b_Ga=d,d=A,q=2\x1b\\")

	entities := anim.GetEntities()
	numPalettes := len(k.fishColors)

	for _, e := range entities {
		if e.Type != "fish" {
			continue
		}
		if e.Position[0]+2 > anim.Width() || e.Position[0] < -2 {
			continue
		}
		if e.Position[1]+1 > anim.Height() || e.Position[1] < 0 {
			continue
		}

		// Pick image id based on entity name hash and direction
		nameHash := 0
		for _, c := range e.Name {
			nameHash += int(c)
		}
		palIdx := nameHash % numPalettes
		goesRight := e.VX >= 0
		var imgID int
		if goesRight {
			imgID = k.imageIDs[palIdx*2]
		} else {
			imgID = k.imageIDs[palIdx*2+1]
		}

		fw, fh := e.Size()
		if fw < 1 {
			fw = 1
		}
		if fh < 1 {
			fh = 1
		}

		// Clamp to screen bounds
		px := e.Position[0]
		py := e.Position[1]
		cols := fw
		rows := fh

		// Clip left
		if px < 0 {
			cols += px
			px = 0
		}
		// Clip right
		if px+cols > anim.Width() {
			cols = anim.Width() - px
		}
		if cols <= 0 || rows <= 0 {
			continue
		}

		// Position cursor and place image
		fmt.Fprintf(k.tty, "\x1b[%d;%dH", py+1, px+1)
		fmt.Fprintf(k.tty, "\x1b_Ga=p,i=%d,c=%d,r=%d,q=2\x1b\\", imgID, cols, rows)
	}
}

// Cleanup removes all Kitty images when the program exits.
func (k *KittyRenderer) Cleanup() {
	if !k.enabled || k.tty == nil {
		return
	}
	// Delete all uploaded images
	fmt.Fprintf(k.tty, "\x1b_Ga=d,d=A,q=2\x1b\\")
	for _, id := range k.imageIDs {
		fmt.Fprintf(k.tty, "\x1b_Ga=d,i=%d,q=2\x1b\\", id)
	}
	if k.tty != os.Stdout {
		k.tty.Close()
	}
}

// uploadImage transmits an image to Kitty using raw RGBA chunked encoding.
func (k *KittyRenderer) uploadImage(id int, img *image.NRGBA) {
	b := img.Bounds()
	w, h := b.Dx(), b.Dy()

	// Raw RGBA bytes
	raw := make([]byte, w*h*4)
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			c := img.NRGBAAt(x, y)
			off := ((y-b.Min.Y)*w + (x - b.Min.X)) * 4
			raw[off], raw[off+1], raw[off+2], raw[off+3] = c.R, c.G, c.B, c.A
		}
	}

	enc := base64.StdEncoding.EncodeToString(raw)
	const chunkSize = 4096
	first := true
	for len(enc) > 0 {
		chunk := enc
		more := 0
		if len(enc) > chunkSize {
			chunk = enc[:chunkSize]
			enc = enc[chunkSize:]
			more = 1
		} else {
			enc = ""
		}
		if first {
			// a=t: transmit; f=32: RGBA; s=width; v=height; i=id; m=more
			fmt.Fprintf(k.tty, "\x1b_Ga=t,f=32,s=%d,v=%d,i=%d,q=2,m=%d;%s\x1b\\",
				w, h, id, more, chunk)
			first = false
		} else {
			fmt.Fprintf(k.tty, "\x1b_Gm=%d;%s\x1b\\", more, chunk)
		}
	}
}

// ── Procedural pixel-art fish ─────────────────────────────────────────────────

// generateFishImage creates a small pixel-art fish facing right (or left if
// mirrored=true).  The image uses a transparent background (A=0).
func generateFishImage(w, h int, pal fishPalette, mirrored bool) *image.NRGBA {
	img := image.NewNRGBA(image.Rect(0, 0, w, h))

	// Body: ellipse occupying the right 60 % of the canvas
	cx := int(float64(w) * 0.62)
	cy := h / 2
	rx := float64(w) * 0.30
	ry := float64(h) * 0.42

	// Tail vertices (left side, triangle)
	tailTip := int(float64(w) * 0.04)
	tailBase := int(float64(w) * 0.30)

	// Dorsal fin vertices
	dorsalBaseL := int(float64(w) * 0.45)
	dorsalBaseR := int(float64(w) * 0.65)
	dorsalPeak := int(float64(h) * 0.12)
	dorsalPeakX := int(float64(w) * 0.55)

	// Draw tail
	for y := range h {
		dy := math.Abs(float64(y-cy)) / float64(cy)
		maxX := tailBase - int(dy*float64(tailBase-tailTip))
		for x := tailTip; x <= maxX && x < w; x++ {
			setPixel(img, x, y, pal.tail, mirrored, w)
		}
	}

	// Draw body
	for y := range h {
		for x := range w {
			dx := float64(x-cx) / rx
			dy := float64(y-cy) / ry
			dist := dx*dx + dy*dy
			if dist <= 1.0 {
				// Slight gradient: lighter toward top-right
				shade := 1.0 - dist*0.25
				c := scaleColor(pal.body, shade)
				setPixel(img, x, y, c, mirrored, w)
			}
		}
	}

	// Draw dorsal fin (filled triangle above body)
	for x := dorsalBaseL; x <= dorsalBaseR; x++ {
		// Fraction along the fin base
		t := float64(x-dorsalBaseL) / float64(dorsalBaseR-dorsalBaseL+1)
		// Peak is at dorsalPeakX
		tPeak := float64(dorsalPeakX-dorsalBaseL) / float64(dorsalBaseR-dorsalBaseL+1)
		var yTop int
		if t < tPeak {
			yTop = cy - int((t/tPeak)*float64(cy-dorsalPeak))
		} else {
			yTop = cy - int(((1-t)/(1-tPeak))*float64(cy-dorsalPeak))
		}
		yBottom := cy - int(ry*math.Sqrt(1-math.Pow(float64(x-cx)/rx, 2)))
		for y := yTop; y <= yBottom && y >= 0 && y < h; y++ {
			if img.NRGBAAt(x, y).A == 0 {
				setPixel(img, x, y, pal.fin, mirrored, w)
			}
		}
	}

	// Draw eye
	eyeX := cx + int(rx*0.4)
	eyeY := cy - int(ry*0.2)
	for dy := -2; dy <= 2; dy++ {
		for dx := -2; dx <= 2; dx++ {
			if dx*dx+dy*dy <= 4 {
				setPixel(img, eyeX+dx, eyeY+dy, pal.eye, mirrored, w)
			}
		}
	}
	// Pupil highlight
	setPixel(img, eyeX-1, eyeY-1, color.NRGBA{255, 255, 255, 200}, mirrored, w)

	// Mouth
	mouthX := cx + int(rx*0.85)
	mouthY := cy
	for dy := -1; dy <= 1; dy++ {
		setPixel(img, mouthX, mouthY+dy, color.NRGBA{200, 50, 50, 255}, mirrored, w)
	}

	// Scale / stripe detail
	for i := range 3 {
		sx := cx - int(rx*0.4) + i*int(rx*0.35)
		for dy := -int(ry * 0.4); dy <= int(ry*0.4); dy++ {
			y := cy + dy
			if y < 0 || y >= h {
				continue
			}
			x := sx
			if img.NRGBAAt(x, y).A > 0 {
				setPixel(img, x, y, darken(pal.body, 0.75), mirrored, w)
			}
		}
	}

	return img
}

func setPixel(img *image.NRGBA, x, y int, c color.NRGBA, mirrored bool, w int) {
	if mirrored {
		x = w - 1 - x
	}
	if x < 0 || x >= img.Bounds().Dx() || y < 0 || y >= img.Bounds().Dy() {
		return
	}
	img.SetNRGBA(x, y, c)
}

func scaleColor(c color.NRGBA, f float64) color.NRGBA {
	clamp := func(v float64) uint8 {
		if v < 0 {
			return 0
		}
		if v > 255 {
			return 255
		}
		return uint8(v)
	}
	return color.NRGBA{
		R: clamp(float64(c.R) * f),
		G: clamp(float64(c.G) * f),
		B: clamp(float64(c.B) * f),
		A: c.A,
	}
}

func darken(c color.NRGBA, f float64) color.NRGBA {
	return scaleColor(c, f)
}
