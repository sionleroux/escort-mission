package main

import (
	"errors"
)

//go:generate ./tools/gen_sprite_tags.sh assets/sprites/Zombie_big.json boss_anim.go boss

// Boss zombie, it is bigger than the rest and you need to kill it twice
type Boss struct {
	*Zombie                     // Inner zombie behaviour
	Daemon    bool              // Whether it has respawned into daemon mode aka Phase 2
	BossState bossAnimationTags // Boss animation state
	BossFrame int               // Current animation frame
	Dead      bool              // Whether the boss has reached its final death
}

// Update boss-specific zombie behaviour
func (z *Boss) Update(g *GameScreen) error {
	if z.Dead {
		return errors.New("Zombie Boss died")
	}

	if z.Daemon {
		switch z.HitToDie {
		case 2:
			switch z.Zombie.State {
			case zombieHit:
				z.BossState = bossDeath1
			case zombieIdle:
				z.BossState = bossIdle4
			case zombieWalking:
				z.BossState = bossRunning
			}
		case 1:
			switch z.Zombie.State {
			case zombieHit:
				z.BossState = bossDeath2
			}
		}
	} else {
		switch z.HitToDie {
		case 7, 6:
			switch z.Zombie.State {
			case zombieIdle:
				z.BossState = bossIdle1
			case zombieWalking:
				z.BossState = bossWalking1
			case zombieHit:
				z.BossState = bossHit1
			}
		case 5:
			switch z.Zombie.State {
			case zombieHit:
				z.BossState = bossHit1
			case zombieIdle:
				z.BossState = bossIdle2
			case zombieWalking:
				z.BossState = bossWalking2
			}
		case 4:
			switch z.Zombie.State {
			case zombieHit:
				z.BossState = bossHit2
			case zombieIdle:
				z.BossState = bossIdle2
			case zombieWalking:
				z.BossState = bossWalking2
			}
		case 3:
			switch z.Zombie.State {
			case zombieHit:
				z.BossState = bossHit2
			case zombieIdle:
				z.BossState = bossIdle3
			case zombieWalking:
				z.BossState = bossWalking3
			}
		case 2:
			switch z.Zombie.State {
			case zombieHit:
				z.BossState = bossDeath1
			case zombieIdle:
				z.BossState = bossIdle4
			case zombieWalking:
				z.BossState = bossRunning
			}
		case 1:
			switch z.Zombie.State {
			case zombieHit:
				z.BossState = bossDeath2
			}
		}
	}

	z.BossFrame = Animate(z.BossFrame, g.Tick, z.Sprite.Meta.FrameTags[z.BossState])
	if z.Frame == z.Sprite.Meta.FrameTags[z.BossState].To {
		z.outterAnimationBasedStateChanges(g)
	}

	if z.BossState == bossDeath1 || z.BossState == bossDeath2 || z.BossState == bossPhase2 {
		return nil
	}

	err := z.Zombie.Update(g)
	return err // probably dead inside, return early without handling
}

// Draw draws the Zombie to the screen
func (z *Boss) Draw(g *GameScreen) {
	z.Frame = z.BossFrame
	z.Zombie.Draw(g)
}

// Animation-trigged state changes
func (z *Boss) outterAnimationBasedStateChanges(g *GameScreen) {
	switch z.BossState {
	case bossHit1:
		z.BossState = bossWalking1
		z.Zombie.State = zombieWalking
	case bossHit2:
		z.BossState = bossWalking2
		z.Zombie.State = zombieWalking
	case bossDeath1:
		z.Daemon = true
		z.Speed = zombieSprinterSpeed * 2
		z.BossState = bossPhase2
		z.Zombie.State = zombieWalking
	case bossPhase2:
		z.BossState = bossRunning
		z.Zombie.State = zombieWalking
	case bossDeath2:
		z.Die(g)
		z.Zombie.State = zombieDead
	}
}

func (z *Boss) Remove() {
	z.SpawnPoint.RemoveZombie(z)
}

func (z *Boss) Die(g *GameScreen) {
	z.Zombie.Die(g)
	z.Dead = true
}
