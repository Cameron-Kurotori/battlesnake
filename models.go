package main

import (
	"math"

	"github.com/Cameron-Kurotori/battlesnake/logging"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
)

type GameState struct {
	Game  Game        `json:"game"`
	Turn  int         `json:"turn"`
	Board Board       `json:"board"`
	You   Battlesnake `json:"you"`
}

func (state GameState) Logger(logger log.Logger) log.Logger {
	return log.With(logger, "game_id", state.Game.ID, "snake_id", state.You.ID, "alive_snakes", len(state.Board.Snakes), "turn", state.Turn)
}

type Game struct {
	ID      string  `json:"id"`
	Ruleset Ruleset `json:"ruleset"`
	Timeout int32   `json:"timeout"`
}

type Ruleset struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type Board struct {
	Height int           `json:"height"`
	Width  int           `json:"width"`
	Food   []Coord       `json:"food"`
	Snakes []Battlesnake `json:"snakes"`

	// Used in non-standard game modes
	Hazards []Coord `json:"hazards"`
}

func (b Board) OutOfBounds(c Coord) bool {
	return c.X >= b.Width ||
		c.X < 0 ||
		c.Y >= b.Height ||
		c.Y < 0
}

// Occupied returns back true if regardless of any movement if the coordinate will be
// occupied by a hazard or a snake body
func (b Board) Occupied(c Coord) bool {
	if CoordSliceContains(c, b.Hazards) {
		return true
	}
	for _, snake := range b.Snakes {
		if CoordSliceContains(c, snake.Body[:snake.Length-1]) {
			return true
		}
	}
	return false
}

func (b Board) PossiblyOccupied(c Coord) bool {
	if CoordSliceContains(c, b.Hazards) {
		return true
	}
	for _, snake := range b.Snakes {
		// if there's a possibility of snake growing, assume it grows
		bodyLen := snake.Length
		for _, move := range snake.Moves() {
			if CoordSliceContains(snake.Next(move, b)[0], b.Food) {
				bodyLen = snake.Length + 1
			}
		}
		if CoordSliceContains(c, snake.Body[:bodyLen-1]) {
			return true
		}
	}
	return false
}

type Battlesnake struct {
	ID      string  `json:"id"`
	Name    string  `json:"name"`
	Health  int32   `json:"health"`
	Body    []Coord `json:"body"`
	Head    Coord   `json:"head"`
	Length  int32   `json:"length"`
	Latency string  `json:"latency"`

	// Used in non-standard game modes
	Shout string `json:"shout"`
	Squad string `json:"squad"`
}

// Next returns back a new slice of coordinates the represents the new snake body
// If addOne is `true` then the body has an additional segment
func (snake Battlesnake) Next(dir Direction, board Board) []Coord {
	nextBody := make([]Coord, 1)
	nextBody[0] = Coord(dir).Add(snake.Body[0])
	nextBody = append(nextBody, snake.Body...)
	if CoordSliceContains(nextBody[0], board.Food) {
		nextBody = nextBody[:len(nextBody)-1]
	}
	return nextBody
}

func (snake Battlesnake) Moves() []Direction {
	moves := []Direction{}
	snakeDirection := snake.Direction()
	for _, dir := range moveToDirection {
		if Coord(dir) != Coord(snakeDirection).Reverse() {
			moves = append(moves, dir)
		}
	}
	_ = level.Debug(logging.GlobalLogger()).Log("msg", "calculating possible moves", "moves", moves, "snake_direction", snakeDirection, "snake_id", snake.ID)
	return moves
}

func (snake Battlesnake) Direction() Direction {
	if snake.Length < 2 {
		return Direction_Right
	}
	head, neck := snake.Head, snake.Body[1]
	return Direction(head.Add(neck.Reverse()))
}

type Direction Coord

var (
	Direction_Up    = Direction{0, 1}
	Direction_Down  = Direction{0, -1}
	Direction_Left  = Direction{-1, 0}
	Direction_Right = Direction{1, 0}
)

var moveToDirection = map[BattlesnakeMove]Direction{
	BattlesnakeMove_Down:  Direction_Down,
	BattlesnakeMove_Up:    Direction_Up,
	BattlesnakeMove_Left:  Direction_Left,
	BattlesnakeMove_Right: Direction_Right,
}

var directionToMove = map[Direction]BattlesnakeMove{
	Direction_Down:  BattlesnakeMove_Down,
	Direction_Up:    BattlesnakeMove_Up,
	Direction_Left:  BattlesnakeMove_Left,
	Direction_Right: BattlesnakeMove_Right,
}

type Coord struct {
	X int `json:"x"`
	Y int `json:"y"`
}

// Add gets the sum of the individual axis of this coordinate and another: {x1 + x2, y1 + y2}
func (c Coord) Add(other Coord) Coord {
	return Coord{c.X + other.X, c.Y + other.Y}
}

// Reverse reverses the coordinate: {-1 * x, -1 * y}
func (c Coord) Reverse() Coord {
	return Coord{-c.X, -c.Y}
}

// Euclidean calculates the euclidean (actual) distance: ((x2 - x1)^2) + (y2 - y1)^2)^0.5
func (c Coord) Euclidean(other Coord) float64 {
	diff := c.Add(other.Reverse())
	return math.Sqrt(math.Pow(float64(diff.X), 2) + math.Pow(float64(diff.Y), 2))
}

// Manhattan calculates the manhattan distance: |x2 - x1| + |y2 - y1|
func (c Coord) Manhattan(other Coord) int {
	diff := c.Add(other.Reverse())
	return int(math.Abs(float64(diff.X)) + math.Abs(float64(diff.Y)))
}

// CoordSliceContains returns back whether elem is contained in slice
func CoordSliceContains(elem Coord, slice []Coord) bool {
	for _, coord := range slice {
		if elem == coord {
			return true
		}
	}
	return false
}

type CoordComparator func(Coord, Coord) bool

var inDirectionOf = map[Direction]CoordComparator{
	Direction_Down: func(source, target Coord) bool {
		return target.Y < source.Y
	},
	Direction_Up: func(source, target Coord) bool {
		return target.Y > source.Y
	},
	Direction_Left: func(source, target Coord) bool {
		return target.X < source.X
	},
	Direction_Right: func(source, target Coord) bool {
		return target.X > source.X
	},
}

// Response Structs

type BattlesnakeInfoResponse struct {
	APIVersion string `json:"apiversion"`
	Author     string `json:"author"`
	Color      string `json:"color"`
	Head       string `json:"head"`
	Tail       string `json:"tail"`
}

type BattlesnakeMove string

const (
	BattlesnakeMove_Up    BattlesnakeMove = "up"
	BattlesnakeMove_Down  BattlesnakeMove = "down"
	BattlesnakeMove_Left  BattlesnakeMove = "left"
	BattlesnakeMove_Right BattlesnakeMove = "right"
)

type BattlesnakeMoveResponse struct {
	Move  BattlesnakeMove `json:"move"`
	Shout string          `json:"shout,omitempty"`
}
