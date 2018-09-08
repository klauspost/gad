package main

import (
	"image"
	_ "image/png"
	"math"
	"math/bits"

	_ "github.com/klauspost/gad/ep03/data" // Load data.
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
	fx := newRotoZoom("data/ashleymcnamara-pride_256.png")
	gfx.Run(func() { gfx.RunTimed(fx) })
}

type RotoZoomer struct {
	img        *image.Paletted
	draw       *image.Paletted
	logW, logH uint
	lines      [][]byte
}

func newRotoZoom(file string) *RotoZoomer {
	var rz RotoZoomer

	// Load picture
	img, err := gfx.LoadPalPicture(file)
	if err != nil {
		panic(err)
	}
	rz.img = img

	// Calculate the with in log(2)
	rz.logW = uint(bits.Len32(uint32(img.Rect.Dx()))) - 1
	rz.logH = uint(bits.Len32(uint32(img.Rect.Dy()))) - 1

	// Ensure that the width and height are actually powers of two
	if img.Rect.Dx() != 1<<rz.logW {
		panic("image " + file + " width is not power of two.")
	}
	if img.Rect.Dy() != 1<<rz.logH {
		panic("image " + file + " width is not power of two.")
	}

	// Create our draw buffer
	rz.draw = image.NewPaletted(image.Rect(0, 0, renderWidth, renderHeight), img.Palette)

	// Store each line as a slice in a slice.
	rz.lines = make([][]byte, rz.draw.Rect.Dy())
	for y := range rz.lines {
		rz.lines[y] = rz.draw.Pix[y*rz.draw.Stride : y*rz.draw.Stride+rz.draw.Rect.Dx()]
	}
	return &rz
}

// Render the effect at time t.
func (rz *RotoZoomer) Render(t float64) image.Image {
	//return rz.RenderTextureSpace(t)
	const (
		DecimalPointLog = 16
		DecimalMul      = 1 << DecimalPointLog
	)

	// Angle of rotation and scale
	ang := (1.0 - t) * math.Pi * 2
	scale := math.Abs(3 * math.Sin(ang*2))

	// Calculate texture slopes.
	uEveryX := int(math.Cos(ang) * scale * DecimalMul)
	vEveryX := int(math.Sin(ang) * scale * DecimalMul)
	uEveryY := int(-math.Sin(ang) * scale * DecimalMul)
	vEveryY := int(math.Cos(ang) * scale * DecimalMul)

	// Center of zoom (screen space)
	centerX, centerY := renderWidth/2, renderHeight/2
	u0 := -uEveryY*centerY - uEveryX*centerX
	v0 := -vEveryY*centerY - vEveryX*centerX

	// Center on texture (texture space)
	texCenterU, texCenterV := 1<<(rz.logW-1), 1<<(rz.logH-1)
	u0 += texCenterU * DecimalMul
	v0 += texCenterV * DecimalMul

	// Texture helpers
	logY := rz.logW
	uMask := 1<<rz.logW - 1
	vMask := 1<<rz.logH - 1
	for _, line := range rz.lines {
		u := u0
		v := v0
		for x := range line {
			srcX := (u >> DecimalPointLog) & uMask
			srcY := (v >> DecimalPointLog) & vMask
			line[x] = rz.img.Pix[srcX+(srcY<<logY)]
			v += vEveryX
			u += uEveryX
		}
		v0 += vEveryY
		u0 += uEveryY
	}
	return rz.draw
}

// RenderTextureSpace the effect in texture space at time t.
func (rz *RotoZoomer) RenderTextureSpace(t float64) image.Image {
	const (
		DecimalPointLog = 16
		DecimalMul      = 1 << DecimalPointLog
	)
	logY := rz.logW
	uMask := 1<<rz.logW - 1
	vMask := 1<<rz.logH - 1

	// Angle of rotation and scale
	ang := (1.0 - t) * math.Pi * 2
	scale := math.Abs(3 * math.Sin(ang))

	uEveryX := int(math.Cos(ang) * scale * DecimalMul)
	vEveryX := int(math.Sin(ang) * scale * DecimalMul)
	uEveryY := int(-math.Sin(ang) * scale * DecimalMul)
	vEveryY := int(math.Cos(ang) * scale * DecimalMul)

	// Center of zoom (screen space)
	centerX, centerY := renderWidth/2, renderHeight/2
	u0 := -uEveryY*centerY - uEveryX*centerX
	v0 := -vEveryY*centerY - vEveryX*centerX

	// Center on texture (texture space)
	texCenterU, texCenterV := 1<<(rz.logW-1), 1<<(rz.logH-1)
	u0 += texCenterU * DecimalMul
	v0 += texCenterV * DecimalMul
	tex := make([]byte, len(rz.img.Pix))
	copy(tex, rz.img.Pix)
	for y, line := range rz.lines {
		u := v0
		v := u0
		for x := range line {
			if x != 0 && y != 0 && x != len(line)-1 && y != len(rz.lines)-1 {
				v += vEveryX
				u += uEveryX
				continue
			}
			srcX := (u >> DecimalPointLog) & uMask
			srcY := (v >> DecimalPointLog) & vMask
			tex[srcX+(srcY<<logY)] = 26
			v += vEveryX
			u += uEveryX
		}
		v0 += vEveryY
		u0 += uEveryY
	}
	for y, line := range rz.lines {
		for x := range line {
			srcX := x & uMask
			srcY := y & vMask
			line[x] = tex[srcX+(srcY<<logY)]
		}
	}

	return rz.draw
}
