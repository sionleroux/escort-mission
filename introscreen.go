// Use of this source code is subject to an MIT-style
// licence which can be found in the LICENSE file.

package main

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/tanema/gween"
	"github.com/tanema/gween/ease"
	"github.com/tinne26/etxt"
)

const introVoiceLength = 900

// IntroScreen is displayed before the actual game starts
type IntroScreen struct {
	textRenderer     *IntroRenderer
	skipTextRenderer *IntroRenderer
	textFader        *gween.Tween
	skipTextFader    *gween.Sequence
	Tick             int
	IntroVoice       *audio.Player
}

func NewIntroScreen(game *Game) *IntroScreen {
	fadeSeq := gween.NewSequence(gween.New(50, 200, 60, ease.OutQuad))
	fadeSeq.SetYoyo(true)

	return &IntroScreen{
		textRenderer:     NewIntroRenderer(),
		skipTextRenderer: NewSkipTextRenderer(),
		textFader:        gween.New(0xff, 0, fadeOutTime, ease.OutQuad),
		skipTextFader:    fadeSeq,
		IntroVoice:       NewSoundPlayer(loadSoundFile("assets/voice/Intro.ogg", sampleRate), context),
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

	alpha, _, _ := s.skipTextFader.Update(1)
	s.skipTextRenderer.alpha = uint8(alpha)

	if s.Tick > introVoiceLength {
		alpha, _ := s.textFader.Update(1)
		s.textRenderer.alpha = uint8(alpha)
	}
	if s.Tick == introVoiceLength+fadeOutTime {
		return gameRunning, nil
	}
	return gameIntro, nil
}

func (s *IntroScreen) Draw(screen *ebiten.Image) {
	s.textRenderer.SetColor(color.RGBA{0xff, 0xff, 0xff, s.textRenderer.alpha})
	s.textRenderer.Renderer.SetTarget(screen)
	s.textRenderer.Renderer.Draw("In the middle of nowhere", screen.Bounds().Dx()/2, screen.Bounds().Dy()/2)

	s.skipTextRenderer.SetColor(color.RGBA{0xff, 0xff, 0xff, s.skipTextRenderer.alpha})
	s.skipTextRenderer.Renderer.SetTarget(screen)
	s.skipTextRenderer.Renderer.Draw("Press S to skip intro", screen.Bounds().Dx()/2, screen.Bounds().Dy()/8*7)
}

// IntroRenderer wraps etxt.Renderer to draw text
type IntroRenderer struct {
	*etxt.Renderer
	alpha uint8
}

// NewIntroRenderer creates a text renderer for intro screens
func NewIntroRenderer() *IntroRenderer {
	font := loadFont("assets/fonts/OptimusPrincepsSemiBold.otf")
	r := etxt.NewStdRenderer()
	r.SetFont(font)
	r.SetAlign(etxt.YCenter, etxt.XCenter)
	r.SetSizePx(22)
	return &IntroRenderer{r, 0xff}
}

// NewSkipTextRenderer creates a text renderer for the skip text
func NewSkipTextRenderer() *IntroRenderer {
	font := loadFont("assets/fonts/PixelOperator8-Bold.ttf")
	r := etxt.NewStdRenderer()
	r.SetFont(font)
	r.SetAlign(etxt.YCenter, etxt.XCenter)
	r.SetSizePx(8)
	return &IntroRenderer{r, 0xff}
}
