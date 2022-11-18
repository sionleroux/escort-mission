// Copyright 2021 Siôn le Roux.  All rights reserved.
// Use of this source code is subject to an MIT-style
// licence which can be found in the LICENSE file.

package main

import (
	"math/rand"
)

// spawnDistance is the distance where the point is activated
const spawnDistance = 250

// SpawnPoints is an array of SpawnPoint
type SpawnPoints []*SpawnPoint

// Update updates all the SpawnPoints
func (sps *SpawnPoints) Update(g *Game) {
	for _, s := range *sps {
		s.Update(g)
	}
}

// SpawnPoint is a point on the map where zombies are spawn
type SpawnPoint struct {
	Position       Coord
	InitialCount   int
	Continuous     bool
	Zombies        Zombies
	InitialSpawned bool
	PrevPosition   int
}

func (s *SpawnPoint) NextPosition() Coord {
	zombiePositions := []Coord{
		{0, 0}, {-1, 0}, {-1, -1},
		{0, -1}, {1, -1}, {1, 0},
		{1, 1}, {0, 1}, {-1, 1},
		{-2, 0}, {-2, -1}, {-2, -2},
		{-1, -2}, {0, -2}, {1, -2},
		{2, -2}, {2, -1}, {2, 0},
		{2, 1}, {2, 2}, {1, 2},
		{0, 2}, {-1, 2},
		{-2, 2}, {-2, 1},
	}
	s.PrevPosition++
	if s.PrevPosition == 25 {
		s.PrevPosition = 0
	}

	return zombiePositions[s.PrevPosition]
}

// Update updates the state of the spawn point
func (s *SpawnPoint) Update(g *Game) error {
	// If the player is close to the spawn point then it is activated
	playerDistance := CalcDistance(s.Position.X, s.Position.Y, g.Player.Object.X, g.Player.Object.Y)

	if playerDistance < spawnDistance {
		// Intial spawning
		if !s.InitialSpawned {
			for i := 0; i < s.InitialCount; i++ {
				np := s.NextPosition()

				z := NewZombie(Coord{ X: s.Position.X + np.X * 16, Y: s.Position.Y + np.Y * 16 }, g.ZombieSprites[rand.Intn(zombieTypes)])
				z.Target = g.Player.Object
			
				g.Space.Add(z.Object)
				g.Zombies = append(g.Zombies, z)
			}
			s.InitialSpawned = true
		}
	}
	
	return nil
}
