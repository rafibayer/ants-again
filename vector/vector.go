package vec

import (
	"math"

	"github.com/rafibayer/ants-again/kdtree"
)

var (
	ZERO = Vector{0, 0}
)

type Vector struct {
	X, Y float64
}

var _ kdtree.Point = Vector{}

func (p Vector) Dimension(i int) float64 {
	// branchless, undefined behavior if 'i' not in {0, 1}
	return p.X*float64(1-i) + p.Y*float64(i)
}

func (p Vector) Dimensions() int {
	return 2
}

func (p Vector) Distance(other Vector) float64 {
	xDiff := p.X - other.X
	yDiff := p.Y - other.Y
	return math.Sqrt((xDiff * xDiff) + (yDiff * yDiff))
}

func (p Vector) Add(other Vector) Vector {
	return Vector{X: p.X + other.X, Y: p.Y + other.Y}
}

func (p Vector) Sub(other Vector) Vector {
	return p.Add(other.Mul(-1.0))
}

func (p Vector) Magnitude() float64 {
	return math.Sqrt(p.X*p.X + p.Y*p.Y)
}

func (p Vector) Normalize() Vector {
	mag := p.Magnitude()
	return Vector{X: p.X / mag, Y: p.Y / mag}
}

func (p Vector) Mul(v float64) Vector {
	return Vector{X: p.X * v, Y: p.Y * v}
}

func (p Vector) CosineSimilarity(other Vector) float64 {
	dot := p.X*other.X + p.Y*other.Y
	return dot / (p.Magnitude() * other.Magnitude())
}

func (v Vector) Rotate(deg float64) Vector {
	rad := deg * math.Pi / 180
	cos := math.Cos(rad)
	sin := math.Sin(rad)

	return Vector{
		X: v.X*cos - v.Y*sin,
		Y: v.X*sin + v.Y*cos,
	}
}

func (v Vector) ToGrid() (row, col int) {
	row = int(math.Round(v.Y))
	col = int(math.Round(v.X))

	return row, col
}
