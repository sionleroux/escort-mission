// Copyright 2021 Siôn le Roux.  All rights reserved.
// Use of this source code is subject to an MIT-style
// licence which can be found in the LICENSE file.

package main

import (
	"image"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/solarlune/resolv"
)

// Coord is the coordinate of a path point
type Coord struct {
	X float64
	Y float64
}

// Path is an array of coordinates
type Path []Coord

// dogSpeed is the distance the dog moves per update cycle
const dogSpeed float64 = 1

// waitingRadius is the maximum distance the dog walks away from the player
const waitingRadius float64 = 96

// zombieDangerRadius: if a zombie is this close to the dog, it runs away
const zombieDangerRadius float64 = 128

// zombieSafeRadius: if a zombie is at least this far from the dog, it stops running
const zombieSafeRadius float64 = 192

// states of the dog
// It would be great to map them to the frameTag.Name from JSON
const (
	dogWalking  int = 0
	dogSniffing     = 1
	dogSitting      = 2
)

// Dog is player's companion
type Dog struct {
	Object   *resolv.Object
	Angle    float64
	Speed    float64
	Frame    int
	State    int
	Path     Path
	NextPath int
	Sprite   *SpriteSheet
	InDanger bool
}

// Update updates the state of the dog
func (d *Dog) Update(g *Game) {
	isSafeAgain := true
	resultantVector := Coord{
		X: 0,
		Y: 0,
	}
	for _, zombie := range g.Zombies {
		zombieDistance, xDistance, yDistance := CalcObjectDistance(d.Object, zombie.Object)
		if zombieDistance < zombieDangerRadius {
			// If zombies are too close then the dog is in danger and will run away
			resultantVector.X += xDistance
			resultantVector.Y += yDistance
			isSafeAgain = false
			d.InDanger = true
			d.State = dogWalking
		} else if d.InDanger && zombieDistance < zombieSafeRadius {
			// If the dog is running away from zombies then it will be safe again when getting far enough from the zombies
			isSafeAgain = false
		}
	}

	// If the dog is not in danger anymore then it turn towards the next path point
	if d.InDanger && isSafeAgain {
		d.TurnTowardsPathPoint()
	}
	d.InDanger = !isSafeAgain

	if (!d.InDanger) {
		playerDistance, _, _ := CalcObjectDistance(d.Object, g.Player.Object)
		if playerDistance < waitingRadius {
			// If the dog is not in danger and it is close to the player then it sniffs towards next path point
			d.State = dogSniffing
			d.FollowPath()
		} else {
		// If the player is not close enough then the dog sits down
		d.State = dogSitting
		}
	} else {
		// If the dog is in danger then it runs away from the zombies
		d.TurnTowardsCoordinate(resultantVector)
		d.Run()
	}

	d.animate(g)
	d.Object.Update()
}

func (d *Dog) animate(g *Game) {
	// Update only in every 5th cycle
	if g.Tick%5 != 0 {
		return
	}

	ft := d.Sprite.Meta.FrameTags[d.State]

	if ft.From == ft.To {
		d.Frame = ft.From
	} else {
		// Contiuously increase the Frame counter between From and To
		d.Frame = (d.Frame-ft.From+1)%(ft.To-ft.From+1) + ft.From
	}
}

// TurnTowardsCoordinate turns the dog to the direction of the point
func (d *Dog) TurnTowardsCoordinate(coord Coord) {
	adjacent := d.Object.X - coord.X
	opposite := d.Object.Y - coord.Y
	// math.Pi is needed only until the dog sprite is looking up
	d.Angle = math.Atan2(opposite, adjacent)+math.Pi/2
}

// TurnTowardsPathPoint turns the dog to the direction of the next path point
func (d *Dog) TurnTowardsPathPoint() {
	adjacent := d.Object.X - float64(d.Path[d.NextPath].X)
	opposite := d.Object.Y - float64(d.Path[d.NextPath].Y)
	// math.Pi is needed only until the dog sprite is looking up
	d.Angle = math.Atan2(opposite, adjacent)-math.Pi
}

// FollowPath moves the dog along the path
func (d *Dog) FollowPath() {
	if d.NextPath < 0 {
		d.NextPath = 0
		d.TurnTowardsPathPoint()
	}

	nextPathCoordDistance := CalcDistance(d.Path[d.NextPath].X, d.Path[d.NextPath].Y, d.Object.X, d.Object.Y)
	if nextPathCoordDistance < 2 {
		d.NextPath++
		if d.NextPath == len(d.Path) {
			d.NextPath = 0
		}
		d.TurnTowardsPathPoint()
	}

	// Temporary until the dog sprite is looking up
	d.move(
		math.Cos(d.Angle)*dogSpeed,
		math.Sin(d.Angle)*dogSpeed,
	)
}

// Run moves the dog faster
func (d *Dog) Run() {
	// Temporary until the dog sprite is looking up
	d.move(
		math.Cos(d.Angle)*dogSpeed*1.2,
		math.Sin(d.Angle)*dogSpeed*1.2,
	)
}

// Move the Dog by the given vector if it is possible to do so
func (d *Dog) move(dx, dy float64) {
	// No collision detection for the time being
	d.Object.X += dx
	d.Object.Y += dy
}

// Draw draws the Dog to the screen
func (d *Dog) Draw(g *Game) {
	s := d.Sprite
	frame := s.Sprite[d.Frame]
	op := &ebiten.DrawImageOptions{}

	op.GeoM.Translate(
		float64(-frame.Position.W/2),
		float64(-frame.Position.H/2),
	)

	op.GeoM.Rotate(d.Angle + math.Pi/2)

	g.Camera.Surface.DrawImage(
		s.Image.SubImage(image.Rect(
			frame.Position.X,
			frame.Position.Y,
			frame.Position.X+frame.Position.W,
			frame.Position.Y+frame.Position.H,
		)).(*ebiten.Image),
		g.Camera.GetTranslation(
			op,
			float64(d.Object.X),
			float64(d.Object.Y)))
}
