package main

import (
	"log"

	"gopkg.in/ini.v1"
)

// ApplyConfigs overrides default values with a config file if available
func ApplyConfigs() {
	log.Println("Looking for INI file...")
	cfg, err := ini.Load("escort-mission.ini")
	if err != nil {
		log.Println("Error parsing INI file:", err)
		return
	}
	deathCoolDownTime, err = cfg.Section("").Key("DeathCoolDownTime").Int()
	if err != nil {
		log.Println("Error parsing INI file:", err)
	}
	hudPadding, err = cfg.Section("").Key("HudPadding").Int()
	if err != nil {
		log.Println("Error parsing INI file:", err)
	}
	startingCheckpoint, err = cfg.Section("").Key("StartingCheckpoint").Int()
	if err != nil {
		log.Println("Error parsing INI file:", err)
	}
	playerSpeed, err = cfg.Section("Player").Key("PlayerSpeed").Float64()
	if err != nil {
		log.Println("Error parsing INI file:", err)
	}
	playerSpeedFactorReverse, err = cfg.Section("Player").Key("PlayerSpeedFactorReverse").Float64()
	if err != nil {
		log.Println("Error parsing INI file:", err)
	}
	playerSpeedFactorSideways, err = cfg.Section("Player").Key("PlayerSpeedFactorSideways").Float64()
	if err != nil {
		log.Println("Error parsing INI file:", err)
	}
	playerSpeedFactorSprint, err = cfg.Section("Player").Key("PlayerSpeedFactorSprint").Float64()
	if err != nil {
		log.Println("Error parsing INI file:", err)
	}
	playerAmmoClipMax, err = cfg.Section("Player").Key("PlayerAmmoClipMax").Int()
	if err != nil {
		log.Println("Error parsing INI file:", err)
	}
	zombieSpeed, err = cfg.Section("Zombie").Key("ZombieSpeed").Float64()
	if err != nil {
		log.Println("Error parsing INI file:", err)
	}
	zombieCrawlerSpeed, err = cfg.Section("Zombie").Key("ZombieCrawlerSpeed").Float64()
	if err != nil {
		log.Println("Error parsing INI file:", err)
	}
	zombieSprinterSpeed, err = cfg.Section("Zombie").Key("ZombieSprinterSpeed").Float64()
	if err != nil {
		log.Println("Error parsing INI file:", err)
	}
	zombieRange, err = cfg.Section("Zombie").Key("ZombieRange").Float64()
	if err != nil {
		log.Println("Error parsing INI file:", err)
	}
	dogWalkingSpeed, err = cfg.Section("Dog").Key("DogWalkingSpeed").Float64()
	if err != nil {
		log.Println("Error parsing INI file:", err)
	}
	dogRunningSpeed, err = cfg.Section("Dog").Key("DogRunningSpeed").Float64()
	if err != nil {
		log.Println("Error parsing INI file:", err)
	}
	waitingRadius, err = cfg.Section("Dog").Key("WaitingRadius").Float64()
	if err != nil {
		log.Println("Error parsing INI file:", err)
	}
	followingRadius, err = cfg.Section("Dog").Key("FollowingRadius").Float64()
	if err != nil {
		log.Println("Error parsing INI file:", err)
	}
	zombieBarkRadius, err = cfg.Section("Dog").Key("ZombieBarkRadius").Float64()
	if err != nil {
		log.Println("Error parsing INI file:", err)
	}
	zombieFleeRadius, err = cfg.Section("Dog").Key("ZombieFleeRadius").Float64()
	if err != nil {
		log.Println("Error parsing INI file:", err)
	}
	zombieSafeRadius, err = cfg.Section("Dog").Key("ZombieSafeRadius").Float64()
	if err != nil {
		log.Println("Error parsing INI file:", err)
	}
	fleeingPathLength, err = cfg.Section("Dog").Key("FleeingPathLength").Float64()
	if err != nil {
		log.Println("Error parsing INI file:", err)
	}
	outOfSightLimit, err = cfg.Section("Dog").Key("OutOfSightLimit").Int()
	if err != nil {
		log.Println("Error parsing INI file:", err)
	}
}
