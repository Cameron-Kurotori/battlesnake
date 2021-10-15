package sdk

import (
	"math"

	"github.com/go-kit/log"
)

type GameState struct {
	Game  Game        `json:"game"`
	Turn  int         `json:"turn"`
	Board Board       `json:"board"`
	You   Battlesnake `json:"you"`
}

// Next gives you the next game state (minus new hazards and new food)
func (state GameState) Next(dirs map[string]Direction) (nextState GameState) {
	var you Battlesnake
	nextSnakes := make([]Battlesnake, len(state.Board.Snakes))
	snakeCoords := map[Coord]bool{}
	for i, snake := range state.Board.Snakes {
		dir, ok := dirs[snake.ID]
		if !ok {
			dir = snake.Direction()
		}
		nextSnake := snake.Next(dir, state.Board.Food, state.Board.Hazards)
		if snake.ID == state.You.ID {
			you = nextSnake
		}
		nextSnakes[i] = nextSnake
		for _, piece := range nextSnake.Body {
			snakeCoords[piece] = true
		}
	}

	newFood := []Coord{}
	for _, food := range state.Board.Food {
		if _, ok := snakeCoords[food]; !ok {
			newFood = append(newFood, food)
		}
	}

	isDead := func(snake Battlesnake) bool {
		if snake.Health <= 0 {
			return true
		}
		if state.Board.OutOfBounds(snake.Head) {
			return true
		}
		for _, oSnake := range nextSnakes {
			if oSnake.Health > 0 {
				// dead snakes can't kill
				if oSnake.ID != snake.ID {
					if oSnake.Head == snake.Head {
						// head on collision
						if oSnake.Length >= snake.Length {
							return true
						}
					}
				}
				if CoordSliceContains(snake.Head, oSnake.Body[1:]) {
					return true
				}
			}
		}
		return false
	}

	aliveSnakes := []Battlesnake{}
	// check death states
	for _, snake := range nextSnakes {
		if !isDead(snake) {
			aliveSnakes = append(aliveSnakes, snake)
		} else if snake.ID == you.ID {
			you.Dead = true
		}
	}

	state.You = you
	state.Board.Snakes = aliveSnakes
	state.Board.otherSnakes = nil
	state.Board.Food = newFood
	state.Turn++
	return state
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
	Height      int           `json:"height"`
	Width       int           `json:"width"`
	Food        []Coord       `json:"food"`
	Snakes      []Battlesnake `json:"snakes"`
	otherSnakes map[string][]Battlesnake

	// Used in non-standard game modes
	Hazards []Coord `json:"hazards"`
}

func (b Board) OtherSnakes(myID string) []Battlesnake {
	if b.otherSnakes == nil {
		b.otherSnakes = map[string][]Battlesnake{}
	}
	others := b.otherSnakes[myID]
	if others == nil {
		others = make([]Battlesnake, len(b.Snakes)-1)
		i := 0
		for _, snake := range b.Snakes {
			if snake.ID == myID {
				continue
			}
			others[i] = snake
			i++
		}
		b.otherSnakes[myID] = others
	}
	return others
}

func (b Board) Moves(snake Battlesnake) []Direction {
	moves := []Direction{}
	for _, move := range snake.Moves() {
		if !b.OutOfBounds(snake.Next(move, b.Food, b.Hazards).Head) {
			moves = append(moves, move)
		}
	}
	return moves
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
		if CoordSliceContains(c, snake.Body[:snake.Length]) {
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
	Dead    bool    `json:"-"`

	// Used in non-standard game modes
	Shout string `json:"shout"`
	Squad string `json:"squad"`
}

// Next returns back a new slice of coordinates the represents the new snake body
func (snake Battlesnake) Next(dir Direction, food []Coord, hazards []Coord) Battlesnake {
	tail := snake.Tail()
	nextBody := append([]Coord{snake.Head.Add(Coord(dir))}, snake.Body[:len(snake.Body)-1]...)
	health := snake.Health

	if CoordSliceContains(nextBody[0], food) {
		nextBody = append(nextBody, tail)
		health = 100
	} else {
		if CoordSliceContains(nextBody[0], hazards) {
			health -= 15
		} else {
			health--
		}
	}

	// no calculations should be done once we start changing struct
	snake.Body = nextBody
	snake.Head = nextBody[0]
	snake.Length = int32(len(nextBody))
	snake.Health = health

	return snake
}

func (snake Battlesnake) Moves() []Direction {
	moves := []Direction{}
	snakeDirection := snake.Direction()
	for _, dir := range MoveToDirection {
		if Coord(dir) != Coord(snakeDirection).Reverse() &&
			snake.Head.Add(Coord(dir)) != snake.Tail() {
			moves = append(moves, dir)
		}
	}
	return moves
}

func (snake Battlesnake) Direction() Direction {
	if snake.Length < 2 || snake.Body[0] == snake.Body[1] {
		return Direction_Right
	}
	head, neck := snake.Head, snake.Body[1]
	return Direction(head.Add(neck.Reverse()))
}

func (snake Battlesnake) Tail() Coord {
	return snake.Body[len(snake.Body)-1]
}

type Direction Coord

var (
	Direction_Up    = Direction{0, 1}
	Direction_Down  = Direction{0, -1}
	Direction_Left  = Direction{-1, 0}
	Direction_Right = Direction{1, 0}
)

var MoveToDirection = map[BattlesnakeMove]Direction{
	BattlesnakeMove_Down:  Direction_Down,
	BattlesnakeMove_Up:    Direction_Up,
	BattlesnakeMove_Left:  Direction_Left,
	BattlesnakeMove_Right: Direction_Right,
}

var DirectionToMove = map[Direction]BattlesnakeMove{
	Direction_Down:  BattlesnakeMove_Down,
	Direction_Up:    BattlesnakeMove_Up,
	Direction_Left:  BattlesnakeMove_Left,
	Direction_Right: BattlesnakeMove_Right,
}

type Coord struct {
	X int `json:"x"`
	Y int `json:"y"`
}

func (c Coord) InDirectionOf(source Coord, dir Direction) bool {
	return inDirectionOf[dir](source, c)
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
