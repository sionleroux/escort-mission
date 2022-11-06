// Copyright 2021 Siôn le Roux.  All rights reserved.
// Use of this source code is subject to an MIT-style
// licence which can be found in the LICENSE file.

package main

import (
	"errors"
	"fmt"
	"log"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	camera "github.com/melonfunction/ebiten-camera"
	"github.com/solarlune/ldtkgo"
	"github.com/solarlune/resolv"
)

// HowManyZombies is how many zombies to generate at the start of the game
const HowManyZombies int = 5

const (
	tagPlayer = "player"
	tagMob    = "mob"
	tagWall   = "wall"
	tagDog    = "dog"
)

func main() {
	gameWidth, gameHeight := 320, 240

	ebiten.SetWindowSize(gameWidth*2, gameHeight*2)
	ebiten.SetWindowTitle("Escort Mission")
	ebiten.SetFPSMode(ebiten.FPSModeVsyncOffMaximum)

	space := resolv.NewSpace(gameWidth, gameHeight, 20, 20)

	game := &Game{
		Width:  gameWidth,
		Height: gameHeight,
		Space:  space,
		Level:  0,
		Tick:   0,
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
	Tick         int
	TileRenderer *TileRenderer
	LDTKProject  *ldtkgo.Project
	Level        int
	Background   *ebiten.Image
	Camera       *camera.Camera
	Sprites      map[SpriteType]*SpriteSheet
	Player       *Player
	Dog          *Dog
	Zombies      []*Zombie
	Space        *resolv.Space
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
	g.Sprites[spritePlayer] = loadSprite("Player")
	g.Sprites[spriteZombie] = loadSprite("Zombie")
	g.Sprites[spriteDog] = loadSprite("Dog")

	//Add player to the game
	g.Player = &Player{
		State:  playerIdle,
		Object: resolv.NewObject(float64(g.Width/2), float64(g.Height/2), 20, 20, tagPlayer),
		Angle:  0,
		Sprite: g.Sprites[spritePlayer],
	}
	g.Space.Add(g.Player.Object)

	// Add dog to the game
	g.Dog = &Dog{
		Object: resolv.NewObject(float64(g.Width/2 + 32), float64(g.Height/2), 32, 32, tagDog),
		Angle:  0,
		Sprite: g.Sprites[spriteDog],
		State:  dogWalking,
	}
	g.Space.Add(g.Dog.Object)

	//Add zombies to the game
	zs := []*Zombie{}
	for i := 0; i < HowManyZombies; i++ {
		z := &Zombie{
			Object: resolv.NewObject(float64(g.Width)/(float64(i)+1)*3, float64(g.Height)/(float64(i)+1*3), 16, 16, tagMob),
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
	g.Tick++

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

	// Update player
	g.Player.Update(g)

	// Update dog
	g.Dog.Update(g)

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

	// Collision detection and response between zombie and player
	if collision := g.Player.Object.Check(0, 0, tagMob); collision != nil {
		if g.Player.Object.Overlaps(collision.Objects[0]) {
			log.Printf("%#v", collision)
			return errors.New("you died")
		}
	}

	// Update zombies
	for _, z := range g.Zombies {
		z.Update(g)
	}

	// Position camera and clamp in to the Map dimensions
	level := g.LDTKProject.Levels[g.Level]
	g.Camera.SetPosition(
		math.Min(math.Max(g.Player.Object.X, float64(g.Width)/2), float64(level.Width)-float64(g.Width)/2),
		math.Min(math.Max(g.Player.Object.Y, float64(g.Height)/2), float64(level.Height)-float64(g.Height)/2))

	return nil
}

// Draw draws the game screen by one frame
func (g *Game) Draw(screen *ebiten.Image) {
	op := &ebiten.DrawImageOptions{}

	g.Camera.Surface.Clear()
	g.Camera.Surface.DrawImage(g.Background, g.Camera.GetTranslation(op, 0, 0))

	// Player
	g.Player.Draw(g)

	// Dog
	g.Dog.Draw(g)

	// Gun
	// sX, sY := g.Camera.GetScreenCoords(
	// 	g.Player.Object.X-math.Cos(g.Player.Angle)*20,
	// 	g.Player.Object.Y-math.Sin(g.Player.Angle)*20)
	// ebitenutil.DrawRect(
	// 	g.Camera.Surface,
	// 	sX,
	// 	sY,
	// 	10,
	// 	10,
	// 	color.White,
	// )

	// Zombies
	for _, z := range g.Zombies {
		z.Draw(g)
	}

	g.Camera.Blit(screen)

	ebitenutil.DebugPrint(screen, fmt.Sprintf(
		"FPS: %.2f\nTPS: %.2f\n",
		ebiten.ActualFPS(),
		ebiten.ActualTPS(),
	))
}
