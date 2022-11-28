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
var zombieSpeed float64 = 0.4

// zombieCrawlerSpeed is the distance the crawler zombie moves per update cycle
var zombieCrawlerSpeed float64 = 0.2

// zombieSprinterSpeed is the distance the sprinter zombie moves per update cycle
var zombieSprinterSpeed float64 = 1.2

// zombieRange is how far away the zombie sees something to attack
var zombieRange float64 = 200

// Types of zombies
type ZombieType uint8

const (
	zombieNormal ZombieType = iota
	zombieCrawler
	zombieSprinter
	zombieBig
)

// Zombielike is anything that behaves like a zombie (e.g. attacks things and dies)
type Zombielike interface {
	Update(*GameScreen) error
	Draw(*GameScreen)
	Hit(*GameScreen)
	Die(*GameScreen)
	Type() ZombieType
	Remove()
	Position() *Coord
}

// Zombies is an array of Zombie
type Zombies []Zombielike

// Update updates all the zombies
func (zs *Zombies) Update(g *GameScreen) {
	for i, z := range *zs {
		err := z.Update(g)
		if err != nil {
			// clear and remove dead zombies
			log.Println(err)
			g.Zombies[i] = nil
			g.Zombies = append((*zs)[:i], (*zs)[i+1:]...)
			z.Remove()
			if len(g.Zombies) == 0 && rand.Float64() < 0.3 {
				g.Voices[voiceKill].Play()
			}
		}
	}
}

func NewZombie(spawnpoint *SpawnPoint, position Coord, zombieType ZombieType, sprites *SpriteSheet) *Zombie {
	// the head and shoulders are about 3px from the middle
	const collisionBoxSize float64 = 6

	var speed float64
	var hitToDie int

	switch zombieType {
	case zombieNormal:
		speed = zombieSpeed
		hitToDie = 1 + rand.Intn(2)
	case zombieCrawler:
		speed = zombieCrawlerSpeed
		hitToDie = 1 + rand.Intn(2)
	case zombieSprinter:
		speed = zombieSprinterSpeed
		hitToDie = 1
	case zombieBig:
		speed = zombieSpeed
		// hitToDie = 10
		hitToDie = 7
	}

	dimensions := sprites.Sprite[0].Position
	object := resolv.NewObject(
		position.X, position.Y,
		float64(dimensions.W), float64(dimensions.H),
		tagMob,
	)
	object.SetShape(resolv.NewRectangle(
		0, 0, // origin
		collisionBoxSize, collisionBoxSize,
	))
	object.Shape.(*resolv.ConvexPolygon).RecenterPoints()

	z := &Zombie{
		Object:     object,
		Angle:      0,
		Sprite:     sprites,
		Speed:      speed * (1 + rand.Float64()),
		HitToDie:   hitToDie,
		ZombieType: zombieType,
		TempSpeed:  1,
	}
	z.Object.Data = z
	z.SpawnPoint = spawnpoint

	return z
}

// Draw draws all the zombies
func (zs Zombies) Draw(g *GameScreen) {
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
	Object     *resolv.Object // Used for collision detection with other objects
	Angle      float64        // The angle the zombies is facing at
	Frame      int            // The current animation frame
	State      int            // The current animation state
	Sprite     *SpriteSheet   // Used for zombie animations
	Speed      float64        // The speed this zombie walks at
	TempSpeed  float64        // Temporary speed multiplier
	Target     *resolv.Object // Target object (player or dog)
	HitToDie   int            // Number of hits needed to die
	ZombieType ZombieType     // Type of the zombie
	SpawnPoint *SpawnPoint    // Reference for the SpawnPoint where the zombie was spawned
}

// Remove the zombie from the game's list of zombies and from the spawn point's
// list of live zombies
func (z *Zombie) Remove() {
	if z.Object.Space != nil {
		z.Object.Space.Remove(z.Object)
	}
	z.SpawnPoint.RemoveZombie(z)
}

// Position returns the current coordinates of the zombie
func (z *Zombie) Position() *Coord {
	return &Coord{
		X: z.Object.X,
		Y: z.Object.Y,
	}
}

// Type returns what type of zombie this is
func (z *Zombie) Type() ZombieType {
	return z.ZombieType
}

// Update updates the state of the zombie
func (z *Zombie) Update(g *GameScreen) error {
	if z.State == zombieDead {
		return errors.New("Zombie died")
	}

	if z.State == zombieIdle || z.State == zombieWalking {
		playerDistance, _, _ := CalcObjectDistance(z.Position(), g.Player.Position())
		dogDistance, _, _ := CalcObjectDistance(z.Position(), g.Dog.Position())

		zShouldWalk := false

		if playerDistance < zombieRange {
			z.Target = g.Player.Object
			zShouldWalk = true
		} else if dogDistance < zombieRange*1.2 {
			z.Target = g.Dog.Object
			zShouldWalk = true
		}

		if zShouldWalk {
			if z.State == zombieIdle {
				// Zombie detects target
				if z.ZombieType == zombieNormal || z.ZombieType == zombieCrawler {
					g.Sounds[soundZombieGrowl].Play()
				} else if z.ZombieType == zombieSprinter {
					g.Sounds[soundZombieScream].Play()
				} else {
					g.Sounds[soundBigZombieSound].Play()
				}
			}
			z.walk()
		} else {
			z.State = zombieIdle
		}
	}

	z.Frame = Animate(z.Frame, g.Tick, z.Sprite.Meta.FrameTags[z.State])
	if z.Frame == z.Sprite.Meta.FrameTags[z.State].To {
		z.animationBasedStateChanges(g)
	}

	z.Object.Shape.SetRotation(-z.Angle)
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

// Animation-trigged state changes
func (z *Zombie) animationBasedStateChanges(g *GameScreen) {
	switch z.State {
	case zombieHit:
		z.State = zombieWalking
	case zombieDeath:
		z.State = zombieDead
	}
}

// MoveUp moves the zombie upwards
func (z *Zombie) MoveUp() {
	z.move(0, -z.Speed*z.TempSpeed)
}

// MoveDown moves the zombie downwards
func (z *Zombie) MoveDown() {
	z.move(0, z.Speed*z.TempSpeed)
}

// MoveLeft moves the zombie left
func (z *Zombie) MoveLeft() {
	z.move(-z.Speed*z.TempSpeed, 0)
}

// MoveRight moves the zombie right
func (z *Zombie) MoveRight() {
	z.move(z.Speed*z.TempSpeed, 0)
}

// Move the Zombie by the given vector if it is possible to do so
func (z *Zombie) move(dx, dy float64) {
	z.State = zombieWalking
	if collision := z.Object.Check(dx, dy, tagMob, tagWall); collision == nil {
		z.Object.X += dx
		z.Object.Y += dy
	}
	// Collision detection and response between sand trap and zombie
	z.TempSpeed = 1
	if collision := z.Object.Check(0, 0, tagSandTrap); collision != nil {
		if z.Object.Overlaps(collision.Objects[0]) {
			if z.Object.Shape.Intersection(0, 0, collision.Objects[0].Shape) != nil {
				z.TempSpeed = sandTrapSpeedMultiplier
			}
		}
	}
}

// Draw draws the Zombie to the screen
func (z *Zombie) Draw(g *GameScreen) {
	// -2, // the centre of the zombie's head is 2px up from the middle
	const centerOffset float64 = 2

	s := z.Sprite
	frame := s.Sprite[z.Frame]
	op := &ebiten.DrawImageOptions{}

	// Centre and rotate
	op.GeoM.Translate(
		float64(-frame.Position.W/2),
		float64(-frame.Position.H/2)+centerOffset/2,
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
			float64(z.Object.X),
			float64(z.Object.Y),
		),
	)

}

// Hit changes zombie state and updates game data in response to it getting shot
func (z *Zombie) Hit(g *GameScreen) {
	g.Stat.CounterZombiesHit++
	z.State = zombieHit
	z.HitToDie--
	if z.HitToDie == 0 {
		z.Die(g)
	} else {
		g.Sounds[soundHit].Play()
		g.Sounds[soundZombieGrowl].Play()
	}
}

// Die changes zombie state and updates game data in case of a deadly shot
func (z *Zombie) Die(g *GameScreen) {
	g.Stat.CounterZombiesKilled++
	g.Sounds[soundZombieDeath].Play()
	z.Remove()
	z.State = zombieDeath
}
