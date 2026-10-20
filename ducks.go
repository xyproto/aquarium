package main

import (
	"fmt"
	"math/rand"

	"github.com/gdamore/tcell/v2"
)

var duckFramesRight = [][]string{
	{
		"      _          _          _  ",
		",____(')=  ,____(')=  ,____(')<",
		" \\~~= ')    \\~~= ')    \\~~= ') ",
	},
	{
		"      _          _          _  ",
		",____(')=  ,____(')<  ,____(')=",
		" \\~~= ')    \\~~= ')    \\~~= ') ",
	},
	{
		"      _          _          _  ",
		",____(')<  ,____(')=  ,____(')=",
		" \\~~= ')    \\~~= ')    \\~~= ') ",
	},
}
var duckFramesLeft = [][]string{
	{
		"  _          _          _",
		">(')____,  =(')____,  =(')____,",
		" (` =~~/    (` =~~/    (` =~~/",
	},
	{
		"  _          _          _",
		"=(')____,  >(')____,  =(')____,",
		" (` =~~/    (` =~~/    (` =~~/",
	},
	{
		"  _          _          _",
		"=(')____,  =(')____,  >(')____,",
		" (` =~~/    (` =~~/    (` =~~/",
	},
}
var duckColorRight = []string{
	"      g          g          g",
	"wwwwwgcgy  wwwwwgcgy  wwwwwgcgy",
	" wwww Ww    wwww Ww    wwww Ww",
}
var duckColorLeft = []string{
	"  g          g          g",
	"ygcgwwwww  ygcgwwwww  ygcgwwwww",
	" wW wwww    wW wwww    wW wwww",
}

func addDucks(anim *Animation) {
	dir := rand.Intn(2)
	speed := 1.0
	var startX int
	var frames [][]string
	var baseColor []string

	if dir == 0 {
		startX = -32
		frames = duckFramesRight
		baseColor = duckColorRight
	} else {
		speed = -speed
		startX = anim.Width() - 2
		frames = duckFramesLeft
		baseColor = duckColorLeft
	}

	colors := make([][]string, len(frames))
	for i := range colors {
		colors[i] = baseColor
	}

	anim.NewEntity(&Entity{
		Name:         fmt.Sprintf("ducks_%d", rand.Int63()),
		Type:         "ducks",
		Shapes:       frames,
		Colors:       colors,
		Position:     [3]int{startX, 5, DepthWaterGap3},
		DefaultColor: tcell.ColorWhite,
		DieOffscreen: true,
		VX:           speed,
		FrameDelay:   4,
		DeathCB:      func(_ *Entity, a *Animation) { randomObject(a) },
		UpdateFn: func(e *Entity, a *Animation) {
			e.FloatX += e.VX
			e.Position[0] = int(e.FloatX)
		},
	})
}
