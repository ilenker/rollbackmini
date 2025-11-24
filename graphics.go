package main

import (
	"os"
	"fmt"
	"time"
	"math"
	"math/rand"
	"github.com/gdamore/tcell/v2"
	//"github.com/kelindar/simd"
)

type colorID = uint8

var COLORTERM_OG string
var stDef  = tcell.StyleDefault
var stText = tcell.StyleDefault
var textCol = tcell.ColorSlateGray

var	vfxLayer   = [MapH+1][MapW+1]tcell.Color{}
var	lightLayer = [MapH+1][MapW+1]Vec3[float32]{}
var lightPoints []Vec2
var lpID int

const FOFactor = 0.5

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
	ShotP1C,
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

	// For each terminal row (board y-coordinates map 2:1 onto terminal y-coordinates)
	for y := range (MapH / 2) {
		lyUpper := y * 2           // Calculate corresponding Logical Row, given Terminal Row
		lyLower := y * 2 + 1

		// For each terminal cell (board x-coordinates map 1:1 onto terminal y-coordinates)
		for x := range MapW {
			upper := cols[board[lyUpper][x].state]
			lower := cols[board[lyLower][x].state]

			upperVfx := vfxLayer[lyUpper][x]
			lowerVfx := vfxLayer[lyLower][x]

			if upperVfx != cols[EmptyC] {
				upper = upperVfx
			}

			if lowerVfx != cols[EmptyC] {
				lower = lowerVfx
			}

			//upper = scaleColor(upper, Vec3[float32]{_foo, _foo, _foo})
			//lower = scaleColor(lower, Vec3[float32]{_foo, _foo, _foo})
			upper = scaleColor(upper, lightLayer[lyUpper][x])
			lower = scaleColor(lower, lightLayer[lyLower][x])

			lightLayer[lyUpper][x].x = clamp(lightLayer[lyUpper][x].x * 0.90, 0.02, 10)
			lightLayer[lyUpper][x].y = clamp(lightLayer[lyUpper][x].y * 0.90, 0.02, 10) 
			lightLayer[lyUpper][x].z = clamp(lightLayer[lyUpper][x].z * 0.90, 0.10, 10) 

			lightLayer[lyLower][x].x = clamp(lightLayer[lyLower][x].x * 0.90, 0.02, 10) 
			lightLayer[lyLower][x].y = clamp(lightLayer[lyLower][x].y * 0.90, 0.02, 10) 
			lightLayer[lyLower][x].z = clamp(lightLayer[lyLower][x].z * 0.90, 0.10, 10) 


			// Raycast from all light points to orange player
			if shadows {
				for _, lPoint := range lightPoints {

					if lPoint.x == 0 && lPoint.y == 0 {
						continue
					}

					p1 := lPoint

					dir := angleTo(p1, player2.pos)

					p2 := player2.pos.translate(dir, 50)
					p1 = player2.pos.translate(dir, 1)

					f := 0.0
					for {
						if f > 1 {
							break
						}
						x := math.Round(lerp(p1.x, p2.x, f))
						y := math.Round(lerp(p1.y, p2.y, f))

						if x > MapW || x < 0 {
							break
						}

						if y > MapH || y < 0 {
							break
						}
						lightLayer[int(y)][int(x)] = Vec3[float32]{0.4, 0.4, 0.4}
						f += 0.010
					}
				}
			}



			if board[lyUpper][x].state == P1Head ||
			   board[lyLower][x].state == P1Head {
				rS, gS, bS := upper.RGB()
				rS_, gS_, bS_ := lower.RGB()
				errorBox(fmt.Sprintf("up%3d:%3d:%3d", rS, gS, bS), 0, 6)
				errorBox(fmt.Sprintf("dn%3d:%3d:%3d", rS_, gS_, bS_), 0, 7)
			}

			r := '▀'
			st := tcell.StyleDefault.Foreground(upper).Background(lower)

			s.SetContent(x + xOffset, y + yOffset, r, nil, st)

		}
	}

	s.Show()
}


func newRGBOscillator(rgbInit VecRGB) func() tcell.Color {
	rPol := 1
	gPol := 1
	bPol := 1

	d := 1

	v := rgbInit

	return func() tcell.Color {

		v.r += int32(rPol * d)
		v.g += int32(gPol * d)
		v.b += int32(bPol * d)

		rPol = rPol + (fB2i(v.r == 255 || v.r == 0) * (rPol * -1) * 2)
		gPol = gPol + (fB2i(v.g == 255 || v.g == 0) * (gPol * -1) * 2)
		bPol = bPol + (fB2i(v.b == 255 || v.b == 0) * (bPol * -1) * 2)

		r := v.r
		g := v.g
		b := v.b

		return tcell.NewRGBColor(r, g, b)
	}
}



func stylesInit() {
	stText = stText.Foreground(tcell.ColorSlateGray)
	cols = map[colorID]tcell.Color{
		EmptyC   : tcell.NewRGBColor(1, 1, 1),
		P1HeadC  : tcell.ColorBlue,
		P2HeadC  : tcell.ColorDarkOrange,
		WallC    : tcell.ColorWhiteSmoke,

		WhiteC  : tcell.ColorWhite,

		ShotP1C : tcell.ColorCornflowerBlue,
		ShotP2C : tcell.ColorOrange,

		Shot2C  : tcell.NewRGBColor(60, 13, 16),
		Shot3C  : tcell.NewRGBColor(66, 18, 26),
	}
}

func beamEffect(start Vec2, dist int, dir Vec2, colorSeq []colorID) {

	start = start.add(dir)
	animLen := dist

	animate := func(col colorID, pos Vec2, d int, delay time.Duration, chance int) {
		c := cols[col]
		for range animLen {
			time.Sleep((SIM_TIME/2) * delay)
			if pos.y >= MapH || pos.y < 0 {
				break
			}
			if col != EmptyC && col != WhiteC {
				c = addRBGtoColor(VecRGB{-5, -5, -5}, c)
			}
			if rand.Intn(20) < chance {
				vfxLayer[pos.y][pos.x] = c
			}
			if d > 0 {
				d--
				pos = pos.addNoWrap(dir)
			}
		}

	}

	// Frame 1
	animate(colorSeq[0], start, dist, 0, 19)
	time.Sleep(SIM_TIME)
	time.Sleep(SIM_TIME)

	animate(colorSeq[1], start, dist, 0, 20)
	time.Sleep(SIM_TIME)

	animate(colorSeq[2], start, dist, 0, 15)

	animate(colorSeq[3], start, dist, 2, 20)

	animate(EmptyC, start, dist, 2, 10)
	animate(EmptyC, start, dist, 3, 15)
	animate(EmptyC, start, dist, 4, 19)
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
				vfxLayer[pos.y][pos.x] = cols[col]
			}

			turns += curve
			pos = pos.translate(turns * math.Pi, 1)
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
			c := cols[col]
			time.Sleep((SIM_TIME/2) * delay)

			if pos.x >= MapW || pos.x < 0 || 
			   pos.y >= MapH || pos.y < 0 {
				break
			}

			if col != EmptyC && col != WhiteC {
				c = addRBGtoColor(VecRGB{-7, -7, -7}, c)
			}

			if rand.Intn(20) < chance {
				vfxLayer[pos.y][pos.x] = c
			}

			turns += curve
			pos = pos.translate(turns * math.Pi, 1)
		}

		if after {
			go hitEffect2nd(pos, 2 * rand.Float64(), colorSeq)
			go hitEffect2nd(pos, 2 * rand.Float64(), colorSeq)
			go hitEffect2nd(pos, 2 * rand.Float64(), colorSeq)
			go hitEffect2nd(pos, 2 * rand.Float64(), colorSeq)
			go flash(pos, 100 + rand.Intn(100), 10, 8, Vec3[float32]{1.0, 0.8, 0.8}, false)
		}
		turns = baseturns

	}

	animate(colorSeq[0], start, 0, 20, true)
	time.Sleep(SIM_TIME)
	time.Sleep(SIM_TIME)

	animate(colorSeq[1], start,  0, 20, true)
	time.Sleep(SIM_TIME)
	time.Sleep(SIM_TIME)

	animate(colorSeq[2], start, 0, 20, false)
	time.Sleep(SIM_TIME)

	animate(colorSeq[2], start, 0, 15, true)

	animate(colorSeq[3], start, 1, 20, false)
	animate(EmptyC, start,  1, 10, false)
	animate(EmptyC, start,  2, 15, false)
	animate(EmptyC, start,  2, 20, true)


}


func hitEffect2nd(start Vec2, baseturns float64, colorSeq []colorID) {

	animLen := 0 + rand.Intn(5)

	turns := baseturns
	curve := -0.1 + rand.Float64() * 0.2

	animate := func(col colorID, pos Vec2, delay time.Duration, chance int) {
		c := cols[col]
		for range animLen {
			time.Sleep((SIM_TIME/2) * delay)
			if pos.x >= MapW || pos.x < 0 || 
			   pos.y >= MapH || pos.y < 0 {
				break
			}

			if col != EmptyC && col != WhiteC {
				c = addRBGtoColor(VecRGB{-7, -7, -7}, c)
			}

			if rand.Intn(20) < chance {
				vfxLayer[pos.y][pos.x] = cols[col]
			}

			turns += curve
			pos = pos.translate(turns * math.Pi, 1)
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

	start = start.add(dir)

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
				pos = pos.add(dir)
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

func addRBGtoColor(v VecRGB, c tcell.Color) tcell.Color {
	r, g, b := c.RGB()
	v = v.add(VecRGB{r, g, b})
	return tcell.NewRGBColor(v.r, v.g, v.b)
}

func scaleColor(c tcell.Color, v Vec3[float32]) tcell.Color {
	r, g, b := c.RGB()

	//if r == 0 { r = 1 }
	//if g == 0 { g = 1 }
	//if b == 0 { b = 1 }

	rS := clamp(float32(r) * v.x, 0, 255)
	gS := clamp(float32(g) * v.y, 0, 255)
	bS := clamp(float32(b) * v.z, 0, 255)

	return tcell.NewRGBColor(int32(rS), int32(gS), int32(bS))
}


func setCOLORTERM() {
	COLORTERM_OG = os.Getenv("COLORTERM")
	os.Setenv("COLORTERM", "truecolor")	
}

func restoreCOLORTERM() {
	os.Setenv("COLORTERM", COLORTERM_OG)	
}

// position, radius, luminance, color
func light(p Vec2, r int, l float32, c Vec3[float32]) {

	l /= 5

	//if l < 1 {
	//	l = 1
	//}

	minL := float32(1.0)

	dMax := int(dist(p, Vec2{p.x+r+1, p.y}))

	//errorBox(fmt.Sprintf("dMaxL: %d", dMax), 0, 1)

	for x := p.x - r;
		x <= p.x + r;
		x++ {

		if x > MapW || x < 0 { continue }

		for y := p.y - r;
			y <= p.y + r;
			y++ {

			if y > MapH || y < 0 { continue }


			d := float32(dist(p, Vec2{x, y}))

			dNormal := iLerp32(0, dMax, d*d*FOFactor)
			if dNormal > 1 { continue }
			
			rL := lerp32(l, minL, dNormal) * c.x
			gL := lerp32(l, minL, dNormal) * c.y
			bL := lerp32(l, minL, dNormal) * c.z

			v := lightLayer[y][x]

			kv := Vec3[float32]{v.x * rL, v.y * gL, v.z * bL}

			if kv.x < 1 { kv.x = 1}
			if kv.y < 1 { kv.y = 1}
			if kv.z < 1 { kv.z = 1}
			lightLayer[y][x] = kv
		}
	}
}


func flash(p Vec2, r int, lum float32, linger int, c Vec3[float32], castShadow bool) {

	lingerFactor := float32(linger)
	l := lum
	mu.Lock()

	if castShadow {
		id := lpID
		lpID++
		if lpID > 9 {
			lpID = 0
		}
		lightPoints[id] = p
		defer func(){
			lightPoints[id] = Vec2{0, 0}
		} ()
	}


	for {
		condLighting.Wait()
		light(p, r, l, c)
		l -= l/lingerFactor
		if l <= 1 {
			mu.Unlock()
			return
		} 
	}


}
