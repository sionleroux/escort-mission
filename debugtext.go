// Use of this source code is subject to an MIT-style
// licence which can be found in the LICENSE file.

//go:build !release

package main

import (
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

func init() {
	debuggers.Add(DebugFunc(DebugText))

	// Uncap FPS so you can see if a code change has had an impact on
	// the game's performance
	ebiten.SetFPSMode(ebiten.FPSModeVsyncOffMaximum)
}

// DebugText prints out general debug information as text
func DebugText(g *GameScreen, screen *ebiten.Image) {
	ebitenutil.DebugPrint(screen, fmt.Sprintf(
		"FPS: %.2f\n"+
			"TPS: %.2f\n"+
			"X: %.2f\n"+
			"Y: %.2f\n"+
			"Zombies: %d\n"+
			"Progress: %.2f%%\n",
		ebiten.ActualFPS(),
		ebiten.ActualTPS(),
		g.Player.Object.Position.X/32,
		g.Player.Object.Position.Y/32,
		len(g.Zombies),
		float64(g.Dog.MainPath.NextPoint)/float64(len(g.Dog.MainPath.Points))*100,
	))
}
