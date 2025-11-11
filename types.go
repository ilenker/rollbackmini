package main

import (
	"fmt"
	"github.com/gdamore/tcell/v2"
)

type cellState uint8
type direction uint8
type input uint8

const (
	iNone input = iota
	iRight 
	iLeft
)

const (
	R direction = iota
	L
)


const (
	SCPT_MIN = 1
	SCPT_MAX = 1024

	SCPT_LOW = 4
	SCPT_HIGH = 32
)

type Vec2 struct {
	x int
	y int
}

func (v1 Vec2) Add(v2 Vec2) Vec2 {
	newX := wrapInt(v1.x + v2.x, MapW)
	newY := wrapInt(v1.y + v2.y, MapH)
	return Vec2{
		newX,
		newY,
	}
}


type Cell struct {
	col cellState
	state cellState 
	connection Vec2
}

type Snake struct {
	pos Vec2
	dir direction
	scpt int16
	subcellDebt int16
	inputQ []input
}


func (s *Snake) move() {
	if s.dir == R {
		s.pos.x = wrapInt(s.pos.x + 1, MapW)
	}
	if s.dir == L {
		s.pos.x = wrapInt(s.pos.x - 1, MapW)
	}
}


func (s *Snake) popInput() (input, bool) {
	if len(s.inputQ) == 0 {
		return iNone, false
	}

	input := s.inputQ[0]
	s.inputQ = s.inputQ[1:]

	return input, true
} 


func (s *Snake) tryInput(inp input) bool {
	if len(s.inputQ) >= 3 {
		return true
	}
	s.inputQ = append(s.inputQ, inp)
	return false
}


var ColEmpty   tcell.Style

var ColP1Head  tcell.Style

var ColP2Head  tcell.Style

var ColDefault tcell.Style

const (
	Empty cellState = iota
	P1Head
	P2Head
	Wall
)

var cols map[cellState]tcell.Style


func assert(a any, b any, aName, bName string) {
	if a == b { // Pass
	} else { F(fmt.Errorf("failed assertion: %s{%d} == %s{%d}", aName, a, bName, b), "") }
}


const MapH = 20
const MapW = 20 * 2

const SUBCELL_SIZE = 32

var	board = [MapH+1][MapW+1]Cell{}

var debugBox func(msg string, args ...int)
var errorBox func(msg string, args ...int)

var SIM_FRAME uint16 = 1

var ROLLBACK bool

var inputName = map[input]string {
	iRight: "right",
	iLeft:  "left",
}
