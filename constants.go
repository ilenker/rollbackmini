package main

import (
	"time"
)

const SCPT = 64           // Subcell resolution
const RB_BUFFER_LEN = 10  // How many frames are stored for rollbacks
const PLAYER_1  = 1      
const PLAYER_2  = 2
const INPUT_BUFFER_LEN = 1


// Possible communication signals
const (
	iNone  signal = 95  // _
	iRight signal = 114 // r
	iLeft  signal = 108 // l 
	iShot  signal = 115 // s

	iHit   signal = 72  // H 
	iMiss  signal = 77  // M
)


// Possible movement directions
const (
	R direction = iota
	L
)


// Logical states the board's cells can be in
const (
	Empty cellState = iota
	P1Head
	P2Head
	Wall
)


// Map settings
const MapH = 30
const MapW = 20 * 2
const SUBCELL_SIZE = 32


// Constants determined by reading config.json at runtime
var SIM_TIME time.Duration
var online = false
var LOCAL int
var PEER  int
