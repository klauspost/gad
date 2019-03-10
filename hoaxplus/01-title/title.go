package title

import (
	"image"
	"image/color"
	"math"

	"github.com/klauspost/gad/hoaxplus/primitive"
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
	b, err := gfx.Load("Basehead.obj")
	if err != nil {
		panic(err)
	}
	t.verts, t.edges = primitive.LoadOBJ(b)
	t.vTransformed = make(primitive.P3Ds, len(t.verts))
	t.vProjected = make(primitive.P2Ds, len(t.verts))
	return &t
}

type title struct {
	draw         *image.Gray
	screen       *image.RGBA
	cleared      bool
	verts        primitive.P3Ds
	edges        [][2]int
	vTransformed primitive.P3Ds
	vProjected   primitive.P2Ds
	color        [2][256]color.RGBA
}

func (fx *title) Render(t float64) image.Image {
	img := fx.draw
	for i := range img.Pix {
		img.Pix[i] = 192
	}
	// Convert to float
	fw, fh := float32(img.Rect.Dx()), float32(img.Rect.Dy())

	// Draw line across screen
	primitive.Line{P2: primitive.Point2D{X: fw, Y: fh}}.DrawAA(img, 255)

	// Draw model
	fx.verts.RotateTo(fx.vTransformed, math.Pi/2, -t*math.Pi*2, 0)
	fx.vTransformed.ProjectTo(fx.vProjected, fw, fh, float32(math.Sin(t*math.Pi))*10)
	for _, edge := range fx.edges {
		p0, p1 := fx.vProjected[edge[0]], fx.vProjected[edge[1]]
		if p0.X == primitive.BehindCamera || p1.X == primitive.BehindCamera {
			continue
		}
		primitive.Line{
			P1: p0,
			P2: p1,
		}.DrawAA(img, 0)
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

func (fx *title) transferGrey() {
	_, h := fx.screen.Bounds().Dx(), fx.screen.Bounds().Dy()
	src := fx.draw
	dst := fx.screen
	for y := 0; y < h; y++ {
		line := src.Pix[y*src.Stride : y*src.Stride+src.Rect.Dx()]
		dLine := dst.Pix[y*dst.Stride : y*dst.Stride+len(line)*4]

		for x, v := range line {
			dLine[x*4] = v
			dLine[x*4+1] = v
			dLine[x*4+2] = v
			dLine[x*4+3] = 255
		}
	}
}
