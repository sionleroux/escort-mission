package main

import (
	"time"
)

// Stat stores the game statistics
type Stat struct {
	GameStarted          time.Time
	GameWon              time.Time
	CounterBulletsFired  int
	CounterDryFires      int
	CounterZombiesHit    int
	CounterZombiesKilled int
	CounterPlayerDied    int
	CounterDogDied       int
}
