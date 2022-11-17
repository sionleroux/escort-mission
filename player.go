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

// playerSpeed is the distance the player moves per update cycle
const playerSpeed float64 = 1.2

// amount to change speed by when the player is reversing backwards
// walking backwards is very slow
const playerSpeedFactorReverse float64 = 0.2

// amount to change speed by when the player is strafing sideways
const playerSpeedFactorSideways float64 = 0.6

const playerSpeedFactorSprint float64 = 2.4

const playerAmmoClipMax int = 7

// states of the player
// It would be great to map them to the frameTag.Name from JSON
const (
	playerIdle     = iota // Waiting for input from the player
	playerWalking         // Walking in some direction
	playerReady           // Readying the gun to shoot
	playerShooting        // Shooting the gun
	playerDryFire         // Trying to shoot with no ammo
	playerReload          // Reloading the gun
	playerUnready         // Unreadying the gun after shooting
)

// Player is the player character in the game
type Player struct {
	Object    *resolv.Object // Used for collision detection with other objects
	Angle     float64        // The angle the player is facing at
	Frame     int            // The current animation frame
	State     int            // The current animation state
	Sprinting bool           // Whether the player is sprinting or not
	Sprite    *SpriteSheet   // Used for player animations
	Range     float64        // How far you can shoot with the gun
	Ammo      int            // How many shots you have left in the gun
}

// NewPlayer constructs a new Player object at the provided location and size
func NewPlayer(position []int, sprites *SpriteSheet) *Player {
	dimensions := sprites.Sprite[0].Position
	player := &Player{
		State: playerIdle,
		Object: resolv.NewObject(
			float64(position[0]), float64(position[1]),
			float64(dimensions.W), float64(dimensions.H),
			tagPlayer,
		),
		Angle:  0,
		Sprite: sprites,
		Range:  200,
		Ammo:   playerAmmoClipMax,
	}
	return player
}

// Update updates the state of the player
func (p *Player) Update(g *Game) {
	p.Sprinting = false

	if p.State == playerIdle || p.State == playerWalking {
		p.State = playerIdle
		p.handleControls()
	}

	// Player gun rotation
	cx, cy := g.Camera.GetCursorCoords()
	adjacent := float64(cx) - p.Object.X
	opposite := float64(cy) - p.Object.Y
	p.Angle = math.Atan2(opposite, adjacent)

	p.animate(g)
	p.Object.Update()
}

func (p *Player) animate(g *Game) {
	// Update only in every 5th cycle
	if g.Tick%5 != 0 {
		return
	}

	ft := p.Sprite.Meta.FrameTags[p.State]

	if ft.From == ft.To {
		p.Frame = ft.From
	} else {
		// Continuously increase the Frame counter between From and To
		p.Frame = (p.Frame-ft.From+1)%(ft.To-ft.From+1) + ft.From
	}

	// Back to idle after shooting animation
	if p.State == playerShooting && p.Frame == ft.To {
		if p.Ammo < 1 {
			p.State = playerReload
			g.Sounds[soundGunReload].Rewind()
			g.Sounds[soundGunReload].Play()	
			return
		}
		p.State = playerIdle
		return
	}

	// Back to idle after reload animation
	if p.State == playerReload && p.Frame == ft.To {
		p.Ammo = playerAmmoClipMax
		p.State = playerIdle
		return
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
	if collision := p.Object.Check(dx, dy, tagWall, tagDog); collision == nil {
		p.Object.X += dx
		p.Object.Y += dy
	}
	p.State = playerWalking
}

// Draw draws the Player to the screen
func (p *Player) Draw(g *Game) {
	s := p.Sprite
	frame := s.Sprite[p.Frame]
	op := &ebiten.DrawImageOptions{}

	op.GeoM.Translate(
		float64(-frame.Position.W/2),
		float64(-frame.Position.H/2),
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
			float64(p.Object.X)+float64(frame.Position.W/2),
			float64(p.Object.Y)+float64(frame.Position.H/2)))
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
