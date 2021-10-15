package main

import (
	"fmt"
	"testing"

	"github.com/Cameron-Kurotori/battlesnake/sdk"
	"github.com/stretchr/testify/assert"
)

func TestDeathFactor(t *testing.T) {
	you := sdk.Battlesnake{
		ID:     "me",
		Health: 100,
		Body:   []sdk.Coord{{X: 0, Y: 3}, {X: 0, Y: 4}, {X: 0, Y: 5}},
	}
	you.Head = you.Body[0]
	you.Length = int32(len(you.Body))
	state := sdk.GameState{
		Board: sdk.Board{
			Width:  5,
			Height: 5,
			Snakes: []sdk.Battlesnake{you},
		},
		You: you,
	}

	type testCase struct {
		expectedDeaths     int
		expectedTotalCount int
		dir                sdk.Direction
		state              sdk.GameState
	}

	test := func(tc testCase) func(*testing.T) {
		return func(t *testing.T) {
			deaths, count := Death(tc.dir, tc.state)
			assert.Equal(t, tc.expectedDeaths, deaths)
			assert.Equal(t, tc.expectedTotalCount, count)
		}
	}

	testCases := []testCase{
		{0, 1, sdk.Direction_Down, state},
		{1, 1, sdk.Direction_Up, state},
		{1, 1, sdk.Direction_Left, state},
		{0, 1, sdk.Direction_Right, state},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("only-me-%s", sdk.DirectionToMove[tc.dir]), test(tc))
	}

	otherSnake := sdk.Battlesnake{
		ID:     "other",
		Health: 100,
		Body:   []sdk.Coord{{X: 2, Y: 3}, {X: 3, Y: 3}, {X: 4, Y: 3}},
	}
	otherSnake.Head = otherSnake.Body[0]
	otherSnake.Length = int32(len(otherSnake.Body))

	state.Board.Snakes = append(state.Board.Snakes, otherSnake)

	testCases = []testCase{
		{0, 3, sdk.Direction_Down, state},
		{3, 3, sdk.Direction_Up, state},
		{3, 3, sdk.Direction_Left, state},
		{1, 3, sdk.Direction_Right, state},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("only-two-%s", sdk.DirectionToMove[tc.dir]), test(tc))
	}

	otherSnake = sdk.Battlesnake{
		ID:     "other",
		Health: 100,
		Body:   []sdk.Coord{{X: 2, Y: 3}, {X: 3, Y: 3}, {X: 3, Y: 4}, {X: 2, Y: 4}},
	}
	otherSnake.Head = otherSnake.Body[0]
	otherSnake.Length = int32(len(otherSnake.Body))

	state.Board.Snakes[1] = otherSnake

	testCases = []testCase{
		{0, 2, sdk.Direction_Down, state},
		{2, 2, sdk.Direction_Up, state},
		{2, 2, sdk.Direction_Left, state},
		{1, 2, sdk.Direction_Right, state},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("other-square-%s", sdk.DirectionToMove[tc.dir]), test(tc))
	}
}
