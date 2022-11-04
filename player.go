// Copyright 2021 Siôn le Roux.  All rights reserved.
// Use of this source code is subject to an MIT-style
// licence which can be found in the LICENSE file.

package main

import (
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
