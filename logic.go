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
func heuristic(state sdk.GameState, pHeadon, openSpaceCount, distanceToClosestFood int) float64 {

	headOnFinal := -float64(5 * pHeadon)

	totalOpenSpace := 0
	for _, snake := range state.Board.Snakes {
		totalOpenSpace += int(snake.Length)
	}

	openSpaceFinal := (3 * float64(openSpaceCount) / float64(totalOpenSpace))

	totalSnakeLengthDiff := 0
	for _, snake := range state.Board.OtherSnakes(state.You.ID) {
		totalSnakeLengthDiff += int(state.You.Length) - int(snake.Length)
	}
	foodDistanceFinal := float64(distanceToClosestFood) / float64(state.Board.Width+state.Board.Height)
	if state.You.Health > 50 && float64(totalSnakeLengthDiff)/float64(len(state.Board.Snakes)-1) >= 2 {
		foodDistanceFinal = -foodDistanceFinal
	}
	foodDistanceFinal = (1.0 + (2.0 / (1.0 + math.Pow(math.E, float64(state.You.Health-60)/10.0)))) * foodDistanceFinal

	return headOnFinal + openSpaceFinal + foodDistanceFinal
}

func snakeWillDie(snake sdk.Battlesnake, board sdk.Board) bool {
	for _, dir := range board.Moves(snake) {
		nextSnake := snake.Next(dir, board.Food, board.Hazards)
		if nextSnake.Dead {
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

func openSpaceCount(dir sdk.Direction, state sdk.GameState) int {
	nextState := state.Next(map[string]sdk.Direction{state.You.ID: dir})

	open := map[sdk.Coord]bool{}
	var recurse func(target sdk.Coord)
	recurse = func(target sdk.Coord) {
		if _, ok := open[target]; ok {
			return
		}
		if target != nextState.You.Head {
			if nextState.Board.OutOfBounds(target) || nextState.Board.Occupied(target) {
				return
			}
		}
		open[target] = true
		for dir := range sdk.DirectionToMove {
			recurse(target.Add(sdk.Coord(dir)))
		}
	}
	recurse(nextState.You.Head)
	return len(open) - 1
}

func distanceToClosestFood(dir sdk.Direction, state sdk.GameState) int {
	nextSnake := state.You.Next(dir, state.Board.Food, state.Board.Hazards)
	closest := state.Board.Width + state.Board.Height + 1
	for _, food := range state.Board.Food {
		if d := food.Manhattan(nextSnake.Head); d < closest {
			closest = d
		}
	}
	return closest
}

// This function is called on every turn of a game. Use the provided GameState to decide
// where to move -- valid moves are "up", "down", "left", or "right".
// We've provided some code and comments to get you started.
func move(state sdk.GameState) sdk.BattlesnakeMoveResponse {
	bestDir := state.You.Direction()
	maxHeuristic := float64(math.MinInt64)
	for _, dir := range state.Board.Moves(state.You) {
		if !safeSpaceRegardlessOfMovement(dir, state) {
			continue
		}
		potentialHeadOn := potentialHeadonLosses(dir, state)
		openSpaceCount := openSpaceCount(dir, state)
		closestFood := distanceToClosestFood(dir, state)
		score := heuristic(state, potentialHeadOn, openSpaceCount, closestFood)
		if score > maxHeuristic {
			maxHeuristic = score
			bestDir = dir
		}
	}
	return sdk.BattlesnakeMoveResponse{
		Move: sdk.DirectionToMove[bestDir],
	}
}
