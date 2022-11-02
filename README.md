# Escort Mission

A basic skeleton for a 2D game using the [Ebiten](https://ebiten.org/) library.

## For game testers

For alpha testing use [this link](https://nightly.link/sinisterstuf/escort-mission/workflows/build/main/escort-mission-bundle.zip) to download the latest development build.

Game controls:
- F: toggle full-screen
- Q: quit the game
- WASD and mouse to move around

## For programmers

Make sure you have [Go 1.19 or later](https://go.dev/) to contribute to the game

To build the game yourself, run: `go build .` it will produce an escort-mission file and on Windows escort-mission.exe.

To run the tests, run: `go test ./...` but there are no tests yet.

The project has a very simple, flat structure, the first place to start looking is the main.go file.
