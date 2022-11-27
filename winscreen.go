// Use of this source code is subject to an MIT-style
// licence which can be found in the LICENSE file.

package main

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

// WinScreen is shown when you reach the end of the game
type WinScreen struct{}

func (s *WinScreen) Update() (GameState, error) {
	// TODO: maybe calculate some cool stats?
	return gameWon, nil
}

func (s *WinScreen) Draw(screen *ebiten.Image) {
	ebitenutil.DebugPrint(screen, "You Won, press Q to quit")
}
