package main

// This file can be a nice home for your sdk.Battlesnake logic and related helper functions.
//
// We have started this for you, with a function to help remove the 'neck' direction
// from the list of possible moves!

import (
	"math"
	"math/rand"
	"sort"
	"time"

	"github.com/Cameron-Kurotori/battlesnake/logging"
	"github.com/Cameron-Kurotori/battlesnake/sdk"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
)

// This function is called when you register your sdk.Battlesnake on play.battlesnake.com
// See https://docs.battlesnake.com/guides/getting-started#step-4-register-your-battlesnake
// It controls your sdk.Battlesnake appearance and author permissions.
// For customization options, see https://docs.battlesnake.com/references/personalization
// TIP: If you open your sdk.Battlesnake URL in browser you should see this data.
func info() sdk.BattlesnakeInfoResponse {
	_ = level.Debug(logging.GlobalLogger()).Log("msg", "INFO")
	return sdk.BattlesnakeInfoResponse{
		APIVersion: "1",
		Author:     "cameron-kurotori", // TODO: Your sdk.Battlesnake username
		Color:      "#0f3d17",          // TODO: Personalize
		Head:       "tiger-king",       // TODO: Personalize
		Tail:       "tiger-tail",       // TODO: Personalize
	}
}

// This function is called everytime your sdk.Battlesnake is entered into a game.
// The provided GameState contains information about the game that's about to be played.
// It's purely for informational purposes, you don't have to make any decisions here.
func start(state sdk.GameState) {
	_ = level.Debug(state.Logger(logging.GlobalLogger())).Log("msg", "START")
}

// This function is called when a game your sdk.Battlesnake was in has ended.
// It's purely for informational purposes, you don't have to make any decisions here.
func end(state sdk.GameState) {
	_ = level.Debug(state.Logger(logging.GlobalLogger())).Log("msg", "END")
}

func nextBody(move sdk.Coord, body []sdk.Coord, board sdk.Board) []sdk.Coord {
	next := make([]sdk.Coord, len(body))
	next[0] = body[0].Add(move)
	for i, coord := range body[0 : len(body)-1] {
		next[i+1] = coord
	}
	if sdk.CoordSliceContains(next[0], board.Food) {
		next = append(next, body[len(body)-1])
	}
	return next
}

func ratioSigmoid(x float64) float64 {
	return 1.0 / (1 + math.Pow(math.E, -10.0*(x-0.5)))
}

func headOnCollision(me, other []sdk.Coord) bool {
	return me[0] == other[0]
}

func bodyCollision(me, other []sdk.Coord) bool {
	return sdk.CoordSliceContains(me[0], other[1:])
}

func avgGuaranteedReduction(logger log.Logger, dir sdk.Direction, mySnakeID string, board sdk.Board) float64 {
	simulatedBoard := board
	simulatedBoard.Snakes = []sdk.Battlesnake{}
	snakeIndices := map[string]int{}
	var mySnake sdk.Battlesnake
	for i, snake := range board.Snakes {
		snakeIndices[snake.ID] = i
		simulatedBoard.Snakes = append(simulatedBoard.Snakes, snake)
		if snake.ID == mySnakeID {
			mySnake = snake
		}
	}

	count := 0
	guaranteedReduction := 0
	for _, snake := range board.Snakes {
		if moves := snake.Moves(logger); len(moves) == 1 {
			newHead := snake.Next(moves[0], simulatedBoard).Head
			spacesWithoutMyMove := numOpenSpaces(logger, newHead, simulatedBoard)
			simulatedBoard.Snakes[snakeIndices[mySnakeID]] = mySnake.Next(dir, board)
			spacesWithMyMove := numOpenSpaces(logger, newHead, simulatedBoard)
			simulatedBoard.Snakes[snakeIndices[mySnakeID]] = mySnake
			guaranteedReduction += spacesWithoutMyMove - spacesWithMyMove
			count++
		}
	}

	if count == 0 {
		return 0.0
	}

	return float64(guaranteedReduction) / float64(count)
}

// the number of spaces available if taking position new head
func numOpenSpaces(logger log.Logger, newHead sdk.Coord, board sdk.Board) int {
	set := map[sdk.Coord]bool{}

	isOccupied := func(target sdk.Coord) bool {
		return board.OutOfBounds(target) ||
			board.Occupied(target)
	}

	var recurse func(target sdk.Coord)
	recurse = func(target sdk.Coord) {
		if _, done := set[target]; done || (target != newHead && isOccupied(target)) {
			return
		}
		set[target] = true

		recurse(sdk.Coord{target.X + 1, target.Y})
		recurse(sdk.Coord{target.X - 1, target.Y})
		recurse(sdk.Coord{target.X, target.Y + 1})
		recurse(sdk.Coord{target.X, target.Y - 1})
	}

	recurse(newHead)

	return len(set) - 1
}

// [0, 1]
// 0 = lots, close
// 1 = few, far
func foodAvailability(dir sdk.Direction, me sdk.Battlesnake, board sdk.Board) (val float64) {
	defer func() {
		_ = level.Debug(logging.GlobalLogger()).Log("msg", "updating food availability val", "value", val)
		val = ratioSigmoid(val)
	}()
	if len(board.Food) == 0 {
		return 1.0
	}

	sum := 0.0
	for _, food := range board.Food {
		if food.InDirectionOf(me.Head, dir) {
			sum += math.Pow(float64(food.Manhattan(me.Head))/float64(board.Height+board.Width), 2.0)
		}
	}
	if sum == 0 {
		return 1.0
	}
	return math.Sqrt(sum / float64(len(board.Food)))
}

func immediateSpace(target sdk.Coord, board sdk.Board) float64 {
	count := 0
	for _, dir := range []sdk.Direction{
		{-1, 2}, {0, 2}, {1, 2},
		{-2, 1}, {-1, 1}, {0, 1}, {1, 1}, {2, 1},
		{-2, 0}, {-1, 0}, {1, 0}, {2, 0},
		{-2, -1}, {-1, -1}, {0, -1}, {1, -1}, {2, -1},
		{-1, -2}, {0, -2}, {1, -2},
	} {
		t := target.Add(sdk.Coord(dir))
		if !(board.OutOfBounds(t) || board.Occupied(t)) {
			count++
		}
	}
	return ratioSigmoid(float64(count) / 20.0) // entry always blocked
}

// [0, 1]
// 1.0 = no collision predicted
// 0.0 = guaranteed collision
func snakeCollisionScore(logger log.Logger, dir sdk.Direction, me sdk.Battlesnake, snake sdk.Battlesnake, board sdk.Board) float64 {
	myNextBody := me.Next(dir, board).Body
	snakeCollisionScore := 1.0
	for _, otherDir := range snake.Moves(logger) {
		nextSnake := snake.Next(otherDir, board).Body
		if headOnCollision(myNextBody, nextSnake) && me.Length < snake.Length {
			snakeCollisionScore -= 1.0 / 3.0
		} else if bodyCollision(myNextBody, nextSnake) {
			snakeCollisionScore -= 1.0 / 3.0
		}
	}
	return ratioSigmoid(snakeCollisionScore)
}

// [0, 1]
// 0 = many snakes (dangerous) in this direction or close 1 = not many snakes
// 1 = not many snakes (dangerous) in this direction or far away
func calculateSnakeWeight(dir sdk.Direction, me sdk.Battlesnake, board sdk.Board) float64 {
	totalSnakeDistances := 0
	directionalDistances := []int{}
	for _, snake := range board.OtherSnakes(me.ID) {
		if me.Length-snake.Length <= 1 {
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
	return ratioSigmoid(math.Sqrt(sum))
}

// [0, 1]
// 0 = on border
// 1 = in center
func edgeWeight(dir sdk.Direction, me sdk.Battlesnake, board sdk.Board) float64 {
	nextHead := me.Next(dir, board).Head
	closestX := math.Min(float64(nextHead.X), float64(board.Width-nextHead.X))
	closestY := math.Min(float64(nextHead.Y), float64(board.Width-nextHead.Y))
	return ratioSigmoid(closestX/math.Floor(float64(board.Width)/2.0)) * (closestY / math.Floor(float64(board.Height)/2.0))
}

type pMove struct {
	dir    sdk.BattlesnakeMove
	weight float64
}

func (p pMove) Weight() float64 {
	if math.IsNaN(p.weight) {
		return -100
	}
	return p.weight
}

type heuristicMover struct{}

var globalMover heuristicMover

// This function is called on every turn of a game. Use the provided GameState to decide
// where to move -- valid moves are sdk.BattlesnakeMove_Up, sdk.BattlesnakeMove_Down, sdk.BattlesnakeMove_Left, or sdk.BattlesnakeMove_Right.
// We've provided some code and comments to get you started.
func (m heuristicMover) Move(state sdk.GameState) sdk.BattlesnakeMoveResponse {
	start := time.Now()
	logger := state.Logger(logging.GlobalLogger())

	possibleMoves := map[sdk.Direction]*pMove{}

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

	snakeHeads := make([]sdk.Coord, len(otherSnakes))
	for i, snake := range otherSnakes {
		snakeHeads[i] = snake.Head
	}

	for _, dir := range state.You.Moves(logger) {
		dirLogger := log.With(logger, "dir", dir)
		nextSnake := state.You.Next(dir, state.Board)
		if state.Board.OutOfBounds(nextSnake.Head) {
			_ = level.Debug(dirLogger).Log("msg", "out of bounds")
			continue
		} else if state.Board.Occupied(nextSnake.Head) {
			_ = level.Debug(dirLogger).Log("msg", "occupied")
			continue
		}
		possibleMoves[dir] = &pMove{
			dir:    sdk.DirectionToMove[dir],
			weight: 1.0,
		}

		foodDistRatio := foodAvailability(dir, state.You, state.Board)
		foodExponent := 1.0
		if state.You.Health < 50 || avgLenDiff > -1.5 {
			foodExponent = 3 * (avgLenDiff + 2) / (float64(state.You.Health))
			foodDistRatio = 1 - foodDistRatio
		}

		foodDistRatio = math.Pow(foodDistRatio, foodExponent)
		possibleMoves[dir].weight *= foodDistRatio
		_ = level.Debug(dirLogger).Log("msg", "updated weight", "after", "food", "weight", possibleMoves[dir].Weight())

		// [0, 1]
		snakeWeight := math.Pow(calculateSnakeWeight(dir, state.You, state.Board), 1.0/4)
		possibleMoves[dir].weight *= snakeWeight
		_ = level.Debug(dirLogger).Log("msg", "updated weight", "after", "snake weight", "weight", possibleMoves[dir].Weight())

		allCollisionWeight := 1.0
		immediateCollisionWeight := 1.0
		for _, snake := range otherSnakes {
			collisionScore := snakeCollisionScore(logger, dir, state.You, snake, state.Board)
			allCollisionWeight *= collisionScore
			if snake.Head.Manhattan(nextSnake.Head) <= 2 {
				immediateCollisionScore := collisionScore
				if snake.Length > nextSnake.Length {
					immediateCollisionScore = 1 - immediateCollisionScore
				} else {
					immediateCollisionScore = math.Pow(immediateCollisionScore, 2.0)
				}
				immediateCollisionWeight *= immediateCollisionScore
			}
		}

		allCollisionWeight = math.Pow(allCollisionWeight, 3.0)
		possibleMoves[dir].weight *= allCollisionWeight
		_ = level.Debug(dirLogger).Log("msg", "updated weight", "after", "all collisions", "weight", possibleMoves[dir].Weight())

		immediateCollisionWeight = math.Pow(immediateCollisionWeight, 4.0)
		possibleMoves[dir].weight *= immediateCollisionWeight
		_ = level.Debug(dirLogger).Log("msg", "updated weight", "after", "immediate collisions", "weight", possibleMoves[dir].Weight())

		immediateSpaceScore := math.Pow(immediateSpace(nextSnake.Head, state.Board), 1.0/70.0)
		possibleMoves[dir].weight *= immediateSpaceScore
		_ = level.Debug(dirLogger).Log("msg", "updated weight", "after", "immediateSpaceScore", "weight", possibleMoves[dir].Weight())

		edgeWeight := edgeWeight(dir, state.You, state.Board)
		edgeWeight = math.Pow(edgeWeight, 1.0/16.0)
		possibleMoves[dir].weight *= edgeWeight
		_ = level.Debug(dirLogger).Log("msg", "updated weight", "after", "edge weight", "weight", possibleMoves[dir].Weight())

		openSpaces := numOpenSpaces(dirLogger, state.You.Next(dir, state.Board).Head, state.Board)
		openSpacesWeight := math.Pow(ratioSigmoid(float64(openSpaces)/float64(openSpacesOnBoard)), 1.2)
		possibleMoves[dir].weight *= openSpacesWeight
		_ = level.Debug(dirLogger).Log("msg", "updated weight", "after", "open spaces", "weight", possibleMoves[dir].Weight())

		avgGuaranteedReductionRatio := ratioSigmoid(avgGuaranteedReduction(dirLogger, dir, state.You.ID, state.Board) / float64(openSpacesOnBoard))
		avgGuaranteedReductionRatio = math.Pow(avgGuaranteedReductionRatio, 1.0/16)
		possibleMoves[dir].weight *= avgGuaranteedReductionRatio
		_ = level.Debug(dirLogger).Log("msg", "updated weight", "after", "guaranteed reduction", "weight", possibleMoves[dir].Weight())

		_ = level.Info(dirLogger).Log(
			"msg", "heuristics calculated",
			"can_kill_weight", avgGuaranteedReductionRatio,
			"collision_weight_all", allCollisionWeight,
			"collision_weight_immediate", immediateCollisionWeight,
			"edge_weight", edgeWeight,
			"final_weight", possibleMoves[dir].Weight(),
			"food_distance_ratio", foodDistRatio,
			"health", state.You.Health,
			"immediate_space", immediateSpaceScore,
			"open_spaces", openSpacesWeight,
			"snake_weight", snakeWeight,
			"total_open_spaces", openSpacesOnBoard,
		)

	}

	possibleMovesList := []*pMove{}
	for _, m := range possibleMoves {
		possibleMovesList = append(possibleMovesList, m)
	}
	sort.Slice(possibleMovesList, func(i, j int) bool {
		return possibleMovesList[i].Weight() > possibleMovesList[j].Weight()
	})

	var nextMove *pMove
	if len(possibleMovesList) > 0 {
		nextMove = possibleMovesList[0]
		if possibleMovesList[0].Weight() == 0.0 {
			_ = level.Debug(logger).Log("msg", "Moving randomly because no viable option")
			nextMove = possibleMovesList[rand.Intn(len(possibleMovesList))]
		}
	} else {
		nextMove = &pMove{
			dir: sdk.BattlesnakeMove_Right,
		}
		_ = level.Debug(logger).Log("msg", "Absolutely no possible moves")
	}

	err := level.Info(logger).Log("msg", "making move", "move", nextMove.dir, "weight", nextMove.Weight(), "took_ms", time.Since(start).Milliseconds())
	if err != nil {
		_ = level.Error(logger).Log("msg", "erorr while logging", "err", err)
	}

	return sdk.BattlesnakeMoveResponse{
		Move: nextMove.dir,
	}
}
