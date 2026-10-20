package main

import (
	"fmt"
	"math/rand"

	"github.com/gdamore/tcell/v2"
)

func addBubble(fish *Entity, anim *Animation) {
	fishW, fishH := fish.Size()
	bx := fish.Position[0]
	if fish.VX > 0 {
		bx += fishW
	}
	by := fish.Position[1] + fishH/2

	// Bubble rises at 1 cell/tick; animate through frames slowly
	e := &Entity{
		Name: fmt.Sprintf("bubble_%d", rand.Int63()),
		Type: "bubble",
		Shapes: [][]string{
			{"."},
			{"o"},
			{"O"},
			{"O"},
			{"O"},
		},
		Position:     [3]int{bx, by, fish.Position[2] - 1},
		DefaultColor: tcell.GetColor("cyan"),
		DieOffscreen: true,
		VY:           -1.0,
		FrameDelay:   4, // advance animation frame every 4 ticks
		UpdateFn:     bubbleUpdate,
	}
	anim.NewEntity(e)
}

func bubbleUpdate(e *Entity, a *Animation) {
	e.FloatY += e.VY
	e.Position[1] = int(e.FloatY)

	// Pop at the water surface
	if e.Position[1] <= 8 {
		a.RemoveEntity(e.Name)
	}
}
