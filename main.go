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
	"strconv"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	camera "github.com/melonfunction/ebiten-camera"
	"github.com/solarlune/ldtkgo"
	"github.com/solarlune/resolv"

	beziercp "github.com/brothertoad/bezier"
	"gonum.org/v1/plot/font"
	beziercurve "gonum.org/v1/plot/tools/bezier"
	"gonum.org/v1/plot/vg"
)

const (
	tagPlayer     = "player"
	tagMob        = "mob"
	tagWall       = "wall"
	tagDog        = "dog"
	tagEnd        = "end"
	tagCheckpoint = "check"
)

const gameWidth, gameHeight = 320, 240

func main() {

	ebiten.SetWindowSize(gameWidth*2, gameHeight*2)
	ebiten.SetWindowTitle("eZcort mission")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
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
	gameOver                     // The game has ended because you died
	gameWon                      // The game has ended because you won
)

// Game represents the main game state
type Game struct {
	Width         int
	Height        int
	Tick          int
	TileRenderer  *TileRenderer
	LDTKProject   *ldtkgo.Project
	Sounds        []*audio.Player
	Level         int
	Background    *ebiten.Image
	Foreground    *ebiten.Image
	Camera        *camera.Camera
	Sprites       map[SpriteType]*SpriteSheet
	ZombieSprites []*SpriteSheet
	Player        *Player
	Dog           *Dog
	SpawnPoints   SpawnPoints
	Zombies       Zombies
	Space         *resolv.Space
	LevelMap      LevelMap
	State         GameState
	Checkpoint    int
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

	// Create level map for A* path planning
	g.LevelMap = CreateMap(level.Width, level.Height)

	// Add wall tiles to space for collision detection
	for _, layer := range level.Layers {
		if layer.Type == ldtkgo.LayerTypeIntGrid && layer.Identifier == "Desert" {
			for _, intData := range layer.IntGrid {
				object := resolv.NewObject(
					float64(intData.Position[0]+layer.OffsetX),
					float64(intData.Position[1]+layer.OffsetY),
					float64(layer.GridSize),
					float64(layer.GridSize),
					tagWall,
				)
				object.SetShape(resolv.NewRectangle(
					float64(intData.Position[0]+layer.OffsetX),
					float64(intData.Position[1]+layer.OffsetY),
					float64(layer.GridSize),
					float64(layer.GridSize),
				))
				g.Space.Add(object)

				g.LevelMap.SetObstacle(intData.Position[0]/layer.GridSize, intData.Position[1]/layer.GridSize)
			}
		}
	}

	// Music
	const sampleRate int = 44100 // assuming "normal" sample rate
	context := audio.NewContext(sampleRate)
	g.Sounds = make([]*audio.Player, 7)
	g.Sounds[soundMusicBackground] = NewMusicPlayer(loadSoundFile("assets/music/BackgroundMusic.ogg", sampleRate), context)
	g.Sounds[soundGunShot] = NewSoundPlayer(loadSoundFile("assets/sfx/Gunshot.ogg", sampleRate), context)
	g.Sounds[soundGunReload] = NewSoundPlayer(loadSoundFile("assets/sfx/Reload.ogg", sampleRate), context)
	g.Sounds[soundDogBark1] = NewSoundPlayer(loadSoundFile("assets/sfx/Dog-bark-1.ogg", sampleRate), context)
	g.Sounds[soundPlayerDies] = NewSoundPlayer(loadSoundFile("assets/sfx/PlayerDies.ogg", sampleRate), context)
	g.Sounds[soundHit1] = NewSoundPlayer(loadSoundFile("assets/sfx/Hit-1.ogg", sampleRate), context)
	g.Sounds[soundDryFire] = NewSoundPlayer(loadSoundFile("assets/sfx/DryFire.ogg", sampleRate), context)
	g.Sounds[soundMusicBackground].Play()

	// Load sprites
	g.Sprites = make(map[SpriteType]*SpriteSheet, 5)
	g.Sprites[spritePlayer] = loadSprite("Player")
	g.Sprites[spriteDog] = loadSprite("Dog")
	g.Sprites[spriteZombieSprinter] = loadSprite("Zombie_sprinter")
	g.Sprites[spriteZombieBig] = loadSprite("Zombie_big")
	g.Sprites[spriteZombieCrawler] = loadSprite("Zombie_crawler")
	g.ZombieSprites = make([]*SpriteSheet, zombieVariants)
	for index := 0; index < zombieVariants; index++ {
		g.ZombieSprites[index] = loadSprite("Zombie_" + strconv.Itoa(index))
	}

	// Load entities from map
	entities := level.LayerByIdentifier("Entities")

	// Add endpoint
	endpoint := entities.EntityByIdentifier("End")
	g.Space.Add(resolv.NewObject(
		float64(endpoint.Position[0]), float64(endpoint.Position[1]),
		float64(endpoint.Width), float64(endpoint.Height),
		tagEnd,
	))

	// Add player to the game
	playerPosition := entities.EntityByIdentifier("Player").Position
	g.Player = NewPlayer(playerPosition, g.Sprites[spritePlayer])
	g.Space.Add(g.Player.Object)

	eid := 0
	for _, e := range entities.Entities {
		if strings.HasPrefix(e.Identifier, "Checkpoint") {
			eid++
			log.Println(e.Identifier, e.Position)
			img := loadEntityImage(e.Identifier)
			w, h := img.Size()
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(float64(e.Position[0]), float64(e.Position[1]))
			g.Background.DrawImage(img, op)
			obj := resolv.NewObject(
				float64(e.Position[0]), float64(e.Position[1]),
				float64(w), float64(h),
				tagCheckpoint,
			)
			obj.Data = eid
			g.Space.Add(obj)
		}
	}

	// Load the dog's path
	dogEntity := entities.EntityByIdentifier("Dog")
	pathArray := dogEntity.PropertyByIdentifier("Path").AsArray()
	// Start with the dog's current position
	pathPoints := []beziercp.PointF{{X: float64(dogEntity.Position[0]), Y: float64(dogEntity.Position[1])}}
	for _, pathCoord := range pathArray {
		pathPoints = append(pathPoints, beziercp.PointF{
			X: (pathCoord.(map[string]any)["cx"].(float64) + 0.5) * float64(entities.GridSize),
			Y: (pathCoord.(map[string]any)["cy"].(float64) + 0.5) * float64(entities.GridSize),
		})
	}
	// Get Bezier control points from the path
	curveCPs := beziercp.GetControlPointsF(pathPoints)
	// Get the Bezier curves through the origiunal path points based on the control points
	var dogpath []Coord
	for _, c := range curveCPs {
		curve := beziercurve.New(
			vg.Point{X: font.Length(int(c.P0.X)), Y: font.Length(c.P0.Y)},
			vg.Point{X: font.Length(int(c.P1.X)), Y: font.Length(c.P1.Y)},
			vg.Point{X: font.Length(int(c.P2.X)), Y: font.Length(c.P2.Y)},
			vg.Point{X: font.Length(int(c.P3.X)), Y: font.Length(c.P3.Y)},
		)

		// 4 points per curve
		for i := 0.0; i < 1; i = i + 0.25 {
			bp := curve.Point(i)
			dogpath = append(dogpath, Coord{X: float64(bp.X), Y: float64(bp.Y)})
		}
	}

	// Add dog to the game
	object := resolv.NewObject(
		float64(dogEntity.Position[0]), float64(dogEntity.Position[1]),
		16, 16,
		tagDog,
	)
	object.SetShape(resolv.NewRectangle(
		0, 0,
		15, 8,
	))
	object.Shape.(*resolv.ConvexPolygon).RecenterPoints()
	g.Dog = &Dog{
		Object:   object,
		Angle:    0,
		Sprite:   g.Sprites[spriteDog],
		Path:     dogpath,
		NextPath: -1,
	}
	g.Space.Add(g.Dog.Object)

	// Add spawnpoints to the game
	for _, e := range entities.Entities {
		if e.Identifier == "Zombie" || e.Identifier == "Zombie_sprinter" || e.Identifier == "Zombie_big" {
			ztype := zombieNormal
			if e.Identifier == "Zombie_sprinter" {
				ztype = zombieSprinter
			} else if e.Identifier == "Zombie_big" {
				ztype = zombieBig
			}
			initialCount := e.PropertyByIdentifier("Initial").AsInt()
			continuous := e.PropertyByIdentifier("Continuous").AsBool()
			g.SpawnPoints = append(g.SpawnPoints, &SpawnPoint{
				Position:     Coord{X: float64(e.Position[0]), Y: float64(e.Position[1])},
				InitialCount: initialCount,
				Continuous:   continuous,
				ZombieType:   ztype,
			})
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

	// Pressing R reloads the ammo
	if inpututil.IsKeyJustPressed(ebiten.KeyR) {
		g.Player.Reload(g)
	}

	if g.State == gameOver || g.State == gameWon {
		return nil // TODO: provide a possibility to restart the game
	}

	// Gun shooting handler
	if clicked() {
		Shoot(g)
	}

	// Update player
	g.Player.Update(g)

	// Update dog
	g.Dog.Update(g)

	// Update zombies
	g.Zombies.Update(g)

	// Update spawn points
	g.SpawnPoints.Update(g)

	// Collision detection and response between zombie and player
	if collision := g.Player.Object.Check(0, 0, tagMob); collision != nil {
		if g.Player.Object.Overlaps(collision.Objects[0]) {
			if g.Player.Object.Shape.Intersection(0, 0, collision.Objects[0].Shape) != nil {
				g.Sounds[soundMusicBackground].Pause()
				g.Sounds[soundMusicBackground].Rewind()
				g.Sounds[soundPlayerDies].Rewind()
				g.Sounds[soundPlayerDies].Play()
				g.State = gameOver
				return nil // return early, no point in continuing, you are dead
			}
		}
	}

	// Do something special when you find a Checkpoint entity
	g.Checkpoint = 0 // maybe we should use actions instead of global state
	if collision := g.Player.Object.Check(0, 0, tagCheckpoint); collision != nil {
		if o := collision.Objects[0]; g.Player.Object.Overlaps(o) {
			g.Checkpoint = o.Data.(int)
		}
	}

	// End game when you reach the End entity
	if collision := g.Player.Object.Check(0, 0, tagEnd); collision != nil {
		if g.Player.Object.Overlaps(collision.Objects[0]) {
			g.State = gameWon
			return nil
		}
	}

	// Collision detection and response between zombie and player
	if collision := g.Dog.Object.Check(0, 0, tagMob); collision != nil {
		if g.Dog.Object.Overlaps(collision.Objects[0]) {
			g.Dog.State = dogDied
		}
	}

	// Game over if the dog dies
	if g.Dog.State == dogDied {
		g.Sounds[soundMusicBackground].Pause()
		g.Sounds[soundMusicBackground].Rewind()
		g.State = gameOver
		return nil
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
		if g.Dog.State == dogDied {
			ebitenutil.DebugPrint(screen, "Your dog died, press Q to quit")
		} else {
			ebitenutil.DebugPrint(screen, "You Died, press Q to quit")
		}
		return // game not loaded yet
	}

	if g.State == gameWon {
		ebitenutil.DebugPrint(screen, "You Won, press Q to quit")
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
			"Reloading: %t\n"+
			"Checkpoint: %d\n",
		ebiten.ActualFPS(),
		ebiten.ActualTPS(),
		g.Player.Object.X/32,
		g.Player.Object.Y/32,
		len(g.Zombies),
		g.Player.Ammo,
		(g.Player.State == playerReload),
		g.Checkpoint,
	))
}

func debugPosition(g *Game, screen *ebiten.Image, o *resolv.Object) {
	verts := o.Shape.(*resolv.ConvexPolygon).Transformed()
	for i := 0; i < len(verts); i++ {
		vert := verts[i]
		next := verts[0]
		if i < len(verts)-1 {
			next = verts[i+1]
		}
		vX, vY := g.Camera.GetScreenCoords(vert.X(), vert.Y())
		nX, nY := g.Camera.GetScreenCoords(next.X(), next.Y())
		ebitenutil.DrawLine(screen, vX, vY, nX, nY, color.White)
	}
}

// Clicked is shorthand for when the left mouse button has just been clicked
func clicked() bool {
	return inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft)
}

// Shoot sets shooting states and also die states for any zombies in range
func Shoot(g *Game) {
	interruptReload := func() {
		g.Sounds[soundGunReload].Pause()
		g.Sounds[soundDryFire].Rewind()
		g.Sounds[soundDryFire].Play()
		g.Player.State = playerDryFire
	}

	switch g.Player.State {
	case playerShooting, playerReady, playerUnready:
		return // no-op
	case playerReload:
		interruptReload()
		return
	default:
		if g.Player.Ammo < 1 {
			interruptReload()
			return
		}

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
					g.Sounds[soundHit1].Rewind()
					g.Sounds[soundHit1].Play()
					o.Data.(*Zombie).Hit()
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
