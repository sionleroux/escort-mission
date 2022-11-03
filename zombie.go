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
	dy := -1.0
	if collision := z.Object.Check(0, dy, "mob", "wall"); collision == nil {
		z.Object.Y += dy
	}
}

// MoveDown moves the player downwards
func (z *Zombie) MoveDown() {
	dy := 1.0
	if collision := z.Object.Check(0, dy, "mob", "wall"); collision == nil {
		z.Object.Y += dy
	}
}

// MoveLeft moves the player left
func (z *Zombie) MoveLeft() {
	dx := -1.0
	if collision := z.Object.Check(dx, 0, "mob", "wall"); collision == nil {
		z.Object.X += dx
	}
}

// MoveRight moves the player right
func (z *Zombie) MoveRight() {
	dx := 1.0
	if collision := z.Object.Check(dx, 0, "mob", "wall"); collision == nil {
		z.Object.X += dx
	}
}
