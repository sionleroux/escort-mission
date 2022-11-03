// Copyright 2021 Siôn le Roux.  All rights reserved.
// Use of this source code is subject to an MIT-style
// licence which can be found in the LICENSE file.

package main

import (
	"errors"
	"image/color"
	"log"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/solarlune/resolv"
)

// HowManyZombies is how many zombies to generate at the start of the game
const HowManyZombies int = 5

func main() {
	gameWidth, gameHeight := 640, 480

	ebiten.SetWindowSize(gameWidth, gameHeight)
	ebiten.SetWindowTitle("Escort Mission")

	space := resolv.NewSpace(gameWidth, gameHeight, 20, 20)

	zs := []*Zombie{}
	for i := 0; i < HowManyZombies; i++ {
		z := &Zombie{
			Object: resolv.NewObject(float64(gameWidth)/(float64(i)+1)*3, float64(gameHeight)/(float64(i)+1*3), 16, 16, "mob"),
			Angle:  0,
		}
		space.Add(z.Object)
		zs = append(zs, z)
	}

	player := &Player{resolv.NewObject(float64(gameWidth/2), float64(gameHeight/2), 20, 20), 0}
	space.Add(player.Object)

	game := &Game{
		Width:   gameWidth,
		Height:  gameHeight,
		Player:  player,
		Zombies: zs,
		Space:   space,
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
	Space   *resolv.Space
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
		if z.Object.X < g.Player.Object.X {
			z.MoveRight()
		}
		if z.Object.X > g.Player.Object.X {
			z.MoveLeft()
		}
		if z.Object.Y < g.Player.Object.Y {
			z.MoveDown()
		}
		if z.Object.Y > g.Player.Object.Y {
			z.MoveUp()
		}
	}

	// Player gun rotation
	cx, cy := ebiten.CursorPosition()
	adjacent := g.Player.Object.X - float64(cx)
	opposite := g.Player.Object.Y - float64(cy)
	g.Player.Angle = math.Atan2(opposite, adjacent)

	// Collision detection and response between zombie and player
	if collision := g.Player.Object.Check(0, 0); collision != nil {
		if g.Player.Object.Overlaps(collision.Objects[0]) {
			log.Printf("%#v", collision)
			return errors.New("you died")
		}
	}

	g.Player.Object.Update()
	for _, z := range g.Zombies {
		z.Object.Update()
	}

	return nil
}

// Draw draws the game screen by one frame
func (g *Game) Draw(screen *ebiten.Image) {
	// Player
	ebitenutil.DrawRect(
		screen,
		g.Player.Object.X,
		g.Player.Object.Y,
		g.Player.Object.W,
		g.Player.Object.H,
		color.White,
	)
	// Gun
	ebitenutil.DrawRect(
		screen,
		g.Player.Object.X-math.Cos(g.Player.Angle)*20,
		g.Player.Object.Y-math.Sin(g.Player.Angle)*20,
		10,
		10,
		color.White,
	)
	// Zombies
	for _, z := range g.Zombies {
		ebitenutil.DrawRect(
			screen,
			z.Object.X,
			z.Object.Y,
			z.Object.W,
			z.Object.H,
			color.RGBA{255, 0, 0, 255},
		)
	}
}
