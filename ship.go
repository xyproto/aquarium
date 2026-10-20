package main

import (
	"fmt"
	"math/rand"

	"github.com/gdamore/tcell/v2"
)

var shipShapeRight = []string{
	"     |    |    |",
	"    )_)  )_)  )_)",
	"   )___))___))___)\\",
	"  )____)____)_____)\\\\ ",
	"_____|____|____|____\\\\\\\\__",
	"\\                   /",
}
var shipColorRight = []string{
	"     y    y    y",
	"",
	"                  w",
	"                   ww",
	"yyyyyyyyyyyyyyyyyyyywwwyy",
	"y                   y",
}

var shipShapeLeft = []string{
	"         |    |    |",
	"        (_(  (_(  (_(",
	"      /(___((___((___((",
	"    //(_____(____(____((",
	"__///____|____|____|_____",
	"    \\                   /",
}
var shipColorLeft = []string{
	"         y    y    y",
	"",
	"      w",
	"    ww",
	"yywwwyyyyyyyyyyyyyyyyyyyy",
	"    y                   y",
}

func addShip(anim *Animation) {
	dir := rand.Intn(2)
	speed := 1.0
	var startX int
	var shape, colorMask []string

	if dir == 0 {
		startX = -26
		shape = shipShapeRight
		colorMask = shipColorRight
	} else {
		speed = -speed
		startX = anim.Width() - 2
		shape = shipShapeLeft
		colorMask = shipColorLeft
	}

	anim.NewEntity(&Entity{
		Name:         fmt.Sprintf("ship_%d", rand.Int63()),
		Type:         "ship",
		Shape:        shape,
		Color:        colorMask,
		Position:     [3]int{startX, 0, DepthWaterGap1},
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
