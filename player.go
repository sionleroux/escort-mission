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
}

// MoveUp moves the player upwards
func (p *Player) MoveUp() {
	p.Object.Y -= playerSpeed
}

// MoveDown moves the player downwards
func (p *Player) MoveDown() {
	p.Object.Y += playerSpeed
}

// MoveLeft moves the player left
func (p *Player) MoveLeft() {
	p.Object.X -= playerSpeed
}

// MoveRight moves the player right
func (p *Player) MoveRight() {
	p.Object.X += playerSpeed
}
