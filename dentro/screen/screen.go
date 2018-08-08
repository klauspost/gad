package screen

import (
	"image"

	"github.com/klauspost/gfx"
)

type Fx struct {
	// main and mipmaps
	screen           [][]byte
	font             [128][16]byte
	fontLU           [256][8]byte
	draw             *image.Gray
	lines            [][]byte
	cpl              int
	fgColor, bgColor uint8
}

func NewFx(file string, dst *image.Gray) *Fx {
	var fx Fx
	fx.draw = dst

	// Load picture
	img, err := gfx.LoadPalPicture(file)
	if err != nil {
		panic(err)
	}
	if img.Rect.Dx() != 512 || img.Rect.Dy() != 32 {
		panic("image " + file + " size is not 512x32")
	}
	for i := 0; i < 64; i++ {
		for y := 0; y < 16; y++ {
			var v, v2 uint8
			for x := 0; x < 8; x++ {
				v <<= 1
				v2 <<= 1
				if img.ColorIndexAt(i*8+x, y) != 0 {
					v |= 1
				}
				if img.ColorIndexAt(i*8+x, y+16) != 0 {
					v2 |= 1
				}
			}
			fx.font[i][y] = v
			fx.font[i+64][y] = v2
		}
	}
	fx.screen = make([][]byte, dst.Rect.Dy()/16)
	fx.ClearScreen()
	fx.SetColor(224, 0)
	// Store each line as a slice in a slice.
	w, h := fx.draw.Rect.Dx(), fx.draw.Rect.Dy()
	fx.lines = make([][]byte, h)
	for y := range fx.lines {
		fx.lines[y] = fx.draw.Pix[y*fx.draw.Stride : y*fx.draw.Stride+w]
	}

	return &fx
}

func (fx *Fx) SetDraw(img *image.Gray) {
	fx.draw = img
}

func (fx *Fx) SetColor(fg, bg uint8) {
	fx.fgColor, fx.bgColor = fg, bg
	for i := 0; i < 256; i++ {
		v := uint8(i)
		for j := 0; j < 8; j++ {
			if v>>7 == 1 {
				fx.fontLU[i][j] = fg
			} else {
				fx.fontLU[i][j] = bg
			}
			v <<= 1
		}
	}
}

func (fx *Fx) DrawText(s string, x, y int) {
	line := fx.line(y)
	for _, ch := range s {
		if y >= len(fx.screen) {
			return
		}
		switch ch {
		case '\r':
			x = 0
		case '\n':
			x = 0
			y++
			line = fx.line(y)
		case '\t':
			x = (x + 8) & 0x7
		default:
			if x >= fx.cpl {
				x = 0
				y++
				line = fx.line(y)
			}
			line[x] = uint8(ch & 127)
			x++
		}
	}
}

func (fx *Fx) line(y int) []byte {
	if y >= len(fx.screen) {
		return nil
	}
	line := fx.screen[y]
	if len(line) == 0 {
		line = line[:fx.cpl]
		for i := range line {
			// Set to spaces
			line[i] = 32
		}
		fx.screen[y] = line
	}
	return line
}

func (fx *Fx) ClearScreen() {
	fx.cpl = fx.draw.Rect.Dx() / 8
	for i := range fx.screen {
		if fx.screen[i] == nil {
			fx.screen[i] = make([]byte, 0, fx.cpl)
		}
		fx.screen[i] = fx.screen[i][:0]
	}
}

// Render the effect at time t.
func (fx *Fx) Render(t float64) image.Image {
	stopAt := int(t * float64(len(fx.screen)*fx.cpl))
	stopAtX := stopAt % fx.cpl
	stopAtY := stopAt / fx.cpl
	addChar := uint8(t * 1000)
	for i, line := range fx.screen {
		if len(line) == 0 {
			// Fast path for empty lines.
			dst := fx.draw.Pix[i*16*fx.draw.Stride : (i+1)*16*fx.draw.Stride]
			c := fx.bgColor
			for j := range dst {
				dst[j] = c
			}
			continue
		}
		dst := fx.draw.Pix[i*16*fx.draw.Stride : (i+1)*16*fx.draw.Stride]
		add := uint8(0)
		if i > stopAtY {
			add = addChar
		}
		for j, c := range line {
			dst := dst[j*8:]
			if i == stopAtY && j >= stopAtX {
				add = addChar
				if j-stopAtX < 1 {
					add = 0
					c = '_'
				}
			}
			if c == 32 {
				for y := range fx.font[32&127][:] {
					// Copy all pixels
					p := y * fx.draw.Stride
					v := int32(fx.bgColor) * 2
					for i := 0; i < 8; i++ {
						org := int32(dst[p+i])
						dst[p+i] = clamp8int32(v - 128 + org)
					}
				}
				continue
			}
			for y, v := range fx.font[(c+add)&127][:] {
				// Copy all pixels
				p := y * fx.draw.Stride
				for i, v := range fx.fontLU[v][:] {
					org := int32(dst[p+i])
					dst[p+i] = clamp8int32(int32(v)*2 - 128 + org)
				}
			}
		}
	}
	return fx.draw
}

func clamp8int32(v int32) uint8 {
	if v <= 0 {
		return 0
	}
	if v >= 255 {
		return 255
	}
	return uint8(v)
}
