// Use of this source code is subject to an MIT-style
// licence which can be found in the LICENSE file.

//go:build !release && debugaim

package main

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

func init() {
	debuggers.Add(DebugFunc(DebugAim))
}

// DebugAim draws a line showing the direction and range of the gun
func DebugAim(g *GameScreen, screen *ebiten.Image) {
	rangeOfFire := g.Player.Range
	sX, sY := g.Camera.GetScreenCoords(
		g.Player.Object.Position.X-math.Cos(g.Player.Angle-math.Pi)*rangeOfFire,
		g.Player.Object.Position.Y-math.Sin(g.Player.Angle-math.Pi)*rangeOfFire,
	)
	pX, pY := g.Camera.GetScreenCoords(
		g.Player.Object.Position.X,
		g.Player.Object.Position.Y,
	)
	ebitenutil.DrawLine(screen, pX, pY, sX, sY, color.Black)
}
