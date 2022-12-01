// Use of this source code is subject to an MIT-style
// licence which can be found in the LICENSE file.

package main

import "github.com/hajimehoshi/ebiten/v2"

// Cursor represents the mouse cursor on the screen
type Cursor struct {
	// state    int
	// tick     int
	images   []*ebiten.Image
	position Coord
}

func NewCursor() *Cursor {
	img := loadImage("assets/sprites/Cursor_1.png")
	return &Cursor{
		images: []*ebiten.Image{img},
	}
}

func (c *Cursor) Update(g *GameScreen) {
	return
}

func (c *Cursor) Draw(screen *ebiten.Image) {
	cx, cy := ebiten.CursorPosition()
	c.position.X, c.position.Y = float64(cx), float64(cy)
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(c.position.X, c.position.Y)
	op.GeoM.Translate( // position image centre around coords
		float64(-c.images[0].Bounds().Dx())/2,
		float64(-c.images[0].Bounds().Dy())/2,
	)
	screen.DrawImage(c.images[0], op)
}
