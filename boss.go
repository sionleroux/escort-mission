package main

import (
	"errors"
	"log"
)

//go:generate ./tools/gen_sprite_tags.sh assets/sprites/Zombie_big.json boss_anim.go boss

// Boss zombie, it is bigger than the rest and you need to kill it twice
type Boss struct {
	*Zombie                   // Inner zombie behaviour
	Daemon  bool              // Whether it has respawned into daemon mode aka Phase 2
	State   bossAnimationTags // Boss animation state
	Frame   int               // Current animation frame
	Dying   bool              // Whether the death and respawn animation is still playing
	Dead    bool              // Whether the boss has reached its final death
}

// Update boss-specific zombie behaviour
func (z *Boss) Update(g *GameScreen) error {
	if z.Dead {
		return errors.New("Zombie Boss died")
	}

	if !z.Dying {
		z.hitpointBasedStateChanges()
	}

	z.Frame = Animate(z.Frame, g.Tick, z.Sprite.Meta.FrameTags[z.State])
	if z.Frame == z.Sprite.Meta.FrameTags[z.State].To {
		z.outterAnimationBasedStateChanges(g)
	}

	if z.State == bossDeath1 || z.State == bossDeath2 || z.State == bossPhase2 {
		return nil
	}

	err := z.Zombie.Update(g)
	return err // probably dead inside, return early without handling
}

// Draw draws the Zombie to the screen
func (z *Boss) Draw(g *GameScreen) {
	z.Zombie.Frame = z.Frame
	z.Zombie.Draw(g)
}

// Animation-trigged state changes
func (z *Boss) outterAnimationBasedStateChanges(g *GameScreen) {
	switch z.State {
	case bossHit1:
		z.State = bossWalking1
		z.Zombie.State = zombieWalking
	case bossHit2:
		z.State = bossWalking2
		z.Zombie.State = zombieWalking
	case bossDeath1:
		g.Sounds[soundBigZombieDeath1].Play()
		z.Daemon = true
		z.Speed = zombieSprinterSpeed * 2
		z.State = bossPhase2
		z.Zombie.State = zombieWalking
	case bossPhase2:
		g.Sounds[soundBigZombieScream].Play()
		z.Dying = false
		z.State = bossRunning
		z.Zombie.State = zombieWalking
	case bossDeath2:
		g.Sounds[soundBigZombieDeath2].Play()
		z.Die(g)
		z.Zombie.State = zombieDead
	}
}

func (z *Boss) hitpointBasedStateChanges() {
	if z.Daemon {
		switch z.HitToDie {
		case 2:
			switch z.Zombie.State {
			case zombieHit:
				z.State = bossDeath1
			case zombieIdle:
				z.State = bossIdle4
			case zombieWalking:
				z.State = bossRunning
			}
		case 1:
			switch z.Zombie.State {
			case zombieHit:
				z.State = bossDeath2
			}
		}
	} else {
		switch z.HitToDie {
		case 10, 9, 8:
			switch z.Zombie.State {
			case zombieIdle:
				z.State = bossIdle1
			case zombieWalking:
				z.State = bossWalking1
			case zombieHit:
				z.State = bossHit1
			}
		case 7: // transition to bleeding arm
			switch z.Zombie.State {
			case zombieHit:
				z.State = bossHit1
			case zombieIdle:
				z.State = bossIdle2
			case zombieWalking:
				z.State = bossWalking2
			}
		case 6, 5:
			switch z.Zombie.State {
			case zombieHit:
				z.State = bossHit2
			case zombieIdle:
				z.State = bossIdle2
			case zombieWalking:
				z.State = bossWalking2
			}
		case 4, 3: // transition to 2 bleeding arms
			switch z.Zombie.State {
			case zombieHit:
				z.State = bossHit2
			case zombieIdle:
				z.State = bossIdle3
			case zombieWalking:
				z.State = bossWalking3
			}
		case 2: // transition to red daemon
			switch z.Zombie.State {
			case zombieHit:
				z.Dying = true
				z.State = bossDeath1
			case zombieIdle:
				z.State = bossIdle4
			case zombieWalking:
				z.State = bossRunning
			}
		case 1:
			switch z.Zombie.State {
			case zombieHit:
				z.State = bossDeath2
			}
		}
	}
}

func (z *Boss) Remove() {
	z.SpawnPoint.RemoveZombie(z)
}

func (z *Boss) Die(g *GameScreen) {
	log.Println("Boss defeated!")
	z.Dead = true
	g.BossDefeated = true
	z.Zombie.Die(g)
}
