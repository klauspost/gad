package primitive

type Point2D struct {
	X, Y float32
}

type P2Ds []Point2D

func (p Point2D) DistSq(p2 Point2D) float32 {
	a, b := p.X-p2.X, p.Y-p2.Y
	return a*a + b*b
}
