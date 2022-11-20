// Copyright 2021 Siôn le Roux.  All rights reserved.
// Use of this source code is subject to an MIT-style
// licence which can be found in the LICENSE file.

package main

import (
	"image"
	"math"
)

type LevelMap [][]int


func CreateMap(w, h int) LevelMap {
	lmap := make([][]int, h)
	for i := range lmap {
		lmap[i] = make([]int, w)
	}
	
	return lmap
}

// Neighbours implements the astar.Graph interface
func (m LevelMap) Neighbours(p image.Point) []image.Point {
	offsets := []image.Point{
		image.Pt(0, -1), // North
		image.Pt(1, 0),  // East
		image.Pt(0, 1),  // South
		image.Pt(-1, 0), // West
	}
	offsetsDiag := []image.Point{
		image.Pt(1, -1),  // NorthEast
		image.Pt(1, 1),   // SouthEast
		image.Pt(-1, 1),  // SouthWest
		image.Pt(-1, -1), // NorthWest
	}
	res := make([]image.Point, 0, 8)
	
	// Check avaialable diagonal neighbours (free only if there is a wide enough corridor)
	for _, off := range offsetsDiag {
		q := p.Add(off)
		q1 := p.Add(image.Pt(0, off.Y))
		q2 := p.Add(image.Pt(off.X, 0))
		if m.isFreeAt(q) && m.isFreeAt(q1) && m.isFreeAt(q2) {
			res = append(res, q)
		}
	}

	// Check avaialable  neighbours
	for _, off := range offsets {
		q := p.Add(off)
		if m.isFreeAt(q) {
			res = append(res, q)
		}
	}
	return res
}

// isFreeAt returns if the cell is free
func (m LevelMap) isFreeAt(p image.Point) bool {
	return m[p.Y][p.X] == 0
}

// distance calculates Euclidean distance between the points
func distance(p, q image.Point) float64 {
	d := q.Sub(p)
	return math.Sqrt(float64(d.X*d.X + d.Y*d.Y))
}

// SetObstacle sets the cell as obstacle
func (m LevelMap) SetObstacle(x, y int) {
	m[y][x] = 1
}
