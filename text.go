package main

import (
	"time"
	"fmt"
	"strings"
	"strconv"
	"github.com/gdamore/tcell/v2"
)

var debugBox func(msg string, args ...int)
var errorBox func(msg string, args ...int)
var variablePage int = 4

type Arg struct {
	name string
	val any
}

func variableDisplay() {
	switch variablePage {
	case 1:
		displayVariables(Arg{"*Player1*", ""},
			Arg{"movdir:", player1.dir},
			Arg{"inputs:", player1.inputBuffer},
			Arg{"inpIdx:", player1.inputIndex},
			Arg{"local :", player1.isLocal},
			Arg{"pos   :", player1.pos},
			Arg{"speed :", player1.scpt},
			Arg{"shot->:", player1.shootDir},
			Arg{"ID    :", player1.stateID},
			Arg{"scDebt:", player1.subcellDebt},
			)
	case 2:
		displayVariables(Arg{"*Player2*", ""},
			Arg{"movdir:", player2.dir},
			Arg{"inputs:", player2.inputBuffer},
			Arg{"inpIdx:", player2.inputIndex},
			Arg{"local :", player2.isLocal},
			Arg{"pos   :", player2.pos},
			Arg{"speed :", player2.scpt},
			Arg{"shot->:", player2.shootDir},
			Arg{"ID    :", player2.stateID},
			Arg{"scDebt:", player2.subcellDebt},
			)
	case 3:
		displayVariables(Arg{"*Gamestate*", ""},
			Arg{"ROLLBACK:", ROLLBACK},
			Arg{"SIM_FRAME:", SIM_FRAME},
			Arg{"RESIM_FRM:", RESIM_FRAME},
			Arg{"[ 0]:", rollbackBuffer.frames[ 0].id},
			Arg{"[ 1]:", rollbackBuffer.frames[ 1].id},
			Arg{"[ 2]:", rollbackBuffer.frames[ 2].id},
			Arg{"[ 3]:", rollbackBuffer.frames[ 3].id},
			Arg{"[ 4]:", rollbackBuffer.frames[ 4].id},
			Arg{"[ 5]:", rollbackBuffer.frames[ 5].id},
			Arg{"[ 6]:", rollbackBuffer.frames[ 6].id},
			Arg{"[ 7]:", rollbackBuffer.frames[ 7].id},
			Arg{"[ 8]:", rollbackBuffer.frames[ 8].id},
			Arg{"[ 9]:", rollbackBuffer.frames[ 9].id},
			)
	case 4:
		displayVariables(Arg{"*Network*", ""},
			Arg{"avgRTT:", fmt.Sprintf("%3dms", avgRTTuSec/1000)})
	}
}

func displayVariables(args ...Arg) {
	debugBox("\\clr")
	for i, arg := range args {
		debugBox(fmt.Sprintf("%s\t%v", arg.name, arg.val), 0, i)
	}
}

func textBoxesInit () {
	debugBox = DrawMessages(scr, MapW + 5 , 4, 30, 30, true)
	errorBox = DrawMessages(scr, MapW + 37, 4, 15, 30, true)
}

// Returns func(msg) with optional args: func(msg, x, y)
// Overrides the internal cursor position
func DrawMessages(scr tcell.Screen, xOrigin, yOrigin, width, height int, drawBox bool) func(msg string, xy ...int) {
	x, y, w, h := xOrigin, yOrigin, width, height
	xO, yO := xOrigin, yOrigin

	if drawBox {
		DrawRoundBox(scr, xOrigin, yOrigin, width, height, tcell.NewHexColor(0xFFFFFF))
	}

	return func(msg string, args ...int) {
		if msg == "\\clr" {
			for i := range h {
				for j := range w {
					scr.SetContent(xO + j, yO + i, ' ', nil, ColDefault)
				}
			}	
			return
		}

		if len(args) > 1 {
			x = (xOrigin + args[0]) % (w + xOrigin)
			y = (yOrigin + args[1]) % (h + yOrigin)
		}
		for _, r := range msg {
			if y > (yOrigin + h) {
				return
			}
			if r != '\n' {
				scr.SetContent(x, y, r, nil, ColDefault)	
				x++
				if x > (xOrigin + w) {
					y++
					x = xO // Line Wrapping
				}
			} else {
				y++     
				if y > height {
					y = yO
				}
				x = xO  // Carriage Return
			}
		}
	}
}


func DrawRoundBox(scr tcell.Screen, x, y, w, h int, col tcell.Color) {
	// Sides
	style := tcell.StyleDefault.Foreground(col).Background(tcell.ColorBlack)
	for i := range h + 1 {
		scr.SetContent(x - 1, y + i, '│', nil, style)
		scr.SetContent(x+w+1, y + i, '│', nil, style)
	}

	// Top/Bottom
	for i := range w + 1 {
		scr.SetContent(x + i, y - 1, '─', nil, style)
		scr.SetContent(x + i, y+h+1, '─', nil, style)
	}

	// Corners
	scr.SetContent(x - 1, y - 1, '╭', nil, style)
	scr.SetContent(x+w+1, y - 1, '╮', nil, style)
	scr.SetContent(x - 1, y+h+1, '╰', nil, style)
	scr.SetContent(x+w+1, y+h+1, '╯', nil, style)
}

func DrawPixelBox(scr tcell.Screen, x, y, w, h int, col tcell.Color) {
	style := tcell.StyleDefault.Foreground(col).Background(tcell.ColorBlack)
	// Sides
	for i := range h + 1 {
		scr.SetContent(x - 1, y + i, '█', nil, style)
		scr.SetContent(x+w+1, y + i, '█', nil, style)
	}

	// Top/Bottom
	for i := range w + 1 {
		scr.SetContent(x + i, y - 1, '▄', nil, style)
		scr.SetContent(x + i, y+h+1, '▀', nil, style)
	}

	// Corners
	scr.SetContent(x - 1, y - 1, '▄', nil, style)
	scr.SetContent(x+w+1, y - 1, '▄', nil, style)
	scr.SetContent(x - 1, y+h+1, '▀', nil, style)
	scr.SetContent(x+w+1, y+h+1, '▀', nil, style)
}


func timeF(d time.Duration) string {
	str := fmt.Sprintf("%v", d)

	if floatStr, found := strings.CutSuffix(str, "ms"); found {
		f, _ := strconv.ParseFloat(floatStr, 64)
		return fmt.Sprintf("%.2fms", f)
	}

	if floatStr, found := strings.CutSuffix(str, "µs"); found {
		f, _ := strconv.ParseFloat(floatStr, 64)
		return fmt.Sprintf("%.2fμs", f)
	}

	if floatStr, found := strings.CutSuffix(str, "s"); found {
		f, _ := strconv.ParseFloat(floatStr, 64)
		return fmt.Sprintf("%.2fs", f)
	}

	return str
}

func intSeps(n int) string {
	if n < 1000 {
		return strconv.Itoa(n)
	}

	s := strconv.Itoa(n)
	result := ""

	for {
		if len(s) <= 3 {
			result = s + result
			return result
		}
		result = "," + s[len(s)-3:] + result
		s = s[:len(s)-3]
	}
}
