package main

import (
	"image"
	_ "image/png"
	"math/bits"

	_ "github.com/klauspost/gad/ep02/data" // Load data.
	"github.com/klauspost/gfx"
)

// Generates binary data.
// To install go-bindata, do: go get -u github.com/jteeuwen/go-bindata/...
//
//go:generate go-bindata -ignore=\.go\z -pkg=data -o ./data/data.go ./data/...

const (
	renderWidth  = 640
	renderHeight = 360
)

func main() {
	fx := newFx("data/ballet-egon-256.png")
	gfx.Run(func() { gfx.RunTimed(fx) })
}

type fx struct {
	img        *image.Paletted
	draw       *image.Paletted
	logW, logH uint
	lines      [][]byte
}

func newFx(file string) *fx {
	var fx fx

	// Load picture
	img, err := gfx.LoadPalPicture(file)
	if err != nil {
		panic(err)
	}
	fx.img = img

	// Calculate the with in log(2)
	fx.logW = uint(bits.Len32(uint32(img.Rect.Dx()))) - 1
	fx.logH = uint(bits.Len32(uint32(img.Rect.Dy()))) - 1

	// Ensure that the width and height are actually powers of two
	if img.Rect.Dx() != 1<<fx.logW {
		panic("image " + file + " width is not power of two.")
	}
	if img.Rect.Dy() != 1<<fx.logH {
		panic("image " + file + " height is not power of two.")
	}

	// Create our draw buffer
	fx.draw = image.NewPaletted(image.Rect(0, 0, renderWidth, renderHeight), img.Palette)

	// Store each line as a slice in a slice.
	fx.lines = make([][]byte, fx.draw.Rect.Dy())
	for y := range fx.lines {
		fx.lines[y] = fx.draw.Pix[y*fx.draw.Stride : y*fx.draw.Stride+fx.draw.Rect.Dx()]
	}
	return &fx
}

// Render the effect at time t.
func (fx *fx) Render(t float64) image.Image {
	const (
		DecimalPointLog = 16
		DecimalMul      = 1 << DecimalPointLog
	)
	// tt is our reverse zoom as 16.16 fixed point
	tt := int(t * t * 12 * DecimalMul)

	xMask := (1 << fx.logW) - 1
	yMask := (1 << fx.logH) - 1

	// Center of zoom (screen space)
	centerX, centerY := renderWidth/2, renderHeight/2
	// Store the reverse transformation for the center of screen.
	x0 := -tt * centerX
	y0 := -tt * centerY

	// Center on texture (texture space)
	texCenterX, texCenterY := 173, 106
	x0 += texCenterX * DecimalMul
	y0 += texCenterY * DecimalMul

	for y, line := range fx.lines {
		srcY := ((y0 + y*tt) >> DecimalPointLog) & yMask
		// Pre-shift, so srcY is offset for x=0 at our line.
		srcY <<= fx.logW
		for x := range line {
			srcX := ((x0 + x*tt) >> DecimalPointLog) & xMask
			line[x] = fx.img.Pix[srcX+srcY]
		}
	}
	return fx.draw
}
