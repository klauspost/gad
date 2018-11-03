package main

import (
	"image"
	"image/color"
	_ "image/png"
	"math"
	"math/rand"
	"time"

	_ "github.com/klauspost/gad/ep05/data" // Load data.
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
	gfx.InitShadedPalette(192, color.RGBA{R: 251, G: 246, B: 158})
	fx := newFx()
	//gfx.RunWriteToDisk(fx, 1, "./saved/dots-%05d.png")
	gfx.Run(func() { gfx.RunTimed(fx) })
}

type coord struct{ x, y, z float32 }

type fx struct {
	draw       *image.Gray
	logW, logH uint
	lines      [][]byte
	dots       []coord
}

func newFx() *fx {
	var fx fx

	// Create our draw buffer
	fx.draw = image.NewGray(image.Rect(0, 0, renderWidth, renderHeight))

	// Store each line as a slice in a slice.
	w, h := fx.draw.Rect.Dx(), fx.draw.Rect.Dy()
	fx.lines = make([][]byte, h)
	for y := range fx.lines {
		fx.lines[y] = fx.draw.Pix[y*fx.draw.Stride : y*fx.draw.Stride+w]
	}

	// Generate some dots in a cylinder along positive z axis.
	const dots = 50000
	fx.dots = make([]coord, dots)
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := range fx.dots {
		// Random angle
		angle := rng.Float64() * math.Pi * 2
		// Random radius.
		r := rng.Float64() * renderWidth * 3
		// Random depth
		z := rng.Float32() * 50
		x := math.Sin(angle)
		y := math.Cos(angle)

		fx.dots[i] = coord{x: float32(x * r), y: float32(y * r), z: z}
	}
	return &fx
}

// Render the effect at time t.
func (fx *fx) Render(t float64) image.Image {
	// Clear the screen or keep trails
	for i := range fx.draw.Pix {
		//fx.draw.Pix[i] = 0
		fx.draw.Pix[i] = fx.draw.Pix[i] >> 1
	}

	angleSin, angleCos := float32(1), float32(0)
	if true {
		// Rotate 0 -> 180 degrees when t goes 0 -> 1
		angleSin = float32(math.Sin(t * math.Pi))
		angleCos = float32(math.Cos(t * math.Pi))
	}

	// t2 goes from 0 -> 2 -> 0
	t2 := float32(math.Sin(t*math.Pi)) * 2
	// Offset z on all points over time, effectively moving the "camera" forward.
	zoff := 3 - 5*t2

	const (
		// We calculate output contribution using the z depth.
		// We use zMaxValue and subtract (z * zFalloff) for each dot.
		zMaxValue = 200
		zFalloff  = 20.0
		// maxZ is where the contribution becomes zero.
		// zMaxValue - zFalloff*z = 0
		maxZ = zMaxValue / zFalloff
	)

	// Draw all our dots
	for _, d := range fx.dots {
		z := d.z + zoff
		if z <= 0 || z >= maxZ {
			// behind the camera or no contribution, skip
			continue
		}
		// Scale x+y and perspective project
		x := (t2 * d.x) / z
		y := (t2 * d.y) / z

		// 2d rotate while still centered on 0,0
		x, y = x*angleCos-y*angleSin, x*angleSin+y*angleCos

		x += renderWidth / 2
		y += renderHeight / 2

		if y := int(y); y >= 0 && y < renderHeight {
			x := int(x)
			if x >= 0 && x < renderWidth {
				l := fx.lines[y]
				l[x] = clamp8(int(l[x]) + (zMaxValue - int(z*zFalloff)))
			}
		}
	}
	return fx.draw
}

func clamp8(v int) uint8 {
	if v >= 255 {
		return 255
	}
	if v <= 0 {
		return 0
	}
	return uint8(v)
}
