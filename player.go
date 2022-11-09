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

// states of the player
// It would be great to map them to the frameTag.Name from JSON
const (
	playerIdle = iota
	playerWalking
	playerReady
	playerShooting
	playerUnready
)

// Player is the player character in the game
type Player struct {
	Object *resolv.Object
	Angle  float64
	Frame  int
	State  int
	Sprite *SpriteSheet
}

// Update updates the state of the player
func (p *Player) Update(g *Game) {
	if p.State != playerShooting {
		p.State = playerIdle

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

	// Player gun rotation
	cx, cy := g.Camera.GetCursorCoords()
	adjacent := p.Object.X - float64(cx)
	opposite := p.Object.Y - float64(cy)
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
		p.State = playerIdle
	}
}

// MoveLeft moves the player left
func (p *Player) MoveLeft() {
	speed := playerSpeed * playerSpeedFactorSideways
	p.move(
		-math.Sin(p.Angle)*speed,
		math.Cos(p.Angle)*speed,
	)
}

// MoveRight moves the player right
func (p *Player) MoveRight() {
	speed := playerSpeed * playerSpeedFactorSideways
	p.move(
		math.Sin(p.Angle)*speed,
		-math.Cos(p.Angle)*speed,
	)
}

// MoveForward moves the player forward towards the pointer
func (p *Player) MoveForward() {
	p.move(
		-math.Cos(p.Angle)*playerSpeed,
		-math.Sin(p.Angle)*playerSpeed,
	)
}

// MoveBackward moves the player backward away from the pointer
func (p *Player) MoveBackward() {
	speed := playerSpeed * playerSpeedFactorReverse
	p.move(
		math.Cos(p.Angle)*speed,
		math.Sin(p.Angle)*speed,
	)
}

// Move the Player by the given vector if it is possible to do so
func (p *Player) move(dx, dy float64) {
	if collision := p.Object.Check(dx, dy, tagWall); collision == nil {
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
