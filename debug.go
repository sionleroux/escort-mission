package main

import (
	"fmt"
	"image/color"
	"math"

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

// Debuggers is a slice of a Debugger to make it easier to handle many debuggers
type Debuggers []Debugger

// Debug passes on the Debug call to all its child Debuggers
func (ds Debuggers) Debug(g *Game, screen *ebiten.Image) {
	for _, d := range ds {
		d.Debug(g, screen)
	}
}

// Add is a shorthand for adding a child Debugger to the Debuggers
func (ds *Debuggers) Add(d Debugger) {
	*ds = append(*ds, d)
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

// DebugAim draws a line showing the direction and range of the gun
func DebugAim(g *Game, screen *ebiten.Image) {
	rangeOfFire := g.Player.Range
	sX, sY := g.Camera.GetScreenCoords(
		g.Player.Object.X-math.Cos(g.Player.Angle-math.Pi)*rangeOfFire,
		g.Player.Object.Y-math.Sin(g.Player.Angle-math.Pi)*rangeOfFire,
	)
	pX, pY := g.Camera.GetScreenCoords(
		g.Player.Object.X+g.Player.Object.W/2,
		g.Player.Object.Y+g.Player.Object.H/2,
	)
	ebitenutil.DrawLine(screen, pX, pY, sX, sY, color.Black)
}
