package main

// Generates binary data.
// To install go-bindata, do: go get -u github.com/jteeuwen/go-bindata/...
//
//go:generate go-bindata -ignore=\.go\z -pkg=data -o ./data/data.go ./data/...

import (
	"image"
	_ "image/png"
	"math"

	_ "./data" // Load data.
	"github.com/klauspost/gfx"
)

const (
	renderWidth  = 640
	renderHeight = 360
)

func main() {
	fx := newFx("data/GOPHER_MIC_DROP-256.png")
	gfx.Run(func() { gfx.RunTimed(fx) })
}

type fx struct {
	img   *image.Paletted
	draw  *image.Paletted
	lines [][]byte
}

func newFx(imageFile string) *fx {
	var fx fx

	// Load picture
	img, err := gfx.LoadPalPicture(imageFile)
	if err != nil {
		panic(err)
	}
	fx.img = img

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
	offsetX := int(t * 100)
	offsetY := offsetX

	// Set screen to index 0 color
	for i := range fx.draw.Pix {
		fx.draw.Pix[i] = 0
	}
	// Width and height of the source image
	w, h := fx.img.Rect.Dx(), fx.img.Rect.Dy()

	offsetX = int(math.Sin(t*math.Pi) * float64(renderWidth-w))
	offsetY = renderHeight - int(math.Abs(math.Sin(t*math.Pi*8)*100)) - h

	// Range over output height
	for y, dst := range fx.lines[offsetY : offsetY+h] {
		// Look up source line
		src := fx.img.Pix[y*fx.img.Stride : y*fx.img.Stride+w]
		// Adjust destination by offset.
		dst := dst[offsetX : offsetX+len(src)]

		//for x := range src {
		//	dst[x] = src[x]
		//}
		copy(dst, src)
	}
	return fx.draw
}
