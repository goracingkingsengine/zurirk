package engine

import (
	"testing"
)

func seeSlow(pos *Position, m Move, score int32) int32 {
	if m == NullMove || score > 0 {
		return score
	}

	// Compute the score change.
	score += seeScore(m)

	// Find the smallest attacker.
	var moves []Move
	pos.GenerateMoves(Violent, &moves)
	next := NullMove
	for _, n := range moves {
		if n.To() != m.To() {
			continue
		}

		// If the move is a promotion, consider the attacker to be a queen.
		fig, sq := n.Target().Figure(), n.From()
		if next == NullMove || fig < next.Target().Figure() || (fig == next.Piece().Figure() && sq < next.From()) {
			next = n
		}
	}

	// Recursively compute the see.
	pos.DoMove(next)
	see := -seeSlow(pos, next, -score)
	pos.UndoMove()

	if see > score {
		return score
	}
	return see
}

func TestSEE(t *testing.T) {
	good, bad := 0, 0
	for i, fen := range testFENs {
		var moves []Move
		pos, _ := PositionFromFEN(fen)
		pos.GenerateMoves(All, &moves)
		for _, m := range moves {
			pos.DoMove(m)
			actual := see(pos, m)
			expected := seeSlow(pos, m, 0)
			pos.UndoMove()

			if expected != actual {
				t.Errorf("#%d expected %d, got %d\nfor %v on %v", i, expected, actual, m, fen)
				bad++
			} else {
				good++
			}
		}
	}

	if bad != 0 {
		t.Errorf("Failed %d out of %d", bad, good+bad)
	}
}

// A benchmark position from http://www.stmintz.com/ccc/index.php?id=60880
var seeBench = "1rr3k1/4ppb1/2q1bnp1/1p2B1Q1/6P1/2p2P2/2P1B2R/2K4R w - - 0 1"

func BenchmarkSEESlow(b *testing.B) {
	var moves []Move
	pos, _ := PositionFromFEN(seeBench)
	pos.GenerateMoves(All, &moves)
	for i := 0; i < b.N; i++ {
		for _, m := range moves {
			seeSlow(pos, m, 0)
		}
	}
}

func BenchmarkSEEFast(b *testing.B) {
	var moves []Move
	pos, _ := PositionFromFEN(seeBench)
	pos.GenerateMoves(All, &moves)
	for i := 0; i < b.N; i++ {
		for _, m := range moves {
			see(pos, m)
		}
	}
}
