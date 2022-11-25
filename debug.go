// Copyright 2021 Siôn le Roux.  All rights reserved.
// Use of this source code is subject to an MIT-style
// licence which can be found in the LICENSE file.

package main

import (
	"github.com/hajimehoshi/ebiten/v2"
)

var debuggers Debuggers

// Debugger provides debug information by rendering it on-screen
type Debugger interface {
	Debug(g *GameScreen, screen *ebiten.Image)
}

// DebugFunc implements the Debugger interface
type DebugFunc func(g *GameScreen, screen *ebiten.Image)

// Debug calls the debug callback with the provided arguments
func (f DebugFunc) Debug(g *GameScreen, screen *ebiten.Image) {
	f(g, screen)
}

// Debuggers is a slice of a Debugger to make it easier to handle many debuggers
type Debuggers []Debugger

// Debug passes on the Debug call to all its child Debuggers
func (ds Debuggers) Debug(g *GameScreen, screen *ebiten.Image) {
	for _, d := range ds {
		d.Debug(g, screen)
	}
}

// Add is a shorthand for adding a child Debugger to the Debuggers
func (ds *Debuggers) Add(d Debugger) {
	*ds = append(*ds, d)
}
