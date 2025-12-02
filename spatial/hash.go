package spatial

import (
	"math"

	"github.com/rafibayer/ants-again/vector"
)

type hashKey struct {
	x, y int
}

type Hash[T vector.Point] struct {
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

func (h *Hash[T]) Chan() chan T {
	dst := make(chan T)

	go func() {
		for _, cell := range h.cells {
			for _, e := range cell {
				dst <- e
			}
		}

		close(dst)
	}()

	return dst
}

func (h *Hash[T]) Insert(p T) {
	k := h.key(p)
	h.cells[k] = append(h.cells[k], p)
}

func (h *Hash[T]) Points() []T {
	result := make([]T, 0)
	for _, cell := range h.cells {
		result = append(result, cell...)
	}

	return result
}

func (h *Hash[T]) RadialSearch(center vector.Point, radius float64) []T {
	result := []T{}

	// Determine the center cell
	key := h.key(center)

	// How many cells to search in each axis
	// ceil(radius / cell_size)
	cellRadius := int(math.Ceil(radius / h.size))

	r2 := radius * radius

	for dx := -cellRadius; dx <= cellRadius; dx++ {
		for dy := -cellRadius; dy <= cellRadius; dy++ {
			k := hashKey{x: key.x + dx, y: key.y + dy}

			points, ok := h.cells[k]
			if !ok {
				continue
			}

			// Check each candidate point
			for _, p := range points {
				// Use provided distance metric
				xDiff := center.GetX() - p.GetX()
				yDiff := center.GetY() - p.GetY()
				if (xDiff*xDiff + yDiff*yDiff) <= r2 {
					result = append(result, p)
				}
			}
		}
	}

	return result
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

	// swap with last and truncate
	deleted := h.cells[k][index]
	h.cells[k][index] = h.cells[k][len(h.cells[k])-1]
	h.cells[k] = h.cells[k][:len(h.cells[k])-1]

	return deleted
}
