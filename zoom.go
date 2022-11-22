package main

import (
	"github.com/tanema/gween"
	"github.com/tanema/gween/ease"
)

const zoomOutLevel = 0.875
const zoomInLevel = 1.125
const zoomTime = 15

// Zoom controls the camera zoom effect
type Zoom struct {
	On       bool
	Amount   float64
	tweenIn  *gween.Tween
	tweenOut *gween.Tween
	dt       int
}

// NewZoom sets up a new zoom state tracker
func NewZoom() *Zoom {
	return &Zoom{
		On:       false,
		Amount:   zoomOutLevel,
		tweenIn:  gween.New(zoomOutLevel, zoomInLevel, zoomTime, ease.OutCubic),
		tweenOut: gween.New(zoomInLevel, zoomOutLevel, zoomTime, ease.OutCubic),
	}
}

// Update updates the camera with the current zoom level
func (zoom *Zoom) Update() {
	if zoom.On {
		zoom.tweenOut.Reset()
		amount, _ := zoom.tweenIn.Update(1)
		zoom.Amount = float64(amount)
	} else {
		zoom.tweenIn.Reset()
		amount, _ := zoom.tweenOut.Update(1)
		zoom.Amount = float64(amount)
	}
}
