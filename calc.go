package main

import (
	"fmt"
	"math"
	"sync/atomic"
	"time"
	"unsafe"
)

const arraySize = 64*64*64*64
var squaredDistances [arraySize]float64
const wordSize = 6

type angleMap map[[4]int]float64
type gammaMap map[float64]float64

var angleCache atomic.Value
var gammaCache atomic.Value
var squaredCache atomic.Value


/*······················································································kVec2    */
type Vec2 struct {
	x int
	y int
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
    
    newX := float64(v1.x) + dx
    newY := float64(v1.y) + dy

	newVec := Vec2{int(math.Round(newX)), int(math.Round(newY))}

    return newVec
}

func dist(v1, v2 Vec2) float64 {
	return math.Sqrt(
		math.Pow(float64(v1.x - v2.x), 2) +
		math.Pow(float64(v1.y - v2.y), 2),
		)
}


/*······················································································kVec3    */
type Vec3[T int | int8 | int16 | int32 | float32 | float64] struct {
	x T
	y T
	z T
}

type Slice3f64 struct {
	rs []float64
	gs []float64
	bs []float64
}

func copyRGB(destSlice *Slice3f64, index int, r, g, b float64) {
	destSlice.rs[index] = r
	destSlice.gs[index] = g
	destSlice.bs[index] = b
}

func (v1 *Vec3[float32]) scale(v2 Vec3[float32]) Vec3[float32] {
	return Vec3[float32]{
		v1.x * v2.x,
		v1.y * v2.y,
		v1.z * v2.z,
	}
}



/*······················································································kVecRGB  */
type VecRGB struct {
	r int32
	g int32
	b int32
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

func newVecRGB[T int | int8 | int16 | int32 | int64] (r T, g T, b T) VecRGB {
	return VecRGB{int32(r), int32(g), int32(b)}
}



/*······················································································kMath    */
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

func lerp(x, y int, f float64) float64 {
	return float64(x) * (1.0-f) + float64(y) * f
}

func lerp64(x, y, f float64) float64 {
	return x * (1.0-f) + (y * f)
}

func iLerp(x, y int, num float64) float64 {
    if x == y { return 0 }
    return (num - float64(x)) / float64(y - x)
}

func iLerp32(x, y int, num float32) float32 {
    if x == y { return 0 }
    return (num - float32(x)) / float32(y - x)
}

func iLerp64[T int | float64](x, y T, num float64) float64 {
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

func clamp[T int | int8 | int16 | int32 | float32 | float64](n, min, max T) T {
	if n > max { return max }
	if n < min { return min }
	return n
}

func clampMin[T int | int8 | int16 | int32 | float32 | float64](n, min T) T {
	if n < min { return min }
	return n
}


func B2i(b bool) int {
	if b { return 1 }
	return 0
}

// Fast boolean to integer
func fB2i(b bool) int {
    return int(*(*byte)(unsafe.Pointer(&b)))
}

// Flat index
func fiVec2(v Vec2) int {
	return v.y * MapW + v.x
}

// Flat index
func fi(x, y int) int {
	return y * MapW + x
}

func floatEq(a, b, ε float64) bool {
    diff := math.Abs(a - b)
    if diff <= ε {
        return true
    }
    return diff <= ε * math.Max(math.Abs(a), math.Abs(b))
}

func gammaCorrection(input float64) float64 {
	current := gammaCache.Load().(gammaMap)
	if CACHE {
		if val, ok := current[input]; ok {
			return val
		}
	}
	newR := math.Pow(input / 255, γRec) * 255

	if CACHE {
		current[input] = newR
		gammaCache.Store(current)
	}

	return newR
}

// Gamma approximation
func gamma06(x float64) float64 {
	y := 0.017 * x
	y2 := y * y
	return 17.4 * y2 * (0.72 + 0.36*y - 0.08*y2)
}


func angleBetween(a, b Vec2) float64 {
	ax := float64(a.x)
	ay := float64(a.y)
	bx := float64(b.x)
	by := float64(b.y)

	dot := ax * bx + ay * by
	mag := math.Hypot(ax, ay) * math.Hypot(bx, by)

	if mag == 0 {
		return 0
	}
	return math.Acos(dot / mag) // radians
}


func angleTo(a, b Vec2) float64 {
	current := angleCache.Load().(angleMap)
	if CACHE {
		if val, ok := current[
			[4]int{ a.x, a.y, b.x, b.y }]; ok {
			return val
		}
	}

	d := Vec2{b.x - a.x, b.y - a.y}
	angleTo := math.Atan2(float64(d.y), float64(d.x)) // radians, from -π to +π

	if CACHE {
		current[[4]int{ a.x, a.y, b.x, b.y }] = angleTo
		angleCache.Store(current)
	}
	return angleTo
}

func encodeIndex(p1, p2 Vec2) int {
	i0 := p1.x
	i1 := p1.y
	i2 := p2.x
	i3 := p2.y
	wordSize := 6
	smooshed := i0 << (wordSize * 3) | i1 << (wordSize * 2) | i2 << (wordSize) | i3
	return smooshed
}

func computeSquareDistances() {
	fmt.Printf("[33m")
	fmt.Println("Precomputing caches...")
	size := 64
	for x := range size {
		for y := range size {
			fmt.Printf("(%2d, %2d)\r", x, y)
			for x2 := range size {
				for y2 := range size {
					squaredDistances[
					encodeIndex(Vec2{x, y}, Vec2{x2, y2})] =
					math.Pow(float64(x - x2), 2) + math.Pow(float64(y - y2), 2)
				}
			}
		}
	}

	//squaredDistances[ encodeIndex(Vec2{43, 34}, Vec2{3, 4}) ] = 34
	//squaredDistances[ encodeIndex(Vec2{1, 21}, Vec2{8, 40}) ] = 69
	//squaredDistances[ encodeIndex(Vec2{61, 35}, Vec2{1, 0}) ] = 420

	fmt.Printf("[33m")
	fmt.Println("√ Finished - validating...[32m")

	// Validate
	pass := true
	for x := range size {
		fmt.Printf("[32m%8d\r", x*size*size*size)	

		for y := range size {
			for x2 := range size {
				for y2 := range size {
					p1 := Vec2{x, y}
					p2 := Vec2{x2, y2}

					calc := math.Pow(float64(x - x2), 2) + math.Pow(float64(y - y2), 2)
					memo := squaredDistances[encodeIndex(p1, p2)]

					if calc != memo {
						pass = false
						fmt.Printf("[31m(Calc): %5.0f  == %5.0f :(Memo)\n", calc, memo)	
					} 
				}
			}
		}
	}

	if pass {
		fmt.Printf("Success! %s entries checked[39m\n", intSeps(size*size*size*size))
		time.Sleep(time.Millisecond*500)
	} else {
		fmt.Printf("[31m")
		fmt.Printf("Failed! Cache invalid - %d entries checked[39m\n", size*size*size*size)
		time.Sleep(time.Millisecond*500)
	}
}


func squareDist(v1, v2 Vec2) float64 {
	if CACHE {
			return squaredDistances[
		v1.x << (wordSize * 3) |
		v1.y << (wordSize * 2) |
		v2.x << (wordSize)     |
		v2.y]
	}
	sqDst := math.Pow(float64(v1.x - v2.x), 2) + math.Pow(float64(v1.y - v2.y), 2)
	return sqDst
};				/*func _squareDist(v1, v2 Vec2) float64 {
						current := squaredCache.Load().(squaredDistMap)
						if CACHE {
							if val, ok := current[[4]int{v1.x, v1.y, v2.x, v2.y}]; ok {
								return val
							}
						}
						sqDst := math.Pow(float64(v1.x - v2.x), 2) + math.Pow(float64(v1.y - v2.y), 2)
						if CACHE {
							current[[4]int{v1.x, v1.y, v2.x, v2.y}] = sqDst
							squaredCache.Store(current)
						}
						return sqDst
					}*/

