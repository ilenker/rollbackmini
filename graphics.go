package main

import (
	"os"
	//"fmt"
	"time"
	"math"
	"math/rand"
	"github.com/gdamore/tcell/v2"
	"github.com/kelindar/simd"
)

type colorID = uint8

var COLORTERM_OG string
var stDef  = tcell.StyleDefault
var stText = tcell.StyleDefault
var textCol = tcell.ColorSlateGray

var renderBuffer	Slice3f64
var vfxLayer 		Slice3f64
var lightVal 		Slice3f64
//var	lightLayer	= [MapH+1][MapW+1]Vec3[float32]{}

var dimmingFactor []float64

var lightPoints []Vec2
var lpID int

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

			i := flatIdx(int(x), int(y))
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

			i := flatIdx(int(x), int(y))
			copyRGB(&lightVal, i, 0.15, 0.15, 0.15)
			f += 0.010
		}
	}

	lightVal.rs[ flatIdx(getLocalPlayerPtr().pos) ] = 2

	calculateLighting()
	// For each terminal row (board y-coordinates map 2:1 onto terminal y-coordinates)
	for y := range (MapH / 2) {
		lyUpper := y * 2           // Calculate corresponding Logical Row, given Terminal Row
		lyLower := y * 2 + 1

		// For each terminal cell (board x-coordinates map 1:1 onto terminal y-coordinates)
		for x := range MapW {
			iU := lyUpper * MapW + x
			iL := lyLower * MapW + x

			//newR := 255.0 * math.Pow((renderBuffer.rs[iU] / 255), γ)

			newR :=
			math.Pow(
				math.Max(renderBuffer.rs[iU] - β, 0) / (ω - β),
				γ) * 255

			newG :=
			math.Pow(
				math.Max(renderBuffer.gs[iU] - β, 0) / (ω - β),
				γ) * 255

			newB :=
			math.Pow(
				math.Max(renderBuffer.bs[iU] - β, 0) / (ω - β),
				γ) * 255


			//newG := int32(float64(255) * math.Pow((renderBuffer.gs[iU] / 255), γ) )
			//newB := int32(float64(255) * math.Pow((renderBuffer.bs[iU] / 255), γ) )
			upper := tcell.NewRGBColor(
				int32(newR),
				int32(newG),
				int32(newB))

			newR =
			math.Pow(
				math.Max(renderBuffer.rs[iL] - β, 0) / (ω - β),
				γ) * 255

			newG =
			math.Pow(
				math.Max(renderBuffer.gs[iL] - β, 0) / (ω - β),
				γ) * 255

			newB =
			math.Pow(
				math.Max(renderBuffer.bs[iL] - β, 0) / (ω - β),
				γ) * 255

			//newR = int32(float64(255) * math.Pow((renderBuffer.rs[iL] / 255), γ) )
			//newG = int32(float64(255) * math.Pow((renderBuffer.gs[iL] / 255), γ) )
			//newB = int32(float64(255) * math.Pow((renderBuffer.bs[iL] / 255), γ) )
			//lower := tcell.NewRGBColor(clamp(newR, 0, 255), clamp(newG, 0, 255), clamp(newB, 0, 255))
			lower := tcell.NewRGBColor(
				int32(newR),
				int32(newG),
				int32(newB))

			//upper = scaleColor(upper, Vec3[float32]{
			//	lightRs[iU],
			//	lightRs[iU] * 0.8,
			//	lightRs[iU] * 0.8,
			//})

			//lower = scaleColor(lower, Vec3[float32]{
			//	lightRs[iL],
			//	lightRs[iL] * 0.8,
			//	lightRs[iL] * 0.8,
			//})

			lightDecay := 0.95
			lightVal.rs[iU] = clampMin(lightVal.rs[iU] * lightDecay, 0.1)
			lightVal.gs[iU] = clampMin(lightVal.gs[iU] * lightDecay, 0.1)
			lightVal.bs[iU] = clampMin(lightVal.bs[iU] * lightDecay, 0.1)

			lightVal.rs[iL] = clampMin(lightVal.rs[iL] * lightDecay, 0.1)
			lightVal.gs[iL] = clampMin(lightVal.gs[iL] * lightDecay, 0.1)
			lightVal.bs[iL] = clampMin(lightVal.bs[iL] * lightDecay, 0.1)

			//lightLayer[lyUpper][x].x = clamp(lightLayer[lyUpper][x].x * 0.90, 0.02, 10)
			//lightLayer[lyUpper][x].y = clamp(lightLayer[lyUpper][x].y * 0.90, 0.02, 10) 
			//lightLayer[lyUpper][x].z = clamp(lightLayer[lyUpper][x].z * 0.90, 0.10, 10) 

			//lightLayer[lyLower][x].x = clamp(lightLayer[lyLower][x].x * 0.90, 0.02, 10) 
			//lightLayer[lyLower][x].y = clamp(lightLayer[lyLower][x].y * 0.90, 0.02, 10) 
			//lightLayer[lyLower][x].z = clamp(lightLayer[lyLower][x].z * 0.90, 0.10, 10) 


			r := '▀'
			st := tcell.StyleDefault.Foreground(upper).Background(lower)

			s.SetContent(x + xOffset, y + yOffset, r, nil, st)

		}
	}

	s.Show()
}

func calculateLighting() {
	simd.MulFloat64s(renderBuffer.rs, vfxLayer.rs, lightVal.rs)
	simd.MulFloat64s(renderBuffer.gs, vfxLayer.gs, lightVal.gs)
	simd.MulFloat64s(renderBuffer.bs, vfxLayer.bs, lightVal.bs)
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
		EmptyC   : tcell.NewRGBColor(int32(β), int32(β), int32(β)),
		P1HeadC  : tcell.ColorBlue,
		P2HeadC  : tcell.ColorDarkOrange,
		WallC    : tcell.ColorWhiteSmoke,

		WhiteC  : tcell.ColorWhite,

		ShotP1C : tcell.ColorCornflowerBlue,
		ShotP2C : tcell.ColorOrange,

		Shot2C  : tcell.NewRGBColor(60, 63, 106),
		Shot3C  : tcell.NewRGBColor(66, 77, 65),
	}
}

func beamEffect(start Vec2, dist int, dir Vec2, colorSeq []colorID) {

	start = start.add(dir)
	animLen := dist

	animate := func(
		col		colorID,
		pos		Vec2,
		d		int,
		delay	time.Duration,
		chance	int,
		fade	bool) {

		c := cols[col]
		for range animLen {
			time.Sleep((SIM_TIME/2) * delay)
			if pos.y >= MapH || pos.y < 0 {
				break
			}
			if fade {
				c = addRBGtoColor(VecRGB{-5, -5, -5}, c)
			}
			if rand.Intn(20) < chance {
				r, g, b := c.RGB()
				vfxLayer.rs[pos.y * MapW + pos.x] = float64(r)
				vfxLayer.gs[pos.y * MapW + pos.x] = float64(g)
				vfxLayer.bs[pos.y * MapW + pos.x] = float64(b)
			}
			if d > 0 {
				d--
				pos = pos.addNoWrap(dir)
			}
		}

	}

	// Frame 1
	animate(colorSeq[0], start, dist, 0, 19, false)
	time.Sleep(SIM_TIME)
	time.Sleep(SIM_TIME)

	animate(colorSeq[1], start, dist, 0, 20, true)
	time.Sleep(SIM_TIME)

	animate(colorSeq[2], start, dist, 0, 15, true)

	animate(colorSeq[3], start, dist, 2, 20, true)

	animate(EmptyC, start, dist, 2, 10, false)
	animate(EmptyC, start, dist, 3, 15, false)
	animate(EmptyC, start, dist, 4, 20, false)
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
				r, g, b := cols[col].RGB()
				vfxLayer.rs[pos.y * MapW + pos.x] = float64(r)
				vfxLayer.gs[pos.y * MapW + pos.x] = float64(g)
				vfxLayer.bs[pos.y * MapW + pos.x] = float64(b)
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

	animate := func(col colorID,
		pos		Vec2,
		delay	time.Duration,
		chance	int,
		after	bool,
		fade	bool) {

		for range animLen {
			c := cols[col]
			time.Sleep((SIM_TIME/2) * delay)

			if pos.x >= MapW || pos.x < 0 || 
			   pos.y >= MapH || pos.y < 0 {
				break
			}

			if fade {
				c = addRBGtoColor(VecRGB{-7, -7, -7}, c)
			}

			if rand.Intn(20) < chance {
				r, g, b := c.RGB()
				vfxLayer.rs[pos.y * MapW + pos.x] = float64(r)
				vfxLayer.gs[pos.y * MapW + pos.x] = float64(g)
				vfxLayer.bs[pos.y * MapW + pos.x] = float64(b)
			}

			turns += curve
			pos = pos.translate(turns * math.Pi, 1)
		}

		if after {
			go hitEffect2nd(pos, 2 * rand.Float64(), colorSeq)
			go hitEffect2nd(pos, 2 * rand.Float64(), colorSeq)
			go hitEffect2nd(pos, 2 * rand.Float64(), colorSeq)
			go hitEffect2nd(pos, 2 * rand.Float64(), colorSeq)
		}
		turns = baseturns

	}

	animate(colorSeq[0], start, 0, 20, true , true)
	time.Sleep(SIM_TIME)
	time.Sleep(SIM_TIME)

	animate(colorSeq[1], start, 0, 20, true , true)
	time.Sleep(SIM_TIME)
	time.Sleep(SIM_TIME)

	animate(colorSeq[2], start, 0, 20, false, true)
	time.Sleep(SIM_TIME)

	animate(colorSeq[2], start, 0, 15, true , true)
	go flash(start, size/4, lum, linger, Vec3[float64]{1, 1, 1}, false)

	animate(colorSeq[3], start, 1, 20, false, true)
	animate(EmptyC,      start, 1, 10, false, false)
	animate(EmptyC,      start, 2, 15, false, false)
	animate(EmptyC,      start, 2, 20, true , false)


}


func hitEffect2nd(start Vec2, baseturns float64, colorSeq []colorID) {

	animLen := 0 + rand.Intn(5)

	turns := baseturns
	curve := -0.1 + rand.Float64() * 0.2

	animate := func(col colorID,
		pos Vec2,
		delay time.Duration,
		chance int,
		fade bool) {

		c := cols[col]
		for range animLen {
			time.Sleep((SIM_TIME/2) * delay)
			if pos.x >= MapW || pos.x < 0 || 
			   pos.y >= MapH || pos.y < 0 {
				break
			}

			if fade {
				c = addRBGtoColor(VecRGB{-7, -7, -7}, c)
			}

			if rand.Intn(20) < chance {
				r, g, b := c.RGB()
				vfxLayer.rs[pos.y * MapW + pos.x] = float64(r)
				vfxLayer.gs[pos.y * MapW + pos.x] = float64(g)
				vfxLayer.bs[pos.y * MapW + pos.x] = float64(b)
			}

			turns += curve
			pos = pos.translate(turns * math.Pi, 1)
		}
		turns = baseturns
	}

	// Frame 1
	animate(colorSeq[0], start,  0, 20, true)
	time.Sleep(SIM_TIME)
	time.Sleep(SIM_TIME)

	animate(colorSeq[1], start,  0, 15, true)
	animate(colorSeq[2], start,  1, 20, true)
	animate(EmptyC,      start,  3, 10, false)
	animate(EmptyC,      start,  3, 15, false)
	animate(EmptyC,      start,  3, 20, false)
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
				r, g, b := col.RGB()
				vfxLayer.rs[pos.y * MapW + pos.x] = float64(r)
				vfxLayer.gs[pos.y * MapW + pos.x] = float64(g)
				vfxLayer.bs[pos.y * MapW + pos.x] = float64(b)
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
		copyRGB(&vfxLayer, flatIdx(origin), 0.0, 0.0, 0.0)
	}

	for i := range length {
		//vfxLayer[origin.y][origin.x + i + 1] = cols[EmptyC]
		copyRGB(&vfxLayer, flatIdx(origin.x + i + 1, origin.y), 0.0, 0.0, 0.0)

		//vfxLayer[origin.y][origin.x + i    ] = cols[colorID]
		r, g, b := cols[colorID].RGB()
		copyRGB(&vfxLayer,
			flatIdx(origin.x + i, origin.y),
			float64(r),
			float64(g),
			float64(b))
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
func _light(lightR float64, p, p2 Vec2, r int, lum float64, c Vec3[float64]) float64 {

	x := p2.x
	y := p2.y

	minL := float64(1.0)

	dMax := r + 1

	//errorBox(fmt.Sprintf("dMaxL: %d", dMax), 0, 1)


	d := dist(p, Vec2{x, y})

	if d > float64(dMax) {
		return 1
	}

	dNormal := iLerp64(0, dMax, d * d * FOFactor)

	if dNormal > 1 { return 1.0 }

	rLum := lerp64(lum, minL, dNormal) * c.x


	//kv := Vec3[float32]{v.x * rL, v.y * gL, v.z * bL}
	kMax := 15.0
	kLightR := clamp(lightR * rLum, 0, kMax)

	//if kv.y < 1 { kv.y = 1}
	//if kv.z < 1 { kv.z = 1}

	return kLightR
}

func light(p Vec2, r int, lum float64, c Vec3[float64]) {

	lum /= 10

	minL := float64(1.0)

	dMax := r + 1

	//errorBox(fmt.Sprintf("dMaxL: %d", dMax), 0, 1)

	for x := p.x - r;
		x <= p.x + r;
		x++ {

		if x > MapW || x < 0 { continue }

		for y := p.y - r;
			y <= p.y + r;
			y++ {
			i := flatIdx(x, y)

			if y > MapH || y < 0 { continue }

			d := dist(p, Vec2{x, y})

			dNormal := iLerp64(0, dMax, d * d * FOFactor)

			if dNormal > 1 { continue }
			
			rLum := lerp64(lum, minL, dNormal) * c.x
			gLum := lerp64(lum, minL, dNormal) * c.y
			bLum := lerp64(lum, minL, dNormal) * c.z

			lightR := lightVal.rs[i]
			lightG := lightVal.gs[i]
			lightB := lightVal.bs[i]


			//kv := Vec3[float32]{v.x * rL, v.y * gL, v.z * bL}
			kMax := 5.0
			rLum = clamp( (rLum + lightR) / 2, 1, kMax)
			gLum = clamp( (gLum + lightG) / 2, 1, kMax)
			bLum = clamp( (bLum + lightB) / 2, 1, kMax)

			//if kv.y < 1 { kv.y = 1}
			//if kv.z < 1 { kv.z = 1}

			if k := vfxLayer.rs[i] * rLum;
			k > 253 { rLum *= 255/k }

			if k := vfxLayer.gs[i] * gLum;
			k > 253 { gLum *= 255/k }

			if k := vfxLayer.bs[i] * bLum;
			k > 253 { bLum *= 255/k }


			lightVal.rs[i] = rLum
			lightVal.gs[i] = gLum
			lightVal.bs[i] = bLum
		}
	}
}


func flash(p Vec2, r int, lum float64, decayRate float64, c Vec3[float64], castShadow bool) {

	mu.Lock()

	if castShadow {
		shadows = true
		defer func(){
			shadows = false
		} ()
	}

	for {
		condLighting.Wait()
		light(p, r, lum, c)
		if lum <= 1 {
			mu.Unlock()
			return
		}
		lum -= lum * decayRate
	}
}
