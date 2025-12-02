package main

import "math/rand/v2"

// remap [-1, 1] to [0, 1]
func LinearRemap(x float64) float64 {
	return (x + 1) * 0.5
}

func Rand(min, max float64) float64 {
	return min + rand.Float64()*(max-min)
}

func Ptr[T any](t T) *T {
	return &t
}
