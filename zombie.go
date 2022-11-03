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
	z.move(0, -1)
}

// MoveDown moves the player downwards
func (z *Zombie) MoveDown() {
	z.move(0, 1)
}

// MoveLeft moves the player left
func (z *Zombie) MoveLeft() {
	z.move(-1, 0)
}

// MoveRight moves the player right
func (z *Zombie) MoveRight() {
	z.move(1, 0)
}

func (z *Zombie) move(dx, dy float64) {
	if collision := z.Object.Check(dx, dy, "mob", "wall"); collision == nil {
		z.Object.X += dx
		z.Object.Y += dy
	}
}
