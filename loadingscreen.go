// Use of this source code is subject to an MIT-style
// licence which can be found in the LICENSE file.

package main

import (
	"errors"
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/tanema/gween"
	"github.com/tanema/gween/ease"
	"github.com/tinne26/etxt"
)

// How long at least to show the loading screen even if everything loads very
// fast so that it isn't just a black flash
var loadingScreenMinTime = 2 * 60

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
	textFader    *gween.Tween
	alpha        uint8
	ticker       int
	Loaded       bool
}

func NewLoadingScreen() *LoadingScreen {
	return &LoadingScreen{
		Counter:      new(uint8),
		textRenderer: NewTextRenderer(),
		textFader:    gween.New(255, 0, float32(loadingScreenMinTime)*0.3, ease.InQuad),
		alpha:        255,
	}
}

var ErrorDoneLoading = errors.New("Done loading")

// Update handles player input to update the start screen
func (s *LoadingScreen) Update() (GameState, error) {
	s.ticker++
	if s.Loaded && s.ticker > loadingScreenMinTime {
		alpha, done := s.textFader.Update(1)
		s.alpha = uint8(math.Ceil(float64(alpha)))
		if done {
			return gameLoading, ErrorDoneLoading
		}
	}
	return gameLoading, nil
}

// Draw renders the start screen to the screen
func (s *LoadingScreen) Draw(screen *ebiten.Image) {
	var whatTxt string
	if int(*s.Counter) < len(loadingWhat) {
		whatTxt = loadingWhat[*s.Counter]
	}
	txt := s.textRenderer
	txt.SetTarget(screen)
	txt.SetColor(color.RGBA{0xff, 0xff, 0xff, s.alpha})
	txt.Draw(
		"An action adventure story by:\nRowan Lindeque\nTristan Le Roux\nSiôn Le Roux\nPéter Kertész",
		screen.Bounds().Dx()/2,
		screen.Bounds().Dy()/2,
	)
	txt.Draw(
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
