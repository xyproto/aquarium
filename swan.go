package main

import (
	"fmt"
	"math/rand"

	"github.com/gdamore/tcell/v2"
)

var swanShapeRight = []string{
	"       ___",
	",_    / _,\\",
	"| \\   \\( \\|",
	"|  \\_  \\\\\\",
	"(_   \\_) \\",
	"(\\_   `   \\",
	" \\   -=~  /",
}
var swanColorRight = []string{
	"",
	"         g",
	"         yy",
	"",
	"",
	"",
	"",
}

var swanShapeLeft = []string{
	" ___",
	"/,_ \\    _,",
	"|/ )/   / |",
	"  //  _/  |",
	" / ( /   _)",
	"/   `   _/)",
	"\\  ~=-   /",
}
var swanColorLeft = []string{
	"",
	" g",
	"yy",
	"",
	"",
	"",
	"",
}

func addSwan(anim *Animation) {
	dir := rand.Intn(2)
	speed := 1.0
	var startX int
	var shape, colorMask []string

	if dir == 0 {
		startX = -12
		shape = swanShapeRight
		colorMask = swanColorRight
	} else {
		speed = -speed
		startX = anim.Width() - 2
		shape = swanShapeLeft
		colorMask = swanColorLeft
	}

	anim.NewEntity(&Entity{
		Name:         fmt.Sprintf("swan_%d", rand.Int63()),
		Type:         "swan",
		Shape:        shape,
		Color:        colorMask,
		Position:     [3]int{startX, 1, DepthWaterGap3},
		DefaultColor: tcell.ColorWhite,
		DieOffscreen: true,
		VX:           speed,
		DeathCB:      func(_ *Entity, a *Animation) { randomObject(a) },
		UpdateFn: func(e *Entity, a *Animation) {
			e.FloatX += e.VX
			e.Position[0] = int(e.FloatX)
		},
	})
}
