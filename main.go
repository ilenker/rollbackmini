package main

import (
	"fmt"
	"os"
	"time"
	"github.com/gdamore/tcell/v2"
)

// Globals
var ROLLBACK bool = false
var BREAK bool = false

var scr tcell.Screen
var err error

var inputFromPeerCh chan PeerPacket
var replyFromPeerCh chan PeerPacket
var packetsToPeerCh chan PeerPacket

var player1 Snake
var player2 Snake

var localPlayerPtr *Snake
var peerPlayerPtr  *Snake


func main() {

							/*#### INIT ####*/
	FrameDiffBuffer = makeAverageIntBuffer(20)
	player1 = snakeMake(Vec2{(MapW/2) ,  6 + MapH/2}, R, P1Head)
	player2 = snakeMake(Vec2{(MapW/2) , -7 + MapH/2}, R, P2Head)
	loadConfig("config.json")

	if online {
		inputFromPeerCh = make(chan PeerPacket, 128)
		replyFromPeerCh = make(chan PeerPacket, 128)
		packetsToPeerCh = make(chan PeerPacket,  32)
		go multiplayer(inputFromPeerCh, replyFromPeerCh, packetsToPeerCh)
		<-inputFromPeerCh
	}

	stylesInit()
	scr, err = tcell.NewScreen(); F(err, "")
	err = scr.Init();             F(err, "")
	scr.SetStyle(ColEmpty)

	boardInit()
	textBoxesInit()

	localInputCh := make(chan signal, 8)
	go readLocalInputs(localInputCh)

	simTick := time.NewTicker(SIM_TIME)

	frameDiffGraph := barGraphInit(2, 19)

	render(scr, 2, 2)

	if online {<-inputFromPeerCh}

/* ············································································· Main Loop       */
	// qwfploop
	for {
		<-simTick.C
/* ············································································· Network Inbound */
		select {
		case pPacket := <-inputFromPeerCh:
			if pPacket.frameID < 5 {
				errorBox("skip", 0, 0)
				goto SkipRollback
			}

			if online {
				avgFrameDiff = FrameDiffBuffer(int(SIM_FRAME - pPacket.frameID))
				frameDiffGraph(int(avgFrameDiff))
				syncFrameDiff()
			}

			// Case of "reporting no inputs"
			if pPacket.content[0] == iNone ||
			   pPacket.content[0] == 0 {
				goto SkipRollback
			}
			errorBox(fmt.Sprintf("inc:%c", pPacket.content[0]), 5, 0)

			// Case of "reporting some inputs"
/* ·····································································┬·············· Rollback */
/*                                                                      └──Net Out - Hit Confirm */
			ROLLBACK = true
			rollbackBuffer.resimFramesWithNewInputs(pPacket)
			ROLLBACK = false

		default:
		// Don't block
		}

		SkipRollback:

/* ············································································Stage Local Input */
		drainLocalInputCh(localInputCh)

/* ·································································· Net Out - Send Local Input */
		sendCurrentFrameInputs()

/* ············································································· Push Save State */
		rollbackBuffer.pushFrame(copyCurrentFrameData(SIM_FRAME))

/* ···················································································· Simulate */
		simulate()

		variableDisplay()
		SIM_FRAME++

/* ············································································· Network Inbound */
		select {
		case reply := <-replyFromPeerCh:
			if reply.content[0] == iHit {
				other := getPeerPlayerPtr()
				dir := 1.5
				if other.stateID == P1Head { dir = 0.5 }
/*                                                                                   Hit Confirm */
				go hitEffect(other.pos, dir, beamCols[other.stateID])
				go hitEffect(other.pos, dir, beamCols[other.stateID])
				go hitEffect(other.pos, dir, beamCols[other.stateID])
				go hitEffect(other.pos, dir, beamCols[other.stateID])
			}
		default:
		}

/* ······················································································ Render */
		render(scr, 2, 2)
	}

}


func Break() {
	
	BREAK = true
	for {
		ev := scr.PollEvent()
		if key, ok := ev.(*tcell.EventKey); ok {

			if key.Key() == tcell.KeyESC {
				scr.Fini()
				os.Exit(0)
			} 

			// Keymap
			switch key.Rune() {

			case 'p':
				BREAK = false
				return

			case '1':
				variablePage = 1
			case '2':
				variablePage = 2
			case '3':
				variablePage = 3
			case '4':
				variablePage = 4
			}

			variableDisplay()
			render(scr, 2, 2)
		} 
	}

}

// This function should not be aware of input sources
// Could be local user, multiplayer peer, bot
// This only updates the snake state.
// We'll see how this works out (input validation based on board state?)


func render(s tcell.Screen, xOffset, yOffset int) {
	// s.Clear()
	// For each terminal row (board y-coordinates map 2:1 onto terminal y-coordinates)
	for y := range (MapH / 2) {
		lyUpper := y * 2           // Calculate corresponding Logical Row, given Terminal Row
		lyLower := y * 2 + 1

		// For each terminal cell (board x-coordinates map 1:1 onto terminal y-coordinates)
		for x := range MapW {
			upper := cols[board[lyUpper][x].col]
			lower := cols[board[lyLower][x].col]

			r := ' '
			style := ColEmpty

			// Blend the two 'styles'
			// take foreground color of each logical cell
			// set foreground of rune to upper color
			// set background of rune to lower color 
			// half-block ▀ displays top color (fg) and bottom color (bg) in one cell
			switch {
			case upper != ColEmpty && lower != ColEmpty:
				fg, _, _ := upper.Decompose()
				bg, _, _ := lower.Decompose()
				blend := ColEmpty.Foreground(fg).Background(bg)
				r, style = '▀', blend

			case upper != ColEmpty:
				r, style = '▀', upper

			case lower != ColEmpty:
				r, style = '▄', lower
			}

			s.SetContent(x + xOffset, y + yOffset, r, nil, style)
		}
	}

	s.Show()
}


// Collect all (local) input and send down a single channel
func readLocalInputs(inputCh chan signal) {

	for {
		if BREAK {
			continue
		}
		if SIM_FRAME < RB_BUFFER_LEN * 5 {
			continue
		}
		ev := scr.PollEvent()
		if key, ok := ev.(*tcell.EventKey); ok {

			if key.Key() == tcell.KeyESC {
				scr.Fini()
				os.Exit(0)
			} 

			switch key.Key() {

			// Special Keys
			case tcell.KeyLeft:
				inputCh <-iLeft

			case tcell.KeyRight:
				inputCh <-iRight

			default:
			}

			// Keymap
			switch key.Rune() {

			case 'x':
				inputCh <-iLeft
			case 'd':
				inputCh <-iRight
			case ' ':
				inputCh <-iShot

			case '!':
				variablePage = 1
			case '@':
				variablePage = 2
			case '#':
				variablePage = 3
			case '$':
				variablePage = 4
			}

		} 
	}

}


func drainLocalInputCh(inputCh chan signal) {

	player := getLocalPlayerPtr()

	select {
	case input := <-inputCh:
		player.tryInput(input) 
	default:
		return
	}

}


func getLocalPlayerCopy() Snake {
	if LOCAL == 2 {
		return player2
	}
	return player1
}

func getPeerPlayerCopy() Snake {
	if LOCAL == 2 {
		return player1
	}
	return player2
}

func getLocalPlayerPtr() *Snake {
	if LOCAL == 2 {
		return &player2
	}
	return &player1
}

func getPeerPlayerPtr() *Snake {
	if LOCAL == 2 {
		return &player1
	}
	return &player2
}
