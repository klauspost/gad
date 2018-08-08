package main

import (
	"bufio"
	"bytes"
	"fmt"
	"image"
	_ "image/png"
	"math"
	"math/bits"
	"strings"

	_ "github.com/klauspost/gad/dentro/data" // Load data.
	"github.com/klauspost/gad/dentro/screen"
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
	fx := newFx("data/scene.obj", "data/flower.png", "data/light.png")
	//gfx.Run(func() { gfx.RunTimedMusic(fx, "music/music.mp3") })
	gfx.RunWriteToDisk(fx, 11, "./saved/frame-%05d.png")
}

type coord struct{ x, y, z float32 }

func (c *coord) scale(f float32) {
	c.x *= f
	c.y *= f
	c.z *= f
}

var texts = [][2]string{
	{
		`--- >> Welcome to  Go After Dark  << ---`,
		`.oOoOoOoOoOo. .oO0Oo.  .oO0OoO0OouoO0Oo.`,
	},
	{
		`A new web series kicking it oldschool   `,
		`with some 90s demoscene effects.    \o/ `,
	},
	{
		`We will look at basic realtime effects  `,
		`and explain everything involved.        `,
	},
	{
		`So you can join the Go After Dark club  `,
		`and code awesome looking effects.       `,
	},
	{
		`Check out the links below to watch      `,
		`the first episode.                      `,
	},
	{
		`I have material ready for at least 10   `,
		`episodes, if there is interest for this.`,
	},
	{
		`>> Credits <<                           `,
		`Coding: sh0dan/VoxPod, aka Klaus Post   `,
	},
	{
		`Music: "Boys can't Fly" by CyberSDF     `,
		`3D Model: "Rubber Duck" by mStuff       `,
	},
	{
		`Tunnel Texture by Tomasz Grabowiecki    `,
		`DOS font by int10h.org                  `,
	},
	{
		`Pure software rendering in Go, compiled `,
		`to Web Assembly by Go 1.11 (beta).      `,
	},
	{
		`There is links to all source code to    `,
		`build and run this demo.                `,
	},
}

type fx struct {
	draw             *image.Gray
	logW, logH       uint
	lines            [][]byte
	dots             []coord
	mipmaps          []*image.Gray
	img              *image.Gray
	tunnelTex        *image.Gray
	lookup           [][]uint32
	tunLogW, tunLogH uint
	text             *screen.Fx
	atText           int
	lastT            float64
}

func newFx(scene, particle, light string) *fx {
	var fx fx

	// Create our draw buffer
	fx.draw = image.NewGray(image.Rect(0, 0, renderWidth, renderHeight))

	// Store each line as a slice in a slice.
	w, h := fx.draw.Rect.Dx(), fx.draw.Rect.Dy()
	fx.lines = make([][]byte, h)
	for y := range fx.lines {
		fx.lines[y] = fx.draw.Pix[y*fx.draw.Stride : y*fx.draw.Stride+w]
	}
	fx.dots = make([]coord, 0)
	b, err := gfx.Load(scene)
	if err != nil {
		panic(err)
	}

	scanner := bufio.NewScanner(bytes.NewBuffer(b))
	for scanner.Scan() {
		t := scanner.Text()
		if !strings.HasPrefix(t, "v ") {
			continue
		}
		var c coord
		n, err := fmt.Sscanf(t, "v %f %f %f", &c.x, &c.y, &c.z)
		if err != nil {
			panic(err)
		}
		if n != 3 {
			panic("not 3")
		}
		c.y *= -1
		c.y += 1.8
		c.scale(150)
		fx.dots = append(fx.dots, c)
	}
	// Load picture
	img, err := gfx.LoadPalPicture(particle)
	if err != nil {
		panic(err)
	}
	fx.img = gfx.ToGray(img)

	// Calculate the with in log(2)
	fx.logW = uint(bits.Len32(uint32(img.Rect.Dx()))) - 1
	fx.logH = uint(bits.Len32(uint32(img.Rect.Dy()))) - 1

	// Ensure that the width and height are actually powers of two
	if img.Rect.Dx() != 1<<fx.logW {
		panic("image " + particle + " width is not power of two.")
	}
	if img.Rect.Dy() != 1<<fx.logH {
		panic("image " + particle + " width is not power of two.")
	}
	if img.Rect.Dx() != img.Rect.Dy() {
		panic("image " + particle + " width is same as height.")
	}

	fx.mipmaps = make([]*image.Gray, fx.logW+1)
	fx.mipmaps[fx.logW] = fx.img
	prev := fx.img
	for i := fx.logW - 1; i > 0; i-- {
		img := image.NewGray(image.Rect(0, 0, prev.Rect.Dx()/2, prev.Rect.Dy()/2))
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
	fx.img, err = gfx.LoadGreyPicture(light)
	if err != nil {
		panic(err)
	}
	// Calculate the with in log(2)
	fx.logW = uint(bits.Len32(uint32(fx.img.Rect.Dx()))) - 1
	fx.logH = uint(bits.Len32(uint32(fx.img.Rect.Dy()))) - 1

	// Store each line as a slice in a slice.
	fx.lookup = make([][]uint32, h)
	for y := range fx.lines {
		lu := make([]uint32, w)
		for x := range lu {
			// Number of bits to represent width/height of texture.
			const prec = 1 << 16
			var angle, distance int
			// Number of distance repetitions.
			const ratio = 16
			distance = int(ratio*prec/
				math.Sqrt(float64(x-w/2)*float64(x-w/2)+float64(y-h/2)*float64(y-h/2))) % prec
			angle = (int)(prec*math.Atan2(float64(y-h/2), float64(x-w/2))/math.Pi) % prec
			// Store distance as lower 16 bits, angle as upper
			lu[x] = uint32(distance) | uint32(angle<<16)
		}
		fx.lookup[y] = lu
	}
	fx.tunnelTex, err = gfx.LoadGreyPicture("data/wildtextures-african-grey.png")
	if err != nil {
		panic(err)
	}
	// Calculate the with in log(2)
	fx.tunLogW = uint(bits.Len32(uint32(fx.tunnelTex.Rect.Dx()))) - 1
	fx.tunLogH = uint(bits.Len32(uint32(fx.tunnelTex.Rect.Dy()))) - 1

	fx.text = screen.NewFx("data/dosfont.png", fx.textArea())
	fx.text.DrawText(strings.Join(texts[0][:], "\n"), 0, 0)
	return &fx
}

func (fx *fx) textArea() *image.Gray {
	//img := image.NewGray(image.Rect(0,0,renderWidth, 32))
	border := 8
	yoff := renderHeight - 32 - border
	img := image.Gray{
		Pix:    fx.draw.Pix[border+yoff*fx.draw.Stride:],
		Stride: fx.draw.Stride,
		Rect:   image.Rect(0, 0, renderWidth/2, renderHeight-yoff-border),
	}
	return &img
}

// Render the effect at time t.
func (fx *fx) Render(t float64) image.Image {
	const (
		fpX, fpY         = 7, 8
		fpXL, fpYL       = 8, 8
		tunLogW, tunLogH = 9, 8
		logW, logH       = 8, 8
		dmX, dmY         = 1 << fpX, 1 << fpY
		xMask            = 1<<tunLogW - 1
		yMask            = 1<<tunLogH - 1
		xMaskL           = 1<<logW - 1
		yMaskL           = 1<<logH - 1
	)
	// Shift in texture space (scrolls)
	shiftX := int(-t * 2 * 256 * float64(dmX))
	shiftY := int(t * 256 * float64(dmY))
	shiftXL := int(2 * -t * 256 * float64(xMaskL+1))
	shiftYL := int(8 * -t * 256 * float64(yMaskL+1))

	seed := uint32(t * math.MaxUint32)
	rng15i := func() int {
		seed = 214013*seed + 2531011
		return int(seed>>16) & 63
	}
	for y, line := range fx.lines {
		lu := fx.lookup[y]
		for x := range line {
			v := int(lu[x])
			srcX := ((rng15i() + shiftX + (v & 0xffff)) >> fpX) & xMask
			srcY := ((rng15i() + shiftY + (v >> 16)) >> fpY) & yMask
			srcXL := ((shiftXL + (v & 0xffff)) >> fpXL) & xMaskL
			srcYL := ((shiftYL + (v >> 16)) >> fpYL) & yMaskL
			line[x] = uint8((uint32(fx.tunnelTex.Pix[srcX+(srcY<<tunLogW)]) *
				(uint32(fx.img.Pix[srcXL+(srcYL<<logH)]))) >> 8)
		}
	}
	const (
		halfWidth  = renderWidth * 0.5
		halfHeight = renderHeight * 0.5
	)

	// Offset z on all points over time, effectively moving the "camera" forward.
	zoff := 400 + float32(math.Sin(t*math.Pi*4)*300)
	// Create rotation matrix
	rot := rotateFn(-math.Pi/2+0.075*math.Cos(t*math.Pi*16), 0.3+(1-t)*math.Pi*2, 0)
	zMul := float32(t * 0.5 * 200)
	if t > 0.5 {
		zMul -= float32((t - 0.5) * 200 * 255 * 16)
	}
	for _, d := range fx.dots {
		rot(&d)
		z := d.z + zoff
		if z <= 0 {
			continue
		}
		invZ := 1.0 / z
		x := 200 * d.x * invZ
		y := 200 * d.y * invZ
		x += halfWidth
		y += halfHeight

		zsize := 30 * 255 * 16 * invZ
		fx.drawParticleMip(int32(x*256), int32(y*256), int32(zsize))
	}
	if fx.lastT > t {
		fx.atText++
		fx.text.ClearScreen()
		txt := texts[fx.atText%len(texts)][:]
		fx.text.DrawText(strings.Join(txt, "\n"), 0, 0)
	}
	fx.lastT = t
	fx.text.Render(math.Min(1, t*2))
	return fx.draw
}

// rotateFn returns a function that rotates around (0,0,0).
// Supply angles in radians.
func rotateFn(xAn, yAn, zAn float64) func(c *coord) {
	var (
		s1 = float32(math.Sin(zAn))
		s2 = float32(math.Sin(xAn))
		s3 = float32(math.Sin(yAn))
		c1 = float32(math.Cos(zAn))
		c2 = float32(math.Cos(xAn))
		c3 = float32(math.Cos(yAn))

		zero  = c1*c3 + s1*s2*s3
		one   = c2 * s3
		two   = -c3*s1 + c1*s2*s3
		three = c2 * s1
		four  = -s2
		five  = c2 * c1
		six   = -c1*s3 + c3*s1*s2
		seven = c2 * c3
		eight = s1*s3 + c1*c3*s2
	)
	return func(c *coord) {
		c.x, c.y, c.z = c.x*zero+c.y*one+c.z*two, c.x*three+c.y*four+c.z*five, c.x*six+c.y*seven+c.z*eight
	}
}

// drawParticle will draw a particle centered at x,y with radius r.
// Input is assumed to be 24.8
func (fx *fx) drawParticleMip(x, y, r int32) {
	m := fx.calcMapping(x, y, r, -1)
	if m.mipSize != 0 && m.mip == nil {
		fx.drawPoint(&m)
		return
	}
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

type mapping struct {
	startX, endX, startY, endY int32
	u0, u1, v0, v1             uint32
	vStep, uStep               uint32
	// Texture coordinates as 16.16 fixed point.
	mipSize uint32
	mip     *image.Gray
}

func (fx *fx) calcMapping(x, y, r, mip int32) mapping {
	var m mapping
	// Quick discard
	if x+r < 0 || x-r > (renderWidth*256) || y+r < 0 || y-r > (renderHeight*256) {
		return m
	}
	// For very small radius we simply draw a point
	if r <= 256 {
		m.startX, m.endX = (x+r)>>8, (x+r)>>8+1
		m.startY, m.endY = (y+r)>>8, (y+r)>>8+1
		if m.startX >= renderWidth || m.startX < 0 || m.startY >= renderHeight || m.startY < 0 {
			return mapping{}
		}
		m.u1 = uint32(x-r) & 0xff
		m.u0 = 256 - m.u1
		m.v1 = uint32(y-r) & 0xff
		m.v0 = 256 - m.v1
		m.u0 = (m.u0 * uint32(r)) >> 7
		m.u1 = (m.u1 * uint32(r)) >> 7
		m.v0 = (m.v0 * uint32(r)) >> 7
		m.v1 = (m.v1 * uint32(r)) >> 7
		m.mipSize = 1
		// leave mip nil
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
	return m
}

type fp8x24 uint32
type fp16x16 uint32

func (f fp8x24) String() string {
	return fmt.Sprintf("(%d,0x%x)", uint32(f)>>8, uint8(f))
}
func (f fp16x16) String() string {
	return fmt.Sprintf("(%d,0x%x)", uint32(f)>>16, uint16(f))
}

func init() {
	var palette [256]uint32
	rC, gC, bC := 180, 180, 255
	flip := 192
	flipScale := int(256.0 * (256.0 / float64(flip)))
	flipScale2 := int(256.0 * (256.0 / float64(255-flip)))
	for i := range palette[:] {
		var r, g, b int
		if i < flip {
			r = (i*rC*flipScale + 128) >> 16
			g = (i*gC*flipScale + 128) >> 16
			b = (i*bC*flipScale + 128) >> 16
		} else {
			r = rC
			g = gC
			b = bC
			r += ((i - flip) * (255 - rC) * flipScale2) >> 16
			g += ((i - flip) * (255 - gC) * flipScale2) >> 16
			b += ((i - flip) * (255 - bC) * flipScale2) >> 16
		}
		palette[i] = uint32(r) | (uint32(g) << 8) | (uint32(b) << 16)
	}
	gfx.InitGreyPalette(palette)
}
