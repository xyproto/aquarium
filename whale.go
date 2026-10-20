package main

import (
	"fmt"
	"math/rand"

	"github.com/gdamore/tcell/v2"
)

var whaleBodyRight = []string{
	"        .-----:",
	"      .'       `.",
	",    /       (o) \\",
	"\\`._/          ,__)",
}
var whaleBodyLeft = []string{
	"    :-----.     ",
	"  .'       `.   ",
	" / (o)       \\    ",
	"(__,          \\_.'/ ",
}
var whaleColorRight = []string{
	"             c c",
	"           ccccccc",
	"           c  c  c",
	"        BBBBBBB",
}
var whaleColorLeft = []string{
	"   c c",
	" ccccccc",
	" c  c  c",
	"    BBBBBBB",
}

// Water spout frames (prepended above the whale body)
var spoutFrames = [][]string{
	{"", "", ""},
	{"", "", ""},
	{"", "", ""},
	{"", "", ""},
	{"", "", ""},
	{"   :", "", ""},
	{"   :", "   :", ""},
	{"  . .", "  -:-", "   :"},
	{"  . .", " .-:-.", "   :"},
	{"  . .", "'.-:-.'", "'  :  '"},
	{"", " .- -.", ";  :  ;"},
	{"", "", ";     ;"},
}

func buildWhaleFrames(body []string, color []string, spoutAlign int) ([][]string, [][]string) {
	var frames [][]string
	var fcolors [][]string
	pad := fmt.Sprintf("%*s", spoutAlign, "")
	for _, spout := range spoutFrames {
		frame := make([]string, 0, len(spout)+len(body))
		for _, sl := range spout {
			if sl == "" {
				frame = append(frame, "")
			} else {
				frame = append(frame, pad+sl)
			}
		}
		frame = append(frame, body...)
		frames = append(frames, frame)
		fcolors = append(fcolors, color)
	}
	return frames, fcolors
}

func addWhale(anim *Animation) {
	dir := rand.Intn(2)
	speed := 1.0
	var startX int
	var frames [][]string
	var fcolors [][]string

	if dir == 0 { // right
		startX = -20
		frames, fcolors = buildWhaleFrames(whaleBodyRight, whaleColorRight, 3)
	} else { // left
		speed = -speed
		startX = anim.Width() - 2
		frames, fcolors = buildWhaleFrames(whaleBodyLeft, whaleColorLeft, 1)
	}

	anim.NewEntity(&Entity{
		Name:         fmt.Sprintf("whale_%d", rand.Int63()),
		Type:         "whale",
		Shapes:       frames,
		Colors:       fcolors,
		Position:     [3]int{startX, 0, DepthWaterGap2},
		DefaultColor: tcell.ColorWhite,
		DieOffscreen: true,
		VX:           speed,
		FrameDelay:   5,
		DeathCB:      func(_ *Entity, a *Animation) { randomObject(a) },
		UpdateFn: func(e *Entity, a *Animation) {
			e.FloatX += e.VX
			e.Position[0] = int(e.FloatX)
		},
	})
}
