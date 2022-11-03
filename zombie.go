// Copyright 2021 Siôn le Roux.  All rights reserved.
// Use of this source code is subject to an MIT-style
// licence which can be found in the LICENSE file.

package main

import "image"

// Zombie is the player character in the game
type Zombie struct {
	Coords image.Point
	Angle  float64
}

// MoveUp moves the player upwards
func (p *Zombie) MoveUp() {
	p.Coords.Y--
}

// MoveDown moves the player downwards
func (p *Zombie) MoveDown() {
	p.Coords.Y++
}

// MoveLeft moves the player left
func (p *Zombie) MoveLeft() {
	p.Coords.X--
}

// MoveRight moves the player right
func (p *Zombie) MoveRight() {
	p.Coords.X++
}
