package main

// This file can be a nice home for your Battlesnake logic and related helper functions.
//
// We have started this for you, with a function to help remove the 'neck' direction
// from the list of possible moves!

import (
	"log"
	"math"
	"sort"
	"time"

	"github.com/Cameron-Kurotori/battlesnake/logging"
	"github.com/go-kit/log/level"
)

// This function is called when you register your Battlesnake on play.battlesnake.com
// See https://docs.battlesnake.com/guides/getting-started#step-4-register-your-battlesnake
// It controls your Battlesnake appearance and author permissions.
// For customization options, see https://docs.battlesnake.com/references/personalization
// TIP: If you open your Battlesnake URL in browser you should see this data.
func info() BattlesnakeInfoResponse {
	_ = level.Debug(logging.GlobalLogger()).Log("msg", "INFO")
	return BattlesnakeInfoResponse{
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
func start(state GameState) {
	_ = level.Debug(logging.GlobalLogger()).Log("msg", "START", "game_id", state.Game.ID, "snake_id", state.You.ID)
	log.Printf("%s START\n", state.Game.ID)
}

// This function is called when a game your Battlesnake was in has ended.
// It's purely for informational purposes, you don't have to make any decisions here.
func end(state GameState) {
	log.Printf("%s END\n\n", state.Game.ID)
}

func food(state GameState, dir Direction) float64 {
	score := 0.0
	for _, food := range state.Board.Food {
		if inDirectionOf[dir](state.You.Head, food) {
			score += float64(state.You.Head.Manhattan(food))
		}
	}

	return score
}

func enemies(state GameState, dir Direction) float64 {
	score := 0.0
	for _, snake := range state.Board.Snakes {
		if snake.ID == state.You.ID {
			continue
		}
		if inDirectionOf[dir](state.You.Head, snake.Head) {
			score -= float64(state.Board.Height * state.Board.Width * state.You.Head.Manhattan(snake.Head))
		}
	}

	return score
}

func trapped(state GameState, dir Direction) float64 {
	set := map[Coord]bool{}
	nextSnake := state.You.Next(dir, state.Board)

	var recurse func(Coord)
	recurse = func(target Coord) {
		if _, done := set[target]; done {
			return
		}

		if state.Board.Occupied(target) || state.Board.OutOfBounds(target) {
			return
		}

		set[target] = true

		for next := range directionToMove {
			recurse(target.Add(Coord(next)))
		}
	}

	for next := range directionToMove {
		if Coord(next) == Coord(dir).Reverse() {
			continue
		}
		recurse(nextSnake[0].Add(Coord(next)))
	}

	return float64(len(set))
}

const (
	foodCoefficient = 1.0
	foodExponent    = 1

	enemyCoefficient = 5.0
	enemyExponent    = 1.0

	trappedCoefficient = 5.0
	trappedExponent    = 1.0
)

// This function is called on every turn of a game. Use the provided GameState to decide
// where to move -- valid moves are "up", "down", "left", or "right".
// We've provided some code and comments to get you started.
func move(state GameState) BattlesnakeMoveResponse {
	start := time.Now()
	logger := state.Logger(logging.GlobalLogger())
	survivability := map[Direction]float64{}
	for _, move := range state.You.Moves() {
		score := 0.0

		logKV := []interface{}{"msg", "heuristics calculated", "direction", move}

		foodScore := food(state, move)
		score += (foodCoefficient * math.Pow(float64(foodScore), foodExponent)) / float64(state.Board.Height*state.Board.Width*len(state.Board.Food))
		logKV = append(logKV, "food_score", foodScore)

		enemyScore := enemies(state, move)
		score += enemyCoefficient * math.Pow(float64(enemyScore), enemyExponent)
		logKV = append(logKV, "enemy_score", enemyScore)

		trappedScore := trapped(state, move)
		score += trappedCoefficient * math.Pow(float64(trappedScore), trappedExponent)
		logKV = append(logKV, "trapped_score", trappedScore)

		_ = level.Info(logger).Log(logKV...)

		survivability[move] = score
	}

	movesList := []Direction{}
	for d := range survivability {
		movesList = append(movesList, d)
	}

	direction := state.You.Direction()
	if len(movesList) > 0 {
		sort.Slice(movesList, func(i, j int) bool {
			return survivability[movesList[i]] > survivability[movesList[j]]
		})
		direction = movesList[0]
	} else {
		_ = level.Warn(logger).Log("msg", "no available moves, continuing in direction", "direction", direction)
	}

	move := directionToMove[direction]
	kv := []interface{}{
		"msg", "making move",
		"move", move,
		"took_ms", time.Since(start).Milliseconds(),
	}

	for dir, score := range survivability {
		kv = append(kv, directionToMove[dir]+"_score", score)
	}
	_ = level.Info(logger).Log(kv...)

	return BattlesnakeMoveResponse{
		Move: move,
	}
}
