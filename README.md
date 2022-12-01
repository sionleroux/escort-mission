# Escort Mission

Follow a dog through a post-apocalytpic wasteland full of zombies; this is our entry for [Game Off 2022](https://itch.io/jam/game-off-2022)

## For game testers

For alpha testing use [⬇️ this link](https://nightly.link/sinisterstuf/escort-mission/workflows/build/main/escort-mission-bundle.zip) to download the latest development build.

Game controls:
- F: toggle full-screen
- Q: quit the game
- WASD and mouse to move around
- click to shoot
- R to reload 
- Hold shift to sprint

If you find an issue with the game [please open a new ticket here](https://github.com/sinisterstuf/escort-mission/issues).

## For programmers

Make sure you have [Go 1.19 or later](https://go.dev/) to contribute to the game.  Get the source code at [github.com/sinisterstuf/escort-mission](https://github.com/sinisterstuf/escort-mission).

To build the game yourself, run: `go build .` it will produce an escort-mission file and on Windows escort-mission.exe.

To run the tests, run: `go test ./...` but there are no tests yet.

The project has a very simple, flat structure, the first place to start looking is the main.go file.
