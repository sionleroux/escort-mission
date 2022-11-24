package main

import (
	"errors"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/tinne26/etxt"
)

const deathCoolDownTime = 4 * 60

// DeathScreen is the screen you see when you or your dog dies in the game, it
// shows you a message for a short while and then respawns you into the game
type DeathScreen struct {
	textRenderer *DeathRenderer
	DogDied      bool
	tick         int
	respawning   bool
	game         *Game
}

func NewDeathScreen(game *Game) *DeathScreen {
	return &DeathScreen{
		textRenderer: NewDeathRenderer(),
		game:         game,
	}
}

func (s *DeathScreen) Update() (GameState, error) {
	s.tick++

	// Pressing Q any time quits immediately
	if ebiten.IsKeyPressed(ebiten.KeyQ) {
		return gameOver, errors.New("game quit by player")
	}

	if s.tick > deathCoolDownTime && !s.respawning {
		s.respawning = true
		go NewGameScreen(s.game)
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
