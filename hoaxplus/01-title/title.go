package title

import (
	"image"
	"image/color"

	"github.com/klauspost/gfx"
)

func NewTitle(draw *image.Gray, screen *image.RGBA) gfx.TimedEffect {
	t := title{draw: draw, screen: screen}
	for i := range t.color[0][:] {
		if i <= 192 {
			t.color[0][i] =
				color.RGBA{
					R: uint8((248 * i) >> 8),
					G: uint8((289 * i) >> 8),
					B: uint8((294 * i) >> 8),
					A: 0xff,
				}
			t.color[1][i] =
				color.RGBA{
					R: uint8((300 * i) >> 8),
					G: uint8((278 * i) >> 8),
					B: uint8((252 * i) >> 8),
					A: 0xff,
				}
		} else {
			pixrem := i - 192
			t.color[0][i] =
				color.RGBA{
					R: 198 + uint8((57*pixrem)>>6),
					G: 218 + uint8((37*pixrem)>>6),
					B: 222 + uint8((33*pixrem)>>6),
					A: 0xff,
				}
			t.color[1][i] =
				color.RGBA{
					R: 226 + uint8((29*pixrem)>>6),
					G: 210 + uint8((45*pixrem)>>6),
					B: 190 + uint8((65*pixrem)>>6),
					A: 0xff,
				}
		}
	}

	return &t
}

type title struct {
	draw    *image.Gray
	screen  *image.RGBA
	cleared bool
	color   [2][256]color.RGBA
}

func (fx *title) Render(t float64) image.Image {
	img := fx.draw
	for i := range img.Pix {
		img.Pix[i] = byte(t * 255)
	}

	fx.transfer()
	return fx.screen
}

func (fx *title) transfer() {
	w, h := fx.screen.Bounds().Dx(), fx.screen.Bounds().Dy()
	src := fx.draw
	dst := fx.screen
	for y := 0; y < h; y++ {
		xch := (int)((float64)(y*w) / (float64)(h))
		line := src.Pix[y*src.Stride : y*src.Stride+src.Rect.Dx()]
		dLine := dst.Pix[y*dst.Stride : y*dst.Stride+len(line)*4]

		// Left of change
		pal := fx.color[0][:]
		for x, v := range line[:xch] {
			col := pal[v]
			dLine[x*4] = col.R
			dLine[x*4+1] = col.G
			dLine[x*4+2] = col.B
			dLine[x*4+3] = 255
		}

		// Right side.
		line = line[xch:]
		dLine = dLine[xch*4:]
		pal = fx.color[1][:]
		for x, v := range line {
			col := pal[v]
			dLine[x*4] = col.R
			dLine[x*4+1] = col.G
			dLine[x*4+2] = col.B
			dLine[x*4+3] = 255
		}
	}
}
