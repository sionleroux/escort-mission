// Use of this source code is subject to an MIT-style
// licence which can be found in the LICENSE file.

package main

import (
	"errors"
	"image"
	"log"
	"sync"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
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

	game := &Game{
		Width:     gameWidth,
		Height:    gameHeight,
		Stat:      &Stat{},
		StateLock: &sync.RWMutex{},
	}
	loadingScreen := NewLoadingScreen()
	game.Screens = []Screen{
		loadingScreen,
		NewStartScreen(game),
		NewIntroScreen(game),
		&GameScreen{},
		NewDeathScreen(game),
		NewWinScreen(game),
	}

	go NewGameScreen(game, loadingScreen.Counter)

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
	StateLock  *sync.RWMutex
	Loaded     bool
	DeathTime  int
	Tick       int
	Checkpoint int
	Stat       *Stat
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

	g.StateLock.Lock()
	if g.State == gameLoading && g.Loaded {
		g.Screens[gameLoading].(*LoadingScreen).Loaded = g.Loaded
	}
	g.StateLock.Unlock()

	prevState := g.State
	state, err := g.Screens[g.State].Update()
	g.State = state

	if errors.Is(err, ErrorDoneLoading) {
		if startingCheckpoint != 0 {
			g.Screens[gameRunning].(*GameScreen).Checkpoint = startingCheckpoint
			g.State = gameOver
		} else {
			g.State = gameStart
		}
		return nil
	}

	if prevState != gameRunning && g.State == gameRunning {
		g.Screens[gameRunning].(*GameScreen).Start()
	}
	if prevState != gameWon && g.State == gameWon {
		g.Stat.GameWon = time.Now()
	}

	switch g.State {
	case gameOver:
		if g.DeathTime == 0 {
			g.DeathTime = g.Tick
			g.Screens[gameOver].(*DeathScreen).DogDied = (g.Screens[gameRunning].(*GameScreen).Dog.Mode == dogDead)
		}
		if g.Tick-g.DeathTime > deathCoolDownTime {
			g.Checkpoint = g.Screens[gameRunning].(*GameScreen).Checkpoint
			g.DeathTime = 0
			g.Screens[gameOver] = NewDeathScreen(g)
			g.Screens[gameRunning].(*GameScreen).Reset(g)
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
