// Copyright 2021 Siôn le Roux.  All rights reserved.
// Use of this source code is subject to an MIT-style
// licence which can be found in the LICENSE file.

package main

import (
	"errors"
	"image"
	"image/color"
	"log"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// HowManyZombies is how many zombies to generate at the start of the game
const HowManyZombies int = 5

func main() {
	gameWidth, gameHeight := 640, 480

	ebiten.SetWindowSize(gameWidth, gameHeight)
	ebiten.SetWindowTitle("Escort Mission")

	zs := []*Zombie{}
	for i := 0; i < HowManyZombies; i++ {
		zs = append(zs, &Zombie{image.Pt(gameWidth/(i+1)*3, gameHeight/(i+1)*3), 0})
	}

	game := &Game{
		Width:   gameWidth,
		Height:  gameHeight,
		Player:  &Player{image.Pt(gameWidth/2, gameHeight/2), 0},
		Zombies: zs,
	}

	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}

// Game represents the main game state
type Game struct {
	Width   int
	Height  int
	Player  *Player
	Zombies []*Zombie
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

	// Movement controls
	if ebiten.IsKeyPressed(ebiten.KeyW) {
		g.Player.MoveUp()
	}
	if ebiten.IsKeyPressed(ebiten.KeyA) {
		g.Player.MoveLeft()
	}
	if ebiten.IsKeyPressed(ebiten.KeyS) {
		g.Player.MoveDown()
	}
	if ebiten.IsKeyPressed(ebiten.KeyD) {
		g.Player.MoveRight()
	}

	// Move zombie towards player
	for _, z := range g.Zombies {
		if z.Coords.X < g.Player.Coords.X {
			z.MoveRight()
		}
		if z.Coords.X > g.Player.Coords.X {
			z.MoveLeft()
		}
		if z.Coords.Y < g.Player.Coords.Y {
			z.MoveDown()
		}
		if z.Coords.Y > g.Player.Coords.Y {
			z.MoveUp()
		}
	}

	// Player gun rotation
	cx, cy := ebiten.CursorPosition()
	adjacent := float64(g.Player.Coords.X - cx)
	opposite := float64(g.Player.Coords.Y - cy)
	g.Player.Angle = math.Atan2(opposite, adjacent)

	// Collision detection and response between zombie and player
	for _, z := range g.Zombies {
		if image.Rect(
			g.Player.Coords.X, g.Player.Coords.Y,
			g.Player.Coords.X+20, g.Player.Coords.Y+20,
		).Overlaps(image.Rect(
			z.Coords.X, z.Coords.Y,
			z.Coords.X+20, z.Coords.Y+20,
		)) {
			return errors.New("you died")
		}
	}

	return nil
}

// Draw draws the game screen by one frame
func (g *Game) Draw(screen *ebiten.Image) {
	// Player
	ebitenutil.DrawRect(
		screen,
		float64(g.Player.Coords.X),
		float64(g.Player.Coords.Y),
		20,
		20,
		color.White,
	)
	// Gun
	ebitenutil.DrawRect(
		screen,
		float64(g.Player.Coords.X)-math.Cos(g.Player.Angle)*20,
		float64(g.Player.Coords.Y)-math.Sin(g.Player.Angle)*20,
		10,
		10,
		color.White,
	)
	// Zombies
	for _, z := range g.Zombies {
		ebitenutil.DrawRect(
			screen,
			float64(z.Coords.X),
			float64(z.Coords.Y),
			20,
			20,
			color.RGBA{255, 0, 0, 255},
		)
	}
}
