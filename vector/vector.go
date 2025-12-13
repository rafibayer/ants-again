package vector

import (
	"math"

	"github.com/rafibayer/ants-again/util"
)

var (
	ZERO = Vector{0, 0}
)

type Vector struct {
	X, Y float64
}

func (p Vector) Distance2(other Vector) float64 {
	xDiff := p.X - other.X
	yDiff := p.Y - other.Y
	return (xDiff * xDiff) + (yDiff * yDiff)
}

func (p Vector) Distance(other Vector) float64 {
	return math.Sqrt(p.Distance2(other))
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

func (p Vector) AngleBetween(other Vector) float64 {
	dot := p.X*other.X + p.Y*other.Y

	cosTheta := dot / (p.Magnitude() * other.Magnitude())
	cosTheta = util.Clamp(-1, cosTheta, 1)

	return math.Acos(cosTheta) * 180 / math.Pi
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
