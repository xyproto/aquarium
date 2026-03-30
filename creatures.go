package main

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/gdamore/tcell/v2"
)

// addBubble adds a bubble rising from a fish
func addBubble(fish *Entity, anim *Animation) {
	if len(fish.CallbackArgs) < 1 {
		return
	}

	fishW, fishH := fish.Size()
	x := fish.Position[0]
	y := fish.Position[1]

	// Bubble goes on right or left depending on direction
	if fish.CallbackArgs[0] > 0 {
		x += fishW
	}
	y += fishH / 2

	e := &Entity{
		Name:         fmt.Sprintf("bubble_%d_%d", rand.Int63(), time.Now().UnixNano()),
		Shape:        []string{".", "o", "O", "O", "O"},
		Position:     [3]int{x, y, fish.Position[2] - 1},
		DefaultColor: tcell.GetColor("cyan"),
		DieOffscreen: true,
		CallbackArgs: []float64{},
		UpdateFn:     updateBubble,
	}
	anim.NewEntity(e)
}

// updateBubble moves bubble upward
func updateBubble(e *Entity, anim *Animation) {
	// Move upward
	e.Position[1]--
	if e.Position[1] < 0 {
		anim.RemoveEntity(e.Name)
	}
}

// fishCallback handles fish and creature animation
func fishCallback(e *Entity, anim *Animation) {
	if len(e.CallbackArgs) < 1 {
		return
	}

	speed := e.CallbackArgs[0]

	// Move horizontally only
	e.Position[0] += int(speed)

	// Wrap around screen
	w, _ := e.Size()
	if e.Position[0] > anim.Width() {
		e.Position[0] = -w
	} else if e.Position[0] < -w {
		e.Position[0] = anim.Width()
	}

	// Occasionally add bubble (3% chance per frame)
	if rand.Float64() < 0.03 {
		addBubble(e, anim)
	}
}

// addFish adds a standard fish with random direction and speed
func addFish(anim *Animation) {
	if len(fishVariants) == 0 {
		return
	}

	variant := fishVariants[rand.Intn(len(fishVariants))]

	// Randomize colors for this specific fish
	colors := randColor(variant.Color)

	// Pick a speed (always positive, direction from position)
	speed := rand.Float64()*1.75 + 0.25
	fishNum := rand.Intn(1000)

	// Create temp entity to get proper dimensions
	tempEntity := &Entity{Shape: variant.Shape}
	fishW, fishH := tempEntity.Size()

	// Placement matching Perl: Odd-numbered fish move left, even move right
	var x int
	if fishNum%2 == 1 {
		speed *= -1          // Negative speed = move left
		x = anim.Width() - 2 // Start from right edge
	} else {
		x = 1 - fishW // Start off-screen on left
	}

	// Y placement: int(rand(screen_height - fish_height - 9)) + 9
	maxHeight := 9
	minHeight := anim.Height() - fishH
	y := rand.Intn(minHeight-maxHeight) + maxHeight

	depth := DepthFishStart + rand.Intn(DepthFishEnd-DepthFishStart)

	e := &Entity{
		Name:         fmt.Sprintf("fish_%d", rand.Int63()),
		Shape:        variant.Shape,
		Color:        colors,
		Position:     [3]int{x, y, depth},
		DefaultColor: tcell.GetColor("white"),
		Physical:     true,
		CallbackArgs: []float64{speed},
		UpdateFn:     fishCallback,
		DieTime:      time.Now().Add(time.Duration(rand.Intn(5)+10) * time.Minute),
	}
	anim.NewEntity(e)
}

// addShark adds a shark entity
func addShark(anim *Animation) {
	if len(sharkVariants) == 0 {
		return
	}

	variant := sharkVariants[rand.Intn(len(sharkVariants))]
	colors := randColor(variant.Color)
	speed := rand.Float64()*0.3 + 0.7
	fishNum := rand.Intn(1000)

	var x int
	if fishNum%2 == 1 {
		speed *= -1
		x = anim.Width() - 2
	} else {
		x = 1
	}

	y := anim.Height() / 2

	e := &Entity{
		Name:         fmt.Sprintf("shark_%d", rand.Int63()),
		Shape:        variant.Shape,
		Color:        colors,
		Position:     [3]int{x, y, DepthShark},
		DefaultColor: tcell.GetColor("red"),
		Physical:     true,
		CallbackArgs: []float64{speed},
		UpdateFn:     fishCallback,
		DieTime:      time.Now().Add(time.Duration(rand.Intn(10)+15) * time.Minute),
	}
	anim.NewEntity(e)
}

// addWhale adds the whale creature
func addWhale(anim *Animation) {
	colors := randColor(whaleVariant.Color)
	speed := rand.Float64()*0.1 + 0.2
	fishNum := rand.Intn(1000)

	var x int
	if fishNum%2 == 1 {
		speed *= -1
		x = anim.Width() - 2
	} else {
		x = 1
	}

	y := anim.Height()/2 - 5

	e := &Entity{
		Name:         fmt.Sprintf("whale_%d", rand.Int63()),
		Shape:        whaleVariant.Shape,
		Color:        colors,
		Position:     [3]int{x, y, DepthFishStart + 5},
		DefaultColor: tcell.GetColor("cyan"),
		Physical:     true,
		CallbackArgs: []float64{speed},
		UpdateFn:     fishCallback,
		DieTime:      time.Now().Add(time.Duration(rand.Intn(5)+8) * time.Minute),
	}
	anim.NewEntity(e)
}

// addMonster adds the monster creature
func addMonster(anim *Animation) {
	colors := randColor(monsterVariant.Color)
	speed := rand.Float64()*0.15 + 0.3
	fishNum := rand.Intn(1000)

	var x int
	if fishNum%2 == 1 {
		speed *= -1
		x = anim.Width() - 2
	} else {
		x = 1
	}

	y := anim.Height()/2 - 5

	e := &Entity{
		Name:         fmt.Sprintf("monster_%d", rand.Int63()),
		Shape:        monsterVariant.Shape,
		Color:        colors,
		Position:     [3]int{x, y, DepthFishStart + 3},
		DefaultColor: tcell.GetColor("magenta"),
		Physical:     true,
		CallbackArgs: []float64{speed},
		UpdateFn:     fishCallback,
		DieTime:      time.Now().Add(time.Duration(rand.Intn(6)+10) * time.Minute),
	}
	anim.NewEntity(e)
}

// addBigFish adds a bigger fish
func addBigFish(anim *Animation) {
	colors := randColor(bigFishVariant.Color)
	speed := rand.Float64()*0.4 + 0.4
	fishNum := rand.Intn(1000)

	var x int
	if fishNum%2 == 1 {
		speed *= -1
		x = anim.Width() - 2
	} else {
		x = 1
	}

	y := 10 + rand.Intn(anim.Height()-15)

	e := &Entity{
		Name:         fmt.Sprintf("bigfish_%d", rand.Int63()),
		Shape:        bigFishVariant.Shape,
		Color:        colors,
		Position:     [3]int{x, y, DepthFishStart + 2},
		DefaultColor: tcell.GetColor("yellow"),
		Physical:     true,
		CallbackArgs: []float64{speed},
		UpdateFn:     fishCallback,
		DieTime:      time.Now().Add(time.Duration(rand.Intn(8)+12) * time.Minute),
	}
	anim.NewEntity(e)
}

// addDucks adds ducks swimming above water
func addDucks(anim *Animation) {
	colors := randColor(ducksVariant.Color)
	speed := rand.Float64()*0.2 + 0.3
	fishNum := rand.Intn(1000)

	var x int
	if fishNum%2 == 1 {
		speed *= -1
		x = anim.Width() - 2
	} else {
		x = 1
	}

	y := 2

	e := &Entity{
		Name:         fmt.Sprintf("ducks_%d", rand.Int63()),
		Shape:        ducksVariant.Shape,
		Color:        colors,
		Position:     [3]int{x, y, DepthFishStart + 1},
		DefaultColor: tcell.GetColor("yellow"),
		CallbackArgs: []float64{speed},
		UpdateFn:     fishCallback,
		DieTime:      time.Now().Add(time.Duration(rand.Intn(6)+8) * time.Minute),
	}
	anim.NewEntity(e)
}

// addDolphins adds dolphins
func addDolphins(anim *Animation) {
	colors := randColor(dolphinsVariant.Color)
	speed := rand.Float64()*0.3 + 0.5
	fishNum := rand.Intn(1000)

	var x int
	if fishNum%2 == 1 {
		speed *= -1
		x = anim.Width() - 2
	} else {
		x = 1
	}

	y := anim.Height()/2 - 3

	e := &Entity{
		Name:         fmt.Sprintf("dolphins_%d", rand.Int63()),
		Shape:        dolphinsVariant.Shape,
		Color:        colors,
		Position:     [3]int{x, y, DepthFishStart + 4},
		DefaultColor: tcell.GetColor("cyan"),
		CallbackArgs: []float64{speed},
		UpdateFn:     fishCallback,
		DieTime:      time.Now().Add(time.Duration(rand.Intn(7)+10) * time.Minute),
	}
	anim.NewEntity(e)
}

// addSwan adds a swan
func addSwan(anim *Animation) {
	colors := randColor(swanVariant.Color)
	speed := rand.Float64()*0.25 + 0.3
	fishNum := rand.Intn(1000)

	var x int
	if fishNum%2 == 1 {
		speed *= -1
		x = anim.Width() - 2
	} else {
		x = 1
	}

	y := 3

	e := &Entity{
		Name:         fmt.Sprintf("swan_%d", rand.Int63()),
		Shape:        swanVariant.Shape,
		Color:        colors,
		Position:     [3]int{x, y, DepthFishStart},
		DefaultColor: tcell.GetColor("white"),
		CallbackArgs: []float64{speed},
		UpdateFn:     fishCallback,
		DieTime:      time.Now().Add(time.Duration(rand.Intn(8)+10) * time.Minute),
	}
	anim.NewEntity(e)
}

// addFishhook adds a fishing hook from above
func addFishhook(anim *Animation) {
	colors := randColor(fishhookVariant.Color)
	x := rand.Intn(anim.Width())
	y := 0

	e := &Entity{
		Name:         fmt.Sprintf("fishhook_%d", rand.Int63()),
		Shape:        fishhookVariant.Shape,
		Color:        colors,
		Position:     [3]int{x, y, DepthFishStart - 1},
		DefaultColor: tcell.GetColor("white"),
		UpdateFn: func(e *Entity, a *Animation) {
			e.Position[1]++
			if e.Position[1] > a.Height() {
				a.RemoveEntity(e.Name)
			}
		},
		DieTime: time.Now().Add(time.Duration(rand.Intn(4)+6) * time.Minute),
	}
	anim.NewEntity(e)
}

// addShip adds a ship sailing above the aquarium
func addShip(anim *Animation) {
	colors := randColor(shipVariant.Color)
	speed := rand.Float64()*0.15 + 0.2
	fishNum := rand.Intn(1000)

	var x int
	if fishNum%2 == 1 {
		speed *= -1
		x = anim.Width() - 2
	} else {
		x = 1
	}

	y := 0

	e := &Entity{
		Name:         fmt.Sprintf("ship_%d", rand.Int63()),
		Shape:        shipVariant.Shape,
		Color:        colors,
		Position:     [3]int{x, y, DepthFishStart - 2},
		DefaultColor: tcell.GetColor("yellow"),
		CallbackArgs: []float64{speed},
		UpdateFn:     fishCallback,
		DieTime:      time.Now().Add(time.Duration(rand.Intn(5)+10) * time.Minute),
	}
	anim.NewEntity(e)
}

// RandomObjectPool holds spawning functions
var randomObjectPool = []func(*Animation){
	addFish,
	addShark,
	addWhale,
	addMonster,
	addBigFish,
	addDucks,
	addDolphins,
	addSwan,
	addFishhook,
	addShip,
}

// randomObject spawns a random creature from the pool
func randomObject(anim *Animation) {
	if len(randomObjectPool) == 0 {
		return
	}
	fn := randomObjectPool[rand.Intn(len(randomObjectPool))]
	fn(anim)
}

// randColor applies random colors to a color mask
func randColor(colorMask []string) []string {
	colors := []rune{'c', 'C', 'r', 'R', 'y', 'Y', 'b', 'B', 'g', 'G', 'm', 'M'}

	// Assign one color per digit (1-9), matching Perl behavior
	colorMap := make(map[rune]rune)
	for digit := '1'; digit <= '9'; digit++ {
		colorMap[digit] = colors[rand.Intn(len(colors))]
	}

	// Special: 4 always maps to white
	colorMap['4'] = 'W'

	result := make([]string, len(colorMask))
	for i, line := range colorMask {
		chars := []rune(line)
		for j, ch := range chars {
			if replacement, ok := colorMap[ch]; ok {
				chars[j] = replacement
			}
		}
		result[i] = string(chars)
	}

	return result
}
