package main

// This file can be a nice home for your Battlesnake logic and related helper functions.
//
// We have started this for you, with a function to help remove the 'neck' direction
// from the list of possible moves!

import (
	"math"
	"math/rand"
	"sort"
	"time"

	"github.com/Cameron-Kurotori/battlesnake/logging"
	"github.com/go-kit/log"
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
	_ = level.Debug(state.Logger(logging.GlobalLogger())).Log("msg", "START")
}

// This function is called when a game your Battlesnake was in has ended.
// It's purely for informational purposes, you don't have to make any decisions here.
func end(state GameState) {
	_ = level.Debug(state.Logger(logging.GlobalLogger())).Log("msg", "END")
}

func nextBody(move Coord, body []Coord, board Board) []Coord {
	next := make([]Coord, len(body))
	next[0] = body[0].Add(move)
	for i, coord := range body[0 : len(body)-1] {
		next[i+1] = coord
	}
	if CoordSliceContains(next[0], board.Food) {
		next = append(next, body[len(body)-1])
	}
	return next
}

func headOnCollision(me, other []Coord) bool {
	return me[0] == other[0]
}

func bodyCollision(me, other []Coord) bool {
	return CoordSliceContains(me[0], other)
}

func dist(c1, c2 Coord) float64 {
	return math.Abs(float64(c1.X-c2.X)) + math.Abs(float64(c1.Y-c2.Y))
}

func numOpenSpaces(body []Coord, board Board) int {
	set := map[Coord]bool{}

	isOccupied := func(target Coord) bool {
		return board.OutOfBounds(target) ||
			board.PossiblyOccupied(target)
	}

	var recurse func(target Coord)
	recurse = func(target Coord) {
		if _, done := set[target]; done {
			return
		}
		if target != body[0] {
			if isOccupied(target) {
				return
			} else {
				set[target] = true
			}
		}
		recurse(Coord{target.X + 1, target.Y})
		recurse(Coord{target.X - 1, target.Y})
		recurse(Coord{target.X, target.Y + 1})
		recurse(Coord{target.X, target.Y - 1})
	}

	recurse(body[0])

	return len(set)
}

var comparator = map[Direction]func(c1, c2 Coord) bool{
	Direction_Up: func(c1, c2 Coord) bool {
		return c1.Y > c2.Y
	},
	Direction_Down: func(c1, c2 Coord) bool {
		return c1.Y < c2.Y
	},
	Direction_Left: func(c1, c2 Coord) bool {
		return c1.X < c2.X
	},
	Direction_Right: func(c1, c2 Coord) bool {
		return c1.X > c2.X
	},
}

func foodWeight(inDirection func(Coord, Coord) bool, head Coord, board Board) float64 {
	count := 1
	distAway := 0.0
	for _, food := range board.Food {
		if inDirection(food, head) {
			count++
			distAway += dist(head, food)
		}
	}
	ratioFood := float64(count) / float64(len(board.Food))
	avgDistAway := 5.0
	if count > 1 {
		avgDistAway = float64(distAway) / float64(count)
	}
	return float64(ratioFood) * (1 / avgDistAway)
}

func otherSnakeWeight(inDirection func(Coord, Coord) bool, me Battlesnake, board Board) float64 {
	head := me.Head
	count := 0
	distAway := 0.0
	for _, snake := range otherSnakes(me.ID, board.Snakes) {
		if inDirection(snake.Head, head) {
			if snake.Length >= me.Length {
				count++
				distAway += dist(head, snake.Head)
			} else {
				_ = level.Debug(logging.GlobalLogger()).Log("msg", "snake is shorter and in this direction... KILL THEM", "other_snake", snake.ID, "their_length", snake.Length, "snake_id", me.ID, "my_length", me.Length)
			}
		}
	}
	if count == 0 {
		return 1
	}
	avgDistAway := distAway / float64(count)

	return (avgDistAway / dist(Coord{0, 0}, Coord{board.Height, board.Width})) / float64(count)
}

func otherSnakes(myID string, snakes []Battlesnake) []Battlesnake {
	otherSnakes := make([]Battlesnake, len(snakes)-1)
	i := 0
	for _, snake := range snakes {
		if snake.ID == myID {
			continue
		}
		otherSnakes[i] = snake
		i++
	}
	return otherSnakes
}

type pMove struct {
	dir    BattlesnakeMove
	weight float64
}

func collisionWeight(dir Direction, me Battlesnake, board Board) float64 {
	weight := 1.0
	myNextBody := me.Next(dir, board)
	for _, snake := range otherSnakes(me.ID, board.Snakes) {
		for _, otherDir := range snake.Moves() {
			nextSnake := snake.Next(otherDir, board)
			if headOnCollision(myNextBody, nextSnake) && me.Length < snake.Length {
				weight *= 1.0 / 3
			}
			if bodyCollision(myNextBody, nextSnake) {
				return 0
			}
		}
	}
	return weight
}

func edgeWeight(dir Direction, me Battlesnake, board Board) float64 {
	nextHead := me.Next(dir, board)[0]
	closestX := math.Min(float64(nextHead.X), float64(board.Width-nextHead.X))
	closestY := math.Min(float64(nextHead.Y), float64(board.Width-nextHead.Y))
	return (closestX / float64(board.Width) / 2.0) * (closestY / float64(board.Height) / 2.0)
}

// This function is called on every turn of a game. Use the provided GameState to decide
// where to move -- valid moves are BattlesnakeMove_Up, BattlesnakeMove_Down, BattlesnakeMove_Left, or BattlesnakeMove_Right.
// We've provided some code and comments to get you started.
func move(state GameState) BattlesnakeMoveResponse {
	start := time.Now()
	logger := state.Logger(logging.GlobalLogger())

	possibleMoves := map[Direction]*pMove{}

	openSpacesOnBoard := state.Board.Height * state.Board.Width
	for _, snake := range state.Board.Snakes {
		openSpacesOnBoard -= int(snake.Length)
	}
	for _, dir := range state.You.Moves() {
		dirLogger := log.With(logger, "dir", dir)
		nextBody := state.You.Next(dir, state.Board)
		if state.Board.OutOfBounds(nextBody[0]) {
			_ = level.Debug(dirLogger).Log("msg", "out of bounds")
			continue
		} else if state.Board.Occupied(nextBody[0]) {
			_ = level.Debug(dirLogger).Log("msg", "occupied")
			continue
		}
		possibleMoves[dir] = &pMove{
			dir:    directionToMove[dir],
			weight: 1.0,
		}
		fWeight := foodWeight(comparator[dir], state.You.Head, state.Board)

		healthExponent := -3*math.Log(float64(state.You.Health-9)) + 10
		possibleMoves[dir].weight *= math.Pow(fWeight, healthExponent)

		snakeWeight := otherSnakeWeight(comparator[dir], state.You, state.Board)
		possibleMoves[dir].weight *= math.Pow(snakeWeight, 1.5)

		collisionWeight := collisionWeight(dir, state.You, state.Board)
		possibleMoves[dir].weight *= math.Pow(collisionWeight, 2)

		edgeWeight := edgeWeight(dir, state.You, state.Board)
		possibleMoves[dir].weight *= edgeWeight

		openSpaces := numOpenSpaces(state.You.Next(dir, state.Board), state.Board)
		possibleMoves[dir].weight *= math.Pow(float64(openSpaces)/float64(openSpacesOnBoard), 2)

		_ = level.Info(dirLogger).Log(
			"msg", "heuristics calculated",
			"collision_weight", collisionWeight,
			"edge_weight", edgeWeight,
			"final_weight", possibleMoves[dir].weight,
			"food_weight", fWeight,
			"health_exponent", healthExponent,
			"health", state.You.Health,
			"open_spaces", openSpaces,
			"snake_weight", snakeWeight,
			"total_open_spaces", openSpacesOnBoard,
		)

	}

	nextMove := possibleMoves[state.You.Direction()]

	possibleMovesList := []*pMove{}
	for _, m := range possibleMoves {
		possibleMovesList = append(possibleMovesList, m)
	}
	sort.Slice(possibleMovesList, func(i, j int) bool {
		return possibleMovesList[i].weight > possibleMovesList[j].weight
	})

	if len(possibleMovesList) > 0 {
		nextMove = possibleMovesList[0]
		if possibleMovesList[0].weight == 0.0 {
			_ = level.Debug(logger).Log("msg", "Moving randomly because no viable option")
			nextMove = possibleMovesList[rand.Intn(len(possibleMovesList))]
		}
	} else {
		_ = level.Debug(logger).Log("msg", "Absolutely no possible moves")
	}

	_ = level.Info(logger).Log("msg", "making move", "move", nextMove.dir, "weight", nextMove.weight, "took_ms", time.Since(start).Milliseconds())

	return BattlesnakeMoveResponse{
		Move: nextMove.dir,
	}
}
