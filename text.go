package main

import (
	"time"
	"fmt"
	"strings"
	"strconv"
	"github.com/gdamore/tcell/v2"
)

type textBox func(msg string, args ...int) (int, int)

func (t textBox) Clear() {
	t("\\clr")
}

type Arg struct {
	name string
	val any
}

var debugBox textBox
var errorBox textBox
var callsBox textBox
var scoreBox textBox
var frameBox textBox
var variablePage int = 3


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
		diffTarget :=
		float64(avgRTTuSec / 2) /
		float64(SIM_TIME.Microseconds())
		adjust :=
		time.Duration(avgFrameDiff -
			diffTarget)

		displayVariables(Arg{"*Gamestate*", ""},
			Arg{"ROLLBACK :", ROLLBACK},
			Arg{"SIM_FRAME:", SIM_FRAME},
			Arg{"Diff Targ:", fmt.Sprintf("%.2f", diffTarget)},
			Arg{"Adjust   :", int(adjust)},
			)
	case 4:
		target := ((avgRTTuSec/1000)/2) / SIM_TIME.Milliseconds()
		displayVariables(Arg{"*Network*", ""},
			Arg{"avgRTT:", fmt.Sprintf("%3dms", avgRTTuSec/1000)},
			Arg{"avgFrameDiff:", fmt.Sprintf("%.2f", avgFrameDiff)},
			Arg{"target:", fmt.Sprintf("%d", target)})
	case 5:
		if len(frameDiffs) < 1 { return }
		displayVariables(Arg{"*Frame Diffs*", ""},
			Arg{"", fmt.Sprintf("%2d | %2d", frameDiffs[0], frameDiffs[10])},
			Arg{"", fmt.Sprintf("%2d | %2d", frameDiffs[1], frameDiffs[11])},
			Arg{"", fmt.Sprintf("%2d | %2d", frameDiffs[2], frameDiffs[12])},
			Arg{"", fmt.Sprintf("%2d | %2d", frameDiffs[3], frameDiffs[13])},
			Arg{"", fmt.Sprintf("%2d | %2d", frameDiffs[4], frameDiffs[14])},
			Arg{"", fmt.Sprintf("%2d | %2d", frameDiffs[5], frameDiffs[15])},
			Arg{"", fmt.Sprintf("%2d | %2d", frameDiffs[6], frameDiffs[16])},
			Arg{"", fmt.Sprintf("%2d | %2d", frameDiffs[7], frameDiffs[17])},
			Arg{"", fmt.Sprintf("%2d | %2d", frameDiffs[8], frameDiffs[18])},
			Arg{"", fmt.Sprintf("%2d | %2d", frameDiffs[9], frameDiffs[19])})
	case 6:
		if len(RTTs) < 1 { return }
		displayVariables(Arg{"*Ping Times*", ""},
			Arg{"", fmt.Sprintf("%3d | %3d", RTTs[0]/1000, RTTs[5]/1000)},
			Arg{"", fmt.Sprintf("%3d | %3d", RTTs[1]/1000, RTTs[6]/1000)},
			Arg{"", fmt.Sprintf("%3d | %3d", RTTs[2]/1000, RTTs[7]/1000)},
			Arg{"", fmt.Sprintf("%3d | %3d", RTTs[3]/1000, RTTs[8]/1000)},
			Arg{"", fmt.Sprintf("%3d | %3d", RTTs[4]/1000, RTTs[9]/1000)})
	}
}


func displayVariables(args ...Arg) {
	debugBox.Clear()
	for i, arg := range args {
		debugBox(fmt.Sprintf("%s\t%v", arg.name, arg.val), 0, i)
	}
}


func textBoxesInit () {
	debugBox = newTextBox(scr, MapW + 5 , 2, 20, 16, true)
	errorBox = newTextBox(scr, MapW + 28, 2, 30, 16, true)
	callsBox = newTextBox(scr, MapW + 61, 2, 20, 16, true)

	scoreBox = newTextBox(
		scr,
		MapW / 2 - 3, 1,
		8, 1, false)

	frameBox = newTextBox(
		scr,
		MapW / 2 - 3, 0,
		8, 1, false)
}

// Returns func(msg) with optional args: func(msg, x, y)
// Overrides the internal cursor position
func newTextBox(
	scr 			 tcell.Screen,
	xOrigin, yOrigin int,
	width  , height  int,
	drawBox 		 bool) textBox {

	x, y, w, h := xOrigin, yOrigin, width, height
	xO, yO := xOrigin, yOrigin

	if drawBox {
		drawRoundBox(scr, xOrigin, yOrigin, width, height, textCol)
	}

	return func(msg string, args ...int) (int, int) {
		if msg == "\\clr" {
			for i := range h {
				for j := range w {
					scr.SetContent(xO + j, yO + i, ' ', nil, tcell.StyleDefault)
				}
			}	
			x = xO
			y = yO
			return x, y
		}

		if len(args) > 1 {
			x = (xOrigin + args[0]) % (w + xOrigin)
			y = (yOrigin + args[1]) % (h + yOrigin)
		}

		for _, r := range msg {
			if y > (yOrigin + h) {
				return x, y
			}
			if r != '\n' {
				scr.SetContent(x, y, r, nil, stText)	
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
		return x - xO, y - yO
	}
}


func drawRoundBox(scr tcell.Screen, x, y, w, h int, col tcell.Color) {
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


func drawPixelBox(scr tcell.Screen, x, y, w, h int, col tcell.Color) {
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


func loadingInfo(x, y int) (int, int) {

	if SIM_FRAME == 70 {
		x, y = debugBox("...OK!\n", x, y)
	}

	if SIM_FRAME == 90 {
		x, y = debugBox("Auto-detecting graphics")
	}

	if SIM_FRAME == 300 {
		x, y = debugBox(" \nPreset to:\n - Ultra High + RTX ON", x, y)
	}

	if SIM_FRAME == 600 - 180 {
		x, y = debugBox(" \nStarting in\n...3", x, y)
	}

	if SIM_FRAME == 600 - 120 {
		x, y = debugBox("\n...2", x, y)
	}

	if SIM_FRAME == 600 - 60 {
		x, y = debugBox("\n...1", x, y)
	}

	if (SIM_FRAME < 70) || 
	(SIM_FRAME > 90 &&
	SIM_FRAME < 420){
		spinner(x, y, 3, debugBox)
	}

	return x, y
}


func spinner(x, y, frameTime int, textBox func(msg string, args ...int) (int, int)) {
	animationFrames := 8
	switch int(SIM_FRAME) % (frameTime * animationFrames) {
	case frameTime * 0: textBox("⣷", x, y)
	case frameTime * 1: textBox("⣯", x, y)
	case frameTime * 2: textBox("⣟", x, y)
	case frameTime * 3: textBox("⡿", x, y)
	case frameTime * 4: textBox("⢿", x, y)
	case frameTime * 5: textBox("⣻", x, y)
	case frameTime * 6: textBox("⣽", x, y)
	case frameTime * 7: textBox("⣾", x, y)
	}
}


func newBarGraph(x, y int) func(int) {

	ColGraph := tcell.ColorSeaGreen
	width  :=  80
	height :=  10

	drawPixelBox(scr, x, y, width, height, tcell.ColorDarkSlateGrey)

		//'0', nil, tcell.StyleDefault)

	//scr.SetContent(x - 2,
	//	y + int(float64(height) * float64(0.5)),
	//	'0', nil, tcell.StyleDefault)
	//	//'1', nil, tcell.StyleDefault)
		//'0', nil, tcell.StyleDefault)

		//'2', nil, tcell.StyleDefault)
		//'0', nil, tcell.StyleDefault)

	counter := x - 1

	return func(n int) {
		if n == 0 { return }

		counter++
		if counter > x + width {
			counter = x
		}

		for i := range height + 1 {
			scr.SetContent(counter,
				(y + height) - i,
				' ', nil, tcell.StyleDefault.Foreground(tcell.ColorBlack))
		}

		v := newVecRGB(ColGraph.RGB())

		for i := range n / 2 {

			i_ := i
			if i > height {
				i_ = wrapInt(i_ - 1, height)
			}

			v = v.add(VecRGB{int32(10*i), int32(-5*i), int32(-5*i)})

			scr.SetContent(counter,
				(y + height) - i_,
				'█', nil, stDef.Foreground(tcell.NewRGBColor(v.r, v.g, v.b)))
		}

		if n % 2 == 1 {
			n /= 2
			col := stDef.Foreground(tcell.NewRGBColor(v.r, v.g, v.b))

			if n > height {
				n = wrapInt(n - 1, height)

				_, _, b, _ := scr.GetContent(counter, (y + height) - n)
				bg, _, _ := b.Decompose()
				col = col.Background(bg)
			}

			scr.SetContent(counter,
				(y + height) - n,
				'▄', nil, col)
		}

		for i := range height + 1 {
			_c := counter
			f := -10
			if counter > x + width - 1 { return }
			scr.SetContent(
				counter + 1,
				(y + height) - i,
				'│', nil, tcell.StyleDefault.Foreground(tcell.Color146).Background(tcell.Color233))

			for j := 1; j < 10; j++ {
				if _c + j == x + width {
					_c = x - j - 1
				}

				lookahead := _c + 1 + j
				r, _, st, _ := scr.GetContent(lookahead, (y + height) - i)
				fg, bg, _ := st.Decompose()

				if fg != tcell.ColorDefault {
					fg = addRBGtoColor(VecRGB{int32(f), int32(f), int32(f)}, fg)
				}
				if bg != tcell.ColorDefault {
					bg = addRBGtoColor(VecRGB{int32(f), int32(f), int32(f)}, bg)
				}
				scr.SetContent(
					lookahead,
					(y + height) - i,
					r, nil, tcell.StyleDefault.Foreground(fg).Background(bg))
			}
		}

	}
}

func newLineGraph(x, y int) func(int) {

	ColGraph := tcell.ColorSeaGreen
	width  := 130
	height :=  10

	drawPixelBox(scr, x, y, width, height, tcell.ColorLightGreen)

	scr.SetContent(x - 1,
		y + int(float64(height) * float64(1)),
		'0', nil, tcell.StyleDefault)
		//'0', nil, tcell.StyleDefault)

	scr.SetContent(x - 2,
		y + int(float64(height) * float64(0.5)),
		'1', nil, tcell.StyleDefault)
		//'1', nil, tcell.StyleDefault)
	scr.SetContent(x - 1,
		y + int(float64(height) * float64(0.5)),
		'0', nil, tcell.StyleDefault)
		//'0', nil, tcell.StyleDefault)

	scr.SetContent(x - 2,
		y + int(float64(height) * float64(0)),
		'2', nil, tcell.StyleDefault)
		//'2', nil, tcell.StyleDefault)
	scr.SetContent(x - 1,
		y + int(float64(height) * float64(0)),
		'0', nil, tcell.StyleDefault)
		//'0', nil, tcell.StyleDefault)

	counter := 1

	return func(x int) {
		if x == 0 { return }

		counter++
		if counter > width + 2 {
			counter = 2 
		}

		for i := range height + 1 {
			scr.SetContent(counter,
				(y + height) - i,
				'█', nil, tcell.StyleDefault.Foreground(tcell.ColorBlack))
		}

		v := newVecRGB(ColGraph.RGB())

		i := x / 2 

		i_ := i

		if i > height {
			i_ = wrapInt(i_ - 1, height)
		}

		v = v.add(VecRGB{int32(20*i), int32(-10*i), int32(-10*i)})

		if x % 2 == 0 {
			scr.SetContent(counter,
				(y + height) - i_,
				'▀', nil, stDef.Foreground(tcell.NewRGBColor(v.r, v.g, v.b)))
		} else {
			scr.SetContent(counter,
				(y + height) - i_,
				'▄', nil, stDef.Foreground(tcell.NewRGBColor(v.r, v.g, v.b)))

		}

		//if x % 2 == 1 {
		//	x /= 2
		//	col := stDef.Foreground(tcell.NewRGBColor(v.r, v.g, v.b))

		//	if x > height {
		//		x = wrapInt(x - 1, height)

		//		_, _, b, _ := scr.GetContent(counter, (y + height) - x)
		//		bg, _, _ := b.Decompose()
		//		col = col.Background(bg)
		//	}

		//	scr.SetContent(counter,
		//		(y + height) - x,
		//		'▄', nil, col)
		//}

		for i := range height + 1 {
			if counter > width { return }
			scr.SetContent(counter + 1,
				(y + height) - i,
				'│', nil, tcell.StyleDefault.Foreground(tcell.Color122).Background(tcell.Color233))
		}

	}
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


func setColor(x, y int, color tcell.Color) {
	r, _, _, _ := scr.GetContent(x, y)
	scr.SetContent(x, y, r, nil, tcell.StyleDefault.Foreground(color))
}
