package util

import "math/rand/v2"

// remap [-1, 1] to [0, 1]
func LinearRemap(x float64) float64 {
	return (x + 1) * 0.5
}

func Chance(odds float64) bool {
	return rand.Float64() < odds
}

func Rand(min, max float64) float64 {
	return min + rand.Float64()*(max-min)
}

// [min, max)
func RandInt(min, max int) int {
	return rand.IntN(max-min) + min
}

func Clamp(low, x, high float64) float64 {
	if x < low {
		return low
	}
	if x > high {
		return high
	}

	return x
}

func Ptr[T any](t T) *T {
	return &t
}
