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

// dogWalkingSpeed is the distance the dog moves per update cycle when walking
const dogWalkingSpeed float64 = 0.7

// dogRunningSpeed is the distance the dog moves per update cycle when running
const dogRunningSpeed float64 = 1.3

// waitingRadius is the maximum distance the dog walks away from the player
const waitingRadius float64 = 96

// zombieBarkRadius: if a zombie is this close to the dog, it barks
const zombieBarkRadius float64 = 150

// zombieDangerRadius: if a zombie is this close to the dog, it runs away
const zombieDangerRadius float64 = 80

// zombieSafeRadius: if a zombie is at least this far from the dog, it stops running
const zombieSafeRadius float64 = 192

// States of the dog
const (
	dogWalkingOnPath = iota
	dogFleeing
	dogWalkingBackToPath
	dogSniffing
	dogBarking
	dogWaiting
	dogBlocked
	dogFinished
	dogDied
)

// Maps dog states to animation frames
var dogStateToFrame = [8]int{0, 0, 0, 1, 1, 2, 2, 2}

// Dog is player's companion
type Dog struct {
	Object            *resolv.Object
	Angle             float64
	Speed             float64
	Frame             int
	State             int
	Path              Path
	NextPath          int
	Sprite            *SpriteSheet
	InDanger          bool
	OnThePath         bool
	LastPathCoord     Coord
	PrevCheckpoint    int
	SniffingCounter   int
	OutOfSightCounter int
}

// Update updates the state of the dog
func (d *Dog) Update(g *Game) {
	if d.State == dogDied {
		return
	}

	if d.State != dogFinished {
		if d.NextPath < 0 {
			d.NextPath = 0
			d.OnThePath = true
			d.TurnTowardsPathPoint()
		}

		zombieInRange := false
		zombieAlert := false
		resultantVectorCoord := Coord{
			X: d.Object.X,
			Y: d.Object.Y,
		}
		closestZombie := 1000.0
		for _, zombie := range g.Zombies {
			zombieDistance, xDistance, yDistance := CalcObjectDistance(d.Object, zombie.Object)
			if zombieDistance < closestZombie {
				closestZombie = zombieDistance
			}
			if zombieDistance < zombieDangerRadius {
				resultantVectorCoord.X += xDistance
				resultantVectorCoord.Y += yDistance
			}
		}

		if closestZombie < zombieDangerRadius {
			// If zombies are too close then the dog is in danger and will run away
			zombieInRange = true
			d.State = dogFleeing
		} else if closestZombie < zombieBarkRadius && d.State != dogFleeing {
			// If zombies gettgin closer then the dog stops walking and start barking
			zombieAlert = true
			if d.State != dogBarking {
				g.Sounds[soundDogBark1].Rewind()
				g.Sounds[soundDogBark1].Play()
			}
			d.State = dogBarking
		}

		// If the dog was in danger then it will be safe again when getting far enough from the zombies
		isSafeAgain := !d.InDanger || (d.InDanger && closestZombie > zombieSafeRadius)

		// If the dog starts fleeing away from the path then the last coordinate is saved to allow navigating back
		if d.State == dogFleeing && d.OnThePath {
			d.LastPathCoord.X = d.Object.X
			d.LastPathCoord.Y = d.Object.Y
			d.OnThePath = false
		}

		// The dog is in danger if there are zombies too close or it did not manage to run far enough
		d.InDanger = zombieInRange || !isSafeAgain

		if !zombieAlert {
			if !d.InDanger {
				playerDistance, _, _ := CalcObjectDistance(d.Object, g.Player.Object)
				if playerDistance < waitingRadius {
					// If the dog is not in danger and it is close to the player then it walks towards next path point
					d.FollowPath()
				} else {
					// If the player is not close enough then the dog sits down
					d.State = dogWaiting
				}
			} else {
				// If the dog is in danger then it runs away from the zombies
				if zombieInRange {
					// If zombies are close then recalculate Angle
					d.TurnTowardsCoordinate(resultantVectorCoord)
				}
				d.Run()
			}
		}
	}

	animationFrame := d.Sprite.Meta.FrameTags[dogStateToFrame[d.State]]
	d.Frame = Animate(d.Frame, g.Tick, animationFrame)

	d.Object.Shape.SetRotation(-d.Angle)
	d.Object.Update()

	sx, sy := g.Camera.GetScreenCoords(d.Object.X, d.Object.Y)
	if sx < 0 || sy < 0 || sx > float64(g.Width) || sy > float64(g.Height) {
		d.OutOfSightCounter++
		if d.OutOfSightCounter > 300 {
			g.Dog.State = dogDied
		}
	} else {
		d.OutOfSightCounter = 0
	}

}

// TurnTowardsCoordinate turns the dog to the direction of the point
func (d *Dog) TurnTowardsCoordinate(coord Coord) {
	adjacent := coord.X - d.Object.X
	opposite := coord.Y - d.Object.Y
	d.Angle = math.Atan2(opposite, adjacent)
}

// TurnTowardsPathPoint turns the dog to the direction of the next path point
func (d *Dog) TurnTowardsPathPoint() {
	d.TurnTowardsCoordinate(d.Path[d.NextPath])
}

// SniffNextCheckpoint starts the dog sniffing for next checkpoint
func (d *Dog) SniffNextCheckpoint() {
	d.SniffingCounter++
	if d.SniffingCounter == 180 {
		d.SniffingCounter = 0
		d.TurnTowardsPathPoint()
		d.State = dogWalkingOnPath
	}
}

// FollowPath moves the dog along the path
func (d *Dog) FollowPath() {
	if !d.OnThePath && d.State != dogWalkingBackToPath {
		d.TurnTowardsCoordinate(d.LastPathCoord)
		d.State = dogWalkingBackToPath
	}

	if d.SniffingCounter != 0 {
		d.State = dogSniffing
	}

	switch d.State {
	case dogWaiting:
		fallthrough
	case dogFleeing:
		fallthrough
	case dogBarking:
		d.State = dogWalkingOnPath
		return
	case dogWalkingOnPath:
		// The dog is following the path
		nextPathCoordDistance := CalcDistance(d.Path[d.NextPath].X, d.Path[d.NextPath].Y, d.Object.X, d.Object.Y)
		if nextPathCoordDistance < 2 {
			d.NextPath++
			if d.NextPath == len(d.Path) {
				d.State = dogFinished
				return
			}
			d.TurnTowardsPathPoint()
			return
		}
	case dogWalkingBackToPath:
		// The dog is getting back to the last know coordinate of the path
		lastPathCoordDistance := CalcDistance(d.LastPathCoord.X, d.LastPathCoord.Y, d.Object.X, d.Object.Y)
		if lastPathCoordDistance < 2 {
			d.OnThePath = true
			d.TurnTowardsPathPoint()
			d.State = dogWalkingOnPath
			return
		}
	case dogSniffing:
		d.SniffNextCheckpoint()
		return
	}

	d.move(
		math.Cos(d.Angle)*dogWalkingSpeed,
		math.Sin(d.Angle)*dogWalkingSpeed,
	)
}

// Run moves the dog faster
func (d *Dog) Run() {
	d.move(
		math.Cos(d.Angle)*dogRunningSpeed,
		math.Sin(d.Angle)*dogRunningSpeed,
	)
}

// Move the Dog by the given vector if it is possible to do so
func (d *Dog) move(dx, dy float64) {
	if d.State == dogWalkingOnPath || d.State == dogWalkingBackToPath {
		// WORKAROUND: If the dog is following the path or going back to the path
		// then collision with walls is not checked
		if collision := d.Object.Check(dx, dy, tagPlayer); collision != nil {
			if d.Object.Shape.Intersection(0, 0, collision.Objects[0].Shape) != nil {
				d.State = dogBlocked
				return
			}
		}
		if collision := d.Object.Check(dx, dy, tagMob); collision != nil {
			return
		}
	} else if d.State == dogBlocked {
		if collision := d.Object.Check(dx, dy, tagPlayer); collision == nil {
			d.State = dogWalkingOnPath
		}
		return
	} else {
		if collision := d.Object.Check(dx, dy, tagWall, tagMob, tagPlayer); collision != nil {
			if d.Object.Shape.Intersection(0, 0, collision.Objects[0].Shape) != nil {
				return
			}
		}
	}

	if d.State == dogWalkingOnPath {
		if collision := d.Object.Check(dx, dy, tagCheckpoint); collision != nil {
			o := collision.Objects[0]
			if o.Data.(int) > d.PrevCheckpoint {
				d.PrevCheckpoint = o.Data.(int)
				d.State = dogSniffing
			}
		}
	}
	d.Object.X += dx
	d.Object.Y += dy
}

// Draw draws the Dog to the screen
func (d *Dog) Draw(g *Game) {
	// the centre of the dog's shoulders is 5px down from the middle
	const centerOffset float64 = 5

	s := d.Sprite
	frame := s.Sprite[d.Frame]
	op := &ebiten.DrawImageOptions{}

	// Centre and rotate
	op.GeoM.Translate(
		float64(-frame.Position.W/2),
		float64(-frame.Position.H/2)+centerOffset/2,
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
			float64(d.Object.Y),
		),
	)
}
