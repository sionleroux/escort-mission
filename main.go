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
	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	camera "github.com/melonfunction/ebiten-camera"
	"github.com/solarlune/ldtkgo"
	"github.com/solarlune/resolv"
)

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

// GameState is global state for the whole game
type GameState int

const (
	gameLoading GameState = iota // Loading files and setting up the game
	gameRunning                  // The game is running the main game code
	gameOver                     // The game has ended
)

// Game represents the main game state
type Game struct {
	Width        int
	Height       int
	Tick         int
	TileRenderer *TileRenderer
	LDTKProject  *ldtkgo.Project
	Sounds       []*audio.Player
	Level        int
	Background   *ebiten.Image
	Foreground   *ebiten.Image
	Camera       *camera.Camera
	Sprites      map[SpriteType]*SpriteSheet
	Player       *Player
	Dog          *Dog
	Zombies      Zombies
	Space        *resolv.Space
	State        GameState
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

	bg := ebiten.NewImage(level.Width, level.Height)
	bg.Fill(level.BGColor)
	fg := ebiten.NewImage(level.Width, level.Height)

	// Render map
	g.TileRenderer.Render(level)
	for _, layer := range g.TileRenderer.RenderedLayers {
		log.Println("Pre-drawing layer:", layer.Layer.Identifier)
		if layer.Layer.Identifier == "Treetops" {
			fg.DrawImage(layer.Image, &ebiten.DrawImageOptions{})
		} else {
			bg.DrawImage(layer.Image, &ebiten.DrawImageOptions{})
		}
	}
	g.Background = bg
	g.Foreground = fg

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

	// Music
	const sampleRate int = 44100 // assuming "normal" sample rate
	context := audio.NewContext(sampleRate)
	g.Sounds = make([]*audio.Player, 3)
	g.Sounds[soundMusicBackground] = NewMusicPlayer(loadSoundFile("assets/music/BackgroundMusic.ogg", sampleRate), context)
	g.Sounds[soundGunShot] = NewSoundPlayer(loadSoundFile("assets/sfx/Gunshot.ogg", sampleRate), context)
	g.Sounds[soundDogBark1] = NewSoundPlayer(loadSoundFile("assets/sfx/Dog-bark-1.ogg", sampleRate), context)
	g.Sounds[soundMusicBackground].Play()

	// Load sprites
	g.Sprites = make(map[SpriteType]*SpriteSheet, 3)
	g.Sprites[spritePlayer] = loadSprite("Player")
	g.Sprites[spriteZombie] = loadSprite("Zombie_1")
	g.Sprites[spriteDog] = loadSprite("Dog")

	// Load entities from map
	entities := level.LayerByIdentifier("Entities")

	// Add player to the game
	playerPosition := entities.EntityByIdentifier("Player").Position
	g.Player = NewPlayer(playerPosition, g.Sprites[spritePlayer])
	g.Space.Add(g.Player.Object)

	// Load the dog's path
	dogEntity := entities.EntityByIdentifier("Dog")

	pathArray := dogEntity.PropertyByIdentifier("Path").AsArray()
	path := make([]Coord, len(pathArray))
	for index, pathCoord := range pathArray {
		// Do we really need to make these crazy castings?
		path[index] = Coord{
			X: (pathCoord.(map[string]any)["cx"].(float64) + 0.5) * float64(entities.GridSize),
			Y: (pathCoord.(map[string]any)["cy"].(float64) + 0.5) * float64(entities.GridSize),
		}
	}

	// Add dog to the game
	g.Dog = &Dog{
		Object:   resolv.NewObject(float64(dogEntity.Position[0]), float64(dogEntity.Position[1]), 8, 8, tagDog),
		Angle:    0,
		Sprite:   g.Sprites[spriteDog],
		Path:     path,
		NextPath: -1,
	}
	g.Space.Add(g.Dog.Object)

	// Add zombies to the game
	zombiePositions := []struct{ X, Y int }{
		{0, 0}, {-1, 0}, {-1, -1},
		{0, -1}, {1, -1}, {1, 0},
		{1, 1}, {0, 1}, {-1, 1},
		{-2, 0}, {-2, -1}, {-2, -2},
		{-1, -2}, {0, -2}, {1, -2},
		{2, -2}, {2, -1}, {2, 0},
		{2, 1}, {2, 2}, {1, 2},
		{0, 2}, {-1, 2}, {-1, 2},
		{-2, 2}, {-2, 1},
	} // XXX: surely there is a smarter way than writing this by hand
	for _, e := range entities.Entities {
		if e.Identifier == "Zombie" {
			howManyZombies := e.PropertyByIdentifier("Initial").AsInt()
			for i := 0; i < howManyZombies; i++ {
				if i >= len(zombiePositions) {
					log.Println("ran out of zombie positions, aborting spawning")
					break
				}
				e.Position[0] += zombiePositions[i].X * 16 // 16px should come from Zombie
				e.Position[1] += zombiePositions[i].Y * 16 // 16px should come from Zombie
				z := NewZombie(e.Position, g.Sprites[spriteZombie])
				g.Space.Add(z.Object)
				g.Zombies = append(g.Zombies, z)
			}
		}
	}

	g.State = gameRunning
}

// Layout is hardcoded for now, may be made dynamic in future
func (g *Game) Layout(outsideWidth int, outsideHeight int) (screenWidth int, screenHeight int) {
	return g.Width, g.Height
}

// Update calculates game logic
func (g *Game) Update() error {
	if g.State == gameLoading {
		return nil // game hasn't loaded yet
	}

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

	if g.State == gameOver {
		return nil // TODO: provide a possibility to restart the game
	}

	// Gun shooting handler
	Shoot(g)

	// Update player
	g.Player.Update(g)

	// Update dog
	g.Dog.Update(g)

	// Update zombies
	g.Zombies.Update(g)

	// Collision detection and response between zombie and player
	if collision := g.Player.Object.Check(0, 0, tagMob); collision != nil {
		if g.Player.Object.Overlaps(collision.Objects[0]) {
			g.State = gameOver
			return nil // return early, no point in continuing, you are dead
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
	if g.State == gameLoading {
		ebitenutil.DebugPrint(screen, "Loading...")
		return // game not loaded yet
	}

	if g.State == gameOver {
		ebitenutil.DebugPrint(screen, "You Died, press Q to quit")
		return // game not loaded yet
	}

	g.Camera.Surface.Clear()

	// Ground, walls and other lowest-level stuff needs to be drawn first
	g.Camera.Surface.DrawImage(
		g.Background,
		g.Camera.GetTranslation(&ebiten.DrawImageOptions{}, 0, 0),
	)

	// Dog
	g.Dog.Draw(g)

	// Player
	g.Player.Draw(g)

	// Zombies
	g.Zombies.Draw(g)

	// Tree tops etc. high-up stuff need to be drawn above the entities
	g.Camera.Surface.DrawImage(
		g.Foreground,
		g.Camera.GetTranslation(&ebiten.DrawImageOptions{}, 0, 0),
	)

	g.Camera.Blit(screen)

	ebitenutil.DebugPrint(screen, fmt.Sprintf(
		"FPS: %.2f\n"+
			"TPS: %.2f\n"+
			"X: %.2f\n"+
			"Y: %.2f\n"+
			"Zombies: %d\n"+
			"Ammo: %d\n"+
			"Reloading: %t\n",
		ebiten.ActualFPS(),
		ebiten.ActualTPS(),
		g.Player.Object.X/32,
		g.Player.Object.Y/32,
		len(g.Zombies),
		g.Player.Ammo,
		(g.Player.State == playerReload),
	))
}

// Clicked is shorthand for when the left mouse button has just been clicked
func clicked() bool {
	return inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft)
}

// Shoot sets shooting states and also die states for any zombies in range
func Shoot(g *Game) {
	if clicked() &&
		g.Player.State != playerShooting &&
		g.Player.State != playerReload &&
		g.Player.State != playerReady &&
		g.Player.State != playerUnready {
		g.Sounds[soundGunShot].Rewind()
		g.Sounds[soundGunShot].Play()

		g.Player.Ammo--
		g.Player.State = playerShooting
		rangeOfFire := g.Player.Range
		sX, sY := g.Space.WorldToSpace(
			g.Player.Object.X-math.Cos(g.Player.Angle-math.Pi)*rangeOfFire,
			g.Player.Object.Y-math.Sin(g.Player.Angle-math.Pi)*rangeOfFire,
		)
		pX, pY := g.Space.WorldToSpace(
			g.Player.Object.X+g.Player.Object.W/2,
			g.Player.Object.Y+g.Player.Object.H/2,
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

// CalcObjectDistance calculates the distance between two Objects
func CalcObjectDistance(obj1, obj2 *resolv.Object) (float64, float64, float64) {
	return CalcDistance(obj1.X, obj1.Y, obj2.X, obj2.Y), obj1.X - obj2.X, obj1.Y - obj2.Y
}

// CalcDistance calculates the distance between two coordinates
func CalcDistance(x1, y1, x2, y2 float64) float64 {
	return math.Sqrt(math.Pow(x1-x2, 2) + math.Pow(y1-y2, 2))
}
