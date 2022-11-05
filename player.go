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
const playerSpeed float64 = 2

//
const (
	playerIdle    int = 0
	playerWalking     = 1   
)

// Player is the player character in the game
type Player struct {
	Object        *resolv.Object
	Angle         float64
	Frame         int
	State         int
	Sprite        *SpriteSheet
}

// Update updates the state of the player
func (p *Player) Update(g *Game) {
	state := playerIdle

	if ebiten.IsKeyPressed(ebiten.KeyW) {
		p.MoveUp()
		state = playerWalking
	}
	if ebiten.IsKeyPressed(ebiten.KeyA) {
		p.MoveLeft()
		state = playerWalking
	}
	if ebiten.IsKeyPressed(ebiten.KeyS) {
		p.MoveDown()
		state = playerWalking
	}
	if ebiten.IsKeyPressed(ebiten.KeyD) {
		p.MoveRight()
		state = playerWalking
	}
	p.State = state

	// Player gun rotation
	cx, cy := g.Camera.GetCursorCoords()
	adjacent := p.Object.X - float64(cx)
	opposite := p.Object.Y - float64(cy)
	p.Angle = math.Atan2(opposite, adjacent)

	p.animate(g)
	p.Object.Update()
}

func (p *Player) animate(g *Game) {
	//It should be changed to frameTags
	switch p.State {
	case playerIdle:
		p.Frame = 0
	case playerWalking:
		//(p.Frame - startFrame + step) % (endFrame - startFrame + 1) + startFrame
		if (g.Tick%5 == 0) {
			p.Frame = (p.Frame - 0 + 1) % (2 - 0 + 1) + 0
		}
	}
}

// MoveUp moves the player upwards
func (p *Player) MoveUp() {
	p.move(0, -playerSpeed)
}

// MoveDown moves the player downwards
func (p *Player) MoveDown() {
	p.move(0, playerSpeed)
}

// MoveLeft moves the player left
func (p *Player) MoveLeft() {
	p.move(-playerSpeed, 0)
}

// MoveRight moves the player right
func (p *Player) MoveRight() {
	p.move(playerSpeed, 0)
}

// Move the Player by the given vector if it is possible to do so
func (p *Player) move(dx, dy float64) {
	if collision := p.Object.Check(dx, dy, "wall"); collision == nil {
		p.Object.X += dx
		p.Object.Y += dy
	}
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
