// Copyright 2021 Siôn le Roux.  All rights reserved.
// Use of this source code is subject to an MIT-style
// licence which can be found in the LICENSE file.

package main

import (
	"image"
	"math"

	"github.com/fzipp/astar"
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

// FindPath finds path between two coordinates on map
func (m LevelMap) FindPath(start, dest Coord) []Coord {
	var result []Coord

	startCell := image.Pt(int(start.X / 32), int(start.Y / 32))
	destCell := image.Pt(int(dest.X / 32), int(dest.Y / 32))
	apath := astar.FindPath[image.Point](m, startCell, destCell, distance, distance)
	apath = simplifyPath(apath)
	for _, p := range apath {
		result = append(result, Coord{X: float64(p.X)*32+16, Y: float64(p.Y)*32+16})
	}
	return result
}

// simplifyPath removes unnecessary points from the path
func simplifyPath(path []image.Point) []image.Point {
	var result []image.Point
	
	prev := image.Pt(0, 0)
	for i, p := range path {
		if i+1 < len(path) {
			next := path[i+1].Sub(p)
			if prev == next {
				continue
			} else {
				prev = next
			}
		}
		result = append(result, p)
	}
	return result
}
