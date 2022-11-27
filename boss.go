package main

//go:generate ./tools/gen_sprite_tags.sh assets/sprites/Zombie_big.json boss_anim.go boss

// Boss zombie, it is bigger than the rest and you need to kill it twice
type Boss struct {
	*Zombie      // Inner zombie behaviour
	Daemon  bool // Whether it has respawned into daemon mode aka Phase 2
}
