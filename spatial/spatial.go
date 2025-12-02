package spatial

// common interface of various "spatial" data structures.
type Spatial[T any] interface {
	Insert(p T)
	Points() []T
	Chan() chan T // same as points, but returns an unbuffered channel to reduce allocation
	Remove(p T) T
	RadialSearch(center T, radius float64, dst func(a T, b T) float64) []T
}
