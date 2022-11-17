// Copyright 2022 Siôn le Roux.  All rights reserved.
// Use of this source code is subject to an MIT-style
// licence which can be found in the LICENSE file.

package main

import (
	"embed"
	"encoding/json"
	"image/png"
	"io/ioutil"
	"log"
	"path"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/audio/vorbis"
	"github.com/solarlune/ldtkgo"
)

//go:embed assets/*
var assets embed.FS

// Frame is a single frame of an animation, usually a sub-image of a larger
// image containing several frames
type Frame struct {
	Duration int           `json:"duration"`
	Position FramePosition `json:"frame"`
}

// FramePosition represents the position of a frame, including the top-left
// coordinates and its dimensions (width and height)
type FramePosition struct {
	X int `json:"x"`
	Y int `json:"y"`
	W int `json:"w"`
	H int `json:"h"`
}

// FrameTags contains tag data about frames to identify different parts of an
// animation, e.g. idle animation, jump animation frames etc.
type FrameTags struct {
	Name      string `json:"name"`
	From      int    `json:"from"`
	To        int    `json:"to"`
	Direction string `json:"direction"`
}

// Frames is a slice of frames used to create sprite animation
type Frames []Frame

// SpriteMeta contains sprite meta-data, basically everything except frame data
type SpriteMeta struct {
	ImageName string      `json:"image"`
	FrameTags []FrameTags `json:"frameTags"`
}

// SpriteSheet is the root-node of sprite data, it contains frames and meta data
// about them
type SpriteSheet struct {
	Sprite Frames     `json:"frames"`
	Meta   SpriteMeta `json:"meta"`
	Image  *ebiten.Image
}

// SpriteType is a unique identifier to load a sprite by name
type SpriteType uint64

const (
	spritePlayer SpriteType = iota
	spriteDog
)

const zombieTypes = 5

// Load a sprite image and associated meta-data given a file name (without
// extension)
func loadSprite(name string) *SpriteSheet {
	name = path.Join("assets", "sprites", name)
	log.Printf("loading %s\n", name)

	file, err := assets.Open(name + ".json")
	if err != nil {
		log.Fatalf("error opening file %s: %v\n", name, err)
	}
	defer file.Close()

	data, err := ioutil.ReadAll(file)
	if err != nil {
		log.Fatal(err)
	}

	var ss SpriteSheet
	json.Unmarshal(data, &ss)
	if err != nil {
		log.Fatal(err)
	}

	ss.Image = loadImage(name + ".png")

	return &ss
}

// Convenience function to load entity/waypoint PNGs from the maps folder
func loadEntityImage(name string) *ebiten.Image {
	name = path.Join("assets", "maps", name) + ".png"
	return loadImage(name)
}

// Load an image from embedded FS into an ebiten Image object
func loadImage(name string) *ebiten.Image {
	log.Printf("loading %s\n", name)

	file, err := assets.Open(name)
	if err != nil {
		log.Fatalf("error opening file %s: %v\n", name, err)
	}
	defer file.Close()

	raw, err := png.Decode(file)
	if err != nil {
		log.Fatalf("error decoding file %s as PNG: %v\n", name, err)
	}

	return ebiten.NewImageFromImage(raw)
}

// Load an project from embedded FS into an LDtk Project object
func loadMaps(name string) *ldtkgo.Project {
	log.Printf("loading %s\n", name)

	file, err := assets.Open(name)
	if err != nil {
		log.Fatalf("error opening file %s: %v\n", name, err)
	}
	defer file.Close()

	data, err := ioutil.ReadAll(file)
	if err != nil {
		log.Fatalf("error reading from file %s: %v\n", name, err)
	}

	// Load the LDtk Project
	maps, err := ldtkgo.Read(data)
	if err != nil {
		log.Fatalf("error parsing file %s as LDtk Project: %v\n", name, err)
	}

	return maps
}

// SoundType is a unique identifier to reference sound by name
type SoundType uint64

const (
	soundMusicBackground SoundType = iota
	soundGunShot
	soundGunReload
	soundDogBark1
)

// NewMusicPlayer loads a sound into an audio player that can be used to play it
// as an infinite loop of music without any additional setup required
func NewMusicPlayer(music *vorbis.Stream, context *audio.Context) *audio.Player {
	musicLoop := audio.NewInfiniteLoop(music, music.Length())
	musicPlayer, err := audio.NewPlayer(context, musicLoop)
	if err != nil {
		log.Fatalf("error making music player: %v\n", err)
	}
	return musicPlayer
}

// NewSoundPlayer loads a sound into an audio player that can be used to play it
// without any additional setup required
func NewSoundPlayer(audioFile *vorbis.Stream, context *audio.Context) *audio.Player {
	audioPlayer, err := audio.NewPlayer(context, audioFile)
	if err != nil {
		log.Fatalf("error making audio player: %v\n", err)
	}
	return audioPlayer
}

// Load an OGG Vorbis sound file with 44100 sample rate and return its stream
func loadSoundFile(name string, sampleRate int) *vorbis.Stream {
	log.Printf("loading %s\n", name)

	file, err := assets.Open(name)
	if err != nil {
		log.Fatalf("error opening file %s: %v\n", name, err)
	}
	defer file.Close()

	music, err := vorbis.DecodeWithSampleRate(sampleRate, file)
	if err != nil {
		log.Fatalf("error decoding file %s as Vorbis: %v\n", name, err)
	}

	return music
}
