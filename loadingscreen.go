// Use of this source code is subject to an MIT-style
// licence which can be found in the LICENSE file.

package main

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
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
	Counter LoadingCounter // what is being loaded
}

func NewLoadingScreen() *LoadingScreen {
	return &LoadingScreen{new(uint8)}
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
	ebitenutil.DebugPrint(screen, "Loading..."+whatTxt)
}
