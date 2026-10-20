package main

import (
	"fmt"
	"math/rand"

	"github.com/gdamore/tcell/v2"
)

var monsterFramesRight = [][]string{
	{
		"                                                          ____",
		"            __                                          /   o  \\",
		"          /    \\        _                     _       /     ____ >",
		"  _      |  __  |     /   \\        _        /   \\    |     |",
		" | \\     |  ||  |    |     |     /   \\     |     |   |     |",
	},
	{
		"                                                          ____",
		"                                             __         /   o  \\",
		"             _                     _       /    \\     /     ____ >",
		"   _       /   \\        _        /   \\    |  __  |   |     |",
		"  | \\     |     |     /   \\     |     |   |  ||  |   |     |",
	},
	{
		"                                                          ____",
		"                                  __                    /   o  \\",
		" _                      _       /    \\        _       /     ____ >",
		"| \\          _        /   \\    |  __  |     /   \\    |     |",
		" \\ \\       /   \\     |     |   |  ||  |    |     |   |     |",
	},
	{
		"                                                          ____",
		"                       __                               /   o  \\",
		"  _          _       /    \\        _                  /     ____ >",
		" | \\       /   \\    |  __  |     /   \\        _      |     |",
		"  \\ \\     |     |   |  ||  |    |     |     /   \\    |     |",
	},
}

var monsterFramesLeft = [][]string{
	{
		"    ____",
		"  /  o   \\                                          __",
		"< ____     \\       _                     _        /    \\",
		"      |     |    /   \\        _        /   \\     |  __  |      _",
		"      |     |   |     |     /   \\     |     |    |  ||  |     / |",
	},
	{
		"    ____",
		"  /  o   \\         __",
		"< ____     \\     /    \\       _                     _",
		"      |     |   |  __  |    /   \\        _        /   \\       _",
		"      |     |   |  ||  |   |     |     /   \\     |     |     / |",
	},
	{
		"    ____",
		"  /  o   \\                    __",
		"< ____     \\       _        /    \\       _                      _",
		"      |     |    /   \\     |  __  |    /   \\        _          / |",
		"      |     |   |     |    |  ||  |   |     |     /   \\       / /",
	},
	{
		"    ____",
		"  /  o   \\                               __",
		"< ____     \\                  _        /    \\       _          _",
		"      |     |      _        /   \\     |  __  |    /   \\       / |",
		"      |     |    /   \\     |     |    |  ||  |   |     |     / /",
	},
}

var monsterColorRight = []string{
	"",
	"                                                            W",
	"",
	"",
	"",
}
var monsterColorLeft = []string{
	"",
	"     W",
	"",
	"",
	"",
}

func addMonster(anim *Animation) {
	dir := rand.Intn(2)
	speed := 2.0
	var startX int
	var frames [][]string
	var baseColor []string

	if dir == 0 {
		startX = -65
		frames = monsterFramesRight
		baseColor = monsterColorRight
	} else {
		speed = -speed
		startX = anim.Width() - 2
		frames = monsterFramesLeft
		baseColor = monsterColorLeft
	}

	colors := make([][]string, len(frames))
	for i := range colors {
		colors[i] = baseColor
	}

	anim.NewEntity(&Entity{
		Name:         fmt.Sprintf("monster_%d", rand.Int63()),
		Type:         "monster",
		Shapes:       frames,
		Colors:       colors,
		Position:     [3]int{startX, 2, DepthWaterGap2},
		DefaultColor: tcell.GetColor("green"),
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
