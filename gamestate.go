package main

import (
	"fmt"
	"log"
	"runtime"
	"github.com/gdamore/tcell/v2"
)

type cellState uint8
type direction uint8
type signal byte

type Cell struct {
	col colorID
	state cellState 
	connection Vec2
}

var	board    = [MapH+1][MapW+1]Cell{}
var	vfxLayer = [MapH+1][MapW+1]colorID{}
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
	for y := range MapH {
		for x := range MapW {
			board[y][x].state = Empty
		}
	}
	drawPixelBox(scr, 2, 2, MapW - 1, MapH/2 - 1, tcell.ColorSteelBlue)
}


func cellSet(vec Vec2, newState cellState) {
	switch board[vec.y][vec.x].state {
	default: 
		board[vec.y][vec.x].state = newState
		board[vec.y][vec.x].col   = colorID(newState)
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

