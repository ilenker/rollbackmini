package main

import (
	"fmt"
	"log"
	"github.com/gdamore/tcell/v2"
)

var slidersX, slidersY int

type Slider struct {
	X, Y        int     // position of the slider
	Width       int     // width of the slider bar
	Percent     float64 // 0.0 - 1.0
	Value		float64 // 0.0 - 1.0
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
		if y == s.Y && x >= s.X && x <= s.X+s.Width {
			s.Dragging = true
			s.Percent = float64(x-s.X) / float64(s.Width)
			s.Value = lerp(int(s.Min), int(s.Max), s.Percent)
		}
	case tcell.ButtonNone:
		if !s.Dragging {
			if y == s.Y {
				// update Percent while dragging
				if x < s.X {
					x = s.X
				}
				if x > s.X+s.Width {
					x = s.X + s.Width
				}
				s.Percent = float64(x-s.X) / float64(s.Width)
				s.Value = lerp(int(s.Min), int(s.Max), s.Percent)
			}
		}
	default:
		s.Dragging = false
	}
}

func test() {
	screen, err := tcell.NewScreen()
	if err != nil {
		log.Fatalf("Error creating screen: %v", err)
	}
	if err = screen.Init(); err != nil {
		log.Fatalf("Error initializing screen: %v", err)
	}
	defer screen.Fini()

	screen.Clear()
	screen.EnableMouse()

	s := make([]*Slider, 4)

	WIDTH := 100
	s[0] = 
	&Slider{
		X:			10,
		Y:			7,
		Width:		WIDTH,
		Percent:	0.5,
		Value:		0.5 * 10,
		Dragging:	true,
		Name:		"Start Lum",
		Min:		0,
		Max:		10,
	}

	s[1] = 
	&Slider{
		X:			10,
		Y:			9,
		Width:		WIDTH,
		Percent:	0.5,
		Value:		0.5 * 40,
		Dragging:	true,
		Name:		"Dist",
		Min:		0,
		Max:		20,
	}

	s[2] = 
	&Slider{
		X:			10,
		Y:			11,
		Width:		WIDTH,
		Percent:	0.5,
		Value:		0.5 * 40,
		Dragging:	true,
		Name:		"Radius",
		Min:		0,
		Max:		30,
	}

	s[3] = 
	&Slider{
		X:			10,
		Y:			13,
		Width:		WIDTH,
		Percent:	0.5,
		Value:		0.5 * 40,
		Dragging:	true,
		Name:		"Lum",
		Min:		0,
		Max:		10,
	}

	// Main loop
	for {
		screen.Clear()

		n :=
		_light(
			s[0].Value,						// Start lum
			Vec2{10, 10},					// Origin pos
			Vec2{10 + int(s[1].Value), 10},		// Canditate pos
			int(s[2].Value),				// Radius
			s[3].Value,						// Lum
			Vec3[float64]{1.0, 1.0, 1.0})	// Tint


		// SLIDERS
		for _, slider := range s {
			str := fmt.Sprintf("%s: %.3f", slider.Name, slider.Value)
			for i, r := range str {
				screen.SetContent(10+i, slider.Y + 1, r, nil, tcell.StyleDefault.Foreground(tcell.ColorYellow))
			}
			slider.Draw(screen)
		}

		str := fmt.Sprintf("Light Output: %.3f", n)
		for i, r := range str {
			screen.SetContent(10+i, 4, r, nil, tcell.StyleDefault.Foreground(tcell.ColorYellow))
		}

		str = fmt.Sprintf("Result (C:100) = %3d", int(n * 100))
		for i, r := range str {
			screen.SetContent(10+i, 5, r, nil, tcell.StyleDefault.Foreground(tcell.ColorYellow))
		}

		i := 4	
		for i := range i {
			screen.SetContent(43, 2 + i, '█', nil, tcell.StyleDefault.Foreground(
				tcell.NewRGBColor(
					int32(100),
					int32(100),
					int32(100),
					)))
		}

		i = 4
		for i := range i {
			screen.SetContent(43 + int(s[1].Value), 2 + i, '█', nil, tcell.StyleDefault.Foreground(
				tcell.NewRGBColor(
					int32(100 * n),
					int32(100 * n),
					int32(100 * n),
					)))
		}

		screen.SetContent(43 + int(s[2].Value), 1, '█', nil, tcell.StyleDefault.Foreground(tcell.ColorYellow))

		screen.Show()

		ev := screen.PollEvent()
		switch tev := ev.(type) {
		case *tcell.EventKey:
			if tev.Key() == tcell.KeyEscape || tev.Key() == tcell.KeyCtrlC {
				return
			}

		case *tcell.EventMouse:
			for _, slider := range s {
				slider.HandleEvent(tev)
			}
		}

	}
}

func slidersInit(x, y int) {
	slidersX = x
	slidersY = y
	WIDTH := 30
	s[0] = 
	&Slider{
		X:			x,
		Y:			y,
		Width:		WIDTH,
		Dragging:	true,
		Name:		"Gamma",
		Min:		0,
		Max:		2,
	}

	s[1] = 
	&Slider{
		X:			x,
		Y:			y + 2,
		Width:		WIDTH,
		Dragging:	true,
		Name:		"Whitepoint",
		Min:		100,
		Max:		255,
	}

	s[2] = 
	&Slider{
		X:			x,
		Y:			y + 4,
		Width:		WIDTH,
		Dragging:	true,
		Name:		"Blackpoint",
		Min:		0,
		Max:		100,
	}

	s[3] = 
	&Slider{
		X:			x,
		Y:			y + 6,
		Width:		WIDTH,
		Dragging:	true,
		Name:		"Fun",
		Min:		69,
		Max:		420,
	}

}

func sliders() {
	// SLIDERS
	for _, slider := range s {
		str := fmt.Sprintf("%s: %.3f", slider.Name, slider.Value)
		for i, r := range str {
			scr.SetContent(slidersX+i, slider.Y + 1, r, nil, tcell.StyleDefault.Foreground(tcell.ColorYellow))
		}
		slider.Draw(scr)
	}
}
