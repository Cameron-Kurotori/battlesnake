package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNeckAvoidance(t *testing.T) {
	// Arrange
	me := Battlesnake{
		// Length 3, facing right
		Head: Coord{X: 2, Y: 0},
		Body: []Coord{{X: 2, Y: 0}, {X: 1, Y: 0}, {X: 0, Y: 0}},
	}
	state := GameState{
		Board: Board{
			Snakes: []Battlesnake{me},
		},
		You: me,
	}

	// Act 1,000x (this isn't a great way to test, but it's okay for starting out)
	for i := 0; i < 1000; i++ {
		nextMove := move(state)
		// Assert never move left
		if nextMove.Move == "left" {
			t.Errorf("snake moved onto its own neck, %s", nextMove.Move)
		}
	}
}

func TestMove(t *testing.T) {
	state := GameState{
		You: Battlesnake{
			Body:   []Coord{{9, 7}, {8, 7}, {8, 8}, {7, 8}, {7, 9}, {8, 9}, {9, 9}, {10, 9}, {10, 8}, {10, 7}, {10, 6}, {10, 5}, {9, 5}},
			Head:   Coord{9, 7},
			Health: 97,
		},
		Board: Board{
			Height: 11,
			Width:  11,
			Food:   []Coord{{8, 0}, {10, 2}, {4, 5}, {5, 5}, {5, 6}, {5, 7}, {7, 7}, {9, 5}},
			Snakes: []Battlesnake{{
				Body: []Coord{{5, 1}, {5, 0}, {4, 0}, {4, 1}, {4, 2}, {3, 2}, {3, 1}, {3, 0}, {2, 0}, {2, 1}, {1, 1}, {1, 0}, {0, 0}, {0, 1}, {0, 2}},
				Head: Coord{5, 1},
			}},
		},
	}

	m := move(state)
	t.Log(m.Move)
}

// TODO: More GameState test cases!
func TestNextBody(t *testing.T) {
	up := Coord{0, 1}
	myBody := []Coord{
		{2, 2}, {1, 2},
	}
	board := Board{
		Height: 5,
		Width:  5,
	}
	nBody := nextBody(up, myBody, board)
	assert.Len(t, nBody, 2)
	assert.EqualValues(t, []Coord{{2, 3}, {2, 2}}, nBody)
}
