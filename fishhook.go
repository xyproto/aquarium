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

func addFishhook(anim *Animation) {
	x := max(10+rand.Intn(anim.Width()-20), 0)

	hookName := fmt.Sprintf("fishhook_%d", rand.Int63())
	lineName := hookName + "_line"
	pointName := hookName + "_point"

	hooked := false // becomes true when a fish is caught

	// Fishing line (long vertical bar above the hook)
	lineShape := make([]string, 60)
	for i := range lineShape {
		if i < 50 {
			lineShape[i] = "|"
		} else {
			lineShape[i] = " "
		}
	}

	anim.NewEntity(&Entity{
		Name:         lineName,
		Type:         "fishline",
		Shape:        lineShape,
		Position:     [3]int{x + 7, -50, DepthWaterLine1},
		DefaultColor: tcell.ColorWhite,
		UpdateFn: func(e *Entity, a *Animation) {
			if hooked {
				e.FloatY--
			} else {
				if e.Position[1]+len(e.Shape) < int(float64(a.Height())*0.75) {
					e.FloatY++
				}
			}
			e.Position[1] = int(e.FloatY)
		},
	})

	anim.NewEntity(&Entity{
		Name:         pointName,
		Type:         "hook_point",
		Shape:        fishhookPointShape,
		Position:     [3]int{x + 1, -2, DepthShark + 1},
		DefaultColor: tcell.GetColor("green"),
		Physical:     true,
		UpdateFn: func(e *Entity, a *Animation) {
			if hooked {
				e.FloatY--
			} else {
				hook := a.GetEntity(hookName)
				if hook != nil {
					e.Position[0] = hook.Position[0] + 1
					e.FloatY = float64(hook.Position[1] + 2)
				}
			}
			e.Position[1] = int(e.FloatY)

			// Check for fish collision
			if !hooked {
				for _, fish := range a.GetEntitiesOfType("fish") {
					fw, fh := fish.Size()
					fx, fy := fish.Position[0], fish.Position[1]
					px, py := e.Position[0], e.Position[1]
					if px >= fx && px < fx+fw && py >= fy && py < fy+fh {
						hooked = true
						// Retract the caught fish with the hook
						fish.Physical = false
						fish.DieOffscreen = false
						fish.VX = 0
						fish.VY = 0
						fish.UpdateFn = func(fe *Entity, fa *Animation) {
							fe.FloatY--
							fe.Position[1] = int(fe.FloatY)
							if fe.Position[1] < 0 {
								fa.RemoveEntity(fe.Name)
							}
						}
						break
					}
				}
			}
		},
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
			if hooked {
				e.FloatY--
			} else {
				if e.Position[1]+len(e.Shape) < int(float64(a.Height())*0.75) {
					e.FloatY++
				}
			}
			e.Position[1] = int(e.FloatY)
		},
	})
}
