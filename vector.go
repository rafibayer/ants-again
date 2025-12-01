package main

import (
	"math"

	"github.com/kyroy/kdtree/kdrange"
	"github.com/rafibayer/ants-again/kdtree"
)

var (
	ZERO = Vector{0, 0}
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

func (p Vector) CosineSimilarity(other Vector) float64 {
	dot := p.x*other.x + p.y*other.y
	return dot / (p.Magnitude() * other.Magnitude())
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

func (v Vector) ToGrid() (row, col int) {
	row = int(math.Round(v.y))
	col = int(math.Round(v.x))

	return row, col
}

func KDSearchRadius(tree *kdtree.KDTree, center Vector, radius float64) []kdtree.Point {
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

func VisitCircle(
	grid *[GAME_SIZE][GAME_SIZE]float32,
	rowF, colF, radius float64,
	callback func(value float32, weight float64, r, c int),
) {
	if radius <= 0 {
		return
	}

	rows := len(grid)
	if rows == 0 {
		return
	}
	cols := len(grid[0])

	// Search bounding box around the float center
	minR := int(math.Floor(rowF - radius))
	maxR := int(math.Ceil(rowF + radius))
	minC := int(math.Floor(colF - radius))
	maxC := int(math.Ceil(colF + radius))

	// Clamp to grid
	if minR < 0 {
		minR = 0
	}
	if minC < 0 {
		minC = 0
	}
	if maxR > rows-1 {
		maxR = rows - 1
	}
	if maxC > cols-1 {
		maxC = cols - 1
	}

	r2 := radius * radius

	for r := minR; r <= maxR; r++ {
		for c := minC; c <= maxC; c++ {

			// Distance from cell center to circle center
			dr := (float64(r) + 0.5) - rowF
			dc := (float64(c) + 0.5) - colF
			dist2 := dr*dr + dc*dc

			if dist2 <= r2 {
				// Fully inside → weight 1
				callback(grid[r][c], 1.0, r, c)
			} else {
				// Optional: compute soft boundary weight
				// Example: linear falloff in a margin of thickness 1 cell
				dist := math.Sqrt(dist2)
				if dist-radius < 1.0 && dist < radius+1.0 {
					// linear blend for partial overlap
					weight := 1.0 - (dist - radius)
					if weight < 0 {
						weight = 0
					}
					callback(grid[r][c], weight, r, c)
				}
			}
		}
	}
}
