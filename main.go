// Use of this source code is subject to an MIT-style
// licence which can be found in the LICENSE file.

package main

import (
	"errors"
	"image"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"gopkg.in/ini.v1"
)

const gameWidth, gameHeight = 320, 240

var deathCoolDownTime = 4 * 60

const sampleRate int = 44100 // assuming "normal" sample rate
var context *audio.Context

func main() {
	ebiten.SetWindowSize(gameWidth*2, gameHeight*2)
	ebiten.SetWindowTitle("eZcort mission")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	ebiten.SetWindowIcon([]image.Image{loadImage("assets/icon.png")})
	ebiten.SetFPSMode(ebiten.FPSModeVsyncOffMaximum)

	context = audio.NewContext(sampleRate)

	ApplyConfigs()

	game := &Game{Width: gameWidth, Height: gameHeight}
	game.Screens = []Screen{
		&LoadingScreen{},
		&StartScreen{},
		NewIntroScreen(game),
		&GameScreen{},
		NewDeathScreen(game),
		&WinScreen{},
	}

	go NewGameScreen(game)

	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}

// GameState is global state for the whole game
type GameState int

const (
	gameLoading GameState = iota // Assets are being loaded
	gameStart                    // Game start screen is shown
	gameIntro                    // Intro is played before game is started
	gameRunning                  // The game is running the main game code
	gameOver                     // The game has ended because you died
	gameWon                      // The game has ended because you won
)

// Game represents the main game state
type Game struct {
	Width      int
	Height     int
	Screens    []Screen
	State      GameState
	Loaded     bool
	DeathTime  int
	Tick       int
	Checkpoint int
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

	prevState := g.State
	state, err := g.Screens[g.State].Update()
	g.State = state

	if prevState != gameRunning && g.State == gameRunning {
		g.Screens[gameRunning].(*GameScreen).Start()
	}

	switch g.State {
	case gameOver:
		if g.DeathTime == 0 {
			g.DeathTime = g.Tick
			g.Screens[gameOver].(*DeathScreen).DogDied = (g.Screens[gameRunning].(*GameScreen).Dog.Mode == dogDead)
		}
		if g.Tick-g.DeathTime > deathCoolDownTime {
			g.Checkpoint = g.Screens[gameRunning].(*GameScreen).Checkpoint
			g.State = gameRunning
			g.DeathTime = 0
			go g.Screens[gameRunning].(*GameScreen).Reset(g)
		}
	}

	return err
}

// Draw draws the game screen by one frame
func (g *Game) Draw(screen *ebiten.Image) {
	g.Screens[g.State].Draw(screen)
}

// Screen is a full-screen UI Screen for some part of the game like a menu or a
// game level
type Screen interface {
	Update() (GameState, error)
	Draw(screen *ebiten.Image)
}

// ApplyConfigs overrides default values with a config file if available
func ApplyConfigs() {
	log.Println("Looking for INI file...")
	cfg, err := ini.Load("escort-mission.ini")
	if err != nil {
		log.Println("Error parsing INI file:", err)
		return
	}
	deathCoolDownTime, err = cfg.Section("").Key("DeathCoolDownTime").Int()
	if err != nil {
		log.Println("Error parsing INI file:", err)
	}
	hudPadding, err = cfg.Section("").Key("HudPadding").Int()
	if err != nil {
		log.Println("Error parsing INI file:", err)
	}
	startingCheckpoint, err = cfg.Section("").Key("StartingCheckpoint").Int()
	if err != nil {
		log.Println("Error parsing INI file:", err)
	}
	playerSpeed, err = cfg.Section("Player").Key("PlayerSpeed").Float64()
	if err != nil {
		log.Println("Error parsing INI file:", err)
	}
	playerSpeedFactorReverse, err = cfg.Section("Player").Key("PlayerSpeedFactorReverse").Float64()
	if err != nil {
		log.Println("Error parsing INI file:", err)
	}
	playerSpeedFactorSideways, err = cfg.Section("Player").Key("PlayerSpeedFactorSideways").Float64()
	if err != nil {
		log.Println("Error parsing INI file:", err)
	}
	playerSpeedFactorSprint, err = cfg.Section("Player").Key("PlayerSpeedFactorSprint").Float64()
	if err != nil {
		log.Println("Error parsing INI file:", err)
	}
	playerAmmoClipMax, err = cfg.Section("Player").Key("PlayerAmmoClipMax").Int()
	if err != nil {
		log.Println("Error parsing INI file:", err)
	}
	zombieSpeed, err = cfg.Section("Zombie").Key("ZombieSpeed").Float64()
	if err != nil {
		log.Println("Error parsing INI file:", err)
	}
	zombieCrawlerSpeed, err = cfg.Section("Zombie").Key("ZombieCrawlerSpeed").Float64()
	if err != nil {
		log.Println("Error parsing INI file:", err)
	}
	zombieSprinterSpeed, err = cfg.Section("Zombie").Key("ZombieSprinterSpeed").Float64()
	if err != nil {
		log.Println("Error parsing INI file:", err)
	}
	zombieRange, err = cfg.Section("Zombie").Key("ZombieRange").Float64()
	if err != nil {
		log.Println("Error parsing INI file:", err)
	}
	dogWalkingSpeed, err = cfg.Section("Dog").Key("DogWalkingSpeed").Float64()
	if err != nil {
		log.Println("Error parsing INI file:", err)
	}
	dogRunningSpeed, err = cfg.Section("Dog").Key("DogRunningSpeed").Float64()
	if err != nil {
		log.Println("Error parsing INI file:", err)
	}
	waitingRadius, err = cfg.Section("Dog").Key("WaitingRadius").Float64()
	if err != nil {
		log.Println("Error parsing INI file:", err)
	}
	followingRadius, err = cfg.Section("Dog").Key("FollowingRadius").Float64()
	if err != nil {
		log.Println("Error parsing INI file:", err)
	}
	zombieBarkRadius, err = cfg.Section("Dog").Key("ZombieBarkRadius").Float64()
	if err != nil {
		log.Println("Error parsing INI file:", err)
	}
	zombieFleeRadius, err = cfg.Section("Dog").Key("ZombieFleeRadius").Float64()
	if err != nil {
		log.Println("Error parsing INI file:", err)
	}
	zombieSafeRadius, err = cfg.Section("Dog").Key("ZombieSafeRadius").Float64()
	if err != nil {
		log.Println("Error parsing INI file:", err)
	}
	fleeingPathLength, err = cfg.Section("Dog").Key("FleeingPathLength").Float64()
	if err != nil {
		log.Println("Error parsing INI file:", err)
	}
}
