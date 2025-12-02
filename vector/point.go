package vector

type Point interface {
	GetX() float64
	GetY() float64
}

var _ Point = Vector{}

func (p Vector) GetX() float64 {
	return p.X
}

func (p Vector) GetY() float64 {
	return p.Y
}
