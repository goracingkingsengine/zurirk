package engine

import (
	"strings"
	"testing"
)

// Tests that the zobrist key is a correct polyglot key.
// Testdata from http://hgm.nubati.net/book_format.html
func TestPolyglotKey(t *testing.T) {
	data := []struct {
		key uint64
		fen string
	}{
		// Starting position and a few moves.
		{0x463b96181691fc9c, "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"},
		{0x823c9b50fd114196, "rnbqkbnr/pppppppp/8/8/4P3/8/PPPP1PPP/RNBQKBNR b KQkq e3 0 1"},
		{0x0756b94461c50fb0, "rnbqkbnr/ppp1pppp/8/3p4/4P3/8/PPPP1PPP/RNBQKBNR w KQkq d6 0 2"},
		{0x662fafb965db29d4, "rnbqkbnr/ppp1pppp/8/3pP3/8/8/PPPP1PPP/RNBQKBNR b KQkq - 0 2"},
		{0x22a48b5a8e47ff78, "rnbqkbnr/ppp1p1pp/8/3pPp2/8/8/PPPP1PPP/RNBQKBNR w KQkq f6 0 3"},
		{0x652a607ca3f242c1, "rnbqkbnr/ppp1p1pp/8/3pPp2/8/8/PPPPKPPP/RNBQ1BNR b kq - 0 3"},
		{0x00fdd303c946bdd9, "rnbq1bnr/ppp1pkpp/8/3pPp2/8/8/PPPPKPPP/RNBQ1BNR w - - 0 4"},
		{0x3c8123ea7b067637, "rnbqkbnr/p1pppppp/8/8/PpP4P/8/1P1PPPP1/RNBQKBNR b KQkq c3 0 3"},
		{0x5c3f9b829b279560, "rnbqkbnr/p1pppppp/8/8/P6P/R1p5/1P1PPPP1/1NBQKBNR b Kkq - 0 4"},

		// Only the kings. White to move.
		{0x5e854d7a97eb14c6, "4k3/8/8/8/8/8/8/4K3 w - - 0 1"},
		// Only the kings. Black to move.
		{0xa6536bd038cc91cf, "4k3/8/8/8/8/8/8/4K3 b - - 0 1"},
		// Only kings and rooks.
		{0x7e4e32cd118c4ab3, "r3k2r/8/8/8/8/8/8/R3K2R b - - 0 1"},
		// Only kings and rooks. Castling possible.
		{0x60b8d416a01a547a, "r3k2r/8/8/8/8/8/8/R3K2R b q - 0 1"},
		{0x8f2b874ace05cb23, "r3k2r/8/8/8/8/8/8/R3K2R b Q - 0 1"},
		{0x91dd61917f93d5ea, "r3k2r/8/8/8/8/8/8/R3K2R b qQ - 0 1"},
		{0x05741f66c60de55a, "r3k2r/8/8/8/8/8/8/R3K2R b KQkq - 0 1"},
		// Enpassant.
		{0x82cb1da07293cfb3, "r3k2r/8/8/8/4P3/8/8/R3K2R b KQkq e3 0 1"},
	}

	for i, d := range data {
		pos, _ := PositionFromFEN(d.fen)
		if d.key != pos.Zobrist() {
			t.Errorf("#%d expected %08x got %08x for %s", i, d.key, pos.Zobrist(), d.fen)
		}
	}
}

func TestZobristUndo(t *testing.T) {
	for g, game := range testGames {
		moves := strings.Fields(game)
		pos, _ := PositionFromFEN(FENStartPos)

		var zob []uint64 // zobrist keys
		var tmp []Move   // moves executed

		for _, move := range moves {
			m, _ := pos.UCIToMove(move)
			zob = append(zob, pos.Zobrist())
			tmp = append(tmp, m)
			pos.DoMove(m)
		}

		for i := len(moves) - 1; i >= 0; i-- {
			pos.UndoMove()
			if zob[i] != pos.Zobrist() {
				t.Errorf("#%d expected zobrist key 0x%x got 0x%x", g, zob[i], pos.Zobrist())
			}
		}
	}
}
