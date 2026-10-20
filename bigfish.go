package main

import (
	"fmt"
	"math/rand"

	"github.com/gdamore/tcell/v2"
)

var bigFishShapeRight = []string{
	" ______",
	"`\"\"\"-. `````-----....__",
	"     `. .      .       `-.",
	"       :     .     .       `.",
	" ,     :   .    .          _ :",
	": `.   :                  (@) `._",
	" `. `..`     .     =`-.       .__).",
	"   ;     .        =  ~  :     .-\"",
	" .' .'`. .    .  =.-'  `._ .'",
	": .'   :               .   .'",
	" '   .'  .    .     .   .-'",
	"   .'____....----''.'=.'",
	"   \"\"             .'.'",
	"               ''\"'`",
}
var bigFishColorRight = []string{
	" 111111",
	"11111  11111111111111111",
	"     11  2      2       111",
	"       1     2     2       11",
	" 1     1   2    2          1 1",
	"1 11   1                  1W1 111",
	" 11 1111     2     1111       1111",
	"   1     2        1  1  1     111",
	" 11 1111   2    2  1111  111 11",
	"1 11   1               2   11",
	" 1   11  2    2     2   111",
	"   111111111111111111111",
	"   11             1111",
	"               11111",
}

var bigFishShapeLeft = []string{
	"                           ______",
	"          __.....-----'''''  .-\"\"'",
	"       .-'       .      .  .'",
	"     .'       .     .     :",
	"    : _          .    .   :     ,",
	" _.' (@)                  :   .' :",
	"(__.       .-'=     .     `..' .'",
	" \"-.     :  ~  =        .     ;",
	"   `. _.'  `-.=  .    .   .'`. `.",
	"     `.   .               :   `. :",
	"       `-.   .     .    .  `.   `",
	"          `.=`.``----....____`.",
	"            `.`.             \"\"",
	"              '`\"``",
}
var bigFishColorLeft = []string{
	"                           111111",
	"          11111111111111111  11111",
	"       111       2      2  11",
	"     11       2     2     1",
	"    1 1          2    2   1     1",
	" 111 1W1                  1   11 1",
	"1111       1111     2     1111 11",
	" 111     1  1  1        2     1",
	"   11 111  1111  2    2   1111 11",
	"     11   2               1   11 1",
	"       111   2     2    2  11   1",
	"          111111111111111111111",
	"            1111             11",
	"              11111",
}

func addBigFish(anim *Animation) {
	dir := rand.Intn(2)
	speed := 3.0
	var startX int
	var shape, colorMask []string

	if dir == 0 {
		startX = -35
		shape = bigFishShapeRight
		colorMask = bigFishColorRight
	} else {
		speed = -speed
		startX = anim.Width() - 2
		shape = bigFishShapeLeft
		colorMask = bigFishColorLeft
	}

	colors := randColor(colorMask)

	minY := 9
	maxY := anim.Height() - 15
	if maxY <= minY {
		maxY = minY + 1
	}
	startY := rand.Intn(maxY-minY) + minY

	anim.NewEntity(&Entity{
		Name:         fmt.Sprintf("bigfish_%d", rand.Int63()),
		Type:         "bigfish",
		Shape:        shape,
		Color:        colors,
		Position:     [3]int{startX, startY, DepthShark},
		DefaultColor: tcell.GetColor("yellow"),
		DieOffscreen: true,
		VX:           speed,
		DeathCB:      func(_ *Entity, a *Animation) { randomObject(a) },
		UpdateFn: func(e *Entity, a *Animation) {
			e.FloatX += e.VX
			e.Position[0] = int(e.FloatX)
		},
	})
}
