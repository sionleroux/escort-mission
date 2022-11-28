// Use of this source code is subject to an MIT-style
// licence which can be found in the LICENSE file.

package main

import (
	"image"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/solarlune/resolv"
)

// playerSpeed is the distance the player moves per update cycle
var playerSpeed float64 = 1.2

// amount to change speed by when the player is reversing backwards
// walking backwards is very slow
var playerSpeedFactorReverse float64 = 0.2

// amount to change speed by when the player is strafing sideways
var playerSpeedFactorSideways float64 = 0.6

var playerSpeedFactorSprint float64 = 2.4

var playerAmmoClipMax int = 7

// states of the player
// It would be great to map them to the frameTag.Name from JSON
type playerState int

const (
	playerIdle     playerState = iota // Waiting for input from the player
	playerWalking                     // Walking in some direction
	playerReady                       // Readying the gun to shoot
	playerShooting                    // Shooting the gun
	playerDryFire                     // Trying to shoot with no ammo
	playerReload                      // Reloading the gun
	playerUnready                     // Unreadying the gun after shooting
)

// Player is the player character in the game
type Player struct {
	Object    *resolv.Object // Used for collision detection with other objects
	Angle     float64        // The angle the player is facing at
	Frame     int            // The current animation frame
	State     playerState    // The current animation state
	PrevState playerState    // The previous animation state
	Sprinting bool           // Whether the player is sprinting or not
	Sprite    *SpriteSheet   // Used for player animations
	Range     float64        // How far you can shoot with the gun
	Ammo      int            // How many shots you have left in the gun
}

// NewPlayer constructs a new Player object at the provided location and size
func NewPlayer(position []int, sprites *SpriteSheet) *Player {
	// the head and shoulders are about 4px from the middle
	const collisionBoxSize float64 = 8

	dimensions := sprites.Sprite[0].Position
	object := resolv.NewObject(
		float64(position[0]), float64(position[1]),
		float64(dimensions.W), float64(dimensions.H),
		tagPlayer,
	)
	object.SetShape(resolv.NewRectangle(
		0, 0, // origin
		collisionBoxSize, collisionBoxSize,
	))
	object.Shape.(*resolv.ConvexPolygon).RecenterPoints()

	player := &Player{
		State:  playerIdle,
		Object: object,
		Angle:  0,
		Sprite: sprites,
		Range:  200,
		Ammo:   playerAmmoClipMax,
	}

	return player
}

// Reload reloads the ammo
func (p *Player) Reload(g *GameScreen) {
	p.State = playerReload
	g.Sounds[soundGunReload].Play()
}

// Update updates the state of the player
func (p *Player) Update(g *GameScreen) {
	p.PrevState = p.State
	p.Sprinting = false

	if p.State == playerIdle || p.State == playerWalking {
		p.State = playerIdle
		p.handleControls()
	}

	if p.Frame == p.Sprite.Meta.FrameTags[p.State].To {
		p.animationBasedStateChanges(g)
	}

	// Player gun rotation
	cx, cy := g.Camera.GetCursorCoords()
	adjacent := float64(cx) - p.Object.X
	opposite := float64(cy) - p.Object.Y
	p.Angle = math.Atan2(opposite, adjacent)

	p.Frame = Animate(p.Frame, g.Tick, p.Sprite.Meta.FrameTags[p.State])
	p.Object.Shape.SetRotation(-p.Angle)
	p.Object.Update()
}

// Animation-trigged state changes
func (p *Player) animationBasedStateChanges(g *GameScreen) {
	switch p.State {
	case playerShooting: // Back to idle after shooting animation
		p.State = playerIdle
		if p.Ammo < 1 {
			p.Reload(g) // Automatic reload if out of ammo
		}
	case playerReload: // Back to idle after reload animation
		p.Ammo = playerAmmoClipMax
		p.State = playerIdle
	case playerDryFire: // Back to idle after reload animation
		p.State = playerIdle
	}
}

// MoveLeft moves the player left
func (p *Player) MoveLeft() {
	speed := playerSpeed * playerSpeedFactorSideways
	p.move(
		math.Sin(p.Angle)*speed,
		-math.Cos(p.Angle)*speed,
	)
}

// MoveRight moves the player right
func (p *Player) MoveRight() {
	speed := playerSpeed * playerSpeedFactorSideways
	p.move(
		-math.Sin(p.Angle)*speed,
		math.Cos(p.Angle)*speed,
	)
}

// MoveForward moves the player forward towards the pointer
func (p *Player) MoveForward() {
	speed := playerSpeed
	if p.Sprinting {
		speed = speed * playerSpeedFactorSprint
	}
	p.move(
		math.Cos(p.Angle)*speed,
		math.Sin(p.Angle)*speed,
	)
}

// MoveBackward moves the player backward away from the pointer
func (p *Player) MoveBackward() {
	speed := playerSpeed * playerSpeedFactorReverse
	p.move(
		-math.Cos(p.Angle)*speed,
		-math.Sin(p.Angle)*speed,
	)
}

// Move the Player by the given vector if it is possible to do so
func (p *Player) move(dx, dy float64) {
	p.State = playerWalking
	if collision := p.Object.Check(dx, dy, tagWall, tagDog); collision != nil {
		if p.Object.Shape.Intersection(dx, dy, collision.Objects[0].Shape) != nil {
			return
		}
	}
	p.Object.X += dx
	p.Object.Y += dy
}

// Draw draws the Player to the screen
func (p *Player) Draw(g *GameScreen) {
	// the centre of the player's head is 2px down from the middle
	const centerOffset float64 = -2

	s := p.Sprite
	frame := s.Sprite[p.Frame]
	op := &ebiten.DrawImageOptions{}

	// Centre and rotate
	op.GeoM.Translate(
		float64(-frame.Position.W/2),
		float64(-frame.Position.H/2)+centerOffset/2,
	)
	op.GeoM.Rotate(p.Angle + math.Pi/2)

	g.Camera.Surface.DrawImage(
		s.Image.SubImage(image.Rect(
			frame.Position.X,
			frame.Position.Y,
			frame.Position.X+frame.Position.W,
			frame.Position.Y+frame.Position.H,
		)).(*ebiten.Image),
		g.Camera.GetTranslation(
			op,
			float64(p.Object.X),
			float64(p.Object.Y),
		),
	)
}

func (p *Player) handleControls() {
	if ebiten.IsKeyPressed(ebiten.KeyShift) {
		p.Sprinting = true
	}
	if ebiten.IsKeyPressed(ebiten.KeyW) {
		p.MoveForward()
	}
	if ebiten.IsKeyPressed(ebiten.KeyA) {
		p.MoveLeft()
	}
	if ebiten.IsKeyPressed(ebiten.KeyS) {
		p.MoveBackward()
	}
	if ebiten.IsKeyPressed(ebiten.KeyD) {
		p.MoveRight()
	}
}

// Position returns the Player's current coordinates
func (p *Player) Position() *Coord {
	return &Coord{
		X: p.Object.X,
		Y: p.Object.Y,
	}
}
