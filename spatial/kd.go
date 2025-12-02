package spatial

import (
	"github.com/rafibayer/ants-again/kdtree"
	"github.com/rafibayer/ants-again/kdtree/kdrange"
)

type KyroyKD[T kdtree.Point] struct {
	inner *kdtree.KDTree
}

var _ Spatial[kdtree.Point] = &KyroyKD[kdtree.Point]{}

func NewKyroyKD[T kdtree.Point]() Spatial[T] {
	return &KyroyKD[T]{inner: kdtree.New(nil)}
}

func (t *KyroyKD[T]) Insert(p T) {
	t.inner.Insert(p)
}

func (t *KyroyKD[T]) KNN(p T, k int) []T {
	return Map(t.inner.KNN(p, k), func(p kdtree.Point) T {
		return p.(T)
	})
}

func (t *KyroyKD[T]) Points() []T {
	return Map(t.inner.Points(), func(p kdtree.Point) T {
		return p.(T)
	})
}

func (t *KyroyKD[T]) Chan() chan T {
	src := t.inner.Chan()
	dst := make(chan T)

	go func() {
		for e := range src {
			dst <- e.(T)
		}

		close(dst)
	}()

	return dst
}

func (t *KyroyKD[T]) RadialSearch(center T, radius float64, dst func(a T, b T) float64) []T {
	// implemented via range search, followed by filter on dist
	// 2d KD tree allows for range search in a rectangle.
	// we find points in the rectangle, then filter down to those within a circle bound by the rectangle
	xMin := center.Dimension(0) - radius
	xMax := center.Dimension(0) + radius
	yMin := center.Dimension(1) - radius
	yMax := center.Dimension(1) + radius

	points := t.RangeSearch(kdrange.New(xMin, xMax, yMin, yMax))
	result := make([]T, 0)

	for _, point := range points {
		if dst(center, point) <= radius {
			result = append(result, point)
		}
	}

	return result

}

func (t *KyroyKD[T]) RangeSearch(r [][2]float64) []T {
	return Map(t.inner.RangeSearch(r), func(p kdtree.Point) T {
		return p.(T)
	})
}

func (t *KyroyKD[T]) Remove(p T) T {
	return t.inner.Remove(p).(T)
}
