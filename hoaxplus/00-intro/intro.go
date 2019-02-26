package intro

import (
	"github.com/klauspost/gfx"
	"image"
)

func NewIntro(screen *image.Gray) gfx.TimedEffect {
	return &intro{draw: screen}
}

type intro struct {
	draw    *image.Gray
	cleared bool
}

func (fx *intro) Render(t float64) image.Image {
	if !fx.cleared {
		for i := range fx.draw.Pix {
			fx.draw.Pix[i] = 0
		}
		fx.cleared = true
	}
	img := fx.draw
	w, h := img.Rect.Dx(), img.Rect.Dy()
	top := img.Pix[:img.Stride]
	bottom := img.Pix[(h-1)*img.Stride:]

	xwhere := (int)(float64(w+1) * t)
	const white = 255
	const grey = 40

	drawLine := func(pix []byte, col byte, start, stop int) {
		if start < 0 {
			start = 0
		}
		if stop > len(pix) {
			stop = len(pix)
		}
		for i := start; i < stop; i++ {
			pix[i] = col
		}
	}
	drawLine(top, white, 0, xwhere+4)
	drawLine(top, grey, 0, xwhere)
	drawLine(bottom, white, w-xwhere-4, w)
	drawLine(bottom, grey, w-xwhere, w)
	return img
}
