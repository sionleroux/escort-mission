// Copyright 2021 Siôn le Roux.  All rights reserved.
// Use of this source code is subject to an MIT-style
// licence which can be found in the LICENSE file.

package main

import (
	"errors"
	"image"
	"log"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/solarlune/resolv"
)

// zombieSpeed is the distance the zombie moves per update cycle
const zombieSpeed float64 = 0.2

// Zombies is an array of Zombie
type Zombies []*Zombie

// Update updates all the zombies
func (zs Zombies) Update(g *Game) {
	for i, z := range zs {
		err := z.Update(g)
		if err != nil {
			// clear and remove dead zombies
			log.Println(err)
			g.Zombies[i] = nil
			g.Zombies = append(zs[:i], zs[i+1:]...)
		}
	}
}

// List of possible zombie states
const (
	zombieAlive = iota
	zombieDead
)

// Zombie is a monster that's trying to eat the player character
type Zombie struct {
	Object *resolv.Object
	Angle  float64
	Frame  int
	State  int
	Sprite *SpriteSheet
}

// Update updates the state of the zombie
func (z *Zombie) Update(g *Game) error {
	if z.State == zombieDead {
		return errors.New("Zombie died")
	}

	// Zombies rotate towards player
	adjacent := z.Object.X - g.Player.Object.X
	opposite := z.Object.Y - g.Player.Object.Y
	z.Angle = math.Atan2(opposite, adjacent)

	// Zombie movement logic
	// TODO: this could be simplified using maths
	if z.Object.X < g.Player.Object.X {
		z.MoveRight()
	}
	if z.Object.X > g.Player.Object.X {
		z.MoveLeft()
	}
	if z.Object.Y < g.Player.Object.Y {
		z.MoveDown()
	}
	if z.Object.Y > g.Player.Object.Y {
		z.MoveUp()
	}

	z.animate(g)
	z.Object.Update()
	return nil
}

func (z *Zombie) animate(g *Game) {
	// Update only in every 5th cycle
	if g.Tick%5 != 0 {
		return
	}

	// No states at the moment, zombies are always walking
	ft := z.Sprite.Meta.FrameTags[1]

	if ft.From == ft.To {
		z.Frame = ft.From
	} else {
		// Contiuously increase the Frame counter between From and To
		z.Frame = (z.Frame-ft.From+1)%(ft.To-ft.From+1) + ft.From
	}
}

// MoveUp moves the zombie upwards
func (z *Zombie) MoveUp() {
	z.move(0, -zombieSpeed)
}

// MoveDown moves the zombie downwards
func (z *Zombie) MoveDown() {
	z.move(0, zombieSpeed)
}

// MoveLeft moves the zombie left
func (z *Zombie) MoveLeft() {
	z.move(-zombieSpeed, 0)
}

// MoveRight moves the zombie right
func (z *Zombie) MoveRight() {
	z.move(zombieSpeed, 0)
}

// Move the Zombie by the given vector if it is possible to do so
func (z *Zombie) move(dx, dy float64) {
	if collision := z.Object.Check(dx, dy, tagMob, tagWall); collision == nil {
		z.Object.X += dx
		z.Object.Y += dy
	}
}

// Draw draws the Zombie to the screen
func (z *Zombie) Draw(g *Game) {
	s := z.Sprite
	frame := s.Sprite[z.Frame]
	op := &ebiten.DrawImageOptions{}

	op.GeoM.Translate(
		float64(-frame.Position.W/2),
		float64(-frame.Position.H/2),
	)

	op.GeoM.Rotate(z.Angle + math.Pi/2)

	g.Camera.Surface.DrawImage(
		s.Image.SubImage(image.Rect(
			frame.Position.X,
			frame.Position.Y,
			frame.Position.X+frame.Position.W,
			frame.Position.Y+frame.Position.H,
		)).(*ebiten.Image),
		g.Camera.GetTranslation(
			op,
			float64(z.Object.X)+float64(frame.Position.W/2),
			float64(z.Object.Y)+float64(frame.Position.H/2)))

}

// Die changes zombie state and updates game data in response to it getting shot
func (z *Zombie) Die() {
	z.Object.Space.Remove(z.Object)
	z.State = zombieDead
}
