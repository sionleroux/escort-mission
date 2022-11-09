// Copyright 2021 Siôn le Roux.  All rights reserved.
// Use of this source code is subject to an MIT-style
// licence which can be found in the LICENSE file.

package main

import (
	"errors"
	"fmt"
	"log"
	"math"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	camera "github.com/melonfunction/ebiten-camera"
	"github.com/solarlune/ldtkgo"
	"github.com/solarlune/resolv"
)

// LayerEntities is the layer to use for entity positions
const LayerEntities int = 0

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
	Dog          *Dog
	Zombies      Zombies
	Space        *resolv.Space
}

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
	g.Sprites[spriteZombie] = loadSprite("Zombie_1")
	g.Sprites[spriteDog] = loadSprite("Dog")

	// Load entities from map
	entities := level.LayerByIdentifier("Entities")

	// Add player to the game
	g.Player = &Player{
		State:  playerIdle,
		Object: resolv.NewObject(float64(g.Width/2), float64(g.Height/2), 20, 20, tagPlayer),
		Angle:  0,
		Sprite: g.Sprites[spritePlayer],
	}
	g.Space.Add(g.Player.Object)

	// Load the dog's path
	dogEntity := entities.EntityByIdentifier("Dog")

	pathArray := dogEntity.Properties[0].AsArray()
	path := make([]Coord, len(pathArray))
	for index, pathCoord := range pathArray {
		// Do we really need to make these crazy castings?
		path[index] = Coord{
			X: (pathCoord.(map[string]interface{})["cx"].(float64) + 0.5)*float64(entities.GridSize),
			Y: (pathCoord.(map[string]interface{})["cy"].(float64) + 0.5)*float64(entities.GridSize),
		}

		if index > 0 {
			ebitenutil.DrawLine(g.Background, path[index-1].X, path[index-1].Y, path[index].X, path[index].Y, color.Black)
		}
	}

	// Add dog to the game
	g.Dog = &Dog{
		Object:   resolv.NewObject(float64(dogEntity.Position[0]), float64(dogEntity.Position[1]), 32, 32, tagDog),
		Angle:    0,
		Sprite:   g.Sprites[spriteDog],
		Path:     path,
		NextPath: -1,
	}
	g.Space.Add(g.Dog.Object)

	// Add zombies to the game
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

	if g.Player.State != playerShooting && clicked() {
		g.Player.State = playerShooting
		// TODO: replace this with zombie dying state, animation, sfx
		// and only remove it *afterwards*
		g.Zombies = g.Zombies[1:]
	}

	// Update player
	g.Player.Update(g)

	// Update dog
	g.Dog.Update(g)

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

	// Dog
	g.Dog.Draw(g)

	// Zombies
	for _, z := range g.Zombies {
		z.Draw(g)
	}

	g.Camera.Blit(screen)

	ebitenutil.DebugPrint(screen, fmt.Sprintf(
		"FPS: %.2f\nTPS: %.2f\nX: %.2f\nY: %.2f",
		ebiten.ActualFPS(),
		ebiten.ActualTPS(),
		g.Player.Object.X / 32,
		g.Player.Object.Y / 32,
	))
}

// Clicked is shorthand for when the left mouse button has just been clicked
func clicked() bool {
	return inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft)
}

// CalcObjectDistance calculates the distance between two Objects
func CalcObjectDistance(obj1, obj2 *resolv.Object) (float64, float64, float64) {
	return CalcDistance(obj1.X, obj1.Y, obj2.X, obj2.Y), obj1.X-obj2.X, obj1.Y-obj2.Y
}

// CalcDistance calculates the distance between two coordinates
func CalcDistance(x1, y1, x2, y2 float64) float64 {
	return math.Sqrt(math.Pow(x1-x2, 2) + math.Pow(y1-y2, 2))
}
