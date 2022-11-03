// Copyright 2021 Siôn le Roux.  All rights reserved.
// Use of this source code is subject to an MIT-style
// licence which can be found in the LICENSE file.

package main

import (
	"github.com/solarlune/resolv"
)

// Zombie is the player character in the game
type Zombie struct {
	Object *resolv.Object
	Angle  float64
}

// MoveUp moves the player upwards
func (z *Zombie) MoveUp() {
	z.Object.Y--
}

// MoveDown moves the player downwards
func (z *Zombie) MoveDown() {
	z.Object.Y++
}

// MoveLeft moves the player left
func (z *Zombie) MoveLeft() {
	z.Object.X--
}

// MoveRight moves the player right
func (z *Zombie) MoveRight() {
	z.Object.X++
}
