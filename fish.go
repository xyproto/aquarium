package main

import (
	"fmt"
	"math/rand"

	"github.com/gdamore/tcell/v2"
)

func addFish(anim *Animation) {
	if len(fishVariants) == 0 {
		return
	}

	// Pick a random pair; pair index determines direction (even=right, odd=left)
	variantIdx := rand.Intn(len(fishVariants))
	variant := fishVariants[variantIdx]
	colors := randColor(variant.Color)

	// Speed: matches Perl rand(2)+0.25 at 10 fps
	speed := rand.Float64()*2.0 + 0.25
	goesRight := variantIdx%2 == 0
	if !goesRight {
		speed = -speed
	}

	// Scratch entity to measure size
	probe := &Entity{Shape: variant.Shape}
	fishW, fishH := probe.Size()

	var startX int
	if goesRight {
		startX = 1 - fishW // just off left edge
	} else {
		startX = anim.Width() - 2 // just off right edge
	}

	maxY := anim.Height() - fishH
	minY := 9
	if maxY <= minY {
		maxY = minY + 1
	}
	startY := rand.Intn(maxY-minY) + minY

	depth := DepthFishStart + rand.Intn(DepthFishEnd-DepthFishStart)

	e := &Entity{
		Name:         fmt.Sprintf("fish_%d", rand.Int63()),
		Type:         "fish",
		Shape:        variant.Shape,
		Color:        colors,
		Position:     [3]int{startX, startY, depth},
		DefaultColor: tcell.ColorWhite,
		Physical:     true,
		DieOffscreen: true,
		DeathCB:      func(_ *Entity, a *Animation) { addFish(a) },
		VX:           speed,
		UpdateFn:     fishUpdate,
	}
	anim.NewEntity(e)
}

func fishUpdate(e *Entity, a *Animation) {
	e.FloatX += e.VX
	e.Position[0] = int(e.FloatX)

	// Occasionally emit a bubble (3 % chance per tick, like Perl)
	if rand.Intn(100) < 3 {
		addBubble(e, a)
	}
}
