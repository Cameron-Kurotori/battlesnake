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
	return CoordSliceContains(me[0], other[1:])
}

// the number of spaces available if taking position new head
func numOpenSpaces(logger log.Logger, newHead Coord, board Board) int {
	set := map[Coord]bool{}

	isOccupied := func(target Coord) bool {
		return board.OutOfBounds(target) ||
			board.Occupied(target)
	}

	var recurse func(target Coord)
	recurse = func(target Coord) {
		if _, done := set[target]; done || (target != newHead && isOccupied(target)) {
			return
		}
		set[target] = true

		recurse(Coord{target.X + 1, target.Y})
		recurse(Coord{target.X - 1, target.Y})
		recurse(Coord{target.X, target.Y + 1})
		recurse(Coord{target.X, target.Y - 1})
	}

	recurse(newHead)

	return len(set) - 1
}

// 1.0 = far away
// -> 0 very close
func findClosest(dir Direction, me Battlesnake, board Board, coords []Coord) float64 {
	max := float64(board.Width + board.Height)
	distance := max
	for _, c := range coords {
		if c.InDirectionOf(me.Head, dir) {
			if d := float64(me.Head.Manhattan(c)); d < distance {
				distance = d
			}
		}
	}
	return distance / max
}

// 1.0 = no collision predicted
// 0.0 = guaranteed collision
func collisionWeight(logger log.Logger, dir Direction, me Battlesnake, board Board) float64 {
	weight := 1.0
	myNextBody := me.Next(dir, board).Body
	for _, snake := range board.OtherSnakes(me.ID) {
		snakeCollisionScore := 1.0
		for _, otherDir := range snake.Moves(logger) {
			nextSnake := snake.Next(otherDir, board).Body
			if headOnCollision(myNextBody, nextSnake) && me.Length < snake.Length {
				snakeCollisionScore -= 1.0 / 3.0
			} else if bodyCollision(myNextBody, nextSnake) {
				snakeCollisionScore -= 1.0 / 3.0
			}
		}
		weight *= snakeCollisionScore
	}
	return weight
}

// 0.0 - many snakes (dangerous) in this direction or close
// 1.0 - not many snakes (dangerous) in this direction or far away
func calculateSnakeWeight(dir Direction, me Battlesnake, board Board) float64 {
	totalSnakeDistances := 0
	directionalDistances := []int{}
	for _, snake := range board.OtherSnakes(me.ID) {
		if me.Length-snake.Length >= 1 {
			snakeDist := snake.Head.Manhattan(me.Head)
			totalSnakeDistances += snakeDist
			if snake.Head.InDirectionOf(me.Head, dir) {
				directionalDistances = append(directionalDistances, snakeDist)
			}
		}
	}
	if len(directionalDistances) == 0 {
		return 1.0
	}

	sum := 0.0
	for _, dist := range directionalDistances {
		sum += math.Pow(float64(dist)/float64(totalSnakeDistances), 2)
	}
	return math.Sqrt(sum)
}

// 1.0 = furthest possibly away
// 0.0 = on border
func edgeWeight(dir Direction, me Battlesnake, board Board) float64 {
	nextHead := me.Next(dir, board).Head
	closestX := math.Min(float64(nextHead.X), float64(board.Width-nextHead.X)) + 1
	closestY := math.Min(float64(nextHead.Y), float64(board.Width-nextHead.Y)) + 1
	return (closestX / float64(board.Width+1) / 2.0) * (closestY / float64(board.Height+1) / 2.0)
}

type pMove struct {
	dir    BattlesnakeMove
	weight float64
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

	otherSnakes := state.Board.OtherSnakes(state.You.ID)

	totalLenDiff := 0.0
	for _, snake := range otherSnakes {
		totalLenDiff += float64(snake.Length - state.You.Length)
	}
	avgLenDiff := totalLenDiff / float64(len(otherSnakes))

	snakeHeads := make([]Coord, len(otherSnakes))
	for i, snake := range otherSnakes {
		snakeHeads[i] = snake.Head
	}

	for _, dir := range state.You.Moves(logger) {
		dirLogger := log.With(logger, "dir", dir)
		nextBody := state.You.Next(dir, state.Board).Body
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

		// further = 1, closer -> 0
		foodDistRatio := findClosest(dir, state.You, state.Board, state.Board.Food)
		foodExponent := 1.0
		if state.You.Health < 60 || avgLenDiff > -1 {
			foodExponent = math.Max(1.0, -math.Log(float64(state.You.Health-5))+5)
			foodDistRatio = 1 - foodDistRatio
		}
		possibleMoves[dir].weight *= math.Pow(foodDistRatio, foodExponent)

		snakeWeight := calculateSnakeWeight(dir, state.You, state.Board)
		possibleMoves[dir].weight *= math.Pow(snakeWeight, 1.5)

		collisionWeight := collisionWeight(dirLogger, dir, state.You, state.Board)
		possibleMoves[dir].weight *= math.Pow(collisionWeight, 2)

		edgeWeight := edgeWeight(dir, state.You, state.Board)
		possibleMoves[dir].weight *= math.Pow(edgeWeight, math.Sqrt(float64(state.Turn+1))/6.0)

		openSpaces := numOpenSpaces(dirLogger, state.You.Next(dir, state.Board).Head, state.Board)
		possibleMoves[dir].weight *= math.Pow(float64(openSpaces)/float64(openSpacesOnBoard+1), 3)

		if math.IsNaN(possibleMoves[dir].weight) {
			possibleMoves[dir].weight = -100
		}
		_ = level.Info(dirLogger).Log(
			"msg", "heuristics calculated",
			"collision_weight", collisionWeight,
			"edge_weight", edgeWeight,
			"final_weight", possibleMoves[dir].weight,
			"food_distance_ratio", foodDistRatio,
			"health", state.You.Health,
			"open_spaces", openSpaces,
			"snake_weight", snakeWeight,
			"total_open_spaces", openSpacesOnBoard,
		)

	}

	possibleMovesList := []*pMove{}
	for _, m := range possibleMoves {
		possibleMovesList = append(possibleMovesList, m)
	}
	sort.Slice(possibleMovesList, func(i, j int) bool {
		return possibleMovesList[i].weight > possibleMovesList[j].weight
	})

	var nextMove *pMove
	if len(possibleMovesList) > 0 {
		nextMove = possibleMovesList[0]
		if possibleMovesList[0].weight == 0.0 {
			_ = level.Debug(logger).Log("msg", "Moving randomly because no viable option")
			nextMove = possibleMovesList[rand.Intn(len(possibleMovesList))]
		}
	} else {
		nextMove = &pMove{
			dir: BattlesnakeMove_Right,
		}
		_ = level.Debug(logger).Log("msg", "Absolutely no possible moves")
	}

	err := level.Info(logger).Log("msg", "making move", "move", nextMove.dir, "weight", nextMove.weight, "took_ms", time.Since(start).Milliseconds())
	if err != nil {
		_ = level.Error(logger).Log("msg", "erorr while logging", "err", err)
	}

	return BattlesnakeMoveResponse{
		Move: nextMove.dir,
	}
}
