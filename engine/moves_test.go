package engine

import (
	"testing"
)

var (
	sanMoves = []struct {
		pos  string
		san  string
		move Move
	}{
		{"r3k2r/p1ppqpb1/bn2pnp1/3PN3/1p2P3/2N2Q1p/PPPBBPPP/R3K2R w KQkq - 0 1",
			"Qxf6", MakeMove(Normal, SquareF3, SquareF6, BlackKnight, WhiteQueen),
		},
		{"r3k2r/p1ppqpb1/bn2pQp1/3PN3/1p2P3/2N4p/PPPBBPPP/R3K2R b KQkq - 0 1",
			"hxg2", MakeMove(Normal, SquareH3, SquareG2, WhitePawn, BlackPawn),
		},
		{"r3k2r/p1ppqpb1/bn2pQp1/3PN3/1p2P3/2N5/PPPBBPpP/R3K2R w KQkq - 0 2",
			"a4", MakeMove(Normal, SquareA2, SquareA4, NoPiece, WhitePawn),
		},
		{"r3k2r/p1ppqpb1/bn2pQp1/3PN3/Pp2P3/2N5/1PPBBPpP/R3K2R b KQkq a3 0 2",
			"bxa3e.p.", MakeMove(Enpassant, SquareB4, SquareA3, WhitePawn, BlackPawn),
		},
		{"r3k2r/p1ppqpb1/bn2pQp1/3PN3/4P3/p1N5/1PPBBPpP/R3K2R w KQkq - 0 3",
			"Qf5", MakeMove(Normal, SquareF6, SquareF5, NoPiece, WhiteQueen),
		},
		{"r3k2r/p1ppqpb1/bn2p1p1/3PNQ2/4P3/p1N5/1PPBBPpP/R3K2R b KQkq - 1 3",
			"gxh1=Q", MakeMove(Promotion, SquareG2, SquareH1, WhiteRook, BlackQueen),
		},
		{"r3k2r/p1ppqpb1/bn2p1p1/3PNQ2/4P3/p1N5/1PPBBP1P/R3K2q w Qkq - 0 4",
			"Bf1", MakeMove(Normal, SquareE2, SquareF1, NoPiece, WhiteBishop),
		},
		{"r3k2r/p1ppqpb1/bn2p1p1/3PNQ2/4P3/p1N5/1PPB1P1P/R3KB1q b Qkq - 1 4",
			"exf5", MakeMove(Normal, SquareE6, SquareF5, WhiteQueen, BlackPawn),
		},
		{"2rqk2b/3bnp2/1p3n2/p2p2p1/1P1Pp3/P1N1P1N1/3B1PP1/2RQKB2 w - - 0 1",
			"Ba6", MakeMove(Normal, SquareF1, SquareA6, NoPiece, WhiteBishop),
		},
		{"2rqk2b/3bnp2/1p3n2/p2p2p1/1P1Pp3/P1N1P1N1/3B1PP1/2RQKB2 w - a6 0 1",
			"Ba6", MakeMove(Normal, SquareF1, SquareA6, NoPiece, WhiteBishop),
		},
		{"2r3k1/6pp/4pp2/3bp3/1Pq5/3R1P2/r1PQ2PP/1K1RN3 b - - 0 1",
			"Ra1", MakeMove(Normal, SquareA2, SquareA1, NoPiece, BlackRook),
		},
	}
)

func TestSANToMovePlay(t *testing.T) {
	for i, test := range sanMoves {
		pos, err := PositionFromFEN(test.pos)
		if err != nil {
			t.Fatalf("#%d invalid position: %s", i, err)
		}

		actual, err := pos.SANToMove(test.san)

		if err != nil {
			t.Fatalf("#%d %s parse error: %v", i, test.san, err)
		} else if test.move != actual {
			t.Fatalf("#%d %s expected %v (%s), got %v (%s)",
				i, test.san, test.move, &test.move, actual, &actual)
		}
	}
}
