// Use of this source code is subject to an MIT-style
// licence which can be found in the LICENSE file.

package main

import (
	"errors"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/tinne26/etxt"
)

// DeathScreen is the screen you see when you or your dog dies in the game, it
// shows you a message for a short while and then respawns you into the game
type DeathScreen struct {
	textRenderer *DeathRenderer
	DogDied      bool
	BellRang     bool
	bellSound    *audio.Player
}

func NewDeathScreen(game *Game) *DeathScreen {
	return &DeathScreen{
		textRenderer: NewDeathRenderer(),
		bellSound:    NewSoundPlayer(loadSoundFile("assets/sfx/Bell.ogg", sampleRate), context),
	}
}

func (s *DeathScreen) Update() (GameState, error) {
	// Pressing Q any time quits immediately
	if ebiten.IsKeyPressed(ebiten.KeyQ) {
		return gameOver, errors.New("game quit by player")
	}

	if !s.BellRang {
		s.bellSound.Play()
		s.BellRang = true
	}

	return gameOver, nil
}

func (s *DeathScreen) Draw(screen *ebiten.Image) {
	if s.DogDied {
		s.textRenderer.DrawCenered(screen, "YOUR DOG DIED")
	} else {
		s.textRenderer.DrawCenered(screen, "YOU DIED")
	}
}

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
