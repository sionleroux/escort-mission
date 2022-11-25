// Use of this source code is subject to an MIT-style
// licence which can be found in the LICENSE file.

package main

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

// LoadingScreen is while the game loads assets and sets up the actual game
type LoadingScreen struct{}

func (s *LoadingScreen) Update() (GameState, error) {
	// TODO: this is where NewGame should happen probably
	return gameLoading, nil
}

func (s *LoadingScreen) Draw(screen *ebiten.Image) {
	ebitenutil.DebugPrint(screen, "Loading...")
}
