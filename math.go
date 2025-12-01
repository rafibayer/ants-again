package main

// remap [-1, 1] to [0, 1]
func LinearRemap(x float64) float64 {
	return (x + 1) * 0.5
}
