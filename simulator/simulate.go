package main

import (
	"os"

	"github.com/Cameron-Kurotori/battlesnake/logging"
	"github.com/Cameron-Kurotori/battlesnake/sdk"
	"github.com/go-kit/log/level"
	"github.com/google/uuid"
)

func getSnakeList(m map[string]sdk.Battlesnake) []sdk.Battlesnake {
	l := []sdk.Battlesnake{}
	for _, snake := range m {
		l = append(l, snake)
	}
	return l
}

// returns true if defender lives
func headOn(defender, attacker sdk.Battlesnake) bool {
	return defender.Length > attacker.Length
}

type simulator struct {
	client BattlesnakeClient
}

func (s simulator) Simulate(width, height int) {

	u := uuid.New().String()

	snakes := map[string]sdk.Battlesnake{}

	state := sdk.GameState{
		Game: sdk.Game{
			ID: u,
		},
		Turn: 0,
		Board: sdk.Board{
			Height: height,
			Width:  width,
			Food:   []sdk.Coord{},
		},
	}

	var err error
	for len(snakes) > 1 {
		state.Board.Snakes = getSnakeList(snakes)
		moves := map[string]sdk.BattlesnakeMoveResponse{}
		for _, snake := range state.Board.Snakes {
			state.You = snake
			moves[snake.ID], err = s.client.Move(state)
			if err != nil {
				moves[snake.ID] = sdk.BattlesnakeMoveResponse{Move: sdk.DirectionToMove[snake.Direction()]}
			}
		}

		nextSnakes := []sdk.Battlesnake{}
		for snakeID, move := range moves {
			snake := snakes[snakeID]
			dir := sdk.MoveToDirection[move.Move]
			nextSnakes = append(nextSnakes, snake.Next(dir, state.Board))
		}

		dies := func(s sdk.Battlesnake) bool {
			for _, other := range nextSnakes {
				if !headOn(s, other) {
					return false
				}
			}
			return true
		}

		snakes = map[string]sdk.Battlesnake{}
		for _, snake := range nextSnakes {
			if !dies(snake) {
				snakes[snake.ID] = snake
			}
		}
		state.Turn++
	}
	_ = level.Info(logging.GlobalLogger()).Log("msg", "DONE")

}

func main() {
	port := os.Getenv("PORT")
	if len(port) == 0 {
		port = "8080"
	}
	host := "0.0.0.0"

	simulator := simulator{
		client: NewClient(host, port),
	}

	simulator.Simulate(11, 11)

}
