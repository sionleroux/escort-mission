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

// Path contains information about a path and its progress
type Path struct {
	Points    []Coord // Coordinates of the path points
	NextPoint int     // Index of the next path point
}

// dogWalkingSpeed is the distance the dog moves per update cycle when walking
const dogWalkingSpeed float64 = 0.7

// dogRunningSpeed is the distance the dog moves per update cycle when running
const dogRunningSpeed float64 = 1.3

// waitingRadius is the maximum distance the dog walks away from the player
const waitingRadius float64 = 96

// followingRadius is the distance within which the dog follows the player after the last checkpoint
const followingRadius float64 = 96

// zombieBarkRadius: if a zombie is this close to the dog, it barks
const zombieBarkRadius float64 = 150

// zombieFleeRadius: if a zombie is this close to the dog, it runs away
const zombieFleeRadius float64 = 80

// zombieSafeRadius: if a zombie is at least this far from the dog, it stops running
const zombieSafeRadius float64 = 192

// fleeingPathLength: the length of the path planned for fleeing
const fleeingPathLength float64 = 200

// Operating modes of the dog
const (
	dogNormal = iota // Dog is alive and no zombies in vicinity
	dogDanger        // Zomibes are close, the dod needs to flee
	dogDead          // Dog died
)

// States of the dog
const (
	dogNormalWaiting = iota
	dogNormalWalking
	dogNormalBlocked
	dogNormalSniffing
	dogNormalWaitingAtCheckpoint

	dogDangerBarking
	dogDangerFleeing
)

// Maps dog states to animation frames
var dogStateToFrame = [7]int{2, 0, 2, 1, 2, 1, 0}

// Dog is player's companion
type Dog struct {
	Object               *resolv.Object
	Angle                float64
	Speed                float64
	Frame                int
	Mode                 int
	State                int
	PrevState            int
	CurrentPath          *Path
	MainPath             *Path
	OnMainPath           bool
	LastPathCoord        Coord
	Sprite               *SpriteSheet
	PrevCheckpoint       int
	LastPathpointReached bool
	AtCheckpointCounter  int
	OutOfSightCounter    int
}

func (d *Dog) Init() {
	d.CurrentPath = d.MainPath
	d.OnMainPath = true
	d.turnTowardsPathPoint()
}

// Resets the dog to a coordinate after death
func (d *Dog) Reset(cp int, x, y float64) {
	d.Mode = dogNormal
	d.State = dogNormalWaiting
	d.CurrentPath = d.MainPath
	d.CurrentPath.NextPoint = d.findClosestPathPoint(x, y)
	d.OnMainPath = true
	d.Object.X, d.Object.Y = x, y
	d.turnTowardsPathPoint()
}

// Finds the closest pathpoint on the dog's path
func (d *Dog) findClosestPathPoint(x, y float64) int {
	minDest := 1000.0
	index := 0
	for i, p := range d.MainPath.Points {
		d := CalcDistance(x, y, p.X, p.Y)
		if d < minDest {
			minDest = d
			index = i
		}
	}
	return index
}

// zombiesInRange checks if there zombies close to the Dog
// Returns
//  - if there are zombies in range
//  - the distance of the closest zombie
//  - the resultant vector of the fleeing path based on all the zombies in range
func (d *Dog) zombiesInRange(zRange float64, g *GameScreen) (bool, float64, Coord) {
	resultantVectorCoord := Coord{X: 0, Y: 0}
	closestZombie := 1000.0
	for _, zombie := range g.Zombies {
		zombieDistance, xDistance, yDistance := CalcObjectDistance(d.Object, zombie.Object)
		if zombieDistance < closestZombie {
			closestZombie = zombieDistance
		}
		if zombieDistance < zRange {
			resultantVectorCoord.X += xDistance
			resultantVectorCoord.Y += yDistance
		}
	}

	return closestZombie < zRange, closestZombie, resultantVectorCoord
}

// planFleeingRoute plans a fleeing route for the dog
func (d *Dog) planFleeingRoute(vector Coord, g *GameScreen) *Path {
	fleeingPath := &Path{}
	nv := NormalizeVector(vector)
	fleeingPath.Points = []Coord{
		{X: d.Object.X + nv.X*fleeingPathLength, Y: d.Object.Y + nv.Y*fleeingPathLength},
	}
	return fleeingPath
}

// planRouteBackToMainPath plans a route back to the main path
func (d *Dog) planRouteBackToMainPath(g *GameScreen) *Path {
	returnPath := &Path{}
	returnPath.Points = g.LevelMap.FindPath(
		Coord{X: d.Object.X, Y: d.Object.Y},
		d.LastPathCoord,
	)
	returnPath.Points[len(returnPath.Points)-1] = d.LastPathCoord
	returnPath.Points = GetBezierPathFromCoords(returnPath.Points, 2)
	return returnPath
}

// updateState updates the state machine of the dog
func (d *Dog) updateState(g *GameScreen) {
	d.PrevState = d.State

	// Does the dog need to change mode? Danger <-> Normal
	switch d.Mode {
	case dogNormal:
		zInRange, _, _ := d.zombiesInRange(zombieBarkRadius, g)
		if zInRange {
			d.Mode = dogDanger
			d.State = dogDangerBarking
		}
	case dogDanger:
		zInRange, _, _ := d.zombiesInRange(zombieSafeRadius, g)
		if !zInRange {
			d.Mode = dogNormal
			d.State = dogNormalWaiting
		}
	}

	// Does the dog need to change state in its current mode?
	switch d.Mode {
	case dogNormal:
		playerDistance, _, _ := CalcObjectDistance(d.Object, g.Player.Object)

		switch d.State {
		case dogNormalWaiting:
			if d.PrevState == dogDangerFleeing {
				d.CurrentPath = d.planRouteBackToMainPath(g)
				d.turnTowardsPathPoint()
				d.OnMainPath = false
			}
			if playerDistance <= waitingRadius {
				d.State = dogNormalWalking
			}
		case dogNormalWalking:
			if playerDistance > waitingRadius {
				d.State = dogNormalWaiting
			}
			// Next state is set elsewhere
			// - In case when dog reaches checkpoint and starts sniffing
			// - In case when dog reaches destination
			// - In case when dog is getting blocked by the player
		case dogNormalBlocked:
			// Next state is set elsewhere
			// - In case when dog is not blocked anymore by the player
		case dogNormalSniffing:
			// If the player has already touched the checkpoint previously
			if (d.PrevCheckpoint == g.Checkpoint) {
				d.State = dogNormalWalking
			}
			// Sniffing for 3 seconds
			if d.AtCheckpointCounter >= 180 {
				d.State = dogNormalWaitingAtCheckpoint
			}
			// Next state is set elsewhere
			// - In case when player is also at the checkpoint
		case dogNormalWaitingAtCheckpoint:
			// Next state is set elsewhere
			// - In case when player is also at the checkpoint
		}
	case dogDanger:
		zInRange, _, _ := d.zombiesInRange(zombieFleeRadius, g)

		switch d.State {
		case dogDangerBarking:
			if zInRange {
				if d.OnMainPath {
					d.LastPathCoord = d.CurrentPath.Points[d.CurrentPath.NextPoint]
				}
				d.State = dogDangerFleeing
			}
		case dogDangerFleeing:
			// Dog is fleeing until it changes mode to Normal
		}
	}
}

// Update updates the state of the dog
func (d *Dog) Update(g *GameScreen) {
	// If the dog is dead, no need to update
	if d.Mode == dogDead {
		return
	}

	d.updateState(g)

	// If the dog is out of the screen for too long then it dies
	sx, sy := g.Camera.GetScreenCoords(d.Object.X, d.Object.Y)
	if sx < 0 || sy < 0 || sx > float64(g.Width) || sy > float64(g.Height) {
		d.OutOfSightCounter++
		if d.OutOfSightCounter > 300 {
			g.Dog.Mode = dogDead
		}
	} else {
		d.OutOfSightCounter = 0
	}

	// Update dog based on current and previous states
	switch d.State {
	case dogNormalWaiting:
		// Do nothing
	case dogNormalWalking:
		// If dog is walking back to the main path and reaches it then change its path to main path
		// Try to move along the path
		d.walk(g)
	case dogNormalBlocked:
		// Try to move along the path
		d.walk(g)
	case dogNormalSniffing:
		d.AtCheckpointCounter++
		// Wait for the player to arrive at the same checkpoint
	case dogNormalWaitingAtCheckpoint:
		d.AtCheckpointCounter++
		// Dog barks at every 5 seconds
		if (d.AtCheckpointCounter % 300 == 0) {
			g.Sounds[soundDogBark].Play()
		}
		// Wait for the player to arrive at the same checkpoint
	case dogDangerBarking:
		// Play barking sound
		if d.PrevState != dogDangerBarking {
			g.Sounds[soundDogBark].Play()
		}
	case dogDangerFleeing:
		zInRange, _, resultantVector := d.zombiesInRange(zombieFleeRadius, g)
		if zInRange {
			d.CurrentPath = d.planFleeingRoute(resultantVector, g)
			d.turnTowardsPathPoint()
			d.OnMainPath = false
		}
		// Try to run along the path
		d.followPath(g)
	}

	// Animate dog
	animationFrame := d.Sprite.Meta.FrameTags[dogStateToFrame[d.State]]
	d.Frame = Animate(d.Frame, g.Tick, animationFrame)
	d.Object.Update()
}

// ContinueFromCheckpoint continues walking on the path when the player is at the same checkpoint
// If the last checkpoint is reached then the dog starts following the player
func (d *Dog) ContinueFromCheckpoint() {
	d.State = dogNormalWalking
}

// walk moves the dog either on a path or by following the player
func (d *Dog) walk(g *GameScreen) {
	if (d.LastPathpointReached) {
		d.followPlayer(g)
	} else {
		d.followPath(g)
	}
}

// TurnTowardsCoordinate turns the dog to the direction of the point
func (d *Dog) turnTowardsCoordinate(coord Coord) {
	adjacent := coord.X - d.Object.X
	opposite := coord.Y - d.Object.Y
	d.Angle = math.Atan2(opposite, adjacent)
}

// TurnTowardsPathPoint turns the dog to the direction of the next path point
func (d *Dog) turnTowardsPathPoint() {
	d.turnTowardsCoordinate(d.CurrentPath.Points[d.CurrentPath.NextPoint])
}

func (d *Dog) followPlayer(g *GameScreen) {
	d.turnTowardsCoordinate(Coord{ X: g.Player.Object.X, Y: g.Player.Object.Y})

	d.move(
		math.Cos(d.Angle) * dogWalkingSpeed,
		math.Sin(d.Angle) * dogWalkingSpeed,
	)
}

// followPath moves the dog along its current path
func (d *Dog) followPath(g *GameScreen) {
	if d.CurrentPath.NextPoint == len(d.CurrentPath.Points) {
		return
	}

	nextPoint := d.CurrentPath.Points[d.CurrentPath.NextPoint]
	nextPathCoordDistance := CalcDistance(nextPoint.X, nextPoint.Y, d.Object.X, d.Object.Y)
	if nextPathCoordDistance < 2 {
		d.CurrentPath.NextPoint++
		if d.CurrentPath.NextPoint == len(d.CurrentPath.Points) {
			if d.State == dogNormalWalking {
				// If the dog is normally walking, not fleeing
				if d.OnMainPath {
					// If the dog reaches the end of the main path then it finishes
					d.LastPathpointReached = true
					return
				} else {
					// If the dog reaches the main path again then it will continue on that
					d.CurrentPath = d.MainPath
					d.OnMainPath = true
				}
			} else {
				// If the dog is fleeing
				zInRange, _, resultantVector := d.zombiesInRange(zombieFleeRadius, g)
				if !zInRange {
					return
				}
				d.CurrentPath = d.planFleeingRoute(resultantVector, g)
			}

		}
		d.turnTowardsPathPoint()
		return
	}

	var speed float64
	if d.State == dogDangerFleeing {
		speed = dogRunningSpeed
	} else {
		speed = dogWalkingSpeed
	}

	d.move(
		math.Cos(d.Angle)*speed,
		math.Sin(d.Angle)*speed,
	)
}

// Move the Dog by the given vector if it is possible to do so
func (d *Dog) move(dx, dy float64) {
	switch d.State {
	case dogNormalWalking:
		// If the dog would collide with the player
		if collision := d.Object.Check(dx, dy, tagPlayer); collision != nil {
			if d.Object.Shape.Intersection(0, 0, collision.Objects[0].Shape) != nil {
				d.State = dogNormalBlocked
				return
			}
		}
		// If the dog reaches a checkpoint
		if collision := d.Object.Check(dx, dy, tagCheckpoint); collision != nil {
			o := collision.Objects[0]
			// If the dog or the player has not been at the checkpoint before
			if o.Data.(int) > d.PrevCheckpoint {
				d.PrevCheckpoint = o.Data.(int)
			}
			d.AtCheckpointCounter = 0
			d.State = dogNormalSniffing
		}
	case dogNormalBlocked:
		// If the dog would not collide with the player anymore
		if collision := d.Object.Check(dx, dy, tagPlayer); collision == nil {
			d.State = dogNormalWalking
		} else {
			return
		}
	default:
		if collision := d.Object.Check(dx, dy, tagWall, tagMob, tagPlayer); collision != nil {
			if d.Object.Shape.Intersection(0, 0, collision.Objects[0].Shape) != nil {
				return
			}
		}
	}

	d.Object.X += dx
	d.Object.Y += dy
}

// Draw draws the Dog to the screen
func (d *Dog) Draw(g *GameScreen) {
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
