package main

import (
	"os"
	"time"
	"math"
	"math/rand"
	"github.com/gdamore/tcell/v2"
)

type colorID = uint8
var COLORTERM_OG string

const (
	EmptyC colorID = iota
	P1HeadC
	P2HeadC
	WallC

	WhiteC

	ShotP1C
	ShotP2C

	Shot2C
	Shot3C
)

var cols map[colorID]tcell.Color

// Color Sequences for beam animation
var p1BeamCols = []colorID {
	WhiteC,
	ShotP1C,
	Shot2C,
	Shot3C,
}

var p1HitCols = []colorID {
	ShotP1C,
	ShotP1C,
	Shot2C,
	Shot3C,
}

var p2BeamCols = []colorID {
	WhiteC,
	ShotP2C,
	ShotP2C,
	Shot3C,
}

var p2HitCols = []colorID {
	ShotP2C,
	ShotP2C,
	Shot2C,
	Shot3C,
}

var beamCols = map[cellState][]colorID{
	P1Head: p1BeamCols,
	P2Head: p2BeamCols,
}

var hitCols = map[cellState][]colorID{
	P1Head: p1HitCols,
	P2Head: p2HitCols,
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

			upperVfx := vfxLayer[lyUpper][x]
			lowerVfx := vfxLayer[lyLower][x]

			if upperVfx != cols[EmptyC] {
				upper = upperVfx
			}

			if lowerVfx != cols[EmptyC] {
				lower = lowerVfx
			}

			r := ' '
			st := tcell.StyleDefault

			r, st = '▀', st.Foreground(upper).Background(lower)

			s.SetContent(x + xOffset, y + yOffset, r, nil, st)

		}
	}

	s.Show()
}

func newRGBOscillator(rgbInit VecRGB) func() tcell.Color {
	rPol := 1
	gPol := 1
	bPol := 1
	rgb := rgbInit
	return
}



func stylesInit() {
	cols = map[colorID]tcell.Color{
		EmptyC   : tcell.ColorBlack,
		P1HeadC  : tcell.ColorBlue,
		P2HeadC  : tcell.ColorOrange,
		WallC    : tcell.ColorWhiteSmoke,

		WhiteC  : tcell.ColorWhite,

		ShotP1C : tcell.ColorBlue,
		ShotP2C : tcell.ColorOrange,

		Shot2C  : tcell.NewRGBColor( 60,  13,  16),
		Shot3C  : tcell.NewRGBColor( 49,  11,  12),
	}
}

func beamEffect(start Vec2, dist int, dir Vec2, colorSeq []colorID) {

	start = start.Add(dir)
	animLen := 26

	animate := func(col colorID, pos Vec2, d int, delay time.Duration, chance int) {
		for range animLen {
			time.Sleep((SIM_TIME/2) * delay)
			if pos.y >= MapH || pos.y < 0 {
				break
			}
			if rand.Intn(20) < chance {
				vfxLayer[pos.y][pos.x] = col
			}
			if d > 0 {
				d--
				pos = pos.AddNoWrap(dir)
			}
		}

	}

	// Frame 1
	animate(colorSeq[0], start, dist, 0, 20)
	time.Sleep(SIM_TIME)
	time.Sleep(SIM_TIME)

	animate(colorSeq[1], start, dist, 0, 20)
	time.Sleep(SIM_TIME)

	animate(colorSeq[2], start, dist, 0, 15)

	animate(colorSeq[3], start, dist, 2, 20)

	animate(EmptyC, start, dist, 2, 10)
	animate(EmptyC, start, dist, 2, 15)
	animate(EmptyC, start, dist, 2, 20)
}


func hitEffect(start Vec2, baseturns float64, colorSeq []colorID) {

	baseturns += -0.3 + rand.Float64() * 0.6

	animLen := 3 + rand.Intn(10)
	turns := baseturns
	curve := -0.05 + rand.Float64() * 0.1

	animate := func(col colorID, pos Vec2, delay time.Duration, chance int, after bool) {
		for range animLen {
			time.Sleep((SIM_TIME/2) * delay)

			if pos.x >= MapW || pos.x < 0 || 
			   pos.y >= MapH || pos.y < 0 {
				break
			}

			if rand.Intn(20) < chance {
				board[pos.y][pos.x].col = col
			}

			turns += curve
			pos = pos.Translate(turns * math.Pi, 1)
		}

		if after {
			go hitEffect2nd(pos, 2 * rand.Float64(), colorSeq)
			go hitEffect2nd(pos, 2 * rand.Float64(), colorSeq)
		}
		turns = baseturns

	}

	animate(colorSeq[0], start, 0, 20, false)
	time.Sleep(SIM_TIME)
	time.Sleep(SIM_TIME)

	animate(EmptyC, start,  0, 20, false)
	time.Sleep(SIM_TIME)
	time.Sleep(SIM_TIME)

	animate(colorSeq[1], start, 0, 20, false)
	time.Sleep(SIM_TIME)

	animate(colorSeq[2], start, 0, 15, false)

	animate(colorSeq[2], start, 1, 20, false)
	animate(EmptyC, start,  1, 10, false)
	animate(EmptyC, start,  2, 15, false)
	animate(EmptyC, start,  2, 20, true)

}

func hitEffectCrit(start Vec2, baseturns float64, colorSeq []colorID) {

	baseturns += -0.2 + rand.Float64() * 0.4

	animLen := 4 + rand.Intn(12)
	turns := baseturns
	curve := -0.05 + rand.Float64() * 0.1

	animate := func(col colorID, pos Vec2, delay time.Duration, chance int, after bool) {
		for range animLen {
			time.Sleep((SIM_TIME/2) * delay)

			if pos.x >= MapW || pos.x < 0 || 
			   pos.y >= MapH || pos.y < 0 {
				break
			}

			if rand.Intn(20) < chance {
				board[pos.y][pos.x].col = col
			}

			turns += curve
			pos = pos.Translate(turns * math.Pi, 1)
		}

		if after {
			go hitEffect2nd(pos, 2 * rand.Float64(), colorSeq)
			go hitEffect2nd(pos, 2 * rand.Float64(), colorSeq)
		}
		turns = baseturns

	}

	animate(colorSeq[0], start, 0, 20, true)
	time.Sleep(SIM_TIME)
	time.Sleep(SIM_TIME)

	animate(EmptyC, start,  0, 20, false)
	time.Sleep(SIM_TIME)
	time.Sleep(SIM_TIME)

	animate(colorSeq[1], start, 0, 20, false)
	time.Sleep(SIM_TIME)

	animate(colorSeq[2], start, 0, 15, true)

	animate(colorSeq[2], start, 1, 20, false)
	animate(EmptyC, start,  1, 10, false)
	animate(EmptyC, start,  2, 15, false)
	animate(EmptyC, start,  2, 20, true)

}


func hitEffect2nd(start Vec2, baseturns float64, colorSeq []colorID) {

	animLen := 0 + rand.Intn(5)

	turns := baseturns
	curve := -0.1 + rand.Float64() * 0.2

	animate := func(col colorID, pos Vec2, delay time.Duration, chance int) {
		for range animLen {
			time.Sleep((SIM_TIME/2) * delay)
			if pos.x >= MapW || pos.x < 0 || 
			   pos.y >= MapH || pos.y < 0 {
				break
			}

			if rand.Intn(20) < chance {
				vfxLayer[pos.y][pos.x] = col
			}

			turns += curve
			pos = pos.Translate(turns * math.Pi, 1)
		}
		turns = baseturns
	}

	// Frame 1
	animate(colorSeq[0], start, 0, 20)
	time.Sleep(SIM_TIME)
	time.Sleep(SIM_TIME)

	animate(colorSeq[1], start, 0, 15)
	animate(colorSeq[2], start, 1, 20)
	animate(EmptyC, start,  3, 10)
	animate(EmptyC, start,  3, 15)
	animate(EmptyC, start,  3, 20)

}


func rollbackStreak(start Vec2, dist int, dir Vec2, colID colorID) {

	start = start.Add(dir)

	animate := func(col tcell.Color, pos Vec2, d int, delay time.Duration, chance int) {
		for range dist {
			time.Sleep((SIM_TIME/2) * delay)
			if pos.y >= MapH || pos.y < 0 {
				break
			}
			if rand.Intn(20) < chance {
				vfxLayer[pos.y][pos.x] = col
			}
			if d > 0 {
				d--
				pos = pos.Add(dir)
			}
		}

	}

	// Frame 1
	animate(cols[WhiteC], start, dist, 0, 20)
	time.Sleep(SIM_TIME)
	animate(cols[EmptyC], start, dist, 1, 20)
	animate(cols[EmptyC], start, dist, 1, 20)

}


func cooldownBar(origin Vec2, length int, colorID colorID) {

	if length == 0 {
		vfxLayer[origin.y][origin.x] = cols[EmptyC]
	}

	for i := range length {
		vfxLayer[origin.y][origin.x + i + 1] = cols[EmptyC]
		vfxLayer[origin.y][origin.x + i    ] = cols[colorID]
	}

}


func setCOLORTERM() {
	COLORTERM_OG = os.Getenv("COLORTERM")
	os.Setenv("COLORTERM", "truecolor")	
}

func restoreCOLORTERM() {
	os.Setenv("COLORTERM", COLORTERM_OG)	
}
