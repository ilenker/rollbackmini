package main

import (
	"time"
	"math"
)

type Vec2 struct {
	x int
	y int
}


func AbsInt(n int) int {
	if n < 0 {
		return ^n + 1
	}
	return n
}


func AbsInt16(n int16) int16 {
	if n < 0 {
		return ^n + 1
	}
	return n
}

// 1  = Extremely Slow
// 2  = Ultra Slow
// 3  = Very Slow
// 4  = Slow
// 5  = Medium
// ..
// 13 = M. Fast
// ..
// 16 = Fast
// .. 
// 32 = Very Fast
// 64 = Ultra Fast
// 128 = Extremely Fast
// 512 = Bullet
// 1024 = Laser

// TODO: figure out a nice conversion between length and speed
// The relationship will probably involve some logs or exponentials
// Apparent speed changes decrease exponentially as SCPT increases.
// Pretty much need to double SCPT to cause the same "amount of change"
// in the apparent speed (see above notes: 1-2 feels about the same as )


func lerp(x, y int, f float64) float64 {
	return float64(x) * (1.0-f) + float64(y) * f
}


func iLerp(x, y int, num float64) float64 {
    if x == y { return 0 }
    return (num - float64(x)) / float64(y - x)
}


func wrapInt(n, n_max int) int {
	wrap := ((n % n_max) + n_max) % n_max
	return wrap
}

func makeAverageDurationBuffer(size int) func(time.Duration) (int64, []time.Duration) {
	buffer := make([]time.Duration, size)
	i := 0

	return func(d time.Duration) (int64, []time.Duration) {
		buffer[i] = d
		i = wrapInt(i + 1, size)
		total := time.Duration(0)

		for n := range size {
			total += buffer[n]
		}

		return (total.Microseconds() / int64(size)), buffer
	}
}

func makeAverageIntBuffer(size int) func(n int) (float64, []int)  {
	buffer := make([]int, size)
	i := 0
	return func(n int) (float64, []int) {
		buffer[i] = n
		i = wrapInt(i + 1, size)

		total := 0

		for n := range size {

			total += buffer[n]
		}

		return (float64(total) / float64(size)), buffer
	}
}


func B2i(b bool) int {
	if b { return 1 }
	return 0
}


func (v1 Vec2) Add(v2 Vec2) Vec2 {
	newX := wrapInt(v1.x + v2.x, MapW)
	newY := wrapInt(v1.y + v2.y, MapH)
	return Vec2{
		newX,
		newY,
	}
}

func (v1 Vec2) AddNoWrap(v2 Vec2) Vec2 {
	return Vec2{
		x: v1.x + v2.x,
		y: v1.y + v2.y,
	}
}

func (v1 Vec2) Translate(angleRad float64, distance float64) Vec2 {
    
    dx := distance * math.Cos(angleRad)
    dy := distance * math.Sin(angleRad)
    
    newX := v1.x + int(math.Round(dx))
    newY := v1.y + int(math.Round(dy))
    
    return Vec2{x: newX, y: newY}
}
