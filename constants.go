package main

import (
	"time"
)

const RB_BUFFER_LEN    = 20	// Number of frames stored for rollbacks
const INPUT_BUFFER_LEN = 2	// Number of inputs that be queued - smoothes out control feel
const RTT_BUFFER_LEN   = 10	// Number of ping times used for averaging


// Communication signals

const (
	iNone  signal = 95		// _	no input
	iRight signal = 114		// r	move right
	iLeft  signal = 108		// l	move left
	iShot  signal = 115		// s	shoot

	iHit   signal = 72		// H	hit confirmed
	iCrit  signal = 67		// C	hit confirmed - crit
	iMiss  signal = 77		// M	hit denied

	iPing  signal = 63		// ?	ping request
	iPong  signal = 33		// !	ping reply
)

// Facing directions
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
const MapW = 40
const MapX = 2
const MapY = 2
const SUBCELL_SIZE = 32			// Resolution of logical subcells - enables finer control over movespeed


// Constants determined by reading config.json at runtime

var SIM_TIME time.Duration		// How long each game loop tick takes
var online = false
var LOCAL int					// Identifier for which player (top / bottom) is the local peer
var PEER  int					// Identifier for which player (top / bottom) is the remote peer
