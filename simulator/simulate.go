package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
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
	client *stateRecorder
}

func newNonCollidingSnakeHead(width, height, padding int, snakes []sdk.Battlesnake) sdk.Coord {
	c := sdk.Coord{X: rand.Intn(width), Y: rand.Intn(height)}
	for _, snake := range snakes {
		if snake.Head.Manhattan(c) < padding {
			return newNonCollidingSnakeHead(width, height, padding, snakes)
		}
	}
	return c
}

func initSnakes(count, width, height int) []sdk.Battlesnake {
	padding := 5
	snakes := make([]sdk.Battlesnake, count)
	for i := 0; i < count; i++ {
		head := newNonCollidingSnakeHead(width, height, padding, snakes)
		snakeID := uuid.NewString()
		snakes[i] = sdk.Battlesnake{
			ID:     snakeID,
			Name:   fmt.Sprintf("my-snake-%d", i),
			Health: 100,
			Body:   []sdk.Coord{head, head, head},
			Head:   head,
			Length: 3,
		}
		log.Printf("initialized snake %d: %s", i, snakeID)
	}
	return snakes
}

func (s simulator) Simulate(count, width, height int) {

	u := uuid.New().String()

	snakes := map[string]sdk.Battlesnake{}
	initializedSnakes := initSnakes(count, width, height)
	for _, snake := range initializedSnakes {
		snakes[snake.ID] = snake
	}

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
		log.Printf("%d snakes left\n", len(snakes))
		state.Board.Snakes = getSnakeList(snakes)
		moves := map[string]sdk.BattlesnakeMoveResponse{}
		for _, snake := range state.Board.Snakes {
			state.You = snake
			moves[snake.ID], err = s.client.Move(state)
			if err != nil {
				log.Printf("error obtaining move: %+v\n", err)
				moves[snake.ID] = sdk.BattlesnakeMoveResponse{Move: sdk.DirectionToMove[snake.Direction()]}
			}
		}

		nextSnakes := []sdk.Battlesnake{}
		for snakeID, move := range moves {
			snake := snakes[snakeID]
			log.Printf("snake %s moves %s\n", snake.Name, move.Move)
			dir := sdk.MoveToDirection[move.Move]
			nextSnakes = append(nextSnakes, snake.Next(dir, state.Board))
		}

		dies := func(s sdk.Battlesnake) bool {
			if state.Board.OutOfBounds(s.Head) {
				log.Printf("snake %s is out of bounds\n", s.Name)
				return true
			}
			for _, other := range nextSnakes {
				if headOn(s, other) {
					log.Printf("snake %s collides with %s\n", s.Name, other.Name)
					return true
				}
			}
			return false
		}

		snakes = map[string]sdk.Battlesnake{}
		for _, snake := range nextSnakes {
			if !dies(snake) {
				snakes[snake.ID] = snake
			}
		}
		state.Turn++
	}
	state.Board.Snakes = getSnakeList(snakes)
	allStates := append(s.client.States, state)
	statesOutput, _ := json.Marshal(allStates)
	_ = level.Info(logging.GlobalLogger()).Log("msg", "DONE")
	fmt.Println(string(statesOutput))
}

type stateRecorder struct {
	BattlesnakeClient
	States []sdk.GameState
}

func (s *stateRecorder) Move(state sdk.GameState) (sdk.BattlesnakeMoveResponse, error) {
	s.States = append(s.States, state)
	return s.BattlesnakeClient.Move(state)
}

func RecordStates(client BattlesnakeClient) *stateRecorder {
	return &stateRecorder{
		BattlesnakeClient: client,
	}
}

func main() {
	port := os.Getenv("PORT")
	if len(port) == 0 {
		port = "8080"
	}
	host := "0.0.0.0"

	simulator := simulator{
		client: RecordStates(NewClient(host, port)),
	}

	simulator.Simulate(4, 11, 11)
}
