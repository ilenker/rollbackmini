package main

import (
	"fmt"
	"log"
	"os"
	"runtime"
	"time"
	"github.com/gdamore/tcell/v2"
)

var MOVES_PER_TICK int
var LastInputT time.Time
const SIM_TIME = time.Duration(time.Millisecond * 16)
const SCPT = 64
const RB_BUFFER_LEN = 15

var FrameSyncCh chan bool
var online = true

var scr tcell.Screen
var err error
var calcAvgRollback func(time.Duration) int64 
var avgRollback int64 = 0

const PLAYER_1  = 0
const PLAYER_2  = 1

var LOCAL int
var PEER  int

func main() {
	var inboundPacketCh chan PeerPacket
	var outboundPacketCh chan PeerPacket

	LOCAL = PLAYER_1

	switch LOCAL {
	case PLAYER_1:
		PEER = PLAYER_2
	case PLAYER_2:
		PEER = PLAYER_1
	}
							/*#### INIT ####*/

	SIM_FRAME = 0
	snakes = make([]*Snake, 2)
	stylesInit()
	boardInit(&board)

	if online {
		inboundPacketCh = make(chan PeerPacket, 64)
		outboundPacketCh = make(chan PeerPacket, 64)

		go multiplayer(inboundPacketCh, outboundPacketCh)
		<-inboundPacketCh
	}

	scr, err = tcell.NewScreen()
	F(err, "")
	err = scr.Init()
	F(err, "")
	defStyle := tcell.StyleDefault.Background(tcell.ColorBlack).Foreground(tcell.ColorBlack)
	scr.SetStyle(defStyle)

	rollbackBuffer := RollbackBuffer{}
							/*##############*/


	// Board frame
	DrawPixelBox(scr, 2, 2, MapW - 1, MapH/2 - 1, tcell.ColorBlue)
	debugBox = DrawMessages(scr, MapW + 5 , 4, 30, 30, true)
	errorBox = DrawMessages(scr, MapW + 37, 4, 15, 30, true)

	snakes[PLAYER_1] = snakeMake(Vec2{MapW/2,  5 + MapH/2}, L, P1Head)
	snakes[PLAYER_2] = snakeMake(Vec2{MapW/2, -5 + MapH/2}, R, P2Head)

	snakes[PLAYER_1].scpt = 8
	snakes[PLAYER_2].scpt = 8

	localInputCh := make(chan input, 8)
	go readLocalInputs(scr, localInputCh)

	simTick := time.NewTicker(SIM_TIME)
	avgSimT := makeAverageDurationBuffer(100)
	calcAvgRollback = makeAverageDurationBuffer(50)
	startT := time.Now()
	FrameSyncCh = make(chan bool, 5)

	var _dMicroSec int64 = 0
	//var rollbackSize uint16 = 0 

	debugPrint := func () {
		{
			debugBox(fmt.Sprintf("Frame Time:\t[%.2f]", float64(_dMicroSec)/1000), 0, 0)
			debugBox(fmt.Sprintf("Avg Rollback:\t[%.2f]", float64(avgRollback)/1000), 0, 1)
			debugBox(fmt.Sprintf("sim_frame:\t%5d", SIM_FRAME), 0, 2)
			//debugBox(fmt.Sprintf("MPT:\t\t%2d(%.2f)", MOVES_PER_TICK, float64(snakes[0].scpt) / float64(SUBCELL_SIZE)), 0, 2)
			//debugBox(fmt.Sprintf("SCPT:\t\t%3d", snakes[0].scpt), 0, 3)
		}
	} 
	render(scr, 2, 2)

	// qwfploop
	for {
		startT = time.Now()
		MOVES_PER_TICK = 0
		<-simTick.C

		// Rollback Check - if packet came in, resimulate
		select {
		case pP := <-inboundPacketCh:
			if pP.frameID == 6969 {
				continue
			}

			if pP.inputQ[0] == iNone {
				goto correctPrediction
			}

			inputQ := []input{
				pP.inputQ[0],
				pP.inputQ[1],
				pP.inputQ[2],
				pP.inputQ[3],
			}

			// Multiple "updateLogic" calls in this function
			rollbackBuffer.resimFramesWithNewInputs(pP.frameID, inputQ, &board, snakes)
			//rollbackSize = SIM_FRAME - pP.frameID

		default:

		}

		correctPrediction:

		// Load up local snake's input queue 
		drainInputChToSnake(localInputCh, snakes, LOCAL)

		// Store current frame in the rollback buffer
		rollbackBuffer.pushFrame(copyCurrentFrameData(&board, snakes, SIM_FRAME))
		if online {
			outboundPacketCh <-makePeerPacket(SIM_FRAME, snakes[LOCAL])
		}

		// Simulate live frame
		// We store the frame before simulating live frame,
		// because the stored frame has the information needed
		// to create this state already - and we won't
		// necessarily use every stored frame.
		updateLogic(snakes)

		SIM_FRAME++
		_dMicroSec = avgSimT(time.Since(startT))
		debugPrint()

		select {
		case FrameSyncCh <-true:
		default:
		}

		render(scr, 2, 2)

	}
}


// This function should not be aware of input sources
// Could be local user, multiplayer peer, bot
// This only updates the snake state.
// We'll see how this works out (input validation based on board state?)
func updateLogic(snakes []*Snake) {

	for _, s := range snakes {
		subcellBudget := s.scpt - s.subcellDebt

		for {
			if subcellBudget <= 0 {
				s.subcellDebt = AbsInt16(subcellBudget)
				break
			}

			controlSnake(s)

			cellSet(&board, s.pos, Empty)

			s.move()

			cellSet(&board, s.pos, s.stateID)

			subcellBudget -= SUBCELL_SIZE
			MOVES_PER_TICK++
		}
	}
}


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


func beamEffect(start Vec2, end Vec2, dir Vec2) {

	start = start.Add(dir)
	pos := start
	animLen := 10

	// Frame 1
	<-FrameSyncCh
	for range animLen {
		board[pos.y][pos.x].col = _Shot1C

		if pos.y != end.y {
			pos = pos.Add(dir)
		}

	}

	time.Sleep(SIM_TIME)
	time.Sleep(SIM_TIME)
	pos = start
	for range animLen {
		board[pos.y][pos.x].col = EmptyC
		if pos.y != end.y {
			pos = pos.Add(dir)
		}

	}

	time.Sleep(SIM_TIME)
	pos = start
	for range animLen {
		board[pos.y][pos.x].col = _Shot2C
		if pos.y != end.y {
			pos = pos.Add(dir)
		}

	}

	pos = start
	for range animLen {
		time.Sleep(SIM_TIME / 2)
		board[pos.y][pos.x].col = _Shot3C
		if pos.y != end.y {
			pos = pos.Add(dir)
		}

	}

	pos = start
	for range animLen {
		time.Sleep(SIM_TIME / 2)
		board[pos.y][pos.x].col = _Shot4C
		if pos.y != end.y {
			pos = pos.Add(dir)
		}

	}

	pos = start
	for range animLen {
		time.Sleep(SIM_TIME)
		board[pos.y][pos.x].col = EmptyC
		if pos.y != end.y {
			pos = pos.Add(dir)
		}

	}

}


// Collect all (local) input and send down a single channel
func readLocalInputs(scr tcell.Screen, inputCh chan input) {

	for {
		ev := scr.PollEvent()
		if key, ok := ev.(*tcell.EventKey); ok {

			if key.Key() == tcell.KeyESC {
				scr.Fini()
				os.Exit(0)
			} 

			// Keymap
			switch key.Rune() {

			case 'x':
				inputCh <-iLeft
			case 'd':
				inputCh <-iRight
			case ' ':
				inputCh <-iShot
			}

		} 
	}

}


func drainInputChToSnake(inputCh chan input, s []*Snake, snakeID int) {
	full := false

	for {
		if full {
			return
		}

		select {
		case input := <-inputCh:
			full = s[snakeID].tryInput(input) 
		default:
			return
		}
	} 

}


func controlSnake(s *Snake) {
	input, ok := s.popInput()
	if ok {
		switch input {
		case iRight:
			s.dir = R

		case iLeft:
			s.dir = L

		case iShot:
			snakes[PEER].tryInput(iShot)
			end := Vec2{s.pos.x, 0}
			dir := Vec2{0, 1}
			if s.stateID == P1Head {
				dir.y = -1
				end.y = MapH
			}
			if s.pos.x == snakes[PEER].pos.x {
				end.y = snakes[PEER].pos.y
			}
			go beamEffect(s.pos, end, dir)

		}
	}

}


func boardInit(board *[MapH+1][MapW+1]Cell) {
	for y := range MapH {
		for x := range MapW {
			//board[y][x].col = tcell.StyleDefault
			board[y][x].state = Empty
		}
	}
}

// Why do we use cellstate for state and color?
// Maybe I need to have separate the color info.
// Hasn't happened yet but maybe surely someday
func cellSet(board *[MapH+1][MapW+1]Cell, vec Vec2, newState cellState) {
	switch board[vec.y][vec.x].state {
	default: 
		board[vec.y][vec.x].state = newState
		board[vec.y][vec.x].col   = colorID(newState)
	}

}

func cellGet(board *[MapH+1][MapW+1]Cell, vec Vec2) *Cell {
	return &board[vec.y][vec.x]
}


func stylesInit() {
	ColBlack := tcell.NewRGBColor(30,  11,  30)
	ColEmpty   = tcell.StyleDefault.Foreground(ColBlack              ).Background(ColBlack)
	ColP1Head  = tcell.StyleDefault.Foreground(tcell.ColorBlue       ).Background(ColBlack)
	ColP2Head  = tcell.StyleDefault.Foreground(tcell.ColorOrange     ).Background(ColBlack)
	ColDefault = tcell.StyleDefault.Foreground(tcell.ColorWhiteSmoke ).Background(ColBlack)

	ColShot1C  = tcell.StyleDefault.Foreground(tcell.ColorWhite      ).Background(ColBlack)
	ColShot2C  = tcell.StyleDefault.Foreground(tcell.ColorRed        ).Background(ColBlack)
	ColShot3C  = tcell.StyleDefault.Foreground(tcell.NewRGBColor( 33*2,  13*2,  16*2)    ).Background(ColBlack)
	ColShot4C  = tcell.StyleDefault.Foreground(tcell.NewRGBColor( 18*2,  11*2,  12*2)    ).Background(ColBlack)

	cols = map[colorID]tcell.Style{
		EmptyC  : ColEmpty  ,
		P1HeadC : ColP1Head ,
		P2HeadC : ColP2Head ,
		WallC   : ColDefault,
		
		_Shot1C : ColShot1C, 
		_Shot2C : ColShot2C, 
		_Shot3C : ColShot3C, 
		_Shot4C : ColShot4C, 
	}

}


func snakeMake(start Vec2, d direction, stateID cellState) *Snake{

	inputQ := make([]input, 0, 4)
	snake := &Snake{
		pos: start,
		dir: d,
		scpt: SCPT,
		subcellDebt: 0,
		inputQ: inputQ,
		stateID: stateID,
	}

	return snake
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

