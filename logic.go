package main

// This file can be a nice home for your Battlesnake logic and related helper functions.
//
// We have started this for you, with a function to help remove the 'neck' direction
// from the list of possible moves!

import (
	"math"

	"github.com/Cameron-Kurotori/battlesnake/logging"
	"github.com/Cameron-Kurotori/battlesnake/sdk"
	"github.com/go-kit/log/level"
)

// This function is called when you register your Battlesnake on play.battlesnake.com
// See https://docs.battlesnake.com/guides/getting-started#step-4-register-your-battlesnake
// It controls your Battlesnake appearance and author permissions.
// For customization options, see https://docs.battlesnake.com/references/personalization
// TIP: If you open your Battlesnake URL in browser you should see this data.
func info() sdk.BattlesnakeInfoResponse {
	_ = level.Info(logging.GlobalLogger()).Log("msg", "info'")
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
	_ = level.Info(state.Logger(logging.GlobalLogger())).Log("msg", "start'")
}

// This function is called when a game your Battlesnake was in has ended.
// It's purely for informational purposes, you don't have to make any decisions here.
func end(state sdk.GameState) {
	_ = level.Info(state.Logger(logging.GlobalLogger())).Log("msg", "end'")
}

// higher means more desirable
func heuristic(pHeadon int) int {
	return -pHeadon
}

func snakeWillDie(snake sdk.Battlesnake, board sdk.Board) bool {
	for _, dir := range board.Moves(snake) {
		nextSnake := snake.Next(dir, board.Food, board.Hazards)
		if !nextSnake.Dead {
			return true
		}
	}
	return false
}

func safeSpaceRegardlessOfMovement(dir sdk.Direction, state sdk.GameState) bool {
	meNext := state.You.Next(dir, state.Board.Food, state.Board.Hazards)
	for _, snake := range state.Board.Snakes {
		if snake.ID == state.You.ID {
			continue
		}
		if !snakeWillDie(snake, state.Board) && sdk.CoordSliceContains(meNext.Head, snake.Body[:snake.Length-1]) {
			return false
		}
	}
	return true
}

func potentialHeadonLosses(dir sdk.Direction, state sdk.GameState) int {
	next := state.You.Next(dir, state.Board.Food, state.Board.Hazards)
	count := 0
	for _, snake := range state.Board.OtherSnakes(state.You.ID) {
		if snake.Length < state.You.Length {
			continue
		}
		for _, oDir := range state.Board.Moves(snake) {
			if snake.Head.Add(sdk.Coord(oDir)) == next.Head {
				count++
			}
		}
	}
	return count
}

// This function is called on every turn of a game. Use the provided GameState to decide
// where to move -- valid moves are "up", "down", "left", or "right".
// We've provided some code and comments to get you started.
func move(state sdk.GameState) sdk.BattlesnakeMoveResponse {
	bestDir := state.You.Direction()
	maxHeuristic := math.MinInt64
	for _, dir := range state.Board.Moves(state.You) {
		if !safeSpaceRegardlessOfMovement(dir, state) {
			continue
		}
		potentialHeadOn := potentialHeadonLosses(dir, state)
		score := heuristic(potentialHeadOn)
		if score > maxHeuristic {
			maxHeuristic = score
			bestDir = dir
		}
	}
	return sdk.BattlesnakeMoveResponse{
		Move: sdk.DirectionToMove[bestDir],
	}
}
