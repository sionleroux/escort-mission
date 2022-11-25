package main

// List of possible boss states
const (
	bossIdle1 = iota
	bossWalking1
	bossInjury1
	bossIdle2
	bossWalking2
	bossInjury2
	bossIdle3
	bossWalking3
	bossDeath
)

// Boss zombie, it is bigger than the rest and you need to kill it twice
type Boss struct {
	Zombie      // inner zombie behaviour
	Daemon bool // whether it has respawned into daemon mode aka Phase 2
}
