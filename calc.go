package main

import (
	"fmt"
	"time"
	"strconv"
	"strings"
)

func AbsInt8(n int8) int8 {
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


func makeAverageDurationBuffer(size int) func(time.Duration) int64 {
	buffer := make([]time.Duration, size)
	i := 0
	return func(d time.Duration) int64 {
		buffer[i] = d
		i = wrapInt(i + 1, size)
		total := time.Duration(0)

		for n := range size {
			total += buffer[n]
		}

		return total.Microseconds() / int64(size)
	}
}

func makeAverageIntBuffer(size int) func(n int) float64 {
	buffer := make([]int, size)
	i := 0
	return func(n int) float64 {
		buffer[i] = n
		i = wrapInt(i + 1, size)

		total := 0

		for n := range size {
			total += buffer[n]
		}

		return float64(total) / float64(size)
	}
}

func timeF(d time.Duration) string {
	str := fmt.Sprintf("%v", d)


	if floatStr, found := strings.CutSuffix(str, "ms"); found {
		f, _ := strconv.ParseFloat(floatStr, 64)
		return fmt.Sprintf("%.2fms", f)
	}

	if floatStr, found := strings.CutSuffix(str, "µs"); found {
		f, _ := strconv.ParseFloat(floatStr, 64)
		return fmt.Sprintf("%.2fμs", f)
	}

	if floatStr, found := strings.CutSuffix(str, "s"); found {
		f, _ := strconv.ParseFloat(floatStr, 64)
		return fmt.Sprintf("%.2fs", f)
	}


	return str
}

func B2i(b bool) int {
	if b { return 1 }
	return 0
}

func intSeps(n int) string {
	if n < 1000 {
		return strconv.Itoa(n)
	}

	s := strconv.Itoa(n)
	result := ""

	for {
		if len(s) <= 3 {
			result = s + result
			return result
		}
		result = "," + s[len(s)-3:] + result
		s = s[:len(s)-3]
	}
}
