// Copyright 2021 Siôn le Roux.  All rights reserved.
// Use of this source code is subject to an MIT-style
// licence which can be found in the LICENSE file.

//go:build !release && debugcol

package main

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/solarlune/resolv"
)

func init() {
	debuggers.Add(DebugFunc(DebugCollision))
}

// DebugCollision draws boxes around objects in collision space to easily
// visualise how they move and collide
func DebugCollision(g *GameScreen, screen *ebiten.Image) {
	debugPosition(g, screen, g.Player.Object)
	debugPosition(g, screen, g.Dog.Object)
	for _, z := range g.Zombies {
		debugPosition(g, screen, z.Object)
	}
	for _, o := range g.Space.Objects() {
		if o.HasTags(tagWall) {
			debugPosition(g, screen, o)
		}
	}
}

func debugPosition(g *GameScreen, screen *ebiten.Image, o *resolv.Object) {
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
