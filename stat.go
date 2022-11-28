package main

import (
	"time"
)

const (
	counterBulletsFired = iota
	counterDryFires
	counterZombiesHit
	counterZombiesKilled
	counterPlayerDied
	counterDogDied
)

// Stat stores the game statistics
type Stat struct {
	GameStarted time.Time
	GameWon     time.Time
	Counters    []int
}
