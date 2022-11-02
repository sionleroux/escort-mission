// Copyright 2021 Siôn le Roux.  All rights reserved.
// Use of this source code is subject to an MIT-style
// licence which can be found in the LICENSE file.

package main

import "image"

// Player is the player character in the game
type Player struct {
	Coords image.Point
}

// MoveUp moves the player upwards
func (p *Player) MoveUp() {
	p.Coords.Y--
}

// MoveDown moves the player downwards
func (p *Player) MoveDown() {
	p.Coords.Y++
}

// MoveLeft moves the player left
func (p *Player) MoveLeft() {
	p.Coords.X--
}

// MoveRight moves the player right
func (p *Player) MoveRight() {
	p.Coords.X++
}
