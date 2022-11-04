// Copyright 2021 Siôn le Roux.  All rights reserved.
// Use of this source code is subject to an MIT-style
// licence which can be found in the LICENSE file.

package main

import (
	"image"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/solarlune/resolv"
)

// playerSpeed is the distance the player moves per update cycle
const playerSpeed float64 = 2

// Player is the player character in the game
type Player struct {
	Object *resolv.Object
	Angle  float64
	Sprite *SpriteSheet
}

// MoveUp moves the player upwards
func (p *Player) MoveUp() {
	p.move(0, -playerSpeed)
}

// MoveDown moves the player downwards
func (p *Player) MoveDown() {
	p.move(0, playerSpeed)
}

// MoveLeft moves the player left
func (p *Player) MoveLeft() {
	p.move(-playerSpeed, 0)
}

// MoveRight moves the player right
func (p *Player) MoveRight() {
	p.move(playerSpeed, 0)
}

// Move the Player by the given vector if it is possible to do so
func (p *Player) move(dx, dy float64) {
	if collision := p.Object.Check(dx, dy, "wall"); collision == nil {
		p.Object.X += dx
		p.Object.Y += dy
	}
}

// Draw draws the Player to the screen
func (p *Player) Draw(g *Game, screen *ebiten.Image) {
	s := p.Sprite
	frame := s.Sprite[0]
	op := &ebiten.DrawImageOptions{}

	op.GeoM.Translate(
		float64(-frame.Position.W/2),
		float64(-frame.Position.H/2),
	)

	op.GeoM.Rotate(p.Angle + math.Pi/2)

	op.GeoM.Translate(
		float64(p.Object.X),
		float64(p.Object.Y),
	)

	screen.DrawImage(s.Image.SubImage(image.Rect(
		frame.Position.X,
		frame.Position.Y,
		frame.Position.X+frame.Position.W,
		frame.Position.Y+frame.Position.H,
	)).(*ebiten.Image), op)
}