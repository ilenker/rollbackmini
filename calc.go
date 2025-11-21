package main

import (
	"time"
	"math"
	"unsafe"
)

type Vec2 struct {
	x int
	y int
}

type VecRGB struct {
	r int32
	g int32
	b int32
}

func newVecRGB[T int | int8 | int16 | int32 | int64] (r T, g T, b T) VecRGB {
	return VecRGB{int32(r), int32(g), int32(b)}
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

func (v1 Vec2) scale(n int) Vec2 {
	newX := wrapInt(v1.x * n, MapW)
	newY := wrapInt(v1.y * n, MapH)
	return Vec2{
		newX,
		newY,
	}
}

func (v1 Vec2) add(v2 Vec2) Vec2 {
	newX := wrapInt(v1.x + v2.x, MapW)
	newY := wrapInt(v1.y + v2.y, MapH)
	return Vec2{
		newX,
		newY,
	}
}

func (v1 Vec2) addNoWrap(v2 Vec2) Vec2 {
	return Vec2{
		x: v1.x + v2.x,
		y: v1.y + v2.y,
	}
}

func (v1 Vec2) translate(angleRad float64, distance float64) Vec2 {
    
    dx := distance * math.Cos(angleRad)
    dy := distance * math.Sin(angleRad)
    
    newX := v1.x + int(math.Round(dx))
    newY := v1.y + int(math.Round(dy))
    
    return Vec2{x: newX, y: newY}
}

func (v1 VecRGB) add(v2 VecRGB) VecRGB {

	v3 := v1
	v3.r = v1.r + v2.r 
	v3.b = v1.b + v2.b
	v3.g = v1.g + v2.g

	if v3.r > 255 { v3.r = 255 }
	if v3.g > 255 { v3.g = 255 }
	if v3.b > 255 { v3.b = 255 }

	if v3.r < 0   { v3.r = 0   }
	if v3.g < 0   { v3.g = 0   }
	if v3.b < 0   { v3.b = 0   }

	return v3
}


func fB2i(b bool) int {
    return int(*(*byte)(unsafe.Pointer(&b)))
}
