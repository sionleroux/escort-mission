// Use of this source code is subject to an MIT-style
// licence which can be found in the LICENSE file.

package main

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/tinne26/etxt"
)

// StartScreen is the first screen you see when you start the game, it shows you
// a menu that lets you start a game or change game options etc.
type StartScreen struct {
	background   *ebiten.Image
	textRenderer *StartTextRenderer
}

func NewStartScreen(game *Game) *StartScreen {
	return &StartScreen{
		background:   loadImage("assets/splash-screen.png"),
		textRenderer: NewStartTextRenderer(),
	}
}

// Update handles player input to update the start screen
func (s *StartScreen) Update() (GameState, error) {
	// Pressing space starts the game
	if ebiten.IsKeyPressed(ebiten.KeySpace) {
		return gameIntro, nil
	}

	return gameStart, nil
}

// Draw renders the start screen to the screen
func (s *StartScreen) Draw(screen *ebiten.Image) {
	screen.DrawImage(s.background, &ebiten.DrawImageOptions{})
	s.textRenderer.Draw(screen, "Press space to start")
}

// StartTextRenderer wraps etxt.Renderer to draw full-screen text
type StartTextRenderer struct {
	*etxt.Renderer
}

// NewStartTextRenderer creates a text renderer for text on the start screen
func NewStartTextRenderer() *StartTextRenderer {
	font := loadFont("assets/fonts/PixelOperator8-Bold.ttf")
	r := etxt.NewStdRenderer()
	r.SetFont(font)
	r.SetColor(color.RGBA{0xff, 0xff, 0xff, 0xff})
	r.SetAlign(etxt.YCenter, etxt.XCenter)
	r.SetSizePx(8)
	return &StartTextRenderer{r}
}

func (r StartTextRenderer) Draw(screen *ebiten.Image, text string) {
	r.SetTarget(screen)
	r.Renderer.Draw(text, screen.Bounds().Dx()/2, screen.Bounds().Dy()/8*7)
}
