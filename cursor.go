// Use of this source code is subject to an MIT-style
// licence which can be found in the LICENSE file.

package main

import "github.com/hajimehoshi/ebiten/v2"

// Cursor represents the mouse cursor on the screen
type Cursor struct {
	Hit   bool // whether the shot hit or not
	state int
	// tick     int
	images   []*ebiten.Image
	position Coord
}

const (
	cursorNormal int = iota
	cursorMiss
	cursorHit
)

func NewCursor() *Cursor {
	return &Cursor{
		images: []*ebiten.Image{
			loadImage("assets/sprites/Cursor_1.png"),
			loadImage("assets/sprites/Cursor_2.png"),
			loadImage("assets/sprites/Cursor_3.png"),
		},
	}
}

func (c *Cursor) Update(g *GameScreen) {
	cx, cy := ebiten.CursorPosition()
	c.position.X, c.position.Y = float64(cx), float64(cy)
	switch g.Player.State {
	case playerDryFire:
		c.state = cursorMiss
	case playerShooting:
		if c.Hit {
			c.state = cursorHit
		} else {
			c.state = cursorMiss
		}
	default:
		c.state = cursorNormal
	}
	return
}

func (c *Cursor) Draw(screen *ebiten.Image) {
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(c.position.X, c.position.Y)
	op.GeoM.Translate( // position image centre around coords
		float64(-c.images[c.state].Bounds().Dx())/2,
		float64(-c.images[c.state].Bounds().Dy())/2,
	)
	screen.DrawImage(c.images[c.state], op)
}
