package main

import (
	"fmt"
)

var COPY_STATE = false
var LOAD_STATE = false


type GameState struct {
	frame uint32
	board [MapH+1][MapW+1]Cell
	snakesData []*Snake
}

const RB_BUFFER_LEN = 20

type RollbackBuffer struct {
	frames [RB_BUFFER_LEN]GameState
	index int
	latestFrame uint32
}

func (r *RollbackBuffer) pushFrame(gs GameState) {
	debugBox(fmt.Sprintf("%x", r.index), 1 + r.index, 2)
	if r.index == 0 { debugBox("                ", r.index, 3) }
	debugBox(" ^", r.index, 3)

	r.index = wrapInt(r.index + 1, RB_BUFFER_LEN)
	r.frames[r.index] = gs
	r.latestFrame = gs.frame
}


/*
	We will be receiving packets with frame numbers.
	Many of these packets will have "no input" and a frame number.
	These packets can be "ignored" for now.

	When we receive a packet that says "frame 1001: left", and we are on frame 1015
	here's what happens:
	Basically, our frames 1001 - 1016 are "stale", right.
	So we need to resimulate.
	What information do we need to resimulate?
	Frame 1001's state. That is the frame that has the peer with *some input queue ready to go*
	Not executed yet. They will be executed on frames 1002, 3 and 4

	What might be useful, is if they send their entire queue (frame 1001: [l, u, r, _])
	we already know what they will do on frames 1002-1004.

	So to accurately resimulate from 1001:
		1001:
	1.	Complete and total unaltered game state - minus peers inputs.

		1001-1004:
	2.	Here we need to *change the peer's inputs*
		GameStates 1001-1004 -> Peer's input queue -> copy in reported inputs.

	3.	Now we need to resimulate.
		Do we need the GameState information from frames 1002-1004?
		We actually don't. We start from 1001, simulate *AND SAVE* 
		frames 1002-1004  (and then 1005-1015, with "no_input", given new state of 1004).
		So anything can happen in this space of resimming -
		Local player can collide with peer, leaving local in completely
		different state compared to original frame 1015.

	4.	The only information we need from frames 1002-1004 is *our local inputs*

	Resimming on *every frame* even if no correction is needed seems very wasteful.
	I'm sure there must be a good reason why it's done (maybe it makes the CPU cycles
	per frame more predictable? Not having light load and then heavy load randomly, 
	potentially causing non-deterministic outcomes?)
*/

func (r *RollbackBuffer) rollBack(b *[MapH+1][MapW+1]Cell, snakes []*Snake, gs GameState) {
	for i := range RB_BUFFER_LEN {
		loadGameState(b, snakes, r.frames[wrapInt(1 + r.index + i, RB_BUFFER_LEN)])
	}	
}

// We pass in the board and snakes to be modified.
// This function will be called on *each packet* that comes in
// that conflicts with "no_input"
// So we go to that frame, resim *everything* from there onwards.
func (r *RollbackBuffer) resimFramesWithNewInputs(frame uint32, inputQ []input, b *[MapH+1][MapW+1]Cell, snakes []*Snake) {

	rollbackFrame := GameState{}
	resimFromIndex := 0

	for i := range RB_BUFFER_LEN {

		// Start from frame after r.index (that is the oldest frame - right after latest)
		idx := wrapInt(1 + r.index + i, RB_BUFFER_LEN)

		if r.frames[idx].frame == frame {
			rollbackFrame = r.frames[i] 
			resimFromIndex = idx
			break
		}

	}

	// TODO: Figure out how this number will be determined
	rollbackFrame.snakesData[1].inputQ = inputQ


	// We are now at "frame 1001", resimming until "frame 1015" (current)
	// Just in time for this frames logic update and rendering
	// so we stop at "frame 1014" - the next frame after that would be the
	// oldest frame in our buffer.
	// rollbackFrame will have had it's contents modified already, just before this func call
	loadGameState(b, snakes, rollbackFrame)

	// Resim all frames
	for i := range RB_BUFFER_LEN {

		idx := wrapInt(resimFromIndex + i, RB_BUFFER_LEN)

		updateLogic(snakes)
		r.frames[idx] = copyCurrentGameState(b, snakes, r.frames[idx].frame)
		debugBox(fmt.Sprintf("%2d resim: %d", i, r.frames[idx].frame), 0, 5 + i)
	}

}

	//loadGameState(b, snakes, r.frames[wrapInt(1 + r.index + i, RB_BUFFER_LEN)])

func (r *RollbackBuffer) rectifyGameState(gs *GameState, inputQ []input) {
}


var GameStateSlot1 GameState

func copyCurrentGameState(b *[MapH+1][MapW+1]Cell, snakes []*Snake, frame uint32) GameState {
	snakesData := make([]*Snake, len(snakes), cap(snakes))

	for i, snake := range snakes {
		snakesData[i] = snakeCopy(snake)
	}

	savedGameState := GameState {
		frame: frame,
		board: *b,
		snakesData: snakesData,
	}

	return savedGameState
}


func loadGameState(b *[MapH+1][MapW+1]Cell, snakes []*Snake, gs GameState) {
	*b = gs.board	
	for i := range snakes {
		snakes[i] = snakeCopy(gs.snakesData[i])
	}
}


func snakeCopy(src *Snake) *Snake {
	newSnake := &Snake{
		pos:         src.pos,
		dir:         src.dir,
		scpt:        src.scpt,
		subcellDebt: src.subcellDebt,
		inputQ: 	 src.inputQ,
	}
	return newSnake
}
