package main

// This file can be a nice home for your Battlesnake logic and related helper functions.
//
// We have started this for you, with a function to help remove the 'neck' direction
// from the list of possible moves!

import (
	"log"
	"math"
	"math/rand"
	"sort"
)

// This function is called when you register your Battlesnake on play.battlesnake.com
// See https://docs.battlesnake.com/guides/getting-started#step-4-register-your-battlesnake
// It controls your Battlesnake appearance and author permissions.
// For customization options, see https://docs.battlesnake.com/references/personalization
// TIP: If you open your Battlesnake URL in browser you should see this data.
func info() BattlesnakeInfoResponse {
	log.Println("INFO")
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
	log.Printf("%s START\n", state.Game.ID)
}

// This function is called when a game your Battlesnake was in has ended.
// It's purely for informational purposes, you don't have to make any decisions here.
func end(state GameState) {
	log.Printf("%s END\n\n", state.Game.ID)
}

func contains(coord Coord, set []Coord) bool {
	for _, c := range set {
		if coord.X == c.X && coord.Y == c.Y {
			return true
		}
	}
	return false
}

func nextBody(move Coord, body []Coord, board Board) []Coord {
	next := make([]Coord, len(body))
	next[0] = Coord{
		X: body[0].X + move.X,
		Y: body[0].Y + move.Y,
	}
	for i, coord := range body[0 : len(body)-1] {
		next[i+1] = coord
	}
	if contains(next[0], board.Food) {
		next = append(next, body[len(body)-1])
	}
	return next
}

func headOnCollision(me, other []Coord) bool {
	return me[0] == other[0]
}

func bodyCollision(me, other []Coord) bool {
	return contains(me[0], other)
}

func dist(c1, c2 Coord) float64 {
	return math.Sqrt(math.Pow(float64(c1.X-c2.X), 2) + math.Pow(float64(c1.Y-c2.Y), 2))
}

func numOpenSpaces(body []Coord, board Board) int {
	set := map[Coord]bool{}

	isOccupied := func(target Coord) bool {
		return target.Y >= board.Height ||
			target.Y < 0 ||
			target.X >= board.Width ||
			target.X < 0 ||
			contains(target, board.Hazards) ||
			func() bool {
				for _, snake := range board.Snakes {
					if contains(target, snake.Body) {
						return true
					}
				}
				return false
			}()
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

func occupied(move Coord, me Battlesnake, board Board) bool {
	nextBody := nextBody(move, me.Body, board)
	nextHead := nextBody[0]

	boardWidth := board.Width
	boardHeight := board.Height

	return nextHead.Y >= boardHeight ||
		nextHead.Y < 0 ||
		nextHead.X >= boardWidth ||
		nextHead.X < 0 ||
		contains(nextHead, nextBody[1:]) ||
		func() bool {
			for _, snake := range otherSnakes(me.ID, board.Snakes) {
				if contains(nextHead, snake.Body) {
					return true
				}
			}
			return false
		}()
}

var comparator = map[string]func(c1, c2 Coord) bool{
	"up": func(c1, c2 Coord) bool {
		return c1.Y > c2.Y
	},
	"down": func(c1, c2 Coord) bool {
		return c1.Y < c2.Y
	},
	"left": func(c1, c2 Coord) bool {
		return c1.X < c2.X
	},
	"right": func(c1, c2 Coord) bool {
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

func direction(body []Coord) Coord {
	return Coord{body[1].X - body[0].X, body[1].Y - body[0].Y}
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
				log.Printf("snake %s is shorter and in this direction... KILL THEM! their_length=%d my_length=%d", snake.ID, snake.Length, me.Length)
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
	dir         string
	weight      float64
	permanentNo bool
}

func collisionWeight(dir Coord, me Battlesnake, board Board) float64 {
	weight := 1.0
	myNextBody := nextBody(dir, me.Body, board)
	for _, snake := range otherSnakes(me.ID, board.Snakes) {
		for _, otherDir := range []Coord{{0, 1}, {0, -1}, {1, 0}, {-1, 0}} {
			nextSnake := nextBody(otherDir, snake.Body, board)
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

// This function is called on every turn of a game. Use the provided GameState to decide
// where to move -- valid moves are "up", "down", "left", or "right".
// We've provided some code and comments to get you started.
func move(state GameState) BattlesnakeMoveResponse {
	possibleMoves := map[string]*pMove{
		"up":    {"up", 1, false},
		"down":  {"down", 1, false},
		"left":  {"left", 1, false},
		"right": {"right", 1, false},
	}

	moves := map[string]Coord{
		"right": {1, 0},
		"left":  {-1, 0},
		"up":    {0, 1},
		"down":  {0, -1},
	}

	for dir, move := range moves {
		if occupied(move, state.You, state.Board) {
			log.Printf("dir=%s occupied", dir)
			possibleMoves[dir].weight = 0
			possibleMoves[dir].permanentNo = true
		}
		fWeight := foodWeight(comparator[dir], state.You.Head, state.Board)

		healthScale := -math.Log2(float64(state.You.Health)) + 5.5
		log.Printf("snake=%s dir=%s health=%d health_scale=%f food_weight=%f", state.You.ID, dir, state.You.Health, healthScale, fWeight)
		possibleMoves[dir].weight *= math.Pow(fWeight, math.Max(healthScale, 1))

		snakeWeight := otherSnakeWeight(comparator[dir], state.You, state.Board)
		possibleMoves[dir].weight *= math.Pow(snakeWeight, 1.5)
		log.Printf("snake=%s dir=%s snake_weight=%f", state.You.ID, dir, snakeWeight)

		collisionWeight := collisionWeight(move, state.You, state.Board)
		possibleMoves[dir].weight *= math.Pow(collisionWeight, 2)
		log.Printf("snake=%s dir=%s collision_weight=%f", state.You.ID, dir, collisionWeight)

		openSpaces := numOpenSpaces(nextBody(moves[dir], state.You.Body, state.Board), state.Board)
		log.Printf("snake=%s dir=%s open_spaces=%d", state.You.ID, dir, openSpaces)
		possibleMoves[dir].weight *= math.Pow(float64(openSpaces)/float64((state.Board.Height)*(state.Board.Width)), 2)
	}

	nextMove := possibleMoves["up"]

	possibleMovesList := []*pMove{}
	for _, m := range possibleMoves {
		if !m.permanentNo {
			possibleMovesList = append(possibleMovesList, m)
		}
	}
	sort.Slice(possibleMovesList, func(i, j int) bool {
		return possibleMovesList[i].weight > possibleMovesList[j].weight
	})

	if len(possibleMovesList) > 0 {
		nextMove = possibleMovesList[0]
		if possibleMovesList[0].weight == 0.0 {
			log.Printf("Moving randomly because no viable option")
			nextMove = possibleMovesList[rand.Intn(len(possibleMovesList))]
		}
	} else {
		log.Printf("snake_id=%s Absolutely no possible moves\n", state.You.ID)
	}

	log.Printf("snake_id=%s %s MOVE %d: %s %f\n", state.You.ID, state.Game.ID, state.Turn, nextMove.dir, nextMove.weight)

	return BattlesnakeMoveResponse{
		Move: nextMove.dir,
	}
}
