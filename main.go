// Copyright 2021 Siôn le Roux.  All rights reserved.
// Use of this source code is subject to an MIT-style
// licence which can be found in the LICENSE file.

package main

import (
	"errors"
	"fmt"
	"image/color"
	"log"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/solarlune/resolv"
	"github.com/solarlune/ldtkgo"
	camera "github.com/melonfunction/ebiten-camera"
)

// HowManyZombies is how many zombies to generate at the start of the game
const HowManyZombies int = 5

func main() {
	gameWidth, gameHeight := 640, 480

	ebiten.SetWindowSize(gameWidth, gameHeight)
	ebiten.SetWindowTitle("Escort Mission")
	ebiten.SetFPSMode(ebiten.FPSModeVsyncOffMaximum)

	space := resolv.NewSpace(gameWidth, gameHeight, 20, 20)

	wall := resolv.NewObject(200, 100, 20, 200, "wall")
	space.Add(wall)

	game := &Game{
		Width:  gameWidth,
		Height: gameHeight,
		Space:  space,
		Wall:   wall,
		Level:  0,
	}

	go NewGame(game)

	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}

// Game represents the main game state
type Game struct {
	Width        int
	Height       int
	TileRenderer *TileRenderer
	LDTKProject  *ldtkgo.Project
	Level        int
	Background   *ebiten.Image
	Camera       *camera.Camera
	Sprites	     map[SpriteType]*SpriteSheet
	Player       *Player
	Zombies      []*Zombie
	Space        *resolv.Space
	Wall         *resolv.Object
}

func NewGame(g *Game) {
	g.Camera = camera.NewCamera(g.Width, g.Height, 0, 0, 0, 1)

	var renderer *TileRenderer
	ldtkProject := loadMaps("assets/maps/maps.ldtk")
	renderer = NewTileRenderer(&EmbedLoader{"assets/maps"})

	g.TileRenderer = renderer
	g.LDTKProject = ldtkProject

	bg := ebiten.NewImage(
		g.LDTKProject.Levels[g.Level].Width,
		g.LDTKProject.Levels[g.Level].Height,
	)
	bg.Fill(g.LDTKProject.Levels[g.Level].BGColor)

	// Render map
	g.TileRenderer.Render(g.LDTKProject.Levels[g.Level])
	for _, layer := range g.TileRenderer.RenderedLayers {
		bg.DrawImage(layer.Image, &ebiten.DrawImageOptions{})
	}
	g.Background = bg

	//Load sprites
	g.Sprites = make(map[SpriteType]*SpriteSheet, 2)
	g.Sprites[spritePlayer] = loadSprite("player")
	g.Sprites[spriteZombie] = loadSprite("zombie")

	//Add player to the game
	g.Player = &Player{
		Object: resolv.NewObject(float64(g.Width/2), float64(g.Height/2), 20, 20),
		Angle:  0,
		Sprite: g.Sprites[spritePlayer],
	}
	g.Space.Add(g.Player.Object)

	//Add zombies to the game
	zs := []*Zombie{}
	for i := 0; i < HowManyZombies; i++ {
		z := &Zombie{
			Object: resolv.NewObject(float64(g.Width)/(float64(i)+1)*3, float64(g.Height)/(float64(i)+1*3), 16, 16, "mob"),
			Angle:  0,
			Sprite: g.Sprites[spriteZombie],
		}
		g.Space.Add(z.Object)
		zs = append(zs, z)
	}
	g.Zombies = zs
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
	cx, cy := g.Camera.GetCursorCoords()
	adjacent := g.Player.Object.X - float64(cx)
	opposite := g.Player.Object.Y - float64(cy)
	g.Player.Angle = math.Atan2(opposite, adjacent)

	// Collision detection and response between zombie and player
	if collision := g.Player.Object.Check(0, 0, "mob"); collision != nil {
		if g.Player.Object.Overlaps(collision.Objects[0]) {
			log.Printf("%#v", collision)
			return errors.New("you died")
		}
	}

	g.Player.Object.Update()
	for _, z := range g.Zombies {
		// Zombies rotate towards player
		adjacent := z.Object.X - g.Player.Object.X
		opposite := z.Object.Y - g.Player.Object.Y
		z.Angle = math.Atan2(opposite, adjacent)
		z.Object.Update()
	}

	// Position camera
	g.Camera.SetPosition(g.Player.Object.X, g.Player.Object.Y)

	return nil
}

// Draw draws the game screen by one frame
func (g *Game) Draw(screen *ebiten.Image) {
	op := &ebiten.DrawImageOptions{}

	g.Camera.Surface.Clear()
	g.Camera.Surface.DrawImage(g.Background, g.Camera.GetTranslation(op, 0, 0))

	// Wall
	sX, sY := g.Camera.GetScreenCoords(g.Wall.X, g.Wall.Y)
	ebitenutil.DrawRect(
		g.Camera.Surface,
		sX,
		sY,
		g.Wall.W,
		g.Wall.H,
		color.RGBA{0, 0, 255, 255},
	)

	// Player
	g.Player.Draw(g)

	// Gun
	sX, sY = g.Camera.GetScreenCoords(
		g.Player.Object.X-math.Cos(g.Player.Angle)*20,
		g.Player.Object.Y-math.Sin(g.Player.Angle)*20)
	ebitenutil.DrawRect(
		g.Camera.Surface,
		sX,
		sY,
		10,
		10,
		color.White,
	)
	// Zombies
	for _, z := range g.Zombies {
		z.Draw(g)
	}

	g.Camera.Blit(screen)

	ebitenutil.DebugPrint(screen, fmt.Sprintf(
		"FPS: %.2f\nTPS: %.2f\n%.2f %.2f %.2f %.2f",
		ebiten.ActualFPS(),
		ebiten.ActualTPS(),
		g.Player.Object.X, g.Player.Object.Y,
		g.Camera.X, g.Camera.Y,
	))
}
