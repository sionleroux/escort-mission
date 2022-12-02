// Use of this source code is subject to an MIT-style
// licence which can be found in the LICENSE file.

package main

import (
	"image/color"
	"log"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"

	beziercp "github.com/brothertoad/bezier"
	camera "github.com/melonfunction/ebiten-camera"
	"github.com/solarlune/ldtkgo"
	"github.com/solarlune/resolv"
	"github.com/tanema/gween"
	"github.com/tanema/gween/ease"
)

const cameraPadding = 1

// For testing it is sometimes useful to start the game at a later checkpoint
var startingCheckpoint int = 0

// Multiplier applied when object is in sand trap
const sandTrapSpeedMultiplier = 0.5

const (
	tagPlayer     = "player"
	tagMob        = "mob"
	tagWall       = "wall"
	tagDog        = "dog"
	tagEnd        = "end"
	tagOutro      = "outro"
	tagCheckpoint = "check"
	tagSandTrap   = "sandtrap"
)

// Length of the fading animation
const fadeOutTime = 180

// Minimum time between two voice lines
const voiceGuardTime = 1200

// Voices are played in this order between the checkpoints: flavour + kill + flavour
const (
	voiceStepFlavour1 uint8 = iota
	voiceStepKill
	voiceStepFlavour2
)

// GameScreen is the screen for the actual main game itself
type GameScreen struct {
	Width          int
	Height         int
	Tick           int
	TileRenderer   *TileRenderer
	LDTKProject    *ldtkgo.Project
	Music          *MusicLoop
	Sounds         Sounds
	Voices         Sounds
	Level          int
	Background     *ebiten.Image
	Foreground     *ebiten.Image
	Camera         *camera.Camera
	Cursor         *Cursor
	Sprites        map[SpriteType]*SpriteSheet
	ZombieSprites  []*SpriteSheet
	Player         *Player
	Dog            *Dog
	SpawnPoints    SpawnPoints
	Zombies        Zombies
	BossDefeated   bool
	Space          *resolv.Space
	LevelMap       LevelMap
	Checkpoint     int
	HUD            *HUD
	Debuggers      Debuggers
	Zoom           *Zoom
	FadeTween      *gween.Tween
	Alpha          uint8
	Stat           *Stat
	VoiceGuardTime int
	NextVoiceStep  uint8
}

// NewGameScreen fills up the main Game data with assets, entities, pre-generated
// tiles and other things that take longer to load and would make the game pause
// before starting if we did it before the first Update loop
func NewGameScreen(game *Game, loadingCount LoadingCounter) {
	g := &GameScreen{
		Width:         game.Width + cameraPadding,
		Height:        game.Height + cameraPadding,
		Checkpoint:    game.Checkpoint,
		Debuggers:     debuggers,
		FadeTween:     gween.New(255, 0, fadeOutTime, ease.OutQuad),
		Alpha:         255,
		Stat:          game.Stat,
		NextVoiceStep: voiceStepFlavour2,
	}

	g.Camera = camera.NewCamera(g.Width, g.Height, 0, 0, 0, 1)
	g.Cursor = NewCursor()

	*loadingCount++
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
			// Draw black, transparent, scaled copy as fake shadows
			op := &ebiten.DrawImageOptions{}
			op.ColorM.Scale(0, 0, 0, 0.1)
			op.GeoM.Translate(8, 8)
			fg.DrawImage(layer.Image, op)
			// Draw real trees
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

	// Add wall tiles and sand traps to space for collision detection
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
		} else if layer.Type == ldtkgo.LayerTypeIntGrid && layer.Identifier == "Sand_Traps" {
			for _, intData := range layer.IntGrid {
				object := resolv.NewObject(
					float64(intData.Position[0]+layer.OffsetX),
					float64(intData.Position[1]+layer.OffsetY),
					16, 16,
					tagSandTrap,
				)
				object.SetShape(resolv.NewRectangle(
					float64(intData.Position[0]+layer.OffsetX),
					float64(intData.Position[1]+layer.OffsetY),
					16, 16,
				))
				g.Space.Add(object)
			}
		}
	}

	// SoundLoops
	*loadingCount++
	g.Music = NewMusicPlayer(loadSoundFile("assets/music/BackgroundMusic.ogg", sampleRate))
	g.Music.SetVolume(0.5)

	// Sound
	*loadingCount++
	howManySounds := 13
	g.Sounds = make([]*Sound, howManySounds)
	for i := 0; i < howManySounds; i++ {
		g.Sounds[i] = &Sound{Volume: 0.7}
	}
	g.Sounds[soundGunShot].AddSound("assets/sfx/Gunshot", sampleRate, context)
	g.Sounds[soundGunReload].AddSound("assets/sfx/Reload", sampleRate, context)
	g.Sounds[soundDogBark].AddSound("assets/sfx/Dog-sound", sampleRate, context, 5)
	g.Sounds[soundPlayerDies].AddSound("assets/sfx/PlayerDies", sampleRate, context)
	g.Sounds[soundHit].AddSound("assets/sfx/Hit", sampleRate, context, 5)
	g.Sounds[soundDryFire].AddSound("assets/sfx/Gun-dry-fire", sampleRate, context)
	g.Sounds[soundZombieScream].AddSound("assets/sfx/Zombie-scream", sampleRate, context)
	g.Sounds[soundZombieGrowl].AddSound("assets/sfx/Zombie-growl", sampleRate, context, 4)
	g.Sounds[soundZombieDeath].AddSound("assets/sfx/Zombie-Death", sampleRate, context, 2)
	g.Sounds[soundBigZombieSound].AddSound("assets/sfx/Big-zombie-sound", sampleRate, context, 4)
	g.Sounds[soundBigZombieDeath1].AddSound("assets/sfx/Big-zombie-death-Phase-1", sampleRate, context)
	g.Sounds[soundBigZombieScream].AddSound("assets/sfx/Big-zombie-scream-Phase-2", sampleRate, context)
	g.Sounds[soundBigZombieDeath2].AddSound("assets/sfx/Big-zombie-death-Phase-2", sampleRate, context)

	// Voices
	howManyVoices := 5
	g.Voices = make([]*Sound, howManyVoices)
	for i := 0; i < howManyVoices; i++ {
		g.Voices[i] = &Sound{Volume: 1}
	}
	g.Voices[voiceCheckpoint].AddSound("assets/voice/Checkpoint", sampleRate, context, 7)
	g.Voices[voiceRespawn].AddSound("assets/voice/Respawn", sampleRate, context, 5)
	g.Voices[voiceKill].AddSound("assets/voice/Kill", sampleRate, context, 6)
	g.Voices[voiceKill].Shuffle()
	g.Voices[voiceFlavour].AddSound("assets/voice/Flavour", sampleRate, context, 12)
	g.Voices[voiceFlavour].Shuffle()
	g.Voices[voiceEndgame].AddSound("assets/voice/Endgame", sampleRate, context)

	// Load sprites
	*loadingCount++
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
	*loadingCount++
	entities := level.LayerByIdentifier("Entities")

	// Add endpoint and nearby outro trigger area
	endpoint := entities.EntityByIdentifier("End")
	g.Space.Add(resolv.NewObject(
		float64(endpoint.Position[0]), float64(endpoint.Position[1]),
		float64(endpoint.Width), float64(endpoint.Height),
		tagEnd,
	))
	g.Space.Add(resolv.NewObject( // much bigger area around the endpoint
		float64(endpoint.Position[0]-endpoint.Width*2), float64(endpoint.Position[1]-endpoint.Height*2),
		float64(endpoint.Width*5), float64(endpoint.Height*5),
		tagOutro,
	))

	// Add player to the game
	playerPosition := entities.EntityByIdentifier("Player").Position
	g.Player = NewPlayer(playerPosition, g.Sprites[spritePlayer])
	g.Space.Add(g.Player.Object)

	for _, e := range entities.Entities {
		if strings.HasPrefix(e.Identifier, "Checkpoint") {
			eid, err := strconv.Atoi(e.Identifier[11:])
			if err != nil {
				log.Printf("Cannot load checkpoint: %s", e.Identifier)
				continue
			}
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

	dogPath := GetBezierPath(pathPoints, 4)

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
		Sprite:   g.Sprites[spriteDog],
		MainPath: &Path{Points: dogPath, NextPoint: 0},
	}
	g.Dog.Init()
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

	g.HUD = NewHUD()
	g.Zoom = NewZoom()

	*loadingCount++
	game.StateLock.Lock()
	game.Loaded = true
	game.Screens[gameRunning] = g
	game.StateLock.Unlock()
}

func (g *GameScreen) Start() {
	g.Music.Play()
	g.Stat.GameStarted = time.Now()
}

// Reset is similar to NewGameScreen but only resets the things that should be
// changed when you reset/restart the game, without reloading all the media
func (g *GameScreen) Reset(game *Game) {
	// How far to spawn dog from player
	dogOffset := 20

	// Load entities from map
	entities := g.LDTKProject.Levels[g.Level].LayerByIdentifier("Entities")

	// Remove zombies
	for i, z := range g.Zombies {
		z.Remove()
		g.Zombies[i] = nil
	}
	g.Zombies = Zombies{}

	// Reset spawnpoints
	for _, s := range g.SpawnPoints {
		s.Reset()
	}

	// Reset some player and dog values
	g.Player.Ammo = playerAmmoClipMax
	startPos := entities.EntityByIdentifier("Player").Position
	if g.Checkpoint > 0 {
		startPos = entities.EntityByIdentifier(
			"Checkpoint_" + strconv.Itoa(g.Checkpoint),
		).Position
	}
	g.Player.Object.X, g.Player.Object.Y = float64(startPos[0]), float64(startPos[1])
	g.Dog.Reset(g.Checkpoint, float64(startPos[0]+dogOffset), float64(startPos[1]))

	g.Music.FadeIn()
	g.Voices[voiceRespawn].Play()
	g.VoiceGuardTime = 0
	g.Zoom = NewZoom()
	game.State = gameRunning
}

func (g *GameScreen) Update() (GameState, error) {
	g.Tick++
	g.VoiceGuardTime++

	// Fade out the black cover
	if g.Tick < fadeOutTime {
		alpha, _ := g.FadeTween.Update(1)
		g.Alpha = uint8(alpha)
	}

	// Pressing R reloads the ammo
	if inpututil.IsKeyJustPressed(ebiten.KeyR) {
		switch g.Player.State {
		case playerShooting, playerReload:
		default:
			g.Player.Reload(g)
		}
	}

	// Gun shooting handler
	if clicked() {
		Shoot(g)
	}

	// Zoom handling
	g.Zoom.Update()
	g.Camera.SetZoom(g.Zoom.Amount)

	// Update player
	g.Player.Update(g)

	// Update dog
	g.Dog.Update(g)

	// Update zombies
	g.Zombies.Update(g)

	// Update spawn points
	g.SpawnPoints.Update(g)

	// Update cursor
	g.Cursor.Update(g)

	// Update music
	g.Music.Update()

	// Collision detection and response between zombie and player
	if collision := g.Player.Object.Check(0, 0, tagMob); collision != nil {
		if g.Player.Object.Overlaps(collision.Objects[0]) {
			if g.Player.Object.Shape.Intersection(0, 0, collision.Objects[0].Shape) != nil {
				g.Music.Pause()
				g.Sounds[soundPlayerDies].Play()
				g.Stat.CounterPlayerDied++
				return gameOver, nil // return early, no point in continuing, you are dead
			}
		}
	}

	// Do something special when you find a Checkpoint entity
	if collision := g.Player.Object.Check(0, 0, tagCheckpoint); collision != nil {
		if o := collision.Objects[0]; g.Player.Object.Overlaps(o) {
			if g.Checkpoint < o.Data.(int) {
				if g.Dog.State == dogNormalSniffing || g.Dog.State == dogNormalWaitingAtCheckpoint {
					g.Checkpoint = o.Data.(int)
					g.Voices[voiceCheckpoint].PlayVariant(g.Checkpoint - 1)
					g.VoiceGuardTime = 0
					g.NextVoiceStep = voiceStepFlavour1
					g.Dog.ContinueFromCheckpoint()
				}
			}

		}
	}

	// End game when you reach the tunnel after defeating the boss zombie
	if g.BossDefeated {
		if collision := g.Player.Object.Check(0, 0, tagEnd); collision != nil {
			if g.Player.Object.Overlaps(collision.Objects[0]) {
				return gameWon, nil
			}
		}
		if collision := g.Player.Object.Check(0, 0, tagOutro); collision != nil {
			if g.Player.Object.Overlaps(collision.Objects[0]) {
				if g.Voices[voiceEndgame].LastPlayed == nil { // only once
					g.Music.FadeOut()
					g.Voices[voiceEndgame].Play()
				}
			}
		}
	}

	// Collision detection and response between zombie and dog
	if collision := g.Dog.Object.Check(0, 0, tagMob); collision != nil {
		if g.Dog.Object.Overlaps(collision.Objects[0]) {
			g.Dog.Mode = dogDead
		}
	}

	// Game over if the dog dies
	if g.Dog.Mode == dogDead {
		g.Stat.CounterDogDied++
		g.Music.Pause()
		return gameOver, nil
	}

	// Position camera and clamp in to the Map dimensions
	level := g.LDTKProject.Levels[g.Level]
	g.Camera.SetPosition(
		math.Min(math.Max(g.Player.Object.X, float64(g.Width)/2), float64(level.Width)-float64(g.Width)/2),
		math.Min(math.Max(g.Player.Object.Y, float64(g.Height)/2), float64(level.Height)-float64(g.Height)/2),
	)

	// Retroactively unstick object that collide from small rotations
	if collision := g.Player.Object.Check(0, 0); collision != nil {
		for _, o := range collision.Objects {
			if cs := g.Player.Object.Shape.Intersection(0, 0, o.Shape); cs != nil {
				g.Player.Object.X += cs.MTV.X()
				g.Player.Object.Y += cs.MTV.Y()
			}
		}
	}

	return gameRunning, nil
}

func (g *GameScreen) Draw(screen *ebiten.Image) {
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

	g.HUD.Draw(g.Player.Ammo, screen)

	if g.Player.State != playerReload {
		g.Cursor.Draw(screen)
	}

	// Fading out black cover
	if g.Tick < fadeOutTime {
		ebitenutil.DrawRect(screen, 0, 0, float64(g.Width), float64(g.Height), color.RGBA{0, 0, 0, g.Alpha})
	}

	g.Debuggers.Debug(g, screen)
}

// Clicked is shorthand for when the left mouse button has just been clicked
func clicked() bool {
	return inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft)
}

// Clicked is shorthand for when the right mouse button has just been clicked
func clickedRight() bool {
	return inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonRight)
}

// Shoot sets shooting states and also die states for any zombies in range
func Shoot(g *GameScreen) {
	interruptReload := func() {
		g.Sounds[soundGunReload].Pause()
		g.Sounds[soundDryFire].Play()
		g.Player.State = playerDryFire
		g.Stat.CounterDryFires++
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

		g.Sounds[soundGunShot].Play()

		g.Stat.CounterBulletsFired++
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
					g.Cursor.Hit = true
					o.Data.(*Zombie).Hit(g)
					return // stop at the first zombie
				} else {
					g.Cursor.Hit = false
				}
			}
		}
	}
}

// CalcObjectDistance calculates the distance between two Objects
func CalcObjectDistance(obj1, obj2 *Coord) (float64, float64, float64) {
	return CalcDistance(obj1.X, obj1.Y, obj2.X, obj2.Y), obj1.X - obj2.X, obj1.Y - obj2.Y
}

// CalcDistance calculates the distance between two coordinates
func CalcDistance(x1, y1, x2, y2 float64) float64 {
	return math.Sqrt(math.Pow(x1-x2, 2) + math.Pow(y1-y2, 2))
}

// NormalizeVector normalizes the vector
func NormalizeVector(vector Coord) Coord {
	magnitude := CalcDistance(vector.X, vector.Y, 0, 0)
	return Coord{X: vector.X / magnitude, Y: vector.Y / magnitude}
}
