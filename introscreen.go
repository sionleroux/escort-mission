// Use of this source code is subject to an MIT-style
// licence which can be found in the LICENSE file.

package main

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

// IntroScreen is displayed before the actual game starts
type IntroScreen struct{
	Tick int
}

func (s *IntroScreen) Update() (GameState, error) {
	s.Tick++

	if (s.Tick == 300) {
		return gameRunning, nil
	}
	return gameIntro, nil
}

func (s *IntroScreen) Draw(screen *ebiten.Image) {
	ebitenutil.DebugPrint(screen, "Intro")
}
