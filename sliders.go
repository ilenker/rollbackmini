package main

import (
	"fmt"
	"github.com/gdamore/tcell/v2"
)

type Slider struct {
	X, Y        int     // position of the slider
	Width       int     // width of the slider bar
	Percent     float64 // 0.0 - 1.0
	Value		float64
	Dragging    bool
	Name		string
	Min, Max	float64
}

func (s *Slider) Draw(screen tcell.Screen) {
	barStart := s.X
	barEnd := s.X + s.Width

	// Draw bar
	for x := barStart; x <= barEnd; x++ {
		r := '-'
		screen.SetContent(x, s.Y, r, nil, tcell.StyleDefault.Foreground(tcell.ColorWhite))
	}
	// Draw handle
	handlePos := barStart + int(s.Percent*float64(s.Width))
	screen.SetContent(handlePos, s.Y, '█', nil, tcell.StyleDefault.Foreground(tcell.ColorGreen))
}

func (s *Slider) HandleEvent(ev *tcell.EventMouse) {
	x, y := ev.Position()
	button := ev.Buttons()

	switch button {
	case tcell.Button1:
		if (AbsInt(y - s.Y) < 2) && x >= s.X && x <= s.X+s.Width {
			s.Percent = float64(x-s.X) / float64(s.Width)
			s.Value = lerp64(s.Min, s.Max, s.Percent)
		}

	default:
	}
}

func slidersDraw() {
	// SLIDERS
	for _, slider := range s {
		str := fmt.Sprintf("%s: %.3f", slider.Name, slider.Value)
		for i, r := range str {
			scr.SetContent(slidersX+i, slider.Y + 1, r, nil, tcell.StyleDefault.Foreground(tcell.ColorYellow))
		}
		slider.Draw(scr)
	}
}
