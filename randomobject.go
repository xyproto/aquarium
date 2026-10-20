package main

import "math/rand"

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
