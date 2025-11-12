package main

import (
	"time"
	"math"
	"math/rand"
)

var p1BeamCols = []colorID {
	_WhiteC,
	_ShotP1C,
	_ShotP1C,
	_ShotP1C,
}

var p2BeamCols = []colorID {
	_WhiteC,
	_ShotP2C,
	_ShotP2C,
	_ShotP2C,
}

func beamEffect(start Vec2, dist int, dir Vec2, colorSeq []colorID) {

	start = start.Add(dir)
	animLen := 17

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

	animLen := 3 + rand.Intn(8)
	turns := baseturns
	curve := -0.1 + rand.Float64() * 0.2

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
			//go hitEffect2nd(pos, 2 * rand.Float64())
			//go hitEffect2nd(pos, 2 * rand.Float64())
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

func hitEffect2nd(start Vec2, baseturns float64) {

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
	<-FrameSyncCh
	animate(_ShotP1C, start, 0, 20)
	time.Sleep(SIM_TIME)
	time.Sleep(SIM_TIME)

	animate(_Shot2C, start, 0, 15)
	animate(_Shot3C, start, 1, 20)
	animate(EmptyC, start,  3, 10)
	animate(EmptyC, start,  3, 15)
	animate(EmptyC, start,  3, 20)

}

