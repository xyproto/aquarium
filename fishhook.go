package main

import (
	"fmt"
	"math/rand"

	"github.com/gdamore/tcell/v2"
)

var fishhookShape = []string{
	"       o",
	"      ||",
	"      ||",
	"/ \\   ||",
	"  \\__//",
	"  `--' ",
}

var fishhookPointShape = []string{
	".",
	" ",
	"\\",
	" ",
}

const (
	fishlineLength   = 50
	fishhookMaxWait  = 300
	fishhookWaitRand = 200
)

func addFishhook(anim *Animation) {
	x := 10 + rand.Intn(max(anim.Width()-20, 1))

	hookName := fmt.Sprintf("fishhook_%d", rand.Int63())
	lineName := hookName + "_line"
	pointName := hookName + "_point"

	retracting := false
	waited := 0
	maxWait := fishhookMaxWait + rand.Intn(fishhookWaitRand)

	lineShape := make([]string, fishlineLength)
	for i := range lineShape {
		lineShape[i] = "|"
	}

	anim.NewEntity(&Entity{
		Name:         lineName,
		Type:         "fishline",
		Shape:        lineShape,
		Position:     [3]int{x + 7, -4 - fishlineLength, DepthWaterLine1},
		DefaultColor: tcell.ColorWhite,
	})

	anim.NewEntity(&Entity{
		Name:         pointName,
		Type:         "hook_point",
		Shape:        fishhookPointShape,
		Position:     [3]int{x + 1, -2, DepthShark + 1},
		DefaultColor: tcell.GetColor("green"),
		Physical:     true,
	})

	anim.NewEntity(&Entity{
		Name:         hookName,
		Type:         "fishhook",
		Shape:        fishhookShape,
		Position:     [3]int{x, -4, DepthWaterLine1},
		DefaultColor: tcell.GetColor("green"),
		DieOffscreen: true,
		DeathCB: func(e *Entity, a *Animation) {
			a.RemoveEntity(lineName)
			a.RemoveEntity(pointName)
			randomObject(a)
		},
		UpdateFn: func(e *Entity, a *Animation) {
			if retracting {
				e.FloatY--
			} else if e.Position[1]+len(e.Shape) < int(float64(a.Height())*0.75) {
				e.FloatY++
			} else {
				waited++
				if waited >= maxWait {
					retracting = true
				}
			}
			e.Position[1] = int(e.FloatY)

			if line := a.GetEntity(lineName); line != nil {
				line.Position[1] = e.Position[1] - fishlineLength
				line.FloatY = float64(line.Position[1])
			}

			point := a.GetEntity(pointName)
			if point == nil {
				return
			}
			point.Position[0] = e.Position[0] + 1
			point.Position[1] = e.Position[1] + 2
			point.FloatY = float64(point.Position[1])

			if retracting {
				return
			}
			px, py := point.Position[0], point.Position[1]
			for _, fish := range a.GetEntitiesOfType("fish") {
				fw, fh := fish.Size()
				fx, fy := fish.Position[0], fish.Position[1]
				if px >= fx && px < fx+fw && py >= fy && py < fy+fh {
					retracting = true
					fish.Type = "hooked_fish"
					fish.Physical = false
					fish.VX = 0
					fish.VY = 0
					fish.UpdateFn = func(fe *Entity, fa *Animation) {
						fe.FloatY--
						fe.Position[1] = int(fe.FloatY)
					}
					break
				}
			}
		},
	})
}
