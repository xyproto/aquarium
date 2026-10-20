package main

import (
	"fmt"
	"math/rand"

	"github.com/gdamore/tcell/v2"
)

// Full Perl-style shark.  The body uses ' ' (rendered in cyan) to fill the
// outline – that is the original artistic choice.
var sharkShapeRight = []string{
	"                              __",
	"                             ( `\\",
	"  ,                          )   `\\",
	";' `.                        (     `\\__",
	" ;   `.             __..---''          `~~~~-._",
	"  `.   `.____...--''                       (b  `--._",
	"    >                     _.-'      .((      ._     )",
	"  .`.-`--...__         .-'     -.___.....-(|/|/|/|/'",
	" ;.'         `. ...----`.___.',,,_______......---'",
	" '           '-'",
}
var sharkColorRight = []string{
	"",
	"",
	"",
	"",
	"",
	"                                           cR",
	"",
	"                                          cWWWWWWWW",
	"",
	"",
}

var sharkShapeLeft = []string{
	"                     __",
	"                    /' )",
	"                  /'   (                          ,",
	"              __/'     )                        .' `;",
	"      _.-~~~~'          ``---..__             .'   ;",
	" _.--'  b)                       ``--...____.'   .'",
	"(     _.      )).      `-._                     <",
	" `\\|\\|\\|\\|)-.....___.-     `-.         __...--'-.'.",
	"   `---......_______,,,`.___.'----... .'         `.;",
	"                                     `-`           `",
}
var sharkColorLeft = []string{
	"",
	"",
	"",
	"",
	"",
	"        Rc",
	"",
	"  WWWWWWWWc",
	"",
	"",
}

func addShark(anim *Animation) {
	dir := rand.Intn(2) // 0 = right, 1 = left
	speed := 2.0
	var startX int
	var shape []string
	var colorMask []string
	var teethOffX int

	if dir == 0 { // right-going
		startX = -53
		shape = sharkShapeRight
		colorMask = sharkColorRight
		teethOffX = 43 // mouth column in right-facing shark
	} else { // left-going
		speed = -speed
		startX = anim.Width() - 2
		shape = sharkShapeLeft
		colorMask = sharkColorLeft
		teethOffX = 9
	}

	maxY := max(anim.Height()-10-9, 9)
	startY := rand.Intn(maxY-9+1) + 9

	sharkName := fmt.Sprintf("shark_%d", rand.Int63())
	teethName := sharkName + "_teeth"

	// Teeth: a single '*' at the shark's mouth, physical for fish collision
	teethX := startX + teethOffX
	teethY := startY + 7 // mouth row in both shark variants

	anim.NewEntity(&Entity{
		Name:         teethName,
		Type:         "teeth",
		Shape:        []string{"*"},
		Position:     [3]int{teethX, teethY, DepthShark + 1},
		DefaultColor: tcell.GetColor("cyan"),
		Physical:     true,
		VX:           speed,
		UpdateFn: func(e *Entity, a *Animation) {
			e.FloatX += e.VX
			e.Position[0] = int(e.FloatX)
		},
	})

	anim.NewEntity(&Entity{
		Name:         sharkName,
		Type:         "shark",
		Shape:        shape,
		Color:        colorMask,
		Position:     [3]int{startX, startY, DepthShark},
		DefaultColor: tcell.GetColor("cyan"),
		DieOffscreen: true,
		VX:           speed,
		DeathCB: func(e *Entity, a *Animation) {
			a.RemoveEntity(teethName)
			randomObject(a)
		},
		UpdateFn: func(e *Entity, a *Animation) {
			e.FloatX += e.VX
			e.Position[0] = int(e.FloatX)

			// Check if any physical fish overlaps the teeth
			teeth := a.GetEntity(teethName)
			if teeth == nil {
				return
			}
			tx, ty := teeth.Position[0], teeth.Position[1]
			for _, other := range a.GetEntitiesOfType("fish") {
				fw, fh := other.Size()
				fx, fy := other.Position[0], other.Position[1]
				if tx >= fx && tx < fx+fw && ty >= fy && ty < fy+fh {
					addSplat(anim, fx+fw/2, fy+fh/2, other.Position[2]-1)
					a.RemoveEntity(other.Name)
				}
			}
		},
	})
}
