package primitive

import (
	"fmt"
	"image"
)

func (l Line) Draw(dst *image.Gray, col byte) {
	w, h := dst.Rect.Dx(), dst.Rect.Dy()
	if !l.clip(w, h) {
		return
	}

	yLonger := false
	shortLen := l.P2.Y - l.P1.Y
	longLen := l.P2.X - l.P1.X
	if absf(shortLen) > absf(longLen) {
		shortLen, longLen = longLen, shortLen
		yLonger = true
	}
	var decInc float32
	if longLen != 0 {
		decInc = shortLen / longLen
	}
	const sigma = 0.001

	if yLonger {
		if longLen > 0 {
			// Top to bottom, one y per loop
			l.P2.Y += sigma
			for j := l.P1.X; l.P1.Y <= l.P2.Y; l.P1.Y++ {
				setGray(dst, j, l.P1.Y, col)
				j += decInc
			}
			return
		}
		// Bottom to top, one y per loop
		l.P2.Y -= sigma
		for j := l.P1.X; l.P1.Y >= l.P2.Y; l.P1.Y-- {
			setGray(dst, j, l.P1.Y, col)
			j -= decInc
		}
		return
	}
	if longLen > 0 {
		// Left to right, one X per loop
		l.P2.X += sigma
		for j := l.P1.Y; l.P1.X <= l.P2.X; l.P1.X++ {
			setGray(dst, l.P1.X, j, col)
			j += decInc
		}
		return
	} else {
		l.P2.X -= sigma
		// Right to left, one X per loop
		for j := l.P1.Y; l.P1.X >= l.P2.X; l.P1.X-- {
			setGray(dst, l.P1.X, j, col)
			j -= decInc
		}
	}
}

func (l Line) DrawAA(dst *image.Gray, col byte) {
	w, h := dst.Rect.Dx(), dst.Rect.Dy()
	if !l.clip(w, h) {
		return
	}

	yLonger := false
	shortLen := l.P2.Y - l.P1.Y
	longLen := l.P2.X - l.P1.X
	if absf(shortLen) > absf(longLen) {
		shortLen, longLen = longLen, shortLen
		yLonger = true
	}
	var decInc float32
	if longLen != 0 {
		decInc = shortLen / longLen
	}
	const sigma = 0.001

	if yLonger {
		if longLen > 0 {
			// Top to bottom, one y per loop
			l.P2.Y += sigma
			for j := l.P1.X; l.P1.Y <= l.P2.Y; l.P1.Y++ {
				setGrayHorizontalAA(dst, j, l.P1.Y, col)
				j += decInc
			}
			return
		}
		// Bottom to top, one y per loop
		l.P2.Y -= sigma
		for j := l.P1.X; l.P1.Y >= l.P2.Y; l.P1.Y-- {
			setGrayHorizontalAA(dst, j, l.P1.Y, col)
			j -= decInc
		}
		return
	}
	if longLen > 0 {
		// Left to right, one X per loop
		l.P2.X += sigma
		for j := l.P1.Y; l.P1.X <= l.P2.X; l.P1.X++ {
			setGrayVerticalAA(dst, l.P1.X, j, col)
			j += decInc
		}
		return
	}
	// Right to left, one X per loop
	l.P2.X -= sigma
	for j := l.P1.Y; l.P1.X >= l.P2.X; l.P1.X-- {
		setGrayVerticalAA(dst, l.P1.X, j, col)
		j -= decInc
	}
}

// setGray will set a pixel.
// It is assumed that X and y are clipped.
func setGray(img *image.Gray, x, y float32, col byte) {
	img.Pix[roundP(x)+roundP(y)*img.Stride] = col
}

// setGray will set a pixel.
// It is assumed that X and y are clipped.
func setGrayAA(img *image.Gray, x, y float32, col byte) {
	// Convert to fixed point
	xx, yy := int(256*x), int(256*y)

	// Calculate weights
	x1, y1 := xx&255, yy&255
	x0, y0 := 256-x1, 256-y1

	// Apply weighted color to pixel
	weight := func(pix *byte, w int) {
		wOrg := 256 - w
		p := int(*pix)*wOrg + int(col)*w
		*pix = byte(p >> 8)
	}

	// Remove fraction from pixels coordinates.
	xx >>= 8
	yy >>= 8

	// Check if we can write to the next pixel
	xOK, yOK := xx < img.Rect.Dx()-1, yy < img.Rect.Dy()-1

	// Pre-multiply with image stride
	yy *= img.Stride

	// draw topleft pixel.
	p := &img.Pix[xx+yy]
	weight(p, (x0*y0)>>8)
	if xOK {
		p = &img.Pix[1+xx+yy]
		weight(p, (x1*y0)>>8)
	}
	if yOK {
		yy += img.Stride
		p = &img.Pix[xx+yy]
		weight(p, (x0*y1)>>8)
		if xOK {
			p = &img.Pix[1+xx+yy]
			weight(p, (x1*y1)>>8)
		}
	}
}

// setGrayHorizontalAA will set a pixel, but only do AA vertically.
// This can be used when each pixel is drawn horizontally.
// It is assumed that X and y are clipped.
func setGrayHorizontalAA(img *image.Gray, x, y float32, col byte) {
	// Convert to fixed point
	xx, yy := int(256*x), int(256*y)

	// Calculate weights
	x1 := xx & 255
	x0 := 256 - x1

	// Apply weighted color to pixel.
	// w0 is color weight, w1 is existing weight.
	// Sum of w0 and w1 must be <= 256
	weight := func(pix *byte, w0, w1 int) {
		p := int(*pix)*w1 + int(col)*w0
		*pix = byte(p >> 8)
	}

	// Remove fraction from pixels coordinates.
	xx >>= 8
	yy >>= 8

	// Check if we can write to the next pixel
	xOK := xx < img.Rect.Dx()-1

	// Pre-multiply with image stride
	yy *= img.Stride

	// draw topleft pixel.
	p := &img.Pix[xx+yy]
	weight(p, x0, x1)
	if xOK {
		p = &img.Pix[1+xx+yy]
		weight(p, x1, x0)
	}
}

// setGrayVerticalAA will set a pixel, but only do AA horizontally.
// This can be used when each pixel is drawn vertically.
// It is assumed that X and y are clipped.
func setGrayVerticalAA(img *image.Gray, x, y float32, col byte) {
	// Convert to fixed point
	xx, yy := int(256*x), int(256*y)

	// Calculate weights
	y1 := yy & 255
	y0 := 256 - y1

	// Apply weighted color to pixel.
	// w0 is color weight, w1 is existing weight.
	// Sum of w0 and w1 must be <= 256
	weight := func(pix *byte, w0, w1 int) {
		p := int(*pix)*w1 + int(col)*w0
		*pix = byte(p >> 8)
	}

	// Remove fraction from pixels coordinates.
	xx >>= 8
	yy >>= 8

	// Check if we can write to the next pixel
	yOK := yy < img.Rect.Dy()-1

	// Pre-multiply with image stride
	yy *= img.Stride

	// draw topleft pixel.
	p := &img.Pix[xx+yy]
	weight(p, y0, y1)
	if yOK {
		yy += img.Stride
		p = &img.Pix[xx+yy]
		weight(p, y1, y0)
	}
}

// roundP will round positive numbers towards nearest integer.
func roundP(x float32) int {
	return int(x + 0.5)
}

func absf(x float32) float32 {
	switch {
	case x < 0:
		return -x
	case x == 0:
		return 0 // return correctly abs(-0)
	}
	return x
}

type Point2D struct {
	X, Y float32
}

type Line struct {
	P1, P2 Point2D
}

// line clipping converted from C++ on
// https://www.geeksforgeeks.org/line-clipping-set-1-cohen-sutherland-algorithm/
// Will also ensure that Y1 is smaller or equal to Y2.
func (l *Line) clip(w, h int) bool {
	// Defining region codes
	const (
		INSIDE = 0
		LEFT   = 1 << iota
		RIGHT
		BOTTOM
		TOP
	)
	if l.P1.Y > l.P2.Y {
		// Swap
		l.P1.X, l.P2.X = l.P2.X, l.P1.X
		l.P1.Y, l.P2.Y = l.P2.Y, l.P1.Y
	}
	fw, fh := float32(w)-0.51, float32(h)-0.51

	// Function to compute region code for a point(X, y)
	var computeCode = func(p Point2D) int {
		// initialized as being inside
		code := INSIDE

		if p.X < 0 {
			// to the left of rectangle
			code |= LEFT
		} else if p.X > fw {
			// to the right of rectangle
			code |= RIGHT
		}
		if p.Y < 0 {
			// below the rectangle
			code |= BOTTOM
		} else if p.Y > fh {
			// above the rectangle
			code |= TOP
		}
		return code
	}

	code1 := computeCode(l.P1)
	code2 := computeCode(l.P2)
	const minLen = (0.1 * 0.1) * 2

	for n := 0; ; n++ {
		if false && n > 5 {
			code1 = computeCode(l.P1)
			code2 = computeCode(l.P2)
			fmt.Println("Could not clip", l.P1, code1, "->", l.P2, code2)
			return false
		}
		if l.P1.DistSq(l.P2) < minLen {
			// Line too small to draw.
			return false
		}
		if (code1 == 0) && (code2 == 0) {
			// If both endpoints lie within rectangle
			return true
		}
		if (code1 & code2) != 0 {
			// If both endpoints are outside rectangle,
			// in same region
			return false
		}

		// Some segment of line lies within the
		// rectangle
		var codeOut int
		var x, y float32

		// At least one endpoint is outside the
		// rectangle, pick it.
		if code1 != 0 {
			codeOut = code1
		} else {
			codeOut = code2
		}

		// Find intersection point;
		// using formulas y = Y1 + slope * (X - X1),
		// X = X1 + (1 / slope) * (y - Y1)
		switch {
		case (codeOut & TOP) != 0:
			// point is above the clip rectangle
			x = l.P1.X + (l.P2.X-l.P1.X)*(fh-l.P1.Y)/(l.P2.Y-l.P1.Y)
			y = fh

		case (codeOut & BOTTOM) != 0:
			// point is below the rectangle
			x = l.P1.X + (l.P2.X-l.P1.X)*(-l.P1.Y)/(l.P2.Y-l.P1.Y)
			y = 0
		case (codeOut & RIGHT) != 0:
			// point is to the right of rectangle
			y = l.P1.Y + (l.P2.Y-l.P1.Y)*(fw-l.P1.X)/(l.P2.X-l.P1.X)
			x = fw
		case (codeOut & LEFT) != 0:
			// point is to the left of rectangle
			y = l.P1.Y + (l.P2.Y-l.P1.Y)*(-l.P1.X)/(l.P2.X-l.P1.X)
			x = 0
		}

		// Now intersection point X,y is found
		// We replace point outside rectangle
		// by intersection point
		if codeOut == code1 {
			l.P1.X = x
			l.P1.Y = y
			code1 = computeCode(l.P1)
		} else {
			l.P2.X = x
			l.P2.Y = y
			code2 = computeCode(l.P2)
		}
	}
}
