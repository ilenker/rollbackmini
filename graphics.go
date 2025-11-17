package main

import (
	"time"
	"math"
	"math/rand"
	"github.com/gdamore/tcell/v2"
)

type colorID uint8

const (
	EmptyC colorID = iota
	P1HeadC
	P2HeadC
	WallC

	_WhiteC

	_ShotP1C
	_ShotP2C

	_Shot2C
	_Shot3C
)


// These are tied 1:1 to a cellState...
var ColEmpty   tcell.Style
var ColP1Head  tcell.Style
var ColP2Head  tcell.Style
var ColDefault tcell.Style

var cols map[colorID]tcell.Style  // ...via this map.


// These have no tie to a cellState
var ColWhite   tcell.Style
var ColShotP1C tcell.Style
var ColShotP2C tcell.Style
var ColShot2C  tcell.Style
var ColShot3C  tcell.Style


// Color Sequences for beam animation
var p1BeamCols = []colorID {
	_WhiteC,
	_ShotP1C,
	_Shot2C,
	_Shot3C,
}

var p1HitCols = []colorID {
	_ShotP1C,
	_ShotP1C,
	_Shot2C,
	_Shot3C,
}

var p2BeamCols = []colorID {
	_WhiteC,
	_ShotP2C,
	_ShotP2C,
	_Shot3C,
}

var p2HitCols = []colorID {
	_ShotP2C,
	_ShotP2C,
	_Shot2C,
	_Shot3C,
}

var beamCols = map[cellState][]colorID{
	P1Head : p1BeamCols,
	P2Head : p2BeamCols,
}

var hitCols = map[cellState][]colorID{
	P1Head : p1HitCols,
	P2Head : p2HitCols,
}

func stylesInit() {
	ColEmpty   = tcell.StyleDefault.Foreground(tcell.ColorBlack     ).Background(tcell.ColorBlack)
	ColP1Head  = tcell.StyleDefault.Foreground(tcell.ColorBlue      ).Background(tcell.ColorBlack)
	ColP2Head  = tcell.StyleDefault.Foreground(tcell.ColorOrange    ).Background(tcell.ColorBlack)
	ColDefault = tcell.StyleDefault.Foreground(tcell.ColorWhiteSmoke).Background(tcell.ColorBlack)

	ColWhite   = tcell.StyleDefault.Foreground(tcell.ColorWhite     ).Background(tcell.ColorBlack)

	ColShotP1C = tcell.StyleDefault.Foreground(tcell.ColorBlue      ).Background(tcell.ColorBlack)
	ColShotP2C = tcell.StyleDefault.Foreground(tcell.ColorOrange    ).Background(tcell.ColorBlack)

	ColShot2C  = tcell.StyleDefault.Foreground(tcell.NewRGBColor( 60,  13,  16)    ).Background(tcell.ColorBlack)
	ColShot3C  = tcell.StyleDefault.Foreground(tcell.NewRGBColor( 49,  11,  12)    ).Background(tcell.ColorBlack)

	cols = map[colorID]tcell.Style{
		EmptyC   : ColEmpty,
		P1HeadC  : ColP1Head,
		P2HeadC  : ColP2Head,
		WallC    : ColDefault,
		
		_WhiteC  : ColWhite,

		_ShotP1C : ColShotP1C, 
		_ShotP2C : ColShotP2C, 

		_Shot2C  : ColShot2C, 
		_Shot3C  : ColShot3C, 
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
				board[pos.y][pos.x].col = col
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
				board[pos.y][pos.x].col = col
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

	animate := func(col colorID, pos Vec2, d int, delay time.Duration, chance int) {
		for range dist {
			time.Sleep((SIM_TIME/2) * delay)
			if pos.y >= MapH || pos.y < 0 {
				break
			}
			if rand.Intn(20) < chance {
				board[pos.y][pos.x].col = col
			}
			if d > 0 {
				d--
				pos = pos.Add(dir)
			}
		}

	}

	// Frame 1
	animate(_WhiteC, start, dist, 0, 20)
	time.Sleep(SIM_TIME)
	animate(EmptyC, start, dist, 1, 20)
	animate(EmptyC, start, dist, 1, 20)

}
