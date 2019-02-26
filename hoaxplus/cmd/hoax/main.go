package main

import (
	"image"
	"time"

	"github.com/klauspost/gfx"
	"github.com/klauspost/hoaxplus/00-intro"
	"github.com/klauspost/hoaxplus/01-title"
)

var (
	renderWidth  = 624
	renderHeight = 240
)

func main() {
	// Create our draw buffer
	screen := image.NewGray(image.Rect(0, 0, renderWidth, renderHeight))
	fullColor := image.NewRGBA(image.Rect(0, 0, renderWidth, renderHeight))
	gfx.SetRenderSize(renderWidth, renderHeight)
	gfx.Fullscreen(false)

	fx := intro.NewIntro(screen)
	if true {
		fx = title.NewTitle(screen, fullColor)
	}
	gfx.Run(func() { gfx.RunTimedDur(fx, 5*time.Second) })
}
