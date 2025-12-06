package spatial

import (
	"iter"

	"github.com/rafibayer/ants-again/vector"
)

// common interface of various "spatial" data structures.
type Spatial[T vector.Point] interface {
	Insert(p T)
	Points() []T
	PointsIter() iter.Seq[T]
	Remove(p T) T
	RadialSearch(center vector.Point, radius float64) []T
	RadialSearchIter(center vector.Point, radius float64) iter.Seq[T]
	Len() int
}
