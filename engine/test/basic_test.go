package engine

import (
	"testing"
)

func TestBitboardNSWE(t *testing.T) {
	data := []struct {
		f    func(bb Bitboard) Bitboard
		i, o Bitboard
	}{
		{North, RankBb(7), 0},
		{North, RankBb(6), RankBb(7)},
		{South, RankBb(1), RankBb(0)},
		{South, RankBb(0), 0},
		{East, FileBb(7), 0},
		{East, FileBb(6), FileBb(7)},
		{West, FileBb(1), FileBb(0)},
		{West, FileBb(0), 0},
		{West, (1 << 0) | (1 << 1), (1 << 0)},
		{East, (1 << 62) | (1 << 63), (1 << 63)},
		{NorthFill, RankBb(7), RankBb(7)},
		{NorthFill, RankBb(6), RankBb(6) | RankBb(7)},
		{NorthFill, 0x80000402002000, 0xa6a6262622202000},
		{SouthFill, RankBb(0), RankBb(0)},
		{SouthFill, RankBb(1), RankBb(0) | RankBb(1)},
		{SouthFill, 0x100218220080, 0x10121a3a3aba},
	}

	for i, d := range data {
		if d.o != d.f(d.i) {
			t.Errorf("#%d expected 0x%08x, got 0x%08x", i, d.o, d.f(d.i))
		}
	}
}

func TestBitboardFB(t *testing.T) {
	data := []struct {
		f    func(c Color, bb Bitboard) Bitboard
		c    Color
		i, o Bitboard
	}{
		{Forward, White, RankBb(4), RankBb(5)},
		{Forward, Black, RankBb(4), RankBb(3)},
		{Backward, White, RankBb(4), RankBb(3)},
		{Backward, Black, RankBb(4), RankBb(5)},
	}

	for i, d := range data {
		if d.o != d.f(d.c, d.i) {
			t.Errorf("#%d expected 0x%08x, got 0x%08x", i, d.o, d.f(d.c, d.i))
		}
	}
}

func TestSquareFromString(t *testing.T) {
	data := []struct {
		sq  Square
		str string
	}{
		{SquareF4, "f4"},
		{SquareA3, "a3"},
		{SquareC1, "c1"},
		{SquareH8, "h8"},
	}

	for _, d := range data {
		if d.sq.String() != d.str {
			t.Errorf("expected %v, got %v", d.str, d.sq.String())
		}
		if sq, err := SquareFromString(d.str); err != nil {
			t.Errorf("parse error: %v", err)
		} else if d.sq != sq {
			t.Errorf("expected %v, got %v", d.sq, sq)
		}
	}
}

func TestRookSquare(t *testing.T) {
	data := []struct {
		kingEnd   Square
		rook      Piece
		rookStart Square
		rookEnd   Square
	}{
		{SquareC1, WhiteRook, SquareA1, SquareD1},
		{SquareC8, BlackRook, SquareA8, SquareD8},
		{SquareG1, WhiteRook, SquareH1, SquareF1},
		{SquareG8, BlackRook, SquareH8, SquareF8},
	}

	for _, d := range data {
		rook, rookStart, rookEnd := CastlingRook(d.kingEnd)
		if rook != d.rook || rookStart != d.rookStart || rookEnd != d.rookEnd {
			t.Errorf("for king to %v, expected rook %v from %v to %v, got rook %v from %v to %v",
				d.kingEnd, d.rook, d.rookStart, d.rookEnd, rook, rookStart, rookEnd)
		}
	}
}

func TestRankFile(t *testing.T) {
	for r := 0; r < 7; r++ {
		for f := 0; f < 7; f++ {
			sq := RankFile(r, f)
			if sq.Rank() != r || sq.File() != f {
				t.Errorf("expected (rank, file) (%d, %d), got (%d, %d)",
					r, f, sq.Rank(), sq.File())
			}
		}
	}
}

func checkPiece(t *testing.T, pi Piece, co Color, fig Figure) {
	if pi.Color() != co || pi.Figure() != fig {
		t.Errorf("for %v expected %v %v, got %v %v", pi, co, fig, pi.Color(), pi.Figure())
	}
}

// TestPiece verifies Piece functionality.
func TestPiece1(t *testing.T) {
	checkPiece(t, NoPiece, NoColor, NoFigure)
	for co := ColorMinValue; co < ColorMaxValue; co++ {
		for fig := FigureMinValue; fig <= FigureMaxValue; fig++ {
			checkPiece(t, ColorFigure(co, fig), co, fig)
		}
	}
}

func TestPiece2(t *testing.T) {
	checkPiece(t, WhitePawn, White, Pawn)
	checkPiece(t, WhiteKnight, White, Knight)
	checkPiece(t, WhiteRook, White, Rook)
	checkPiece(t, WhiteKing, White, King)
	checkPiece(t, BlackPawn, Black, Pawn)
	checkPiece(t, BlackBishop, Black, Bishop)
}

func TestCastlingRook(t *testing.T) {
	data := []struct {
		kingEnd Square
		rook    Piece
	}{
		{SquareC1, WhiteRook},
		{SquareC8, BlackRook},
		{SquareG1, WhiteRook},
		{SquareG8, BlackRook},
	}

	for _, d := range data {
		rook, _, _ := CastlingRook(d.kingEnd)
		if rook != d.rook {
			t.Errorf("for king to %v, expected %v, got %v", d.kingEnd, d.rook, rook)
		}
	}
}

func TestKingHomeRank(t *testing.T) {
	data := []struct {
		col  Color
		rank int
	}{
		{NoColor, 0},
		{White, 0},
		{Black, 7},
	}

	for _, d := range data {
		if d.rank != d.col.KingHomeRank() {
			t.Errorf("for color %v, expected king home rank %d, got %d", d.col, d.rank, d.col.KingHomeRank())
		}
	}
}
