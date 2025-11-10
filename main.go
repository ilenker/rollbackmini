package main

import (
	"fmt"
	"log"
	"os"
	"runtime"
	"time"
	"github.com/gdamore/tcell/v2"
	"math/rand"
)

var MOVES_PER_TICK int
var LastInputT time.Time
const SIM_TIME = time.Duration(time.Millisecond * 16)
const SCPT = 64
const RB_BUFFER_LEN = 15

var scr tcell.Screen
var err error
var lBorder = (MapW / 2) - 5
var rBorder = (MapW / 2) + 5

const LOCAL = 0
const PEER  = 1

func main() {
							/*#### INIT ####*/
	SIM_FRAME = 0
	snakes := make([]*Snake, 2)
	stylesInit()
	boardInit(&board)
	scr, err = tcell.NewScreen()
	F(err, "")
	err = scr.Init()
	F(err, "")
	defStyle := tcell.StyleDefault.Background(tcell.ColorBlack).Foreground(tcell.ColorBlack)
	scr.SetStyle(defStyle)

	rollbackBuffer := RollbackBuffer{}
							/*##############*/


	snakes[LOCAL] = snakeMake(Vec2{MapW/2, 8}, R)
	snakes[PEER ] = snakeMake(Vec2{0,  5}, R)
	localInputCh := make(chan input, 8)

	snakes[LOCAL].scpt = 32
	snakes[PEER ].scpt = 16

	board[MapH/2][rBorder].col = Wall
	board[MapH/2][lBorder].col = Wall

	go readInputs(scr, localInputCh)

	// Here well have "go readPeer()" or something
	simTick := time.NewTicker(SIM_TIME)


	DrawPixelBox(scr, 2, 2, MapW - 1, MapH/2 - 1, tcell.ColorBlue)

	//avgSimT := makeAverageDurationBuffer(100)
	//startT := time.Now()

	debugBox = DrawMessages(scr, MapW + 5, 4, 30, 30, true)

	//var _dMicroSec int64 = 0
	rollbackSize := 0 

	debugPrint := func () {
		{
			//debugBox(fmt.Sprintf("Frame Time:\t[%.2f]", float64(_dMicroSec)/1000), 0, 0)
			debugBox(fmt.Sprintf("Rollback Size:\t[%2d]", rollbackSize), 0, 0)
			debugBox(fmt.Sprintf("sim_frame:\t%5d", SIM_FRAME), 0, 1)
			//debugBox(fmt.Sprintf("MPT:\t\t%2d(%.2f)", MOVES_PER_TICK, float64(snakes[0].scpt) / float64(SUBCELL_SIZE)), 0, 2)
			//debugBox(fmt.Sprintf("SCPT:\t\t%3d", snakes[0].scpt), 0, 3)
		}
	} 
	render(scr, 2, 2)

	// qwfploop
	for {
		//startT = time.Now()
		MOVES_PER_TICK = 0
		<-simTick.C

		if SIM_FRAME > RB_BUFFER_LEN {
			
			if SIM_FRAME % 10 == 0 {
				
				rollbackSize = 1 + rand.Intn(RB_BUFFER_LEN)

				var q []input = make([]input, 4)
				ilen := rand.Intn(3)
				for i := range ilen {
					q[i] = input(1 + rand.Intn(3))
				}

				rollbackBuffer.resimFramesWithNewInputs(SIM_FRAME - uint32(rollbackSize),
					q,
					&board, snakes)
			}

		}

		if snakes[LOCAL].pos.x >= rBorder {
			localInputCh <-iLeft
		}
		if snakes[LOCAL].pos.x <= lBorder {
			localInputCh <-iRight
		}
		drainInputChToSnake(localInputCh, snakes, LOCAL)

		rollbackBuffer.pushFrame(copyCurrentFrameData(&board, snakes, SIM_FRAME))

		updateLogic(snakes)

		SIM_FRAME++
		//_dMicroSec = avgSimT(time.Since(startT))
		debugPrint()
		render(scr, 2, 2)

		if snakes[LOCAL].pos.x > rBorder ||
			snakes[LOCAL].pos.x < lBorder {
			scr.PollEvent()
		}

	}
}


// This function should not be aware of input sources
// Could be local user, multiplayer peer, bot
// This only updates the snake state.
// We'll see how this works out (input validation based on board state?)
func updateLogic(snakes []*Snake) {

	for _, s := range snakes {
		subcellBudget := s.scpt - s.subcellDebt
		//debugBox2(fmt.Sprintf("subcellBudget:\t[%2d]", subcellBudget), 0, 1)
		//debugBox2(fmt.Sprintf("scpt[%3d]-debt[%3d]", s.scpt, s.subcellDebt), 0, 2)

		for {
			if subcellBudget <= 0 {
				s.subcellDebt = AbsInt16(subcellBudget)
				break
			}

			// Here we pop from the input queue
			// If our scpt is greater than the subcell size
			// Then we may pop more than one input per simtick.
			controlSnake(s)

			//if s.halving {
			//	s.half()
			//}

			//tailCell := cellGet(&board, s.tail.pos)
			cellSet(&board, s.pos, Empty, Empty)
			//cellSet(&board, s.tail.pos, Empty, Empty)

			//if tailCell.state == P1Food {
			//	s.grow()
			//	tailCell.state = Empty
			//} else {
			s.move()
			//}

			//headCell := cellGet(&board, s.head.pos)
			//if headCell.state == Portal {
			//	tailPos := s.tail.pos
			//	s.port(headCell.connection)
			//	cellSet(&board, tailPos, Empty, Empty)
			//	headCell = cellGet(&board, s.head.pos)
			//}

			//if headCell.state == P1Body {
			//	s.eatSelf()
			//	cellSet(&board, s.head.pos, Empty, Empty)
			//}

			cellSet(&board, s.pos, P1Head, P1Head)
			//cellSet(&board, s.tail.pos, P1Tail, P1Tail)
			//debugBox2(fmt.Sprintf("len:%2d\t", s.length), 0, 0)

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
			style := tcell.StyleDefault

			// Blend the two 'styles'
			// take foreground color of each logical cell
			// set foreground of rune to upper color
			// set background of rune to lower color 
			// half-block ▀ displays top color (fg) and bottom color (bg) in one cell
			switch {
			case upper != tcell.StyleDefault && lower != tcell.StyleDefault:
				fg, _, _ := upper.Decompose()
				bg, _, _ := lower.Decompose()
				blend := tcell.StyleDefault.Foreground(fg).Background(bg)
				r, style = '▀', blend

			case upper != tcell.StyleDefault:
				r, style = '▀', upper

			case lower != tcell.StyleDefault:
				r, style = '▄', lower
			}

			s.SetContent(x + xOffset, y + yOffset, r, nil, style)
		}
	}

	for i := range 20 {
		s.SetContent(2+i, MapH, rune(((i+1) % 10)+48), nil, ColDefault)
	}

	s.Show()
}


// Collect all (local) input and send down a single channel
func readInputs(scr tcell.Screen, inputCh chan input) {

	for {
		ev := scr.PollEvent()
		if key, ok := ev.(*tcell.EventKey); ok {

			if key.Key() == tcell.KeyESC {
				scr.Fini()
				os.Exit(0)
			} 

			// Keymap
			switch key.Rune() {


				// "WASD"
			case 'x':
				inputCh <-iLeft
			case 'd':
				inputCh <-iRight
			case 'f':
				inputCh <-iRight
				inputCh <-iLeft
				inputCh <-iRight
				inputCh <-iLeft

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
func cellSet(board *[MapH+1][MapW+1]Cell, vec Vec2, newState cellState, newCol cellState) {
	switch board[vec.y][vec.x].state {
	default: 
		board[vec.y][vec.x].state = newState
		board[vec.y][vec.x].col   = newCol
	}

}

func cellGet(board *[MapH+1][MapW+1]Cell, vec Vec2) *Cell {
	return &board[vec.y][vec.x]
}


func stylesInit() {
	ColEmpty   = tcell.StyleDefault.Foreground(tcell.ColorBlack      ).Background(tcell.ColorBlack)
	ColP1Head  = tcell.StyleDefault.Foreground(tcell.ColorGreen      ).Background(tcell.ColorBlack)
	ColP2Head  = tcell.StyleDefault.Foreground(tcell.ColorGreen      ).Background(tcell.ColorBlack)
	ColDefault = tcell.StyleDefault.Foreground(tcell.ColorOrange     ).Background(tcell.ColorBlack)

	cols = map[cellState]tcell.Style{
		Empty  : ColEmpty  ,

		P1Head : ColP1Head ,

		P2Head : ColP2Head ,

		Wall   : ColDefault,
	}

}


func snakeMake(start Vec2, d direction) *Snake{

	inputQ := make([]input, 0, 4)
	snake := &Snake{
		pos: start,
		dir: d,
		scpt: SCPT,
		subcellDebt: 0,
		inputQ: inputQ,
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

