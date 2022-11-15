package main

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

// StartScreen is the first screen you see when you start the game, it shows you
// a menu that lets you start a game or change game options etc.
type StartScreen struct{}

// Update handles player input to update the start screen
func (s *StartScreen) Update() (GameState, error) {
	// Pressing Q any time quits immediately
	if ebiten.IsKeyPressed(ebiten.KeySpace) {
		return gameRunning, nil
	}

	return gameStart, nil
}

// Draw renders the start screen to the screen
func (s *StartScreen) Draw(screen *ebiten.Image) {
	ebitenutil.DebugPrint(screen, "Press space to start")
}
