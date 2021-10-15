package main

// This file can be a nice home for your Battlesnake logic and related helper functions.
//
// We have started this for you, with a function to help remove the 'neck' direction
// from the list of possible moves!

import (
	"log"
	"sort"

	"github.com/Cameron-Kurotori/battlesnake/sdk"
)

// This function is called when you register your Battlesnake on play.battlesnake.com
// See https://docs.battlesnake.com/guides/getting-started#step-4-register-your-battlesnake
// It controls your Battlesnake appearance and author permissions.
// For customization options, see https://docs.battlesnake.com/references/personalization
// TIP: If you open your Battlesnake URL in browser you should see this data.
func info() sdk.BattlesnakeInfoResponse {
	log.Println("INFO")
	return sdk.BattlesnakeInfoResponse{
		APIVersion: "1",
		Author:     "cameron-kurotori", // TODO: Your Battlesnake username
		Color:      "#0f3d17",          // TODO: Personalize
		Head:       "tiger-king",       // TODO: Personalize
		Tail:       "tiger-tail",       // TODO: Personalize
	}
}

// This function is called everytime your Battlesnake is entered into a game.
// The provided GameState contains information about the game that's about to be played.
// It's purely for informational purposes, you don't have to make any decisions here.
func start(state sdk.GameState) {
	log.Printf("%s START\n", state.Game.ID)
}

// This function is called when a game your Battlesnake was in has ended.
// It's purely for informational purposes, you don't have to make any decisions here.
func end(state sdk.GameState) {
	log.Printf("%s END\n\n", state.Game.ID)
}

// This function is called on every turn of a game. Use the provided GameState to decide
// where to move -- valid moves are "up", "down", "left", or "right".
// We've provided some code and comments to get you started.
func move(state sdk.GameState) sdk.BattlesnakeMoveResponse {
	deathRatios := map[sdk.Direction]float64{}
	dirs := []sdk.Direction{}
	for _, dir := range state.Board.Moves(state.You) {
		death, total := DeathDepth(dir, state, 4)
		deathRatios[dir] = float64(death) / float64(total)
		dirs = append(dirs, dir)
	}
	if len(dirs) > 0 {
		sort.Slice(dirs, func(i, j int) bool {
			di, dj := dirs[i], dirs[j]
			return deathRatios[di] < deathRatios[dj]
		})

		return sdk.BattlesnakeMoveResponse{
			Move: sdk.DirectionToMove[dirs[0]],
		}

	}
	return sdk.BattlesnakeMoveResponse{
		Move: sdk.BattlesnakeMove_Right,
	}
}
