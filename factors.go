package main

import (
	"math"

	"github.com/Cameron-Kurotori/battlesnake/sdk"
)

func AllPossibleGameStates(dir sdk.Direction, state sdk.GameState) []sdk.GameState {
	cProduct := []map[string]sdk.Direction{{state.You.ID: dir}}
	for _, snake := range state.Board.Snakes {
		if snake.ID == state.You.ID {
			continue
		}
		newCProduct := []map[string]sdk.Direction{}
		for _, m := range cProduct {
			for _, snakeDir := range state.Board.Moves(snake) {
				newM := map[string]sdk.Direction{
					snake.ID: snakeDir,
				}
				for k, v := range m {
					newM[k] = v
				}
				newCProduct = append(newCProduct, newM)
			}
		}
		cProduct = newCProduct
	}

	gameStates := []sdk.GameState{}
	for _, dirs := range cProduct {
		nextState := state.Next(dirs)
		gameStates = append(gameStates, nextState)
	}
	return gameStates
}

// Death returns number of death states and total number of states checked
func Death(dir sdk.Direction, state sdk.GameState) (deathStates int, totalStates int) {
	gameStates := AllPossibleGameStates(dir, state)
	deaths := 0
	for _, gameState := range gameStates {
		if gameState.You.Dead {
			deaths++
		}
	}
	return deaths, len(gameStates)
}

func DeathDepth(dir sdk.Direction, state sdk.GameState, depth int) (deathStates int, totalStates int) {
	gameStates := AllPossibleGameStates(dir, state)
	deaths := 0
	total := 0
	for _, gameState := range gameStates {
		if !gameState.You.Dead {
			if depth > 0 {
				for nextDir := range sdk.DirectionToMove {
					nextDeaths, nextTotalCount := DeathDepth(nextDir, gameState, depth-1)
					deaths += nextDeaths
					total += nextTotalCount
				}
			} else {
				total++
			}
		} else {
			if depth > 0 {
				numSubsequentStates := int(4 * math.Pow(3, float64(depth*len(state.Board.Snakes)-1)))
				deaths += numSubsequentStates
				total += numSubsequentStates
			} else {
				total++
				deaths++
			}
		}
	}
	return deaths, total
}
