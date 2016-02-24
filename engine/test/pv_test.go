package engine

import (
	"strings"
	"testing"
)

func TestPV(t *testing.T) {
	pos, _ := PositionFromFEN(FENStartPos)
	pvTable := newPvTable()
	for _, game := range testGames {
		var moves []Move
		movesStr := strings.Fields(game)
		for _, moveStr := range movesStr {
			move, _ := pos.UCIToMove(moveStr)
			pos.DoMove(move)
			moves = append(moves, move)
		}

		for i := len(moves) - 1; i >= 0; i-- {
			pos.UndoMove()
			pvTable.Put(pos, moves[i])
		}

		pv := pvTable.Get(pos)
		if len(pv) == 0 {
			t.Errorf("expected at least on move on principal variation")
		}
		if len(pv) > len(moves) {
			// This can actually happen during the game.
			t.Errorf("got more moves on pv than in the game")
		}
		for i := range pv {
			if moves[i] != pv[i] {
				t.Errorf("#%d Expected move %v, got %v", i, pv[i], moves[i])
			}
		}
	}
}
