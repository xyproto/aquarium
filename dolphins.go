package main

import (
	"fmt"
	"math/rand"

	"github.com/gdamore/tcell/v2"
)

var dolphinShapeRight = [][]string{
	{
		"        ,",
		"      __)\\_ ",
		"(\\_.-'    a`-.",
		"(/~~````(/~^^`",
	},
	{
		"        ,",
		"(\\__  __)\\_ ",
		"(/~.''    a`-.",
		"    ````\\)~^^`",
	},
}
var dolphinShapeLeft = [][]string{
	{
		"     ,",
		"   _/(__",
		".-'a    `-._/)",
		"'^^~\\)''''~~\\)",
	},
	{
		"     ,",
		"   _/(__  __/)",
		".-'a    ``.~\\)",
		"'^^~(/''''",
	},
}
var dolphinColorRight = []string{
	"",
	"",
	"          W",
	"",
}
var dolphinColorLeft = []string{
	"",
	"",
	"   W",
	"",
}

func addDolphins(anim *Animation) {
	dir := rand.Intn(2)
	speed := 1.0
	distance := 15

	if dir == 1 {
		speed = -speed
		distance = -distance
	}

	var startX int
	if dir == 0 {
		startX = -15
	} else {
		startX = anim.Width() - 2
	}

	// Build shape frames and colors (cycle through 2 animation frames)
	var frames [][]string
	var fcolors [][]string
	var baseShape [][]string
	var baseColor []string

	if dir == 0 {
		baseShape = dolphinShapeRight
		baseColor = dolphinColorRight
	} else {
		baseShape = dolphinShapeLeft
		baseColor = dolphinColorLeft
	}
	for _, f := range baseShape {
		frames = append(frames, f)
		fcolors = append(fcolors, baseColor)
	}

	// Build jumping path: up 14 steps, glide 2, down 14 steps, glide 6
	var path [][2]float64
	for range 14 {
		path = append(path, [2]float64{speed, -0.5})
	}
	for range 2 {
		path = append(path, [2]float64{speed, 0})
	}
	for range 14 {
		path = append(path, [2]float64{speed, 0.5})
	}
	for range 6 {
		path = append(path, [2]float64{speed, 0})
	}

	makeDolphin := func(name string, x, y, pathOffset int, dieOffscreen bool, deathCB func(*Entity, *Animation)) *Entity {
		return &Entity{
			Name:         name,
			Type:         "dolphin",
			Shapes:       frames,
			Colors:       fcolors,
			Position:     [3]int{x, y, DepthWaterGap3},
			DefaultColor: tcell.GetColor("cyan"),
			DieOffscreen: dieOffscreen,
			FrameDelay:   3,
			DeathCB:      deathCB,
			Path:         path,
			PathIndex:    pathOffset % len(path),
			UpdateFn: func(e *Entity, a *Animation) {
				step := e.Path[e.PathIndex%len(e.Path)]
				e.FloatX += step[0]
				e.FloatY += step[1]
				e.Position[0] = int(e.FloatX)
				e.Position[1] = int(e.FloatY)
				e.PathIndex++
			},
		}
	}

	d3Name := fmt.Sprintf("dolphin3_%d", rand.Int63())
	d2Name := fmt.Sprintf("dolphin2_%d", rand.Int63())
	d1Name := fmt.Sprintf("dolphin1_%d", rand.Int63())

	d3 := makeDolphin(d3Name, startX-2*distance, 8, 0, false, nil)
	d2 := makeDolphin(d2Name, startX-distance, 2, 12, false, nil)
	d1 := makeDolphin(d1Name, startX, 5, 24, true, func(_ *Entity, a *Animation) {
		// When lead dolphin exits, enable die-offscreen for trailing ones
		if d3e := a.GetEntity(d3Name); d3e != nil {
			d3e.DieOffscreen = true
		}
		if d2e := a.GetEntity(d2Name); d2e != nil {
			d2e.DieOffscreen = true
		}
		randomObject(a)
	})

	anim.NewEntity(d3)
	anim.NewEntity(d2)
	anim.NewEntity(d1)
}
