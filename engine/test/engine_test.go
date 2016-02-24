package engine

import (
	"strings"
	"testing"
)

var ()

func TestGame(t *testing.T) {
	pos, _ := PositionFromFEN(FENStartPos)
	eng := NewEngine(pos, nil, Options{})
	for i := 0; i < 1; i++ {
		tc := NewFixedDepthTimeControl(pos, 3)
		tc.Start(false)
		move := eng.Play(tc)
		eng.DoMove(move[0])
	}
}

func TestMateIn1(t *testing.T) {
	for i, d := range mateIn1 {
		pos, _ := PositionFromFEN(d.fen)
		bm, err := pos.SANToMove(d.bm)
		if err != nil {
			t.Errorf("#%d cannot parse move %s", i, d.bm)
			continue
		}

		tc := NewFixedDepthTimeControl(pos, 2)
		tc.Start(false)
		eng := NewEngine(pos, nil, Options{})
		pv := eng.Play(tc)

		if len(pv) != 1 {
			t.Errorf("#%d Expected at most one move, got %d", i, len(pv))
			t.Errorf("position is %v", pos)
			continue
		}

		if pv[0] != bm {
			t.Errorf("#%d expected move %v, got %v", i, bm, pv[0])
			t.Errorf("position is %v", pos)
			continue
		}
	}
}

// Test score is the same if we start with the position or move.
func TestScore(t *testing.T) {
	for _, game := range testGames {
		pos, _ := PositionFromFEN(FENStartPos)
		dynamic := NewEngine(pos, nil, Options{})
		static := NewEngine(pos, nil, Options{})

		moves := strings.Fields(game)
		for _, move := range moves {
			m, _ := pos.UCIToMove(move)
			if !pos.IsPseudoLegal(m) {
				// t.Fatalf("bad bad bad")
			}

			dynamic.DoMove(m)
			static.SetPosition(pos)
			if dynamic.Score() != static.Score() {
				t.Fatalf("expected static score %v, got dynamic score %v", static.Score(), dynamic.Score())
			}
		}
	}
}

func TestEndGamePosition(t *testing.T) {
	pos, _ := PositionFromFEN("6k1/5p1p/4p1p1/3p4/5P1P/8/3r2q1/6K1 w - - 2 55")
	tc := NewFixedDepthTimeControl(pos, 3)
	tc.Start(false)
	eng := NewEngine(pos, nil, Options{})
	moves := eng.Play(tc)
	if 0 != len(moves) {
		t.Errorf("expected no pv, got %d moves", len(moves))
	}
}

func passedPawns(pos *Position) Bitboard {
	wp := pos.ByPiece(White, Pawn)
	bp := pos.ByPiece(Black, Pawn)
	wpp := wp &^ SouthSpan(wp)
	bpp := bp &^ NorthSpan(bp)

	wp |= East(wp) | West(wp)
	bp |= East(bp) | West(bp)
	wpp = wpp &^ SouthSpan(bp)
	bpp = bpp &^ NorthSpan(wp)

	return wpp | bpp
}

func TestPassed(t *testing.T) {
	for _, fen := range testFENs {
		pos, _ := PositionFromFEN(fen)
		var moves []Move
		pos.GenerateMoves(All, &moves)
		before := passedPawns(pos)

		for _, m := range moves {
			pos.DoMove(m)
			after := passedPawns(pos)
			if passed(pos, m) && before == after {
				t.Errorf("expected no passed pawn, got passed pawn: move = %v, position = %v", m, pos)
			}

			pos.UndoMove()
			if passed(pos, m) && before == after {
				t.Errorf("expected no passed pawn, got passed pawn: move = %v, position = %v", m, pos)
			}
		}
	}
}

func BenchmarkStallingFENs(b *testing.B) {
	fens := []string{
		// Causes quiscence search to explode.
		"rnb1kbnr/pppp1ppp/8/8/3PPp1q/6P1/PPP4P/RNBQKBNR b KQkq -1 0 4",
		"r2qr1k1/2pn1ppp/pp2pn2/3b4/3P4/B2BPN2/P1P1QPPP/R4RK1 w - -1 4 13",
		"r1bq2k1/ppp4p/2n5/2bpPr2/5pQ1/2P5/PP4PP/RNB1NR1K b - -1 4 15",
	}

	for i := 0; i < b.N; i++ {
		for _, fen := range fens {
			pos, _ := PositionFromFEN(fen)
			eng := NewEngine(pos, nil, Options{})
			tc := NewFixedDepthTimeControl(pos, 5)
			tc.Start(false)
			eng.Play(tc)
		}
	}
}

func BenchmarkGame(b *testing.B) {
	for i := 0; i < b.N; i++ {
		pos, _ := PositionFromFEN(FENStartPos)
		eng := NewEngine(pos, nil, Options{})
		for j := 0; j < 20; j++ {
			tc := NewFixedDepthTimeControl(pos, 4)
			tc.Start(false)
			move := eng.Play(tc)
			eng.DoMove(move[0])
		}
	}
}

func BenchmarkScore(b *testing.B) {
	pos, _ := PositionFromFEN(FENStartPos)
	eng := NewEngine(pos, nil, Options{})

	for i := 0; i < b.N; i++ {
		for _, g := range testGames {
			var done []Move
			todo := strings.Fields(g)

			for j := range todo {
				move, _ := eng.Position.UCIToMove(todo[j])
				done = append(done, move)
				eng.DoMove(move)
				_ = eng.Score()
			}

			for range done {
				eng.UndoMove()
				_ = eng.Score()
			}
		}
	}
}
