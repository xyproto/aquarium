package main

import "math/rand"

var randomObjectFns []func(*Animation)

const maxRandomObjects = 6

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

func objectCount(anim *Animation) int {
	n := 0
	for _, t := range []string{"ship", "whale", "monster", "bigfish", "shark", "fishhook", "swan", "ducks"} {
		n += len(anim.GetEntitiesOfType(t))
	}
	return n + (len(anim.GetEntitiesOfType("dolphin"))+2)/3
}

func randomObject(anim *Animation) {
	if objectCount(anim) >= maxRandomObjects {
		return
	}
	fn := randomObjectFns[rand.Intn(len(randomObjectFns))]
	fn(anim)
}
