// Copyright 2021 Siôn le Roux.  All rights reserved.
// Use of this source code is subject to an MIT-style
// licence which can be found in the LICENSE file.

package main

import (
	"errors"
	"image"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

const gameWidth, gameHeight = 320, 240

func main() {
	ebiten.SetWindowSize(gameWidth*2, gameHeight*2)
	ebiten.SetWindowTitle("eZcort mission")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	ebiten.SetWindowIcon([]image.Image{loadImage("assets/icon.png")})
	ebiten.SetFPSMode(ebiten.FPSModeVsyncOffMaximum)

	game := &Game{Width: gameWidth, Height: gameHeight}
	game.DeathScreen = NewDeathScreen(false)

	go NewGameScreen(game)

	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}

// GameState is global state for the whole game
type GameState int

const (
	gameLoading GameState = iota // Loading files and setting up the game
	gameStart                    // Game start screen is shown
	gameRunning                  // The game is running the main game code
	gameOver                     // The game has ended because you died
	gameWon                      // The game has ended because you won
)

// Game represents the main game state
type Game struct {
	Width         int
	Height        int
	StartScreen   Screen
	DeathScreen   Screen
	LoadingScreen Screen
	WinScreen     Screen
	GameScreen    Screen
	State         GameState
}

// Layout is hardcoded for now, may be made dynamic in future
func (g *Game) Layout(outsideWidth int, outsideHeight int) (screenWidth int, screenHeight int) {
	return g.Width, g.Height
}

// Update calculates game logic
func (g *Game) Update() error {
	// Pressing Q any time quits immediately
	if ebiten.IsKeyPressed(ebiten.KeyQ) {
		return errors.New("game quit by player")
	}

	// Pressing F toggles full-screen
	if inpututil.IsKeyJustPressed(ebiten.KeyF) {
		if ebiten.IsFullscreen() {
			ebiten.SetFullscreen(false)
		} else {
			ebiten.SetFullscreen(true)
		}
	}

	if g.State == gameStart {
		state, err := g.StartScreen.Update()
		g.State = state
		return err
	}

	if g.State == gameLoading {
		state, err := g.LoadingScreen.Update()
		g.State = state
		return err
	}

	if g.State == gameOver {
		state, err := g.DeathScreen.Update()
		g.State = state
		return err
	}

	if g.State == gameWon {
		state, err := g.WinScreen.Update()
		g.State = state
		return err
	}

	if g.State == gameRunning {
		state, err := g.GameScreen.Update()
		g.State = state
		return err
	}

	return nil
}

// Draw draws the game screen by one frame
func (g *Game) Draw(screen *ebiten.Image) {
	if g.State == gameLoading {
		g.LoadingScreen.Draw(screen)
		return
	}

	if g.State == gameOver {
		g.DeathScreen.Draw(screen)
		return
	}

	if g.State == gameWon {
		g.WinScreen.Draw(screen)
		return
	}

	if g.State == gameStart {
		g.StartScreen.Draw(screen)
		return
	}

	if g.State == gameRunning {
		g.GameScreen.Draw(screen)
		return
	}

}

// Screen is a full-screen UI Screen for some part of the game like a menu or a
// game level
type Screen interface {
	Update() (GameState, error)
	Draw(screen *ebiten.Image)
}
