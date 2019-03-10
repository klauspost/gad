package primitive

type P2D struct {
	X, Y float32
}

type P2Ds []P2D

func (p P2D) DistSq(p2 P2D) float32 {
	a, b := p.X-p2.X, p.Y-p2.Y
	return a*a + b*b
}
