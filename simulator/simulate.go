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

func newAvoidingCoord(width, height, padding int, avoid []sdk.Coord) sdk.Coord {
	c := sdk.Coord{X: rand.Intn(width), Y: rand.Intn(height)}
	for _, coord := range avoid {
		if coord.Manhattan(c) < padding {
			return newAvoidingCoord(width, height, padding, avoid)
		}
	}
	return c
}

func initSnakes(count, width, height int) []sdk.Battlesnake {
	padding := 5
	snakes := make([]sdk.Battlesnake, count)
	heads := []sdk.Coord{}
	for i := 0; i < count; i++ {
		head := newAvoidingCoord(width, height, padding, heads)
		heads = append(heads, head)
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

func addFood(width, height int, snakes []sdk.Battlesnake, food []sdk.Coord) []sdk.Coord {
	foodDistribution := []int{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 1, 1, 1, 2, 2, 3}

	if len(food) > 10 {
		return food
	}
	avoid := []sdk.Coord{}
	for _, snake := range snakes {
		avoid = append(avoid, snake.Body...)
	}
	avoid = append(avoid, food...)
	for i := 0; i < foodDistribution[rand.Intn(len(foodDistribution))]; i++ {
		c := newAvoidingCoord(width, height, 2, avoid)
		food = append(food, c)
		avoid = append(avoid, c)
	}
	return food
}

func (s simulator) Simulate(count, width, height, maxTurns int) {

	u := uuid.New().String()

	snakes := map[string]sdk.Battlesnake{}
	initializedSnakes := initSnakes(count, width, height)
	for _, snake := range initializedSnakes {
		snakes[snake.ID] = snake
	}

	food := []sdk.Coord{}
	food = addFood(width, height, initializedSnakes, food)

	state := sdk.GameState{
		Game: sdk.Game{
			ID: u,
		},
		Turn: 0,
		Board: sdk.Board{
			Height: height,
			Width:  width,
			Food:   food,
		},
	}

	var err error

	for len(snakes) > 1 && (maxTurns < 0 || state.Turn < maxTurns) {
		log.Printf("%d snakes left: turn %d\n", len(snakes), state.Turn)
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
			if sdk.CoordSliceContains(s.Head, s.Body[1:]) {
				return true
			}
			for _, other := range nextSnakes {
				if other.ID == s.ID {
					continue
				}
				if other.Head == s.Head && !headOn(s, other) {
					log.Printf("snake %s head-to-head collides with %s\n", s.Name, other.Name)
					return true
				} else if sdk.CoordSliceContains(s.Head, other.Body) {
					log.Printf("snake %s collides with %s\n", s.Name, other.Name)
					return true
				}
			}
			return s.Health <= 0
		}

		newFood := []sdk.Coord{}
		for _, food := range state.Board.Food {
			eaten := false
			for _, snake := range nextSnakes {
				if snake.Head == food {
					eaten = true
					break
				}
			}
			if !eaten {
				newFood = append(newFood, food)
			}
		}

		newFood = addFood(width, height, initializedSnakes, newFood)
		state.Board.Food = newFood

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
	Seen   map[int]bool
	States []sdk.GameState
}

func (s *stateRecorder) Move(state sdk.GameState) (sdk.BattlesnakeMoveResponse, error) {
	if _, ok := s.Seen[state.Turn]; !ok {
		s.States = append(s.States, state)
		s.Seen[state.Turn] = true
	}
	return s.BattlesnakeClient.Move(state)
}

func RecordStates(client BattlesnakeClient) *stateRecorder {
	return &stateRecorder{
		Seen:              map[int]bool{},
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

	simulator.Simulate(4, 11, 11, 1000)
}
