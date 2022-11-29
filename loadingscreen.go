// Use of this source code is subject to an MIT-style
// licence which can be found in the LICENSE file.

package main

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/tinne26/etxt"
)

// LoadingCounter is for tracking how much of the assets have been loaded
type LoadingCounter *uint8

var loadingWhat = []string{
	"",
	"map",
	"music",
	"sounds",
	"sprites",
	"entities",
	"done",
}

// LoadingScreen is shown while all the assets are loading.
// When loading is ready it switches to Intro screen
type LoadingScreen struct {
	Counter      LoadingCounter // what is being loaded
	textRenderer *etxt.Renderer
}

func NewLoadingScreen() *LoadingScreen {
	return &LoadingScreen{
		Counter:      new(uint8),
		textRenderer: NewTextRenderer(),
	}
}

// Update handles player input to update the start screen
func (s *LoadingScreen) Update() (GameState, error) {
	return gameLoading, nil
}

// Draw renders the start screen to the screen
func (s *LoadingScreen) Draw(screen *ebiten.Image) {
	var whatTxt string
	if int(*s.Counter) < len(loadingWhat) {
		whatTxt = loadingWhat[*s.Counter]
	}
	s.textRenderer.SetTarget(screen)
	s.textRenderer.Draw(
		"An action adventure story by:\nRowan Lindeque\nTristan Le Roux\nSiôn Le Roux\nPéter Kertész",
		screen.Bounds().Dx()/2,
		screen.Bounds().Dy()/2,
	)
	s.textRenderer.Draw(
		"Loading..."+whatTxt,
		screen.Bounds().Dx()/2,
		screen.Bounds().Dy()/8*7,
	)
}

func NewTextRenderer() *etxt.Renderer {
	font := loadFont("assets/fonts/PixelOperator8.ttf")
	r := etxt.NewStdRenderer()
	r.SetFont(font)
	r.SetAlign(etxt.YCenter, etxt.XCenter)
	r.SetColor(color.RGBA{0xff, 0xff, 0xff, 0xff})
	r.SetSizePx(8)
	return r
}
