package main

import (
	"fmt"
	"log"
	"runtime"
	"github.com/gdamore/tcell/v2"
)

type cellState = uint8
type direction = uint8
type signal byte

type Cell struct {
	state cellState 
}

var	_board      = [MapH+1][MapW+1]Cell{}
var	board       = [(MapH+1) * (MapW+1)]Cell{}
var SIM_FRAME uint16 = 1


func simulate() {

	player1.cooldown()
	player2.cooldown()

	subcellBudget := player1.scpt - player1.subcellDebt
	for subcellBudget > 0 {
		player1.control()
		cellSet(player1.pos, Empty)
		player1.move()
		cellSet(player1.pos, player1.stateID)
		subcellBudget -= SUBCELL_SIZE
	}
	player1.subcellDebt = AbsInt16(subcellBudget)


	subcellBudget = player2.scpt - player2.subcellDebt
	for subcellBudget > 0 {
		player2.control()
		cellSet(player2.pos, Empty)
		player2.move()
		cellSet(player2.pos, player2.stateID)
		subcellBudget -= SUBCELL_SIZE
	}
	player2.subcellDebt = AbsInt16(subcellBudget)

	player1.shoot()
	player2.shoot()

}


func boardInit() {
	mapSize := (MapH + 1) * (MapW + 1)
	lightRs         = make([]float64, mapSize)
	renderBuffer.rs = make([]float64, mapSize)
	renderBuffer.gs = make([]float64, mapSize)
	renderBuffer.bs = make([]float64, mapSize)
	vfxLayerRs      = make([]float64, mapSize)
	vfxLayerGs      = make([]float64, mapSize)
	vfxLayerBs      = make([]float64, mapSize)

	for y := range MapH {
		for x := range MapW {
			board[y * MapW + x].state = Empty
			vfxLayerRs[y * MapW + x] = 40
			vfxLayerGs[y * MapW + x] = 40
			vfxLayerBs[y * MapW + x] = 40
			//lightLayer[y][x] = Vec3[float32]{1, 1, 1}
			lightRs[y * MapW + x] = float64(1)
			//dimmingFactor[y * MapW + x] = float64(0.99)
		}
	}
	drawPixelBox(scr, 2, 2, MapW - 1, MapH/2 - 1, tcell.ColorSteelBlue)
}


func cellSet(vec Vec2, newState cellState) {
	switch board[vec.y * MapW + vec.x].state {
	default: 
		board[vec.y * MapW + vec.x].state = newState

		r, g, b := cols[newState].RGB()
		vfxLayerRs[flatIdx(vec)] = float64(r)
		vfxLayerGs[flatIdx(vec)] = float64(g)
		vfxLayerBs[flatIdx(vec)] = float64(b)
	}

}


func assert(a any, b any, aName, bName string) {
	if a == b { // Pass
	} else { F(fmt.Errorf("failed assertion: %s{%d} == %s{%d}", aName, a, bName, b), "") }
}


func E(err error, msg string) {
	if err != nil {
		_, _, line, _ := runtime.Caller(1)
		log.Printf("%s at line[%d]: %v\n", msg, err, line)
	}
}


func F(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %v\n", msg, err)
	}
}

