// Use of this source code is subject to an MIT-style
// licence which can be found in the LICENSE file.

package main

import (
	"github.com/hajimehoshi/ebiten/v2"
)

// HudImage are images for use in the HUD
type HudImage int

const (
	hudBullet HudImage = iota
	hudCasing
)

var hudPadding int = 5

// HUD is a display showing information during the game
// So far it only shows how much ammo you have left
type HUD struct {
	Images []*ebiten.Image
}

// NewHUD initialises a new HUD with its graphics
func NewHUD() *HUD {
	return &HUD{
		Images: []*ebiten.Image{
			loadImage("assets/sprites/Bullet.png"),
			loadImage("assets/sprites/Casing.png"),
		},
	}
}

// Draw draws the HUD onto the screen, this is intended to be drawn onto the
// game screen with current camera view *after* the camera surface has been
// blitted to the screen
func (hud HUD) Draw(ammo int, screen *ebiten.Image) {
	corner := screen.Bounds().Max
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(
		float64(corner.X),
		float64(corner.Y-hud.Images[hudBullet].Bounds().Dy()-hudPadding),
	)
	for i := 0; i < playerAmmoClipMax; i++ {
		var bullet *ebiten.Image
		if i < ammo {
			bullet = hud.Images[hudBullet]
		} else {
			bullet = hud.Images[hudCasing]
		}
		op.GeoM.Translate(float64(-bullet.Bounds().Dx()-hudPadding), 0)
		screen.DrawImage(bullet, op)
	}
}
