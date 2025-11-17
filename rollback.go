package main

import (
	"fmt"
)


var rollbackBuffer RollbackBuffer
var FrameDiffBuffer func(int) (float64, []int)
var avgFrameDiff float64
var frameDiffs []int

type HitConfirm	struct {
	pos Vec2
	confirm bool
}

type FrameData struct {
	id uint16
	board [MapH+1][MapW+1]Cell
	local Snake
	peer Snake
}


type RollbackBuffer struct {
	frames [RB_BUFFER_LEN]FrameData
	idxLatest int
	latestFrameID uint16
}


func (rbb *RollbackBuffer) pushFrame(frame FrameData) {


	//debugBox(fmt.Sprintf("%x", rbb.idxLatest), 1 + rbb.idxLatest, 3)
	//if rbb.idxLatest == 0 { debugBox("                ", rbb.idxLatest, 4) }
	//debugBox(" ^", rbb.idxLatest, 4)
	rbb.frames[rbb.idxLatest] = frame
	rbb.idxLatest = (rbb.idxLatest + 1) % RB_BUFFER_LEN
	rbb.latestFrameID = frame.id

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




// We pass in the board and snakes to be modified.
// This function will be called on *each packet* that comes in
// that conflicts with "no_input"
// So we go to that frame, resim *everything* from there onwards.
func (rbb *RollbackBuffer) resimFramesWithNewInputs(pPacket PeerPacket) {

	rollbackFrame := FrameData{}
	resimFromBufferIdx := 0
	frameExists := false

	// Search for the frame to be resimmed
	for i := range RB_BUFFER_LEN {
		if rbb.frames[i].id == pPacket.frameID {
			frameExists = true
			resimFromBufferIdx = i
			rollbackFrame = rbb.frames[i]
			break
		}
	}

	if !frameExists {
		errorBox(fmt.Sprintf("Frame %04d not found\n", pPacket.frameID), 0, 0)
		return
		//pPacket.content = [4]signal{95, 95, 95, 95}
		//pPacket.frameID = SIM_FRAME
	}

	for _, b := range pPacket.content {
		rollbackFrame.peer.tryInput(b)
	}

	currentRollbackFrameID := rollbackFrame.id

	loadFrameData(rollbackFrame)

	// Resim from rollbackFrame, until latest frame 
	i := resimFromBufferIdx
	i_ := 0
	errorBox("\\clr")
	for {

		*getLocalPlayerPtr() = rbb.frames[i % RB_BUFFER_LEN].local

		rbb.frames[i % RB_BUFFER_LEN] = copyCurrentFrameData(currentRollbackFrameID)

		simulate()

		currentRollbackFrameID++
		
		errorBox(fmt.Sprintf("resim: %3d\n", currentRollbackFrameID))

		//render(scr, 2, 2)
		//Break()

		if currentRollbackFrameID == rbb.latestFrameID + 1 {
			return
		}

		i++; i_++
	}
}

		/*

		"received at frame 2:  frame 0: left "
		0 1 2                 3 4 5  *buffer full* 6 . . . . .
drain   x x x                 x x x                x
resim   - - x→ 0′→ 0″→ 1″→ 2″ - - -                -
copy    x x    ┃   x   x   x  x x x                x
		┌──────┘   ┃   ┃   ┃  ┃ ┃ ┃                ┃
		0′1′       0″  1″  2″ 3′4′5′               6′1″2″3′4′5′
sim		x x        x   x   x  x x x
render  x x        -   -   x  x x x
		drain inputs
		resim
		copy current state
		sim (update)
		render

		*/


func copyCurrentFrameData(frameID uint16) FrameData {

	savedFrameData := FrameData {
		id: frameID,
		board: board,
		local: getLocalPlayerCopy(),
		peer: getPeerPlayerCopy(),
	}

	return savedFrameData
}


func loadFrameData(fd FrameData) {
	board = fd.board

	*getLocalPlayerPtr() = fd.local
	*getPeerPlayerPtr() = fd.peer

}


