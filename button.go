package main

import (
	"github.com/gdamore/tcell/v2"
)

type Button struct {
	X, Y   int
	Width  int
	Label  string
	Hover  bool
	Down   bool
	OnClick func()
}

func (b *Button) Draw(screen tcell.Screen) {
	style := tcell.StyleDefault.Foreground(tcell.ColorBlack).Background(tcell.ColorGray)
	if b.Hover {
		style = style.Background(tcell.ColorLightGray)
	}
	if b.Down {
		style = style.Background(tcell.ColorGreen)
	}

	// Button box
	for i := 0; i < b.Width; i++ {
		screen.SetContent(b.X+i, b.Y, ' ', nil, style)
	}

	// Centered text
	start := b.X + (b.Width-len(b.Label))/2
	for i, r := range b.Label {
		screen.SetContent(start+i, b.Y, r, nil, style)
	}
}

func (b *Button) HandleEvent(ev *tcell.EventMouse) {
	x, y := ev.Position()
	inside := y == b.Y && x >= b.X && x < b.X+b.Width

	switch ev.Buttons() {
	case tcell.Button1: // mouse down
		if inside {
			b.Down = true
		}
	case tcell.ButtonNone: // mouse up
		wasDown := b.Down
		b.Down = false
		if inside && wasDown {
			if b.OnClick != nil {
				b.OnClick()
			}
		}
	}

	b.Hover = inside
}

