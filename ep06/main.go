package main

import (
	"fmt"
	"image"
	"image/color"
	_ "image/png"
	"math"
	"math/bits"

	_ "github.com/klauspost/gad/ep06/data" // Load data.
	"github.com/klauspost/gfx"
	"golang.org/x/image/draw"
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
	//rz := newFx("./data/pride_circle_grey.png")
	//rz := newFx("./data/light.png")
	gfx.InitShadedPalette(180, color.RGBA{R: 158, G: 240, B: 158})

	fx := newFx("data/flower.png")
	gfx.Run(func() { gfx.RunTimed(fx) })
	//gfx.RunWriteToDisk(fx, 1, "./saved/particle-%05d.png")

}

type fx struct {
	// main and mipmaps
	img        *image.Gray
	mipmaps    []*image.Gray
	draw       *image.Gray
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
	// Ensure we are grayscale.
	grey := image.NewGray(img.Rect)
	draw.Draw(grey, grey.Rect, img, image.Pt(0, 0), draw.Src)
	fx.img = grey

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
	if img.Rect.Dx() != img.Rect.Dy() {
		panic("image " + file + " width is same as height.")
	}

	// Create our draw buffer
	fx.draw = image.NewGray(image.Rect(0, 0, renderWidth, renderHeight))

	// Store each line as a slice in a slice.
	w, h := fx.draw.Rect.Dx(), fx.draw.Rect.Dy()
	fx.lines = make([][]byte, h)
	for y := range fx.lines {
		fx.lines[y] = fx.draw.Pix[y*fx.draw.Stride : y*fx.draw.Stride+w]
	}
	// Calculate mipmaps. Mipmaps must be square.
	// Size of a mipmap in pixels  =  1 << mipLevel
	fx.mipmaps = make([]*image.Gray, fx.logW+1)
	fx.mipmaps[fx.logW] = grey
	prev := grey
	for i := int(fx.logW - 1); i >= 0; i-- {
		img := image.NewGray(image.Rect(0, 0, prev.Rect.Dx()/2, prev.Rect.Dy()/2))
		fmt.Println(i, img.Rect)
		fx.mipmaps[i] = img
		for y := 0; y < img.Rect.Dy(); y++ {
			src0, src1 := prev.Pix[y*2*prev.Stride:], prev.Pix[(y*2+1)*prev.Stride:]
			dst := img.Pix[y*img.Stride : (y+1)*img.Stride]
			for x := range dst {
				// Average 4 pixels.
				dst[x] = uint8((uint(src0[x*2]) + uint(src0[x*2+1]) + uint(src1[x*2]) + uint(src0[x*2+1]) + 2) >> 2)
			}
		}
		prev = img
	}
	return &fx
}

// Render the effect at time t.
func (fx *fx) Render(t float64) image.Image {
	for i := range fx.draw.Pix {
		fx.draw.Pix[i] = 0
	}

	const (
		halfWidth  = renderWidth * 0.5
		halfHeight = renderHeight * 0.5
	)
	drawfunc := fx.drawSpriteFast
	//drawfunc := fx.drawSpriteMip
	//drawfunc := fx.drawSpriteNice
	//drawfunc := fx.drawSpriteGo

	if false {
		for i := int32(0); i < 1024; i++ {
			x := 10 + 20*(i%32-16)
			y := 20 * (i/32 - 16)
			r := 7 * math.Abs(math.Sin(t+t*float64(x)/5)+math.Cos(t+t*float64(y)/3))
			drawfunc(x*256+halfWidth*256, y*256+halfHeight*256, int32(256*r))
		}
		return fx.draw
	}
	drawfunc(
		int32(halfWidth*256+256*50*math.Sin(t*math.Pi*2*4)),
		int32(halfHeight*256+256*50*math.Cos(t*math.Pi*2*4)),
		//int32(256*50),
		int32(256*130*math.Abs(math.Sin(t*math.Pi))),
	)
	drawfunc(
		int32(150*256+256*t*100),
		int32(55*256),
		int32(256*50),
	)
	drawfunc(
		int32(55*256),
		int32(256*50+256*t*100),
		int32(256*50),
	)

	drawfunc(
		int32(halfWidth*1.5*256+255),
		int32(halfHeight*256+255),
		int32(256*150*(1-t)),
	)
	drawfunc(
		int32(55*256),
		int32(255*256),
		int32(256*15*math.Abs(math.Cos(t*math.Pi*10))),
	)
	drawfunc(
		int32(100*256+int32(t*200*256)),
		int32(255*256),
		int32(256*2*math.Abs(math.Cos(t*math.Pi*2))),
	)
	return fx.draw
}

// drawSpriteFast will draw a particle centered at x,y with radius r.
// The image is drawn with nearest neighbor scaling, but subpixel position and radius.
// Input is assumed to be 24.8
func (fx *fx) drawSpriteFast(x, y, r int32) {
	m := fx.calcMapping(x, y, r, int32(len(fx.mipmaps)-1))
	if m.mip == nil {
		return
	}

	v := m.v0
	for y := m.startY; y < m.endY; y++ {
		// First destination pixel
		dst := fx.draw.Pix[int(y)*fx.draw.Stride:]
		// Line in mipmap to read from.
		mipLine := m.mip.Pix[(v>>16)*uint32(m.mip.Stride):]
		// Reset u
		u := m.u0
		for x := m.startX; x < m.endX; x++ {
			// Read from u
			xPos := u >> 16
			// Add input to output.
			dst[x] = clamp8uint32(uint32(mipLine[xPos]) + uint32(dst[x]))
			// Offset u for every pixel
			u += m.uStep
		}
		// Offset v for every line
		v += m.vStep
	}
}

// drawSpriteMip will draw a particle centered at x,y with radius r.
// The image is drawn with nearest neighbor scaling, but choosing from a mipmap
// and with subpixel position and radius.
// Input is assumed to be 24.8
func (fx *fx) drawSpriteMip(x, y, r int32) {
	m := fx.calcMapping(x, y, r, -1)
	if m.mip == nil {
		return
	}

	// This is similar to drawSpriteFast
	v := m.v0
	for y := m.startY; y < m.endY; y++ {
		dst := fx.draw.Pix[int(y)*fx.draw.Stride:]
		mipLine := m.mip.Pix[(v>>16)*uint32(m.mip.Stride):]
		u := m.u0
		for x := m.startX; x < m.endX; x++ {
			xPos := u >> 16
			dst[x] = clamp8uint32(uint32(mipLine[xPos]) + uint32(dst[x]))
			u += m.uStep
		}
		v += m.vStep
	}
}

// drawSpriteNice will draw a particle centered at x,y with radius r.
// The input is selected from the appropriate mipmap and bilinear interpolation is used.
// The image position and size have subpixel precision.
// Input is assumed to be 24.8
func (fx *fx) drawSpriteNice(x, y, r int32) {
	m := fx.calcMapping(x, y, r, -1)
	if m.mip == nil {
		return
	}
	// Mipmap must be at least 2x2.
	v := m.v0
	for y := m.startY; y < m.endY; y++ {
		dst := fx.draw.Pix[int(y)*fx.draw.Stride:]

		// Input line above and below the current desired input.
		mipLine0 := m.mip.Pix[(v>>16)*uint32(m.mip.Stride):]
		mipLine1 := mipLine0

		// Calculate weight for lines above/below desired input.
		vf1 := (v & 0xffff) >> 8
		vf0 := 256 - vf1
		if (v + 65536) < m.mipSize {
			// Set mipline 1 to next line unless last line.
			mipLine1 = mipLine0[m.mip.Stride:]
		}
		u := m.u0
		for x := m.startX; x < m.endX; x++ {
			// Calculate pixel offset before and after desired pixel.
			xPos0 := u >> 16
			xPos1 := xPos0
			if u+65536 < m.mipSize {
				xPos1++
			}
			// Calculate weights as fp24.8.
			uf1 := (u & 0xffff) >> 8
			uf0 := 256 - uf1
			// Using the calculated weights, calculate output pixel, scaled up 16 bits.
			pix := uint32(mipLine0[xPos0]) * uf0 * vf0
			pix += uint32(mipLine0[xPos1]) * uf1 * vf0
			pix += uint32(mipLine1[xPos0]) * uf0 * vf1
			pix += uint32(mipLine1[xPos1]) * uf1 * vf1

			// Add output to current pixel value.
			dst[x] = clamp8uint32((pix >> 16) + uint32(dst[x]))
			u += m.uStep
		}
		v += m.vStep
	}
}

type mapping struct {
	// Start and end coordinates in screen space.
	// This is where we will be drawing the pixels.
	// This is directly translatable to a screen coordinate.
	startX, endX, startY, endY int32

	// Start and end coordinates in 16.16 fixed point coordinates on the source texture.
	// One pixel on source is equal to 65536.
	u0, u1, v0, v1 uint32

	// Every pixel increment u and v by this when moving one pixel in screen space.
	vStep, uStep uint32

	// The size (width/height) of the chosen mip.
	mipSize uint32

	// The image to draw from.
	// If nil, do not draw anything.
	mip *image.Gray
}

// calcMapping will return a mapping for a sprite with radius r placed at (x,y)
// at the specified mip level.
func (fx *fx) calcMapping(x, y, r, mip int32) mapping {
	var m mapping
	// Quick discard
	if x+r < 0 || x-r > (renderWidth*256) || y+r < 0 || y-r > (renderHeight*256) {
		return m
	}
	// For very small radius we simply draw a point
	if r <= 128 {
		m.startX, m.endX = (x-r)>>8, (x-r)>>8+1
		m.startY, m.endY = (y-r)>>8, (y-r)>>8+1
		if m.startX >= renderWidth || m.startX < 0 || m.startY >= renderHeight || m.startY < 0 {
			return mapping{}
		}
		m.u1 = uint32(x-r) & 0xff
		m.u0 = 256 - m.u1
		m.v1 = uint32(y-r) & 0xff
		m.v0 = 256 - m.v1

		// Radius times 1x1 pixel value.
		rmip := (r * int32(fx.mipmaps[0].Pix[0])) >> 5
		m.u0 = (m.u0 * uint32(rmip)) >> 8
		m.u1 = (m.u1 * uint32(rmip)) >> 8
		m.v0 = (m.v0 * uint32(rmip)) >> 8
		m.v1 = (m.v1 * uint32(rmip)) >> 8
		m.mipSize = 1
		// leave mip nil
		fx.drawPoint(&m)
		return m
	}
	mipLevel := mip
	if mip < 0 {
		mipLevel = int32(bits.Len32(uint32(r >> 7)))
		if int(mipLevel) >= len(fx.mipmaps) {
			mipLevel = int32(len(fx.mipmaps)) - 1
		}
	}
	m.mip = fx.mipmaps[mipLevel]
	m.mipSize = uint32(1<<16) << uint(mipLevel)

	// Texture pixels per output pixel
	textureScale := float64(m.mip.Rect.Dx()) / (float64(r * 2 / 256))

	// Screen space start, rounded down
	m.startX, m.startY = (x-r)>>8, (y-r)>>8
	// Screen space, rounded up.
	m.endX, m.endY = (x+r+255)>>8, (y+r+255)>>8

	// Calculate rounded difference and convert to texture space.
	m.u0, m.v0 = 256-uint32((x-r)-(m.startX<<8)), 256-uint32((y-r)-(m.startY<<8))
	m.u0, m.v0 = uint32(textureScale*float64(m.u0*256)), uint32(textureScale*float64(m.v0)*256)

	// Calculate rounded difference and convert to texture space.
	m.u1, m.v1 = 256-uint32((m.endX<<8)-(x+r)), 256-uint32((m.endY<<8)-(y+r))
	m.u1, m.v1 = m.mipSize-uint32(textureScale*float64(m.u1)*256), m.mipSize-uint32(textureScale*float64(m.v1)*256)

	// Calculate step size per screen space pixel.
	m.uStep = uint32(float64(m.u1-m.u0) / float64(m.endX-m.startX-1))
	m.vStep = uint32(float64(m.v1-m.v0) / float64(m.endY-m.startY-1))

	// Clip
	if m.startX < 0 {
		m.u0 += m.uStep * uint32(-m.startX)
		m.startX = 0
	}
	if m.startY < 0 {
		m.v0 += m.vStep * uint32(-m.startY)
		m.startY = 0
	}
	if m.endX > renderWidth {
		// Not needed for most
		m.v1 -= uint32(m.endX-renderWidth) * m.uStep
		m.endX = renderWidth
	}
	if m.endY > renderHeight {
		// Not needed for most
		m.v1 -= uint32(m.endY-renderHeight) * m.vStep
		m.endY = renderHeight
	}
	// Final sanity to make sure we don't go over due to rounding.
	for m.u0+m.uStep*uint32(m.endX-m.startX) >= m.mipSize {
		m.endX--
	}
	for m.v0+m.vStep*uint32(m.endY-m.startY) >= m.mipSize {
		m.endY--
	}
	return m
}

// drawPoint will draw a single point between the pixel.
func (fx *fx) drawPoint(m *mapping) {
	dst := fx.draw.Pix[m.startX+m.startY*int32(fx.draw.Stride):]
	dst[0] = clamp8uint32((m.u0*m.v0)>>8 + uint32(dst[0]))
	if m.endX < renderWidth {
		dst[1] = clamp8uint32((m.u1*m.v0)>>8 + uint32(dst[1]))
	}
	if m.endY >= renderHeight {
		return
	}
	dst = dst[fx.draw.Stride:]
	dst[0] = clamp8uint32((m.u0*m.v1)>>8 + uint32(dst[0]))
	if m.endX < renderWidth {
		dst[1] = clamp8uint32((m.u1*m.v1)>>8 + uint32(dst[1]))
	}
}

func clamp8uint32(v uint32) uint8 {
	if v >= 255 {
		return 255
	}
	return uint8(v)
}

// drawParticle will draw a particle centered at x,y with radius r.
// The sprite
// Input is assumed to be 24.8
func (fx *fx) drawSpriteGo(x, y, r int32) {
	m := fx.calcMapping(x, y, r, -1)
	if m.startX == m.endX || m.startY == m.endY || m.mip == nil {
		return
	}
	draw.ApproxBiLinear.Scale(fx.draw, image.Rect(int(m.startX), int(m.startY), int(m.endX), int(m.endY)),
		image.NewUniform(color.White), image.Rect(int(m.u0>>16), int(m.v0>>16), int(m.u1>>16), int(m.v1>>16)), draw.Over, &draw.Options{
			SrcMask: grayToShallowAlpha(m.mip),
		})
}

// grayToShallowAlpha converts the grey image data to an alpha image.
// The image data is a shallow (non-copy) representation of the input pixels.
func grayToShallowAlpha(src *image.Gray) *image.Alpha {
	return &image.Alpha{
		Pix:    src.Pix,
		Stride: src.Stride,
		Rect:   src.Rect,
	}
}

type fp8x24 uint32
type fp16x16 uint32

func (f fp8x24) String() string {
	return fmt.Sprintf("(%d,0x%x)", uint32(f)>>8, uint8(f))
}
func (f fp16x16) String() string {
	return fmt.Sprintf("(%d,0x%x)", uint32(f)>>16, uint16(f))
}
