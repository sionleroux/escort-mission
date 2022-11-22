package main

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/tinne26/etxt"
)

// DeathRenderer wraps etxt.Renderer to draw full-screen text
type DeathRenderer struct {
	*etxt.Renderer
}

// NewDeathRenderer creates a text renderer for death screens
func NewDeathRenderer() *DeathRenderer {
	font := loadFont("assets/fonts/OptimusPrincepsSemiBold.otf")
	r := etxt.NewStdRenderer()
	r.SetFont(font)
	r.SetColor(color.RGBA{0x4f, 0x0, 0x1, 0xff})
	r.SetAlign(etxt.YCenter, etxt.XCenter)
	r.SetSizePx(36)
	return &DeathRenderer{r}
}

// DrawCenered draws the death text to the centre of the screen
func (r DeathRenderer) DrawCenered(screen *ebiten.Image, text string) {
	r.SetTarget(screen)
	r.Draw(text, screen.Bounds().Dx()/2, screen.Bounds().Dy()/2)
}
