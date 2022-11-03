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
func (z *Zombie) MoveUp() {
	z.Coords.Y--
}

// MoveDown moves the player downwards
func (z *Zombie) MoveDown() {
	z.Coords.Y++
}

// MoveLeft moves the player left
func (z *Zombie) MoveLeft() {
	z.Coords.X--
}

// MoveRight moves the player right
func (z *Zombie) MoveRight() {
	z.Coords.X++
}
