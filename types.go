package main

import (
	"fmt"
	"math"
	"github.com/gdamore/tcell/v2"
)

type cellState uint8
type colorID uint8
type direction uint8
type input uint8

const (
	iNone input = iota
	iRight 
	iLeft
	iShot
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

func (v1 Vec2) AddNoWrap(v2 Vec2) Vec2 {
	return Vec2{
		x: v1.x + v2.x,
		y: v1.y + v2.y,
	}
}

func (v1 Vec2) Translate(angleRad float64, distance float64) Vec2 {
    
    dx := distance * math.Cos(angleRad)
    dy := distance * math.Sin(angleRad)
    
    newX := v1.x + int(math.Round(dx))
    newY := v1.y + int(math.Round(dy))
    
    return Vec2{x: newX, y: newY}
}


type Cell struct {
	col colorID
	state cellState 
	connection Vec2
}

type Snake struct {
	pos Vec2
	dir direction
	scpt int16
	subcellDebt int16
	inputQ []input
	stateID cellState
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

var ColShot1C  tcell.Style
var ColShot2C  tcell.Style
var ColShot3C  tcell.Style
var ColShot4C  tcell.Style

const (
	Empty cellState = iota
	P1Head
	P2Head
	Wall
)

const (
	EmptyC colorID = iota
	P1HeadC
	P2HeadC
	WallC

	_Shot1C
	_Shot2C
	_Shot3C
	_Shot4C
)

var cols map[colorID]tcell.Style


func assert(a any, b any, aName, bName string) {
	if a == b { // Pass
	} else { F(fmt.Errorf("failed assertion: %s{%d} == %s{%d}", aName, a, bName, b), "") }
}


const MapH = 20
const MapW = 20 * 2

const SUBCELL_SIZE = 32

var	board = [MapH+1][MapW+1]Cell{}
var snakes []*Snake


var debugBox func(msg string, args ...int)
var errorBox func(msg string, args ...int)

var SIM_FRAME uint16 = 1
var RESIM_FRAME uint16 = 1

var ROLLBACK bool

var inputName = map[input]string {
	iRight: "right",
	iLeft:  "left",
}
