package main

import (
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

// Debugger provides debug information by rendering it on-screen
type Debugger interface {
	Debug(g *Game, screen *ebiten.Image)
}

// DebugFunc implements the Debugger interface
type DebugFunc func(g *Game, screen *ebiten.Image)

// Debug calls the debug callback with the provided arguments
func (f DebugFunc) Debug(g *Game, screen *ebiten.Image) {
	f(g, screen)
}

// DebugText prints out general debug information as text
func DebugText(g *Game, screen *ebiten.Image) {
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
