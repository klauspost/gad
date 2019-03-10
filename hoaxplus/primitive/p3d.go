package primitive

import "math"

type P3D struct{ X, Y, Z float32 }
type P3Ds []P3D

func (c *P3D) Scale(f float32) {
	c.X *= f
	c.Y *= f
	c.Z *= f
}

func (p P3Ds) Scale(f float32) {
	for i := range p {
		p[i].Scale(f)
	}
}

// rotateFn returns a function that rotates around (0,0,0).
// Supply angles in radians.
// dst must be same size or bigger than p.
func (p P3Ds) RotateTo(dst P3Ds, xAn, yAn, zAn float64) {
	fn := rotateFn(xAn, yAn, zAn)
	for i, v := range p {
		dst[i] = fn(v)
	}
}

// BehindCamera is magic
const BehindCamera = 10e21

// rotateFn returns a function that rotates around (0,0,0).
// dst must be same size or bigger than p.
func (p P3Ds) ProjectTo(dst []Point2D, w, h, zoff float32) {
	halfWidth := w * 0.5
	halfHeight := h * 0.5
	for i, v := range p {
		z := v.Z + zoff
		if z <= 0 {
			dst[i] = Point2D{X: BehindCamera, Y: BehindCamera}
			continue
		}
		invZ := 1 / z
		x := halfWidth * v.X * invZ
		y := halfWidth * v.Y * invZ
		x += halfWidth
		y += halfHeight
		dst[i] = Point2D{X: x, Y: y}
	}
}

// rotateFn returns a function that rotates around (0,0,0).
// Supply angles in radians.
func rotateFn(xAn, yAn, zAn float64) func(c P3D) P3D {
	var (
		s1 = float32(math.Sin(zAn))
		s2 = float32(math.Sin(xAn))
		s3 = float32(math.Sin(yAn))
		c1 = float32(math.Cos(zAn))
		c2 = float32(math.Cos(xAn))
		c3 = float32(math.Cos(yAn))

		zero  = c1*c3 + s1*s2*s3
		one   = c2 * s3
		two   = -c3*s1 + c1*s2*s3
		three = c2 * s1
		four  = -s2
		five  = c2 * c1
		six   = -c1*s3 + c3*s1*s2
		seven = c2 * c3
		eight = s1*s3 + c1*c3*s2
	)
	return func(c P3D) P3D {
		c.X, c.Y, c.Z = c.X*zero+c.Y*one+c.Z*two, c.X*three+c.Y*four+c.Z*five, c.X*six+c.Y*seven+c.Z*eight
		return c
	}
}
