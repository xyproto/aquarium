package main

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/gdamore/tcell/v2"
)

// ── Fish ─────────────────────────────────────────────────────────────────────

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

// ── Bubbles ───────────────────────────────────────────────────────────────────

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

// ── Shark ─────────────────────────────────────────────────────────────────────

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

// ── Splat (fish eaten) ────────────────────────────────────────────────────────

var splatFrames = [][]string{
	{
		"",
		"   .",
		"  ***",
		"   '",
		"",
	},
	{
		"",
		" \",*;`",
		" \"*,**",
		" *\"'~'",
		"",
	},
	{
		"  , ,",
		" \" \",\"'",
		" *\" *'\"",
		"  \" ; .",
		"",
	},
	{
		"* ' , ' `",
		"' ` * . '",
		" ' `' \",'",
		"* ' \" * .",
		"\" * ', '",
	},
}

func addSplat(anim *Animation, x, y, z int) {
	e := &Entity{
		Name:         fmt.Sprintf("splat_%d", rand.Int63()),
		Type:         "splat",
		Shapes:       splatFrames,
		Position:     [3]int{x - 4, y - 2, z},
		DefaultColor: tcell.GetColor("red"),
		FrameDelay:   3,
		DieTime:      time.Now().Add(2 * time.Second),
	}
	anim.NewEntity(e)
}

// ── Ship ──────────────────────────────────────────────────────────────────────

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

// ── Whale ─────────────────────────────────────────────────────────────────────

var whaleBodyRight = []string{
	"        .-----:",
	"      .'       `.",
	",    /       (o) \\",
	"\\`._/          ,__)",
}
var whaleBodyLeft = []string{
	"    :-----.     ",
	"  .'       `.   ",
	" / (o)       \\    ",
	"(__,          \\_.'/ ",
}
var whaleColorRight = []string{
	"             c c",
	"           ccccccc",
	"           c  c  c",
	"        BBBBBBB",
}
var whaleColorLeft = []string{
	"   c c",
	" ccccccc",
	" c  c  c",
	"    BBBBBBB",
}

// Water spout frames (prepended above the whale body)
var spoutFrames = [][]string{
	{"", "", ""},
	{"", "", ""},
	{"", "", ""},
	{"", "", ""},
	{"", "", ""},
	{"   :", "", ""},
	{"   :", "   :", ""},
	{"  . .", "  -:-", "   :"},
	{"  . .", " .-:-.", "   :"},
	{"  . .", "'.-:-.'", "'  :  '"},
	{"", " .- -.", ";  :  ;"},
	{"", "", ";     ;"},
}

func buildWhaleFrames(body []string, color []string, spoutAlign int) ([][]string, [][]string) {
	var frames [][]string
	var fcolors [][]string
	pad := fmt.Sprintf("%*s", spoutAlign, "")
	for _, spout := range spoutFrames {
		frame := make([]string, 0, len(spout)+len(body))
		for _, sl := range spout {
			if sl == "" {
				frame = append(frame, "")
			} else {
				frame = append(frame, pad+sl)
			}
		}
		frame = append(frame, body...)
		frames = append(frames, frame)
		fcolors = append(fcolors, color)
	}
	return frames, fcolors
}

func addWhale(anim *Animation) {
	dir := rand.Intn(2)
	speed := 1.0
	var startX int
	var frames [][]string
	var fcolors [][]string

	if dir == 0 { // right
		startX = -20
		frames, fcolors = buildWhaleFrames(whaleBodyRight, whaleColorRight, 3)
	} else { // left
		speed = -speed
		startX = anim.Width() - 2
		frames, fcolors = buildWhaleFrames(whaleBodyLeft, whaleColorLeft, 1)
	}

	anim.NewEntity(&Entity{
		Name:         fmt.Sprintf("whale_%d", rand.Int63()),
		Type:         "whale",
		Shapes:       frames,
		Colors:       fcolors,
		Position:     [3]int{startX, 0, DepthWaterGap2},
		DefaultColor: tcell.ColorWhite,
		DieOffscreen: true,
		VX:           speed,
		FrameDelay:   5,
		DeathCB:      func(_ *Entity, a *Animation) { randomObject(a) },
		UpdateFn: func(e *Entity, a *Animation) {
			e.FloatX += e.VX
			e.Position[0] = int(e.FloatX)
		},
	})
}

// ── Monster ───────────────────────────────────────────────────────────────────

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

// ── Big Fish ──────────────────────────────────────────────────────────────────

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

// ── Fishhook ──────────────────────────────────────────────────────────────────

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

// ── Ducks ─────────────────────────────────────────────────────────────────────

var duckFramesRight = [][]string{
	{
		"      _          _          _  ",
		",____(')=  ,____(')=  ,____(')<",
		" \\~~= ')    \\~~= ')    \\~~= ') ",
	},
	{
		"      _          _          _  ",
		",____(')=  ,____(')<  ,____(')=",
		" \\~~= ')    \\~~= ')    \\~~= ') ",
	},
	{
		"      _          _          _  ",
		",____(')<  ,____(')=  ,____(')=",
		" \\~~= ')    \\~~= ')    \\~~= ') ",
	},
}
var duckFramesLeft = [][]string{
	{
		"  _          _          _",
		">(')____,  =(')____,  =(')____,",
		" (` =~~/    (` =~~/    (` =~~/",
	},
	{
		"  _          _          _",
		"=(')____,  >(')____,  =(')____,",
		" (` =~~/    (` =~~/    (` =~~/",
	},
	{
		"  _          _          _",
		"=(')____,  =(')____,  >(')____,",
		" (` =~~/    (` =~~/    (` =~~/",
	},
}
var duckColorRight = []string{
	"      g          g          g",
	"wwwwwgcgy  wwwwwgcgy  wwwwwgcgy",
	" wwww Ww    wwww Ww    wwww Ww",
}
var duckColorLeft = []string{
	"  g          g          g",
	"ygcgwwwww  ygcgwwwww  ygcgwwwww",
	" wW wwww    wW wwww    wW wwww",
}

func addDucks(anim *Animation) {
	dir := rand.Intn(2)
	speed := 1.0
	var startX int
	var frames [][]string
	var baseColor []string

	if dir == 0 {
		startX = -32
		frames = duckFramesRight
		baseColor = duckColorRight
	} else {
		speed = -speed
		startX = anim.Width() - 2
		frames = duckFramesLeft
		baseColor = duckColorLeft
	}

	colors := make([][]string, len(frames))
	for i := range colors {
		colors[i] = baseColor
	}

	anim.NewEntity(&Entity{
		Name:         fmt.Sprintf("ducks_%d", rand.Int63()),
		Type:         "ducks",
		Shapes:       frames,
		Colors:       colors,
		Position:     [3]int{startX, 5, DepthWaterGap3},
		DefaultColor: tcell.ColorWhite,
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

// ── Dolphins ──────────────────────────────────────────────────────────────────

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

	d3 := makeDolphin(d3Name, startX-2*distance, 8, 0, false, func(_ *Entity, a *Animation) { randomObject(a) })
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

// ── Swan ─────────────────────────────────────────────────────────────────────

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

// ── Random object pool ────────────────────────────────────────────────────────

var randomObjectFns []func(*Animation)

func init() {
	randomObjectFns = []func(*Animation){
		addShip,
		addWhale,
		addMonster,
		addBigFish,
		addShark,
		addFishhook,
		addSwan,
		addDucks,
		addDolphins,
	}
}

func randomObject(anim *Animation) {
	fn := randomObjectFns[rand.Intn(len(randomObjectFns))]
	fn(anim)
}

// ── Color randomization ───────────────────────────────────────────────────────

// randColor replaces digit placeholders 1-9 in a color mask with random
// color letters, matching the Perl rand_color() function.
// Digit '4' (eye) is always mapped to 'W' (white/gray) first.
func randColor(colorMask []string) []string {
	availColors := []rune{'c', 'C', 'r', 'R', 'y', 'Y', 'b', 'B', 'g', 'G', 'm', 'M'}

	assignment := make(map[rune]rune)
	for d := '1'; d <= '9'; d++ {
		assignment[d] = availColors[rand.Intn(len(availColors))]
	}
	assignment['4'] = 'W' // eye is always white/gray

	result := make([]string, len(colorMask))
	for i, line := range colorMask {
		runes := []rune(line)
		for j, ch := range runes {
			if c, ok := assignment[ch]; ok {
				runes[j] = c
			}
		}
		result[i] = string(runes)
	}
	return result
}
