package main

import (
	"fmt"
	"math/rand"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gdamore/tcell/v2"
)

// Depth constants for Z-ordering
const (
	DepthGUIText   = 0
	DepthGUI       = 1
	DepthShark     = 2
	DepthFishStart = 3
	DepthFishEnd   = 20
	DepthSeaweed   = 21
	DepthCastle    = 22

	DepthWaterLine3 = 2
	DepthWaterGap3  = 3
	DepthWaterLine2 = 4
	DepthWaterGap2  = 5
	DepthWaterLine1 = 6
	DepthWaterGap1  = 7
	DepthWaterLine0 = 8
	DepthWaterGap0  = 9
)

// Entity represents an object in the aquarium
type Entity struct {
	Name         string
	Type         string
	Shape        []string // multi-line ASCII art (or single-char frames for animation)
	Position     [3]int   // x, y, z (depth)
	Color        []string // color mask matching shape
	DefaultColor tcell.Color
	Physical     bool // can collide
	DieOffscreen bool
	DieTime      time.Time
	DeathCB      func(*Entity, *Animation)
	CallbackArgs []float64                 // for animation state (vx, vy, etc)
	UpdateFn     func(*Entity, *Animation) // per-frame update function
	FrameIndex   int                       // for animated entities
	FrameDelay   int                       // frames between animation updates
}

// Size returns the width and height of the entity
func (e *Entity) Size() (int, int) {
	if len(e.Shape) == 0 {
		return 0, 0
	}

	// For single-character animated frames, return (1, 1)
	if len(e.Shape) > 0 && len(e.Shape[0]) == 1 && len(e.Shape) > 3 {
		// This looks like animation frames (multiple single characters)
		return 1, 1
	}

	width := 0
	for _, line := range e.Shape {
		if len(line) > width {
			width = len(line)
		}
	}
	return width, len(e.Shape)
}

// CurrentFrame returns the current animation frame
func (e *Entity) CurrentFrame() string {
	if len(e.Shape) == 0 {
		return ""
	}
	// Check if this is animated (all elements are single characters or very short)
	isAnimated := true
	for _, s := range e.Shape {
		if len(s) > 1 {
			isAnimated = false
			break
		}
	}

	if isAnimated {
		return e.Shape[e.FrameIndex%len(e.Shape)]
	}
	// Not animated, return as multi-line
	return ""
}

// Animation manages the aquarium state
type Animation struct {
	screen      tcell.Screen
	entities    map[string]*Entity
	entityMutex sync.RWMutex
	width       int
	height      int
	paused      bool
	colorMap    map[byte]tcell.Color
	version     string
}

// NewAnimation creates a new animation with terminal
func NewAnimation() (*Animation, error) {
	s, err := tcell.NewScreen()
	if err != nil {
		return nil, err
	}

	if err := s.Init(); err != nil {
		return nil, err
	}

	s.SetStyle(tcell.StyleDefault.Background(tcell.GetColor("black")).Foreground(tcell.GetColor("white")))
	s.Clear()

	w, h := s.Size()
	anim := &Animation{
		screen:   s,
		entities: make(map[string]*Entity),
		width:    w,
		height:   h,
		version:  "1.1",
		colorMap: makeColorMap(),
	}

	return anim, nil
}

// makeColorMap creates the color code to tcell.Color mapping
func makeColorMap() map[byte]tcell.Color {
	return map[byte]tcell.Color{
		'c': tcell.GetColor("cyan"),
		'C': tcell.GetColor("darkcyan"),
		'r': tcell.GetColor("red"),
		'R': tcell.GetColor("maroon"),
		'y': tcell.GetColor("yellow"),
		'Y': tcell.GetColor("olive"),
		'b': tcell.GetColor("blue"),
		'B': tcell.GetColor("navy"),
		'g': tcell.GetColor("green"),
		'G': tcell.GetColor("darkgreen"),
		'm': tcell.GetColor("magenta"),
		'M': tcell.GetColor("purple"),
		'w': tcell.GetColor("white"),
		'W': tcell.GetColor("gray"),
		'k': tcell.GetColor("black"),
		'K': tcell.GetColor("black"),
	}
}

// Width returns the terminal width
func (a *Animation) Width() int {
	return a.width
}

// Height returns the terminal height
func (a *Animation) Height() int {
	return a.height
}

// UpdateTermSize updates width/height from current terminal
func (a *Animation) UpdateTermSize() {
	a.width, a.height = a.screen.Size()
}

// NewEntity creates and adds an entity to the animation
func (a *Animation) NewEntity(e *Entity) {
	a.entityMutex.Lock()
	defer a.entityMutex.Unlock()
	if e.Name == "" {
		e.Name = fmt.Sprintf("entity_%d", len(a.entities))
	}
	a.entities[e.Name] = e
}

// RemoveEntity removes an entity by name
func (a *Animation) RemoveEntity(name string) {
	a.entityMutex.Lock()
	defer a.entityMutex.Unlock()
	delete(a.entities, name)
}

// RemoveAllEntities clears all entities
func (a *Animation) RemoveAllEntities() {
	a.entityMutex.Lock()
	defer a.entityMutex.Unlock()
	a.entities = make(map[string]*Entity)
}

// GetEntities returns a copy of all entities
func (a *Animation) GetEntities() []*Entity {
	a.entityMutex.RLock()
	defer a.entityMutex.RUnlock()
	var result []*Entity
	for _, e := range a.entities {
		result = append(result, e)
	}
	return result
}

// GetEntity retrieves an entity by name
func (a *Animation) GetEntity(name string) *Entity {
	a.entityMutex.RLock()
	defer a.entityMutex.RUnlock()
	return a.entities[name]
}

// Animate updates all entities
func (a *Animation) Animate() {
	if a.paused {
		return
	}

	entities := a.GetEntities()
	now := time.Now()

	// First, update all entities
	for _, e := range entities {
		// Check if entity should die
		if !e.DieTime.IsZero() && now.After(e.DieTime) {
			if e.DeathCB != nil {
				e.DeathCB(e, a)
			}
			a.RemoveEntity(e.Name)
			continue
		}

		// Check if entity went offscreen
		if e.DieOffscreen {
			w, h := e.Size()
			if e.Position[0]+w < 0 || e.Position[0] > a.width ||
				e.Position[1]+h < 0 || e.Position[1] > a.height {
				a.RemoveEntity(e.Name)
				continue
			}
		}

		// Update animation frame for animated entities
		if len(e.Shape) > 0 {
			// Check if animated (all single characters)
			isAnimated := true
			for _, s := range e.Shape {
				if len(s) > 1 {
					isAnimated = false
					break
				}
			}
			if isAnimated {
				e.FrameIndex++
			}
		}

		// Call update function if present
		if e.UpdateFn != nil {
			e.UpdateFn(e, a)
		}
	}

	// Check for collisions between physical entities
	a.checkCollisions()
}

// checkCollisions detects collisions between physical entities
func (a *Animation) checkCollisions() {
	entities := a.GetEntities()
	physicalEntities := make([]*Entity, 0)

	// Gather physical entities
	for _, e := range entities {
		if e.Physical {
			physicalEntities = append(physicalEntities, e)
		}
	}

	// Check each pair for collision
	for i := 0; i < len(physicalEntities); i++ {
		for j := i + 1; j < len(physicalEntities); j++ {
			e1 := physicalEntities[i]
			e2 := physicalEntities[j]

			if a.isColliding(e1, e2) {
				// Handle collision - for now just remove one
				// In the future could add damage, special effects, etc
				if e1.DeathCB != nil {
					e1.DeathCB(e1, a)
				}
				a.RemoveEntity(e1.Name)
				break
			}
		}
	}
}

// isColliding checks if two entities are colliding
func (a *Animation) isColliding(e1, e2 *Entity) bool {
	w1, h1 := e1.Size()
	w2, h2 := e2.Size()

	// Simple bounding box collision
	return !(e1.Position[0]+w1 < e2.Position[0] ||
		e2.Position[0]+w2 < e1.Position[0] ||
		e1.Position[1]+h1 < e2.Position[1] ||
		e2.Position[1]+h2 < e1.Position[1])
}

// RenderScreen draws all entities to the screen
func (a *Animation) RenderScreen() {
	a.screen.Clear()

	entities := a.GetEntities()

	// Sort entities by depth (Z)
	for i := range entities {
		for j := i + 1; j < len(entities); j++ {
			if entities[i].Position[2] > entities[j].Position[2] {
				entities[i], entities[j] = entities[j], entities[i]
			}
		}
	}

	// Draw each entity
	for _, e := range entities {
		a.drawEntity(e)
	}

	a.screen.Show()
}

// drawEntity renders a single entity
func (a *Animation) drawEntity(e *Entity) {
	x := e.Position[0]
	y := e.Position[1]

	// Check if this is an animated single-frame entity
	isAnimated := true
	if len(e.Shape) > 0 {
		for _, s := range e.Shape {
			if len(s) > 1 {
				isAnimated = false
				break
			}
		}
	} else {
		isAnimated = false
	}

	// Render animated frame (single character)
	if isAnimated && len(e.Shape) > 0 {
		frameChar := e.Shape[e.FrameIndex%len(e.Shape)]
		if len(frameChar) > 0 && frameChar != "\n" && frameChar != "\r" {
			px := x
			py := y

			if px >= 0 && px < a.width && py >= 0 && py < a.height {
				style := tcell.StyleDefault.Foreground(e.DefaultColor).Background(tcell.GetColor("black"))
				a.screen.SetContent(px, py, rune(frameChar[0]), nil, style)
			}
		}
		return
	}

	// Render multi-line shape
	for row, line := range e.Shape {
		for col, ch := range line {
			if ch == '\n' || ch == '\r' {
				continue
			}

			px := x + col
			py := y + row

			// Skip if out of bounds
			if px < 0 || px >= a.width || py < 0 || py >= a.height {
				continue
			}

			// Determine color
			color := e.DefaultColor
			if len(e.Color) > row && col < len(e.Color[row]) {
				colorByte := byte(e.Color[row][col])
				if c, ok := a.colorMap[colorByte]; ok && colorByte != ' ' {
					color = c
				}
			}

			style := tcell.StyleDefault.Foreground(color).Background(tcell.GetColor("black"))
			a.screen.SetContent(px, py, ch, nil, style)
		}
	}
}

// Close cleans up the terminal
func (a *Animation) Close() {
	a.screen.Fini()
}

func main() {
	anim, err := NewAnimation()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create animation: %v\n", err)
		os.Exit(1)
	}
	defer anim.Close()

	// Initialize entities
	initializeAquarium(anim)

	// Set up terminal input
	done := make(chan bool)
	go func() {
		for {
			ev := anim.screen.PollEvent()
			if ev == nil {
				done <- true
				return
			}

			switch ev := ev.(type) {
			case *tcell.EventKey:
				switch ev.Rune() {
				case 'q':
					done <- true
					return
				case 'p':
					anim.paused = !anim.paused
				case 'r':
					anim.UpdateTermSize()
					anim.RemoveAllEntities()
					initializeAquarium(anim)
				}
			case *tcell.EventResize:
				anim.UpdateTermSize()
				anim.RemoveAllEntities()
				initializeAquarium(anim)
			}
		}
	}()

	// Main animation loop
	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()

	spawnTicker := time.NewTicker(2 * time.Second)
	defer spawnTicker.Stop()

	for {
		select {
		case <-done:
			return
		case <-ticker.C:
			anim.Animate()
			anim.RenderScreen()
		case <-spawnTicker.C:
			// Occasionally spawn random creatures
			if rand.Float64() < 0.6 {
				randomObject(anim)
			}
		}
	}
}

// initializeAquarium sets up the static aquarium environment
func initializeAquarium(anim *Animation) {
	addEnvironment(anim)
	addCastle(anim)
	addAllSeaweed(anim)
	addAllFish(anim)
}

// addEnvironment adds water line segments
func addEnvironment(anim *Animation) {
	waterLines := []string{
		"~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~",
		"^^^^ ^^^  ^^^   ^^^    ^^^^      ",
		"^^^^      ^^^^     ^^^    ^^     ",
		"^^      ^^^^      ^^^    ^^^^^^  ",
	}

	segmentSize := len(waterLines[0])
	segmentRepeat := anim.Width()/segmentSize + 2

	for i, seg := range waterLines {
		var repeated strings.Builder
		for range segmentRepeat {
			repeated.WriteString(seg)
		}

		e := &Entity{
			Name:         fmt.Sprintf("water_seg_%d", i),
			Type:         "waterline",
			Shape:        []string{repeated.String()},
			Position:     [3]int{0, 5 + i, DepthWaterLine0 + i},
			DefaultColor: tcell.GetColor("cyan"),
			Physical:     true,
		}
		anim.NewEntity(e)
	}
}

// addCastle adds the castle decoration
func addCastle(anim *Animation) {
	castleImage := []string{
		"               T~~",
		"               |",
		"              /^\\",
		"             /   \\",
		" _   _   _  /     \\  _   _   _",
		"[ ]_[ ]_[ ]/ _   _ \\[ ]_[ ]_[ ]",
		"|_=__-_ =_|_[ ]_[ ]_|_=-___-__|",
		" | _- =  | =_ = _    |= _=   |",
		" |= -[]  |- = _ =    |_-=_[] |",
		" | =_    |= - ___    | =_ =  |",
		" |=  []- |-  /| |\\   |=_ =[] |",
		" |- =_   | =| | | |  |- = -  |",
		" |_______|__|_|_|_|__|_______|",
	}

	castleColor := []string{
		"                RR",
		"",
		"              yyy",
		"             y   y",
		"            y     y",
		"           y       y",
		"",
		"",
		"",
		"              yyy",
		"             yy yy",
		"            y y y y",
		"            yyyyyyy",
	}

	e := &Entity{
		Name:         "castle",
		Shape:        castleImage,
		Color:        castleColor,
		Position:     [3]int{anim.Width() - 32, anim.Height() - 13, DepthCastle},
		DefaultColor: tcell.GetColor("black"),
	}
	anim.NewEntity(e)
}

// addAllSeaweed adds seaweed randomly across screen
func addAllSeaweed(anim *Animation) {
	seaweedCount := anim.Width() / 15
	for range seaweedCount {
		addSeaweed(anim)
	}
}

// addSeaweed adds a single seaweed entity
func addSeaweed(anim *Animation) {
	height := rand.Intn(4) + 3
	left := make([]string, height)
	right := make([]string, height)

	for i := range height {
		if i%2 == 0 {
			left[i] = "("
			right[i] = ""
		} else {
			left[i] = ""
			right[i] = ")"
		}
	}

	// Interleave the two sides
	shape := make([]string, 0)
	for i := range height {
		shape = append(shape, left[i]+right[i])
	}

	x := rand.Intn(anim.Width()-2) + 1
	y := anim.Height() - height
	animSpeed := rand.Float64()*0.05 + 0.25

	e := &Entity{
		Name:         fmt.Sprintf("seaweed_%d", rand.Int63()),
		Shape:        shape,
		Position:     [3]int{x, y, DepthSeaweed},
		DefaultColor: tcell.GetColor("green"),
		CallbackArgs: []float64{0, 0, 0, animSpeed},
		DieTime:      time.Now().Add(time.Duration(rand.Intn(4*60)+8*60) * time.Second),
		DeathCB:      func(e *Entity, a *Animation) { addSeaweed(a) },
	}
	anim.NewEntity(e)
}

// addAllFish adds fish based on screen size
func addAllFish(anim *Animation) {
	screenSize := (anim.Height() - 9) * anim.Width()
	fishCount := screenSize / 350
	for range fishCount {
		addFish(anim)
	}
}
