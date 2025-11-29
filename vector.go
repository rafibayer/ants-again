package main

import (
	"math"

	"github.com/kyroy/kdtree"
	"github.com/kyroy/kdtree/kdrange"
)

var (
	UP    = Vector{x: 0.0, y: 1.0}
	DOWN  = UP.Mul(-1)
	RIGHT = Vector{x: 1.0, y: 0.0}
	LEFT  = RIGHT.Mul(-1)
)

type Vector struct {
	x, y float64
}

var _ kdtree.Point = Vector{}

func (p Vector) Dimension(i int) float64 {
	// branchless, undefined behavior if 'i' not in {0, 1}
	return p.x*float64(1-i) + p.y*float64(i)
}

func (p Vector) Dimensions() int {
	return 2
}

func (p Vector) Distance(other Vector) float64 {
	xDiff := p.x - other.x
	yDiff := p.y - other.y
	return math.Sqrt((xDiff * xDiff) + (yDiff * yDiff))
}

func (p Vector) Add(other Vector) Vector {
	return Vector{x: p.x + other.x, y: p.y + other.y}
}

func (p Vector) Sub(other Vector) Vector {
	return p.Add(other.Mul(-1.0))
}

func (p Vector) Magnitude() float64 {
	return math.Sqrt(p.x*p.x + p.y*p.y)
}

func (p Vector) Normalize() Vector {
	mag := p.Magnitude()
	return Vector{x: p.x / mag, y: p.y / mag}
}

func (p Vector) Mul(v float64) Vector {
	return Vector{x: p.x * v, y: p.y * v}
}

func (v Vector) Rotate(deg float64) Vector {
	rad := deg * math.Pi / 180
	cos := math.Cos(rad)
	sin := math.Sin(rad)

	return Vector{
		x: v.x*cos - v.y*sin,
		y: v.x*sin + v.y*cos,
	}
}

func FindWithin(tree *kdtree.KDTree, center Vector, radius float64) []kdtree.Point {
	// 2d KD tree allows for range search in a rectangle.
	// we find points in the rectangle, then filter down to those within a circle bound by the rectangle
	xMin := center.x - radius
	xMax := center.x + radius
	yMin := center.y - radius
	yMax := center.y + radius

	points := tree.RangeSearch(kdrange.New(xMin, xMax, yMin, yMax))
	result := make([]kdtree.Point, 0)

	for _, point := range points {
		vec := Vector{x: point.Dimension(0), y: point.Dimension(1)}
		if vec.Distance(center) <= radius {
			result = append(result, point)
		}
	}

	return result
}
