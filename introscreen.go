// Use of this source code is subject to an MIT-style
// licence which can be found in the LICENSE file.

package main

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/tanema/gween"
	"github.com/tanema/gween/ease"
	"github.com/tinne26/etxt"
)

// IntroScreen is displayed before the actual game starts
type IntroScreen struct {
	textRenderer *IntroRenderer
	fadeTween    *gween.Tween
	Alpha        float64
	Tick         int
	IntroVoice   *audio.Player
}

func NewIntroScreen(game *Game) *IntroScreen {
	return &IntroScreen{
		textRenderer: NewIntroRenderer(),
		fadeTween:    gween.New(255, 0, 180, ease.OutExpo),
		Alpha:        255,
		IntroVoice:   NewSoundPlayer(loadSoundFile("assets/voice/Intro.ogg", sampleRate), context),
	}
}

func (s *IntroScreen) Update() (GameState, error) {
	if s.Tick == 0 {
		s.IntroVoice.Play()
	}

	// Pressing S skips the intrr
	if ebiten.IsKeyPressed(ebiten.KeyS) {
		s.IntroVoice.Pause()
		return gameRunning, nil
	}

	s.Tick++

	if s.Tick > 1080 {
		alpha, _ := s.fadeTween.Update(1)
		s.Alpha = float64(alpha)
	}
	if s.Tick == 1260 {
		return gameRunning, nil
	}
	return gameIntro, nil
}

func (s *IntroScreen) Draw(screen *ebiten.Image) {
	ebitenutil.DebugPrint(screen, "Press S to skip intro")

	s.textRenderer.SetColor(color.RGBA{0xff, 0xff, 0xff, uint8(s.Alpha)})
	s.textRenderer.DrawCentered(screen, "In the middle of nowhere")
}

// IntroRenderer wraps etxt.Renderer to draw full-screen text
type IntroRenderer struct {
	*etxt.Renderer
}

// NewIntroRenderer creates a text renderer for intro screens
func NewIntroRenderer() *IntroRenderer {
	font := loadFont("assets/fonts/OptimusPrincepsSemiBold.otf")
	r := etxt.NewStdRenderer()
	r.SetFont(font)
	r.SetAlign(etxt.YCenter, etxt.XCenter)
	r.SetSizePx(22)
	return &IntroRenderer{r}
}

// DrawCentered draws the text to the centre of the screen
func (r IntroRenderer) DrawCentered(screen *ebiten.Image, text string) {
	r.SetTarget(screen)
	r.Draw(text, screen.Bounds().Dx()/2, screen.Bounds().Dy()/2)
}
