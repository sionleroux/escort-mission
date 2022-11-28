// Use of this source code is subject to an MIT-style
// licence which can be found in the LICENSE file.

package main

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/tanema/gween"
	"github.com/tanema/gween/ease"
	"github.com/tinne26/etxt"
)

const startText = "Press space to start"

// StartScreen is the first screen you see when you start the game, it shows you
// a menu that lets you start a game or change game options etc.
type StartScreen struct {
	background   *ebiten.Image
	textRenderer *StartTextRenderer
	textFader    *gween.Sequence
}

func NewStartScreen(game *Game) *StartScreen {
	fadeSeq := gween.NewSequence(gween.New(50, 200, 60, ease.OutQuad))
	fadeSeq.SetYoyo(true)
	return &StartScreen{
		background:   loadImage("assets/splash-screen.png"),
		textRenderer: NewStartTextRenderer(),
		textFader:    fadeSeq,
	}
}

// Update handles player input to update the start screen
func (s *StartScreen) Update() (GameState, error) {
	// Pressing space starts the game
	if ebiten.IsKeyPressed(ebiten.KeySpace) {
		return gameIntro, nil
	}

	alpha, _, _ := s.textFader.Update(1)
	s.textRenderer.alpha = uint8(alpha)

	return gameStart, nil
}

// Draw renders the start screen to the screen
func (s *StartScreen) Draw(screen *ebiten.Image) {
	screen.DrawImage(s.background, &ebiten.DrawImageOptions{})
	s.textRenderer.Draw(screen, startText)
}

// StartTextRenderer wraps etxt.Renderer to draw full-screen text
type StartTextRenderer struct {
	*etxt.Renderer
	alpha uint8
}

// NewStartTextRenderer creates a text renderer for text on the start screen
func NewStartTextRenderer() *StartTextRenderer {
	font := loadFont("assets/fonts/PixelOperator8-Bold.ttf")
	r := etxt.NewStdRenderer()
	r.SetFont(font)
	r.SetAlign(etxt.YCenter, etxt.XCenter)
	r.SetSizePx(8)
	return &StartTextRenderer{r, 0}
}

func (r StartTextRenderer) Draw(screen *ebiten.Image, text string) {
	r.SetTarget(screen)
	r.SetColor(color.RGBA{0x0, 0x0, 0x0, r.alpha})
	r.Renderer.Draw(text, screen.Bounds().Dx()/2+1, screen.Bounds().Dy()/8*7+1)
	r.SetColor(color.RGBA{0xff, 0xff, 0xff, r.alpha})
	r.Renderer.Draw(text, screen.Bounds().Dx()/2, screen.Bounds().Dy()/8*7)
}
