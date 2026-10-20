package main

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/gdamore/tcell/v2"
)

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
