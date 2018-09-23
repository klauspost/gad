package main

import (
	"image"
	_ "image/png"
	"math"
	"math/bits"

	_ "github.com/klauspost/gad/ep04/data" // Load data.
	"github.com/klauspost/gfx"
)

const (
	renderWidth  = 640
	renderHeight = 360
)

// Generates binary data.
// To install go-bindata, do: go get -u github.com/jteeuwen/go-bindata/...
//
//go:generate go-bindata -ignore=\.go\z -pkg=data -o ./data/data.go ./data/...

func main() {
	fx := newTunnel("data/wildtextures-african-inspir.png")
	//gfx.RunWriteToDisk(fx, 1, "./saved/tunnel-%05d.png")
	gfx.Run(func() { gfx.RunTimed(fx) })
}

type tunnel struct {
	img        *image.Paletted
	draw       *image.Paletted
	logW, logH uint
	lines      [][]byte
	lookup     [][]uint32
}

func newTunnel(file string) *tunnel {
	var fx tunnel

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
		panic("image " + file + " width is not power of two.")
	}

	// Create our draw buffer
	fx.draw = image.NewPaletted(image.Rect(0, 0, renderWidth, renderHeight), img.Palette)

	// Store each line as a slice in a slice.
	w, h := fx.draw.Rect.Dx(), fx.draw.Rect.Dy()
	fx.lines = make([][]byte, h)

	fx.lookup = make([][]uint32, h)
	for y := range fx.lines {
		fx.lines[y] = fx.draw.Pix[y*fx.draw.Stride : y*fx.draw.Stride+fx.draw.Rect.Dx()]
		lu := make([]uint32, w)
		for x := range lu {
			// Number of distance repetitions.
			const ratio = 8
			// texture coordinates for screen space coordinates (x,y)
			centerX, centerY := float64(x-w/2), float64(y-h/2)

			// distance is u coordinate, 0 -> coordPrec
			distance := int(ratio*coordPrec/math.Sqrt(centerX*centerX+centerY*centerY)) % coordPrec
			// angle is v coordinate 0 -> coordPrec
			angle := (int)(0.5*coordPrec*math.Atan2(centerY, centerX)/math.Pi) % coordPrec

			// Store distance as lower 16 bits, angle as upper
			lu[x] = uint32(distance) | uint32(angle<<16)
		}
		fx.lookup[y] = lu
	}
	return &fx
}

// Number of bits to represent width/height of texture.
const coordBits = 16
const coordPrec = 1 << coordBits

// Render the effect at time t.
func (fx *tunnel) Render(t float64) image.Image {
	// fixed point precision
	fpX, fpY := coordBits-fx.logW, coordBits-fx.logH
	dmX, dmY := 1<<fpX, 1<<fpY

	uMask := 1<<fx.logW - 1
	vMask := 1<<fx.logH - 1

	// Shift in texture space (scrolls)
	offsetU := int(math.Sin(t*math.Pi*2+6*math.Pi/4) * 80 * float64(dmX))
	offsetV := int(t * 256 * float64(dmY))
	logY := fx.logW
	for y, line := range fx.lines {
		lu := fx.lookup[y]
		for x := range line {
			pos := int(lu[x])
			u := ((offsetU + (pos & 0xffff)) >> fpX) & uMask
			v := ((offsetV + (pos >> 16)) >> fpY) & vMask
			line[x] = fx.img.Pix[u+(v<<logY)]
		}
	}
	return fx.draw
}
