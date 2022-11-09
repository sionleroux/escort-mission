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

const (
	tagMob  = "mob"
	tagWall = "wall"
)

func main() {
	gameWidth, gameHeight := 320, 240

	ebiten.SetWindowSize(gameWidth*2, gameHeight*2)
	ebiten.SetWindowTitle("Escort Mission")
	ebiten.SetFPSMode(ebiten.FPSModeVsyncOffMaximum)

	game := &Game{
		Width:  gameWidth,
		Height: gameHeight,
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
	Zombies      Zombies
	Space        *resolv.Space
}

// NewGame fills up the main Game data with assets, entities, pre-generated
// tiles and other things that take longer to load and would make the game pause
// before starting if we did it before the first Update loop
func NewGame(g *Game) {
	g.Camera = camera.NewCamera(g.Width, g.Height, 0, 0, 0, 1)

	var renderer *TileRenderer
	ldtkProject := loadMaps("assets/maps/maps.ldtk")
	renderer = NewTileRenderer(&EmbedLoader{"assets/maps"})

	g.TileRenderer = renderer
	g.LDTKProject = ldtkProject

	level := g.LDTKProject.Levels[g.Level]

	bg := ebiten.NewImage(
		level.Width,
		level.Height,
	)
	bg.Fill(level.BGColor)

	// Render map
	g.TileRenderer.Render(level)
	for _, layer := range g.TileRenderer.RenderedLayers {
		bg.DrawImage(layer.Image, &ebiten.DrawImageOptions{})
	}
	g.Background = bg

	// Create space for collision detection
	g.Space = resolv.NewSpace(level.Width, level.Height, 16, 16)

	// Add wall tiles to space for collision detection
	for _, layer := range level.Layers {
		switch layer.Type {
		case ldtkgo.LayerTypeIntGrid:

			for _, intData := range layer.IntGrid {
				g.Space.Add(resolv.NewObject(
					float64(intData.Position[0]+layer.OffsetX),
					float64(intData.Position[1]+layer.OffsetY),
					float64(layer.GridSize),
					float64(layer.GridSize),
					tagWall,
				))
			}
		}
	}

	// Load sprites
	g.Sprites = make(map[SpriteType]*SpriteSheet, 2)
	g.Sprites[spritePlayer] = loadSprite("Player")
	g.Sprites[spriteZombie] = loadSprite("Zombie")

	// Load entities from map
	entities := level.LayerByIdentifier("Entities")

	// Add player to the game
	playerPosition := entities.EntityByIdentifier("Player").Position
	g.Player = NewPlayer(playerPosition, g.Sprites[spritePlayer])
	g.Space.Add(g.Player.Object)

	// Add zombies to the game
	for _, e := range entities.Entities {
		if e.Identifier == "Zombie" {
			z := NewZombie(e.Position, g.Sprites[spriteZombie])
			g.Space.Add(z.Object)
			g.Zombies = append(g.Zombies, z)
		}
	}
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

	// Gun shooting handler
	Shoot(g)

	// Update player
	g.Player.Update(g)

	// Update zombies
	g.Zombies.Update(g)

	// Collision detection and response between zombie and player
	if collision := g.Player.Object.Check(0, 0, tagMob); collision != nil {
		if g.Player.Object.Overlaps(collision.Objects[0]) {
			log.Printf("%#v", collision)
			return errors.New("you died")
		}
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

	// Zombies
	g.Zombies.Draw(g)

	g.Camera.Blit(screen)

	ebitenutil.DebugPrint(screen, fmt.Sprintf(
		"FPS: %.2f\n"+
			"TPS: %.2f\n"+
			"X: %.2f\n"+
			"Y: %.2f\n"+
			"Zombies: %d\n",
		ebiten.ActualFPS(),
		ebiten.ActualTPS(),
		g.Player.Object.X/32,
		g.Player.Object.Y/32,
		len(g.Zombies),
	))
}

// Clicked is shorthand for when the left mouse button has just been clicked
func clicked() bool {
	return inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft)
}

// Shoot sets shooting states and also die states for any zombies in range
func Shoot(g *Game) {
	if g.Player.State != playerShooting && clicked() {
		g.Player.State = playerShooting
		var rangeOfFire float64 = 100 // XXX this should come from Player
		sX, sY := g.Space.WorldToSpace(
			g.Player.Object.X-math.Cos(g.Player.Angle)*rangeOfFire,
			g.Player.Object.Y-math.Sin(g.Player.Angle)*rangeOfFire,
		)
		pX, pY := g.Space.WorldToSpace(
			g.Player.Object.X+8, // XXX this should come from Player
			g.Player.Object.Y+8,
		)
		cells := g.Space.CellsInLine(pX, pY, sX, sY)
		for _, c := range cells {
			for _, o := range c.Objects {
				if o.HasTags(tagMob) {
					log.Println("HIT!")
					o.Data.(*Zombie).Die()
					return // stop at the first zombie
				}
			}
		}
	}
}
