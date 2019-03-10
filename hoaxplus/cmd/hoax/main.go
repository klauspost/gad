package main

import (
	"image"
	"time"

	"github.com/klauspost/gad/hoaxplus/00-intro"
	"github.com/klauspost/gad/hoaxplus/01-title"
	_ "github.com/klauspost/gad/hoaxplus/data"
	"github.com/klauspost/gfx"
)

var (
	renderWidth  = 624
	renderHeight = 240
)

// Generates binary data.
// To install go-bindata, do: go get -u github.com/jteeuwen/go-bindata/...
//
//go:generate go-bindata -ignore=\.go\z -pkg=data -o ../../data/data.go --prefix=../../data ../../data/...

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
