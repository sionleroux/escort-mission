// Copyright 2021 Siôn le Roux.  All rights reserved.
// Use of this source code is subject to an MIT-style
// licence which can be found in the LICENSE file.

package main

import (
	"errors"
	"image"
	"log"
	"math"
	"math/rand"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/solarlune/resolv"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

// zombieSpeed is the distance the zombie moves per update cycle
const zombieSpeed float64 = 0.4

// zombieRange is how far away the zombie sees something to attack
const zombieRange float64 = 200

// Zombies is an array of Zombie
type Zombies []*Zombie

// Update updates all the zombies
func (zs *Zombies) Update(g *Game) {
	for i, z := range *zs {
		err := z.Update(g)
		if err != nil {
			// clear and remove dead zombies
			log.Println(err)
			g.Zombies[i] = nil
			g.Zombies = append((*zs)[:i], (*zs)[i+1:]...)
		}
	}
}

// Draw draws all the zombies
func (zs Zombies) Draw(g *Game) {
	for _, z := range zs {
		z.Draw(g)
	}
}

// List of possible zombie states
const (
	zombieIdle    = iota // Doesn't have any target to attack
	zombieWalking        // Walking in some direction
	zombieHit            // Hit by a shot, but not deadly
	zombieDeath          // Plays the death animation
	zombieDead           // Marked as dead, will be removed on next Update
)

// Zombie is a monster that's trying to eat the player character
type Zombie struct {
	Object    *resolv.Object // Used for collision detection with other objects
	Angle     float64        // The angle the zombies is facing at
	Frame     int            // The current animation frame
	State     int            // The current animation state
	Sprite    *SpriteSheet   // Used for zombie animations
	Speed     float64        // The speed this zombie walks at
	Target    *resolv.Object // Target object (player or dog)
	HitToDie  int            // Number of hits needed to die
}

// NewZombie constructs a new Zombie object
func NewZombie(position Coord, sprites *SpriteSheet) *Zombie {
	dimensions := sprites.Sprite[0].Position
	zombie := &Zombie{
		Object: resolv.NewObject(
			position.X, position.Y,
			float64(dimensions.W), float64(dimensions.H),
			tagMob,
		),
		Angle:    0,
		Sprite:   sprites,
		Speed:    zombieSpeed * (1 + rand.Float64()),
		HitToDie: 1 + rand.Intn(2),
	}
	zombie.Object.Data = zombie // self-reference for later
	return zombie
}

// Update updates the state of the zombie
func (z *Zombie) Update(g *Game) error {
	if z.State == zombieDead {
		return errors.New("Zombie died")
	}

	if z.State == zombieDeath || z.State == zombieHit {
		z.animate(g)
		return nil
	}

	playerDistance, _, _ := CalcObjectDistance(z.Object, g.Player.Object)
	dogDistance, _, _ := CalcObjectDistance(z.Object, g.Dog.Object)

	if playerDistance < zombieRange {
		z.Target = g.Player.Object
		z.walk()
	} else if dogDistance < zombieRange * 1.2 {
		z.Target = g.Dog.Object
		z.walk()
	} else {
		z.State = zombieIdle
	}

	z.animate(g)
	z.Object.Update()
	return nil
}

func (z *Zombie) walk() {
	// Zombies rotate towards their target
	adjacent := z.Target.X - z.Object.X
	opposite := z.Target.Y - z.Object.Y
	z.Angle = math.Atan2(opposite, adjacent)

	// Zombie movement logic
	// TODO: this could be simplified using maths
	if z.Object.X < z.Target.X {
		z.MoveRight()
	}
	if z.Object.X > z.Target.X {
		z.MoveLeft()
	}
	if z.Object.Y < z.Target.Y {
		z.MoveDown()
	}
	if z.Object.Y > z.Target.Y {
		z.MoveUp()
	}
}

func (z *Zombie) animate(g *Game) {
	// Update only in every 5th cycle
	if g.Tick%5 != 0 {
		return
	}

	// No states at the moment, zombies are always walking
	ft := z.Sprite.Meta.FrameTags[z.State]

	if ft.From == ft.To {
		z.Frame = ft.From
	} else {
		// Contiuously increase the Frame counter between From and To
		z.Frame = (z.Frame-ft.From+1)%(ft.To-ft.From+1) + ft.From
	}

	// Set as walking after hit animation
	if z.State == zombieHit && z.Frame == ft.To {
		z.State = zombieWalking
	}

	// Set as dead after death animation
	if z.State == zombieDeath && z.Frame == ft.To {
		z.State = zombieDead
	}
}

// MoveUp moves the zombie upwards
func (z *Zombie) MoveUp() {
	z.move(0, -z.Speed)
}

// MoveDown moves the zombie downwards
func (z *Zombie) MoveDown() {
	z.move(0, z.Speed)
}

// MoveLeft moves the zombie left
func (z *Zombie) MoveLeft() {
	z.move(-z.Speed, 0)
}

// MoveRight moves the zombie right
func (z *Zombie) MoveRight() {
	z.move(z.Speed, 0)
}

// Move the Zombie by the given vector if it is possible to do so
func (z *Zombie) move(dx, dy float64) {
	z.State = zombieWalking
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
			float64(z.Object.Y)+float64(frame.Position.H/2),
		),
	)

}

// Hit changes zombie state and updates game data in response to it getting shot
func (z *Zombie) Hit() {
	z.State = zombieHit
	z.HitToDie--
	if z.HitToDie == 0 {
		z.Die()
	}
}

// Die changes zombie state and updates game data in case of a deadly shot
func (z *Zombie) Die() {
	z.Object.Space.Remove(z.Object)
	z.State = zombieDeath
}
