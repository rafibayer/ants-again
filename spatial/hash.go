package spatial

import (
	"iter"
	"math"

	"github.com/rafibayer/ants-again/vector"
)

type hashKey struct {
	x, y int
}

type Hash[T vector.Point] struct {
	len   int
	size  float64
	cells map[hashKey][]T
}

var _ Spatial[vector.Point] = &Hash[vector.Point]{}

func NewHash[T vector.Point](size float64) Spatial[T] {
	return &Hash[T]{
		size:  size,
		cells: map[hashKey][]T{},
	}
}

func (h *Hash[T]) key(p vector.Point) hashKey {
	x := int(math.Floor(p.GetX() / h.size))
	y := int(math.Floor(p.GetY() / h.size))

	return hashKey{x: int(x), y: int(y)}
}

func (h *Hash[T]) Insert(p T) {
	k := h.key(p)
	h.cells[k] = append(h.cells[k], p)
	h.len++
}

func (h *Hash[T]) Points() []T {
	result := make([]T, h.Len())
	for _, cell := range h.cells {
		result = append(result, cell...)
	}

	return result
}

func (h *Hash[T]) PointsIter() iter.Seq[T] {
	return func(yield func(T) bool) {
		for _, cell := range h.cells {
			for _, p := range cell {
				if !yield(p) {
					return
				}
			}
		}
	}
}

func (h *Hash[T]) RadialSearch(center vector.Point, radius float64) []T {
	result := []T{}
	for p := range h.RadialSearchIter(center, radius) {
		result = append(result, p)
	}

	return result
}

// todo: dedup with RadialSearch
func (h *Hash[T]) RadialSearchIter(center vector.Point, radius float64) iter.Seq[T] {
	return func(yield func(T) bool) {
		// 1. Determine the center cell and search bounds
		cx := int(math.Floor(center.GetX() / h.size))
		cy := int(math.Floor(center.GetY() / h.size))

		// ceil(radius / cell_size)
		cellRadius := int(math.Ceil(radius / h.size))
		r2 := radius * radius

		// 2. Iterate the square grid of candidates
		for dx := -cellRadius; dx <= cellRadius; dx++ {
			for dy := -cellRadius; dy <= cellRadius; dy++ {
				key := hashKey{x: cx + dx, y: cy + dy}

				// Optimization 1: Check map existence first.
				// If the cell is empty, we don't care if it overlaps.
				points, ok := h.cells[key]
				if !ok || len(points) == 0 {
					continue
				}

				// Optimization 2: Cell-Circle Intersection Test
				// We find the closest point on the grid cell to the search center.
				// Calculate cell bounds
				minX := float64(key.x) * h.size
				maxX := minX + h.size
				minY := float64(key.y) * h.size
				maxY := minY + h.size

				// Clamp the center to the cell bounds to find the closest point
				closestX := math.Max(minX, math.Min(center.GetX(), maxX))
				closestY := math.Max(minY, math.Min(center.GetY(), maxY))

				// Calculate squared distance from center to that closest point
				distX := center.GetX() - closestX
				distY := center.GetY() - closestY
				distSq := (distX * distX) + (distY * distY)

				// If the closest point on the square is outside the radius,
				// the whole square is outside.
				if distSq > r2 {
					continue
				}

				// 3. Point-Circle Intersection Test (Standard)
				for _, p := range points {
					xDiff := center.GetX() - p.GetX()
					yDiff := center.GetY() - p.GetY()
					if (xDiff*xDiff + yDiff*yDiff) <= r2 {
						if !yield(p) {
							return
						}
					}
				}
			}
		}
	}

}

func (h *Hash[T]) Remove(p T) T {
	k := h.key(p)

	index := -1
	for i, c := range h.cells[k] {
		if c.GetX() == p.GetX() && c.GetY() == p.GetY() {
			index = i
			break
		}
	}

	if index == -1 {
		var zero T
		return zero
	}

	h.len--

	// swap with last and truncate
	deleted := h.cells[k][index]
	h.cells[k][index] = h.cells[k][len(h.cells[k])-1]
	h.cells[k] = h.cells[k][:len(h.cells[k])-1]

	return deleted
}

func (h *Hash[T]) Len() int {
	return h.len
}
