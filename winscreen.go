// Use of this source code is subject to an MIT-style
// licence which can be found in the LICENSE file.

package main

import (
	"image/color"
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/tinne26/etxt"
)

// WinScreen is shown when you reach the end of the game
type WinScreen struct{
	textRenderer *WinTextRenderer
	statRenderer *WinTextRenderer
	Stat         *Stat
}

func NewWinScreen(game *Game) *WinScreen {
	return &WinScreen{
		textRenderer: NewWinTextRenderer(),
		statRenderer: NewStatTextRenderer(),
		Stat:         game.Stat,
	}
}


// WinTextRenderer wraps etxt.Renderer to draw full-screen text
type WinTextRenderer struct {
	*etxt.Renderer
}

// NewWinTextRenderer creates a text renderer for main text on the win screen
func NewWinTextRenderer() *WinTextRenderer {
	font := loadFont("assets/fonts/OptimusPrincepsSemiBold.otf")
	r := etxt.NewStdRenderer()
	r.SetFont(font)
	r.SetColor(color.RGBA{0xff, 0xff, 0xff, 0xff})
	r.SetAlign(etxt.YCenter, etxt.XCenter)
	r.SetSizePx(24)
	return &WinTextRenderer{r}
}

// NewStatTextRenderer creates a text renderer for stat texts on the win screen
func NewStatTextRenderer() *WinTextRenderer {
	font := loadFont("assets/fonts/PixelOperator8-Bold.ttf")
	r := etxt.NewStdRenderer()
	r.SetFont(font)
	r.SetColor(color.RGBA{0xff, 0xff, 0xff, 0xff})
	r.SetAlign(etxt.YCenter, etxt.XCenter)
	r.SetSizePx(8)
	return &WinTextRenderer{r}
}

func (s *WinScreen) Update() (GameState, error) {
	// TODO: maybe calculate some cool stats?
	return gameWon, nil
}

func (s *WinScreen) Draw(screen *ebiten.Image) {
	s.textRenderer.Renderer.SetTarget(screen)
	s.textRenderer.Renderer.Draw("You survived... for now", screen.Bounds().Dx()/2, screen.Bounds().Dy()/4)

	timePlayed := s.Stat.GameWon.Sub(s.Stat.GameStarted).Seconds()
	statText := fmt.Sprintf("You played %d min %d sec\n", int(timePlayed/60), int(timePlayed)%60)
	statText = statText + fmt.Sprintf("You died %d times\n", s.Stat.CounterPlayerDied)
	statText = statText + fmt.Sprintf("Rover died %d times\n", s.Stat.CounterDogDied)
	statText = statText + fmt.Sprintf("You fired %d bullets\n", s.Stat.CounterBulletsFired)
	statText = statText + fmt.Sprintf("You killed %d zombies\n", s.Stat.CounterZombiesKilled)
	s.statRenderer.Renderer.SetTarget(screen)
	s.statRenderer.Renderer.Draw(statText, screen.Bounds().Dx()/2, screen.Bounds().Dy()/5*3)
}
