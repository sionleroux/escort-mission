package main

type bossAnimationTags uint8

const (
	bossIdle1 bossAnimationTags = iota
	bossWalking1
	bossHit1
	bossIdle2
	bossWalking2
	bossHit2
	bossWalking3
	bossIdle3
	bossDeath1
	bossPhase2
	bossIdle4
	bossRunning
	bossDeath2
)

// Boss zombie, it is bigger than the rest and you need to kill it twice
type Boss struct {
	Zombie      // inner zombie behaviour
	Daemon bool // whether it has respawned into daemon mode aka Phase 2
}
