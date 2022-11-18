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

// NextPosition gives the offset of the next spawning to the center of the point
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

// SpawnZombie spawns one zombie
func (s *SpawnPoint) SpawnZombie(g *Game) {
	np := s.NextPosition()

	z := NewZombie(Coord{ X: s.Position.X + np.X * 16, Y: s.Position.Y + np.Y * 16 }, g.ZombieSprites[rand.Intn(zombieTypes)])
	z.Target = g.Player.Object
	z.SpawnPoint = s

	g.Space.Add(z.Object)
	g.Zombies = append(g.Zombies, z)
	s.Zombies = append(s.Zombies, z)
}

// Update updates the state of the spawn point
func (s *SpawnPoint) Update(g *Game) {
	if s.InitialSpawned && !s.Continuous {
		return
	}

	// If the player is close to the spawn point then it is activated
	playerDistance := CalcDistance(s.Position.X, s.Position.Y, g.Player.Object.X, g.Player.Object.Y)

	if playerDistance < spawnDistance {
		// Intial spawning
		if !s.InitialSpawned {
			for i := 0; i < s.InitialCount; i++ {
				s.SpawnZombie(g)
			}
			s.InitialSpawned = true
		} else {
			// Continuous spawning one zombie if needed after a while
			if (g.Tick % 120 == 0) {
				if len(s.Zombies) < s.InitialCount {
					s.SpawnZombie(g)
				}
			}		
		}
	}
}

// RemoveZombie removes a dead zombie from the zombie array of the SpawnPoint
func (s *SpawnPoint) RemoveZombie(z *Zombie) {
	for i, sz := range s.Zombies {
		if sz == z {
			s.Zombies[i] = nil
			s.Zombies = append((s.Zombies)[:i], (s.Zombies)[i+1:]...)
		}
	}
}
