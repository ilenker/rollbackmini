package main

import (
	"github.com/gdamore/tcell/v2"
)


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

