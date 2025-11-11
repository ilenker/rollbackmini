package main

import (
	"time"
	"math"
	"math/rand"
)

func beamEffect(start Vec2, end Vec2, dir Vec2) {

	start = start.Add(dir)
	animLen := 17

	animate := func(col colorID, pos, end Vec2, delay time.Duration, chance int) {
		for range animLen {
			time.Sleep((SIM_TIME/2) * delay)
			if pos.y >= MapH || pos.y < 0 {
				break
			}
			if chance == 20 {
				board[pos.y][pos.x].col = col
			} else {
				if rand.Intn(20) < chance {
					board[pos.y][pos.x].col = col
				}
			}

			if pos.y != end.y {
				pos = pos.AddNoWrap(dir)
			}
		}
	}

	// Frame 1
	<-FrameSyncCh
	animate(_Shot1C, start, end, 0, 20)
	time.Sleep(SIM_TIME)
	time.Sleep(SIM_TIME)
	animate(_Shot1C, start, end, 0, 20)
	time.Sleep(SIM_TIME)
	animate(_Shot3C, start, end, 0, 15)

	animate(_Shot4C, start, end, 2, 20)
	animate(EmptyC, start, end, 2, 10)
	animate(EmptyC, start, end, 2, 15)
	animate(EmptyC, start, end, 2, 20)

}

func hitEffect(start Vec2, baseturns float64) {

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
			go hitEffect2nd(pos, 2 * rand.Float64())
			go hitEffect2nd(pos, 2 * rand.Float64())
		}
		turns = baseturns

	}

	// Frame 1
	<-FrameSyncCh
	animate(_Shot1C, start, 0, 20, false)
	time.Sleep(SIM_TIME)
	time.Sleep(SIM_TIME)
	animate(EmptyC, start,  0, 20, false)
	time.Sleep(SIM_TIME)
	time.Sleep(SIM_TIME)
	animate(_Shot1C, start, 0, 20, false)
	time.Sleep(SIM_TIME)
	animate(_Shot3C, start, 0, 15, false)

	animate(_Shot4C, start, 1, 20, false)
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
	animate(_Shot1C, start, 0, 20)
	time.Sleep(SIM_TIME)
	time.Sleep(SIM_TIME)

	animate(_Shot3C, start, 0, 15)
	animate(_Shot4C, start, 1, 20)
	animate(EmptyC, start,  3, 10)
	animate(EmptyC, start,  3, 15)
	animate(EmptyC, start,  3, 20)

}

