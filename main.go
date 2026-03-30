package main

import (
	"fmt"
	"math/rand"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/gdamore/tcell/v2"
)

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

// Entity represents any object in the aquarium.
type Entity struct {
	Name         string
	Type         string
	Shape        []string   // single-frame multi-line shape
	Shapes       [][]string // multi-frame animation (each element = one frame = []string of lines)
	Color        []string   // color mask for Shape
	Colors       [][]string // color masks for each frame in Shapes
	Position     [3]int     // x, y, z (z: higher = further back, rendered first)
	DefaultColor tcell.Color
	Physical     bool
	DieOffscreen bool
	DieTime      time.Time
	DeathCB      func(*Entity, *Animation)
	UpdateFn     func(*Entity, *Animation)
	FrameIndex   int
	FrameDelay   int // update-ticks between frame advances (0 = advance every tick)
	FrameTick    int
	FloatX       float64 // sub-cell x accumulator
	FloatY       float64 // sub-cell y accumulator
	VX           float64 // x velocity (cells per update tick)
	VY           float64 // y velocity (cells per update tick)
	// Path animation (dolphins): sequence of (vx,vy) velocity steps
	Path      [][2]float64
	PathIndex int
}

// currentShape returns the active shape and color mask for the current frame.
func (e *Entity) currentShape() ([]string, []string) {
	if len(e.Shapes) > 0 {
		idx := e.FrameIndex % len(e.Shapes)
		shape := e.Shapes[idx]
		var col []string
		if idx < len(e.Colors) {
			col = e.Colors[idx]
		} else if len(e.Color) > 0 {
			col = e.Color
		}
		return shape, col
	}
	return e.Shape, e.Color
}

// Size returns the bounding box of the entity in terminal cells.
func (e *Entity) Size() (int, int) {
	shape, _ := e.currentShape()
	if len(shape) == 0 {
		return 0, 0
	}
	maxW := 0
	for _, line := range shape {
		w := len([]rune(line))
		if w > maxW {
			maxW = w
		}
	}
	return maxW, len(shape)
}

// Animation manages all entities and drives the render loop.
type Animation struct {
	screen      tcell.Screen
	entities    map[string]*Entity
	entityMutex sync.RWMutex
	width       int
	height      int
	paused      bool
	colorMap    map[byte]tcell.Color
	version     string
	kitty       *KittyRenderer
}

func NewAnimation() (*Animation, error) {
	s, err := tcell.NewScreen()
	if err != nil {
		return nil, err
	}
	if err := s.Init(); err != nil {
		return nil, err
	}

	var bgColor tcell.Color
	if isKittyTerminal() {
		bgColor = tcell.NewRGBColor(0, 8, 24) // deep ocean blue
	} else {
		bgColor = tcell.ColorBlack
	}
	s.SetStyle(tcell.StyleDefault.Background(bgColor).Foreground(tcell.ColorWhite))
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

	if isKittyTerminal() {
		anim.kitty = NewKittyRenderer()
	}

	return anim, nil
}

func makeColorMap() map[byte]tcell.Color {
	if isKittyTerminal() {
		return makeKittyColorMap()
	}
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
		'k': tcell.ColorBlack,
		'K': tcell.ColorBlack,
	}
}

func (a *Animation) Width() int  { return a.width }
func (a *Animation) Height() int { return a.height }

func (a *Animation) UpdateTermSize() {
	a.width, a.height = a.screen.Size()
}

func (a *Animation) NewEntity(e *Entity) {
	a.entityMutex.Lock()
	defer a.entityMutex.Unlock()
	if e.Name == "" {
		e.Name = fmt.Sprintf("entity_%d_%d", len(a.entities), rand.Int63())
	}
	// Sync float positions with integer positions on creation
	e.FloatX = float64(e.Position[0])
	e.FloatY = float64(e.Position[1])
	a.entities[e.Name] = e
}

func (a *Animation) RemoveEntity(name string) {
	a.entityMutex.Lock()
	defer a.entityMutex.Unlock()
	delete(a.entities, name)
}

func (a *Animation) RemoveAllEntities() {
	a.entityMutex.Lock()
	defer a.entityMutex.Unlock()
	a.entities = make(map[string]*Entity)
}

func (a *Animation) GetEntities() []*Entity {
	a.entityMutex.RLock()
	defer a.entityMutex.RUnlock()
	result := make([]*Entity, 0, len(a.entities))
	for _, e := range a.entities {
		result = append(result, e)
	}
	return result
}

func (a *Animation) GetEntity(name string) *Entity {
	a.entityMutex.RLock()
	defer a.entityMutex.RUnlock()
	return a.entities[name]
}

func (a *Animation) GetEntitiesOfType(t string) []*Entity {
	a.entityMutex.RLock()
	defer a.entityMutex.RUnlock()
	var result []*Entity
	for _, e := range a.entities {
		if e.Type == t {
			result = append(result, e)
		}
	}
	return result
}

// Animate advances all entity logic by one update tick.
func (a *Animation) Animate() {
	if a.paused {
		return
	}

	entities := a.GetEntities()
	now := time.Now()

	for _, e := range entities {
		// Check die-time
		if !e.DieTime.IsZero() && now.After(e.DieTime) {
			if e.DeathCB != nil {
				e.DeathCB(e, a)
			}
			a.RemoveEntity(e.Name)
			continue
		}

		// Check die-offscreen
		if e.DieOffscreen {
			w, h := e.Size()
			x, y := e.Position[0], e.Position[1]
			if x+w < 0 || x > a.width || y+h < 0 || y > a.height {
				if e.DeathCB != nil {
					e.DeathCB(e, a)
				}
				a.RemoveEntity(e.Name)
				continue
			}
		}

		// Advance multi-frame animation
		if len(e.Shapes) > 1 {
			e.FrameTick++
			delay := e.FrameDelay
			if delay <= 0 {
				delay = 1
			}
			if e.FrameTick >= delay {
				e.FrameTick = 0
				e.FrameIndex = (e.FrameIndex + 1) % len(e.Shapes)
			}
		}

		// Call per-entity update
		if e.UpdateFn != nil {
			e.UpdateFn(e, a)
		}
	}
}

// RenderScreen draws all entities and shows the frame.
func (a *Animation) RenderScreen() {
	a.screen.Clear()

	entities := a.GetEntities()

	// Higher Z value = further back = rendered first (lower Z is drawn on top)
	sort.Slice(entities, func(i, j int) bool {
		return entities[i].Position[2] > entities[j].Position[2]
	})

	for _, e := range entities {
		a.drawEntity(e)
	}

	a.screen.Show()

	// Kitty graphics overlay (rendered after tcell frame)
	if a.kitty != nil && a.kitty.enabled {
		a.kitty.Render(a)
	}
}

func (a *Animation) bgStyle() tcell.Style {
	if isKittyTerminal() {
		return tcell.StyleDefault.Background(tcell.NewRGBColor(0, 8, 24))
	}
	return tcell.StyleDefault.Background(tcell.ColorBlack)
}

func (a *Animation) drawEntity(e *Entity) {
	x := e.Position[0]
	y := e.Position[1]
	shape, colorMask := e.currentShape()

	for row, line := range shape {
		col := 0
		for _, ch := range line {
			px := x + col
			py := y + row
			col++

			// Space = transparent (skip, show entity beneath)
			if ch == ' ' {
				continue
			}

			if px < 0 || px >= a.width || py < 0 || py >= a.height {
				continue
			}

			// Determine color from mask
			color := e.DefaultColor
			if row < len(colorMask) {
				maskRunes := []rune(colorMask[row])
				if col-1 < len(maskRunes) {
					cb := byte(maskRunes[col-1])
					if c, ok := a.colorMap[cb]; ok {
						color = c
					}
				}
			}

			style := a.bgStyle().Foreground(color)
			a.screen.SetContent(px, py, ch, nil, style)
		}
	}
}

func (a *Animation) Close() {
	if a.kitty != nil && a.kitty.enabled {
		a.kitty.Cleanup()
	}
	a.screen.Fini()
}

func main() {
	anim, err := NewAnimation()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create animation: %v\n", err)
		os.Exit(1)
	}
	defer anim.Close()

	// Upload fish images for Kitty mode
	if anim.kitty != nil && anim.kitty.enabled {
		anim.kitty.Init()
	}

	initializeAquarium(anim)

	done := make(chan bool, 1)
	go func() {
		for {
			ev := anim.screen.PollEvent()
			if ev == nil {
				done <- true
				return
			}
			switch ev := ev.(type) {
			case *tcell.EventKey:
				switch {
				case ev.Rune() == 'q' || ev.Rune() == 'Q':
					done <- true
					return
				case ev.Key() == tcell.KeyEscape || ev.Key() == tcell.KeyCtrlC:
					done <- true
					return
				case ev.Rune() == 'p' || ev.Rune() == 'P':
					anim.paused = !anim.paused
				case ev.Rune() == 'r' || ev.Rune() == 'R':
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

	renderTicker := time.NewTicker(50 * time.Millisecond)  // 20 fps
	updateTicker := time.NewTicker(100 * time.Millisecond) // 10 fps entity updates (matches Perl)
	spawnTicker := time.NewTicker(6 * time.Second)
	defer renderTicker.Stop()
	defer updateTicker.Stop()
	defer spawnTicker.Stop()

	for {
		select {
		case <-done:
			return
		case <-updateTicker.C:
			anim.Animate()
		case <-renderTicker.C:
			anim.RenderScreen()
		case <-spawnTicker.C:
			if rand.Float64() < 0.7 {
				randomObject(anim)
			}
		}
	}
}

// initializeAquarium sets up the static and initial dynamic elements.
func initializeAquarium(anim *Animation) {
	addEnvironment(anim)
	addCastle(anim)
	addAllSeaweed(anim)
	addAllFish(anim)
	// Spawn a couple of random objects to populate the screen immediately
	randomObject(anim)
	randomObject(anim)
}

// addEnvironment adds tiled water-line segments across the screen.
func addEnvironment(anim *Animation) {
	waterLineSegments := []string{
		"~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~",
		"^^^^ ^^^  ^^^   ^^^    ^^^^      ",
		"^^^^      ^^^^     ^^^    ^^     ",
		"^^      ^^^^      ^^^    ^^^^^^  ",
	}

	segLen := len(waterLineSegments[0])
	repeat := anim.Width()/segLen + 2

	depths := []int{DepthWaterLine0, DepthWaterLine1, DepthWaterLine2, DepthWaterLine3}

	for i, seg := range waterLineSegments {
		var b strings.Builder
		for range repeat {
			b.WriteString(seg)
		}
		line := b.String()

		anim.NewEntity(&Entity{
			Name:         fmt.Sprintf("water_seg_%d", i),
			Type:         "waterline",
			Shape:        []string{line},
			Position:     [3]int{0, 5 + i, depths[i]},
			DefaultColor: tcell.GetColor("cyan"),
			Physical:     true,
		})
	}
}

// addCastle adds the castle decoration.
func addCastle(anim *Animation) {
	castleShape := []string{
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

	anim.NewEntity(&Entity{
		Name:         "castle",
		Shape:        castleShape,
		Color:        castleColor,
		Position:     [3]int{anim.Width() - 32, anim.Height() - 13, DepthCastle},
		DefaultColor: tcell.ColorBlack,
	})
}

// addAllSeaweed plants seaweed proportional to screen width.
func addAllSeaweed(anim *Animation) {
	count := anim.Width() / 15
	for range count {
		addSeaweed(anim)
	}
}

// addSeaweed plants a single seaweed entity.
func addSeaweed(anim *Animation) {
	height := rand.Intn(4) + 3
	shape := make([]string, height)
	for i := range height {
		if i%2 == 0 {
			shape[i] = "("
		} else {
			shape[i] = " )"
		}
	}

	x := rand.Intn(anim.Width()-2) + 1
	y := anim.Height() - height

	anim.NewEntity(&Entity{
		Name:         fmt.Sprintf("seaweed_%d", rand.Int63()),
		Type:         "seaweed",
		Shape:        shape,
		Position:     [3]int{x, y, DepthSeaweed},
		DefaultColor: tcell.GetColor("green"),
		DieTime:      time.Now().Add(time.Duration(rand.Intn(4*60)+8*60) * time.Second),
		DeathCB:      func(_ *Entity, a *Animation) { addSeaweed(a) },
	})
}

// addAllFish adds fish proportional to the underwater screen area.
func addAllFish(anim *Animation) {
	screenSize := (anim.Height() - 9) * anim.Width()
	count := max(screenSize/350, 1)
	for range count {
		addFish(anim)
	}
}
