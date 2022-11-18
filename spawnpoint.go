// Copyright 2021 Siôn le Roux.  All rights reserved.
// Use of this source code is subject to an MIT-style
// licence which can be found in the LICENSE file.

package main

// SpawnPoints is an array of SpawnPoint
type SpawnPoints []*SpawnPoint

// SpawnPoint is a point on the map where zombies are spawn
type SpawnPoint struct {
	Position     Coord
	InitialCount int
	SpawnCount   int
	Continuous   bool
}

// Update updates the state of the spawn point
func (s *SpawnPoint) Update(g *Game) error {
	return nil
}
