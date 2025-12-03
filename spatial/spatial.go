package spatial

import "github.com/rafibayer/ants-again/vector"

// common interface of various "spatial" data structures.
type Spatial[T vector.Point] interface {
	Insert(p T)
	Points() []T

	// same as points, but returns an unbuffered channel to reduce allocation
	// must be fully consumed by caller, cannot modify while in use
	Chan() chan T
	Remove(p T) T
	RadialSearch(center vector.Point, radius float64) []T
	Len() int
}
