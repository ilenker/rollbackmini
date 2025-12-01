package main

import (
	"fmt"
	"math"
	"github.com/gdamore/tcell/v3"

)

var slidersX, slidersY int

func startLightTestBench() {
							/*#### INIT ####*/
	setCOLORTERM()
	defer restoreCOLORTERM()
	scr, err = tcell.NewScreen()	;F(err, "")
	err = scr.Init()				;F(err, "")
	scr.EnableMouse()
	defer scr.Fini()
	stylesInit()

	boardInit()

	const (
		gamma = iota
		dist
		radius
		lum
		falloff
	)

	// kSliders
	s := make([]*Slider, 5)
	WIDTH := 70
	spacing := 3
	x := MapW * 2 + 5
	y := 2
	i := 0
	s[i] = 
	&Slider{
		X:			x,
		Y:			y + i * spacing,
		Width:		WIDTH,
		Dragging:	true,
		Name:		"γ",
		Min:		0,
		Max:		2,
	}

	i++
	s[i] = 
	&Slider{
		X:			x,
		Y:			y + i * spacing,
		Width:		WIDTH,
		Dragging:	true,
		Name:		"Dist",
		Min:		0,
		Max:		50,
	}

	i++
	s[i] = 
	&Slider{
		X:			x,
		Y:			y + i * spacing,
		Width:		WIDTH,
		Dragging:	true,
		Name:		"Radius",
		Min:		0,
		Max:		30,
	}

	i++
	s[i] = 
	&Slider{
		X:			x,
		Y:			y + i * spacing,
		Width:		WIDTH,
		Dragging:	true,
		Name:		"Lum",
		Min:		0,
		Max:		4,
	}

	i++
	s[i] = 
	&Slider{
		X:			x,
		Y:			y + i * spacing,
		Width:		WIDTH,
		Dragging:	true,
		Name:		"Falloff",
		Min:		0,
		Max:		3,
	}

	s[gamma  ].Value = 1.0
	s[dist   ].Value = 1.0
	s[radius ].Value = 10.0
	s[lum    ].Value = 1.5
	s[falloff].Value = 0.01

	s[gamma  ].Percent = iLerp(int(s[0].Min), int(s[0].Max), s[0].Value)
	s[dist   ].Percent = iLerp(int(s[1].Min), int(s[1].Max), s[1].Value)
	s[radius ].Percent = iLerp(int(s[2].Min), int(s[2].Max), s[2].Value)
	s[lum    ].Percent = iLerp(int(s[3].Min), int(s[3].Max), s[3].Value)
	s[falloff].Percent = iLerp(int(s[4].Min), int(s[4].Max), s[4].Value)


	checkerBoard := func() {
		cellSize := 2
		for y := range MapH {
			for x := range MapW {
				i := fi(x, y)
				if (x/cellSize + y/cellSize) % 2 == 0 {
					copyRGB(&vfxLayer, i, ω, β, β)
				} else {
					copyRGB(&vfxLayer, i, β, β, β)
				}
			}
		}
	}
	checkerBoard()

	button := Button{
		X: MapW * 2 + 4,
		Y: 20,
		Width: 10,
		Label: "reset",
		OnClick: func(){
			boardInit()
			checkerBoard()
		},
	}

	// Main loop
	for {
		scr.Clear()
		FOFactor = s[falloff].Value
		for _, slider := range s {
			str := fmt.Sprintf("%s: %.3f", slider.Name, slider.Value)
			for i, r := range str {
				scr.SetContent(x+i, slider.Y + 1, r, nil, tcell.StyleDefault.Foreground(tcell.ColorYellow))
			}
			slider.Draw(scr)
		}

		button.Draw(scr)

		// Radius indicator
		scr.SetContent(
			1 + MapW / 2 + int(s[radius].Value),
			MapH / 2 + 1,
			'▲', nil,
			tcell.StyleDefault.Foreground(tcell.ColorYellow))

		scr.SetContent(
			1 + MapW / 2,
			MapH / 2 + 1,
			'△', nil,
			tcell.StyleDefault.Foreground(tcell.ColorOrange))

		ev := <-scr.EventQ()
		switch tev := ev.(type) {
		case *tcell.EventKey:
			if tev.Key() == tcell.KeyEscape || tev.Key() == tcell.KeyCtrlC {
				return
			}

		case *tcell.EventMouse:
			for _, slider := range s {
				slider.HandleEvent(tev)
			}
			button.HandleEvent(tev)
		}

		γ = s[0].Value
		light(
			Vec2{MapW/2, MapH/2},
			int(s[radius].Value),
			s[lum].Value,
			Vec3[float64]{0.5, 0.5, 1.0},)

		_render(scr, 1       , 1, false)
		_render(scr, 3 + MapW, 1, true)
		scr.Show()
	}
}



// position, radius, luminance, color
func _light(lightR float64, p, p2 Vec2, r int, lum float64, c Vec3[float64]) float64 {

	x := p2.x
	y := p2.y

	minL := float64(1.0)

	dMax := (r + 1) * (r + 1)

	d := squareDist(p, Vec2{x, y})

	//if d > float64(dMax) {
	//	return 1
	//}

	dNormal := iLerp64(0, dMax, d * d * FOFactor)

	if dNormal > 1 { return 1.0 }

	rLum := lerp64(lum, minL, dNormal) * c.x

	kMax := 5.0
	kLightR := clamp( (rLum + lightR) * 0.5, 1, kMax)

	return kLightR
}


func _render(s tcell.Screen, xOffset, yOffset int, gammaOn bool) {

	if shadows {
		lPoint := Vec2{MapW / 2, MapH / 2}
		dir := angleTo(lPoint, player2.pos)
		p2 := player2.pos.translate(dir, 50)
		p1 := player2.pos.translate(dir, 1)
		f := 0.0

		for {
			if f > 1 { break }

			x := math.Round(lerp(p1.x, p2.x, f))
			y := math.Round(lerp(p1.y, p2.y, f))

			if x > MapW || x < 0 { break }
			if y > MapH || y < 0 { break }

			i := fi(int(x), int(y))
			copyRGB(&lightVal, i, 0.1, 0.1, 0.1)
			f += 0.010
		}

		dir = angleTo(lPoint, player1.pos)
		p2 = player1.pos.translate(dir, 50)
		p1 = player1.pos.translate(dir, 1)
		f = 0.0

		for {
			if f > 1 { break }

			x := math.Round(lerp(p1.x, p2.x, f))
			y := math.Round(lerp(p1.y, p2.y, f))

			if x > MapW || x < 0 { break }
			if y > MapH || y < 0 { break }

			i := fi(int(x), int(y))
			copyRGB(&lightVal, i, 0.15, 0.15, 0.15)
			f += 0.010
		}
	}

	lightVal.rs[ fiVec2(getLocalPlayerPtr().pos) ] = 1

	//calculateLighting()

	var upper, lower tcell.Color

	// For each terminal row (board y-coordinates map 2:1 onto terminal y-coordinates)
	for y := range (MapH / 2) {
		lyUpper := y * 2           // Calculate corresponding Logical Row, given Terminal Row
		lyLower := y * 2 + 1

		// For each terminal cell (board x-coordinates map 1:1 onto terminal y-coordinates)
		for x := range MapW {
			iU := lyUpper * MapW + x
			iL := lyLower * MapW + x

			//newR := 255.0 * math.Pow((renderBuffer.rs[iU] / 255), γ)

			// Gamma correction
			if gammaOn {
				newR := gammaCorrection(renderBuffer.rs[iU])
				newG := gammaCorrection(renderBuffer.gs[iU])
				newB := gammaCorrection(renderBuffer.bs[iU])
				upper = tcell.NewRGBColor( int32(newR), int32(newG), int32(newB) )

				newR  = gammaCorrection(renderBuffer.rs[iL])
				newG  = gammaCorrection(renderBuffer.gs[iL])
				newB  = gammaCorrection(renderBuffer.bs[iL])
				lower = tcell.NewRGBColor( int32(newR), int32(newG), int32(newB))
			} else {
				upper = tcell.NewRGBColor(
					int32(renderBuffer.rs[iU]),
					int32(renderBuffer.gs[iU]),
					int32(renderBuffer.bs[iU]))
				lower = tcell.NewRGBColor(
					int32(renderBuffer.rs[iL]),
					int32(renderBuffer.gs[iL]),
					int32(renderBuffer.bs[iL]))
			}

			decay := false
			if decay {
				lightDecay := 0.99
				limit := 0.5
				lightVal.rs[iU] = clampMin(lightVal.rs[iU] * lightDecay, limit)
				lightVal.gs[iU] = clampMin(lightVal.gs[iU] * lightDecay, limit)
				lightVal.bs[iU] = clampMin(lightVal.bs[iU] * lightDecay, limit)

				lightVal.rs[iL] = clampMin(lightVal.rs[iL] * lightDecay, limit)
				lightVal.gs[iL] = clampMin(lightVal.gs[iL] * lightDecay, limit)
				lightVal.bs[iL] = clampMin(lightVal.bs[iL] * lightDecay, limit)
			}

			r := '▀'
			st := tcell.StyleDefault.Foreground(upper).Background(lower)
			s.SetContent(x + xOffset, y + yOffset, r, nil, st)
		}
	}

}



func slidersInit(x, y int) {
	s = make([]*Slider, 4)

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

	s[0].Value = γ 
	s[0].Percent = iLerp(int(s[0].Min), int(s[0].Max), γ)

	s[1].Value = ω 
	s[1].Percent = iLerp(int(s[1].Min), int(s[1].Max), ω)

	s[2].Value = β 
	s[2].Percent = iLerp(int(s[2].Min), int(s[2].Max), β)
}
