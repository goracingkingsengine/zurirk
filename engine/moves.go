// moves.go deals with move parsing.

package engine

import (
	"fmt"
)

var (
	errorWrongLength       = fmt.Errorf("SAN string is too short")
	errorUnknownFigure     = fmt.Errorf("unknown figure symbol")
	errorBadDisambiguation = fmt.Errorf("bad disambiguation")
	errorBadPromotion      = fmt.Errorf("only pawns on the last rank can be promoted")
	errorNoSuchMove        = fmt.Errorf("no such move")

	// Maps runes to figures.
	symbolToFigure = map[rune]Figure{
		'p': Pawn,
		'n': Knight,
		'b': Bishop,
		'r': Rook,
		'q': Queen,
		'k': King,

		'P': Pawn,
		'N': Knight,
		'B': Bishop,
		'R': Rook,
		'Q': Queen,
		'K': King,
	}
)

// SANToMove converts a move from SAN format to internal representation.
// SAN stands for standard algebraic notation and
// its description can be found in FIDE handbook.
//
// The set of strings accepted is a slightly different.
//   x (capture) presence or correctness is ignored.
//   + (check) and # (checkmate) is ignored.
//   e.p. (enpassant) is ignored
//
// TODO: verify that the returned move is legal.
func (pos *Position) SANToMove(s string) (Move, error) {
	moveType := Normal
	rank, file := -1, -1 // from
	to := SquareA1
	capture := NoPiece
	target := NoPiece

	// s[b:e] is the part that still needs to be parsed.
	b, e := 0, len(s)
	if b == e {
		return Move(0), errorWrongLength
	}
	// Skip + (check) and # (checkmate) at the end.
	for e > b && (s[e-1] == '#' || s[e-1] == '+') {
		e--
	}

	if s[b:e] == "o-o" || s[b:e] == "O-O" { // king side castling
		moveType = Castling
		if pos.SideToMove == White {
			rank, file = SquareE1.Rank(), SquareE1.File()
			to = SquareG1
			target = WhiteKing
		} else {
			rank, file = SquareE8.Rank(), SquareE8.File()
			to = SquareG8
			target = BlackKing
		}
	} else if s[b:e] == "o-o-o" || s[b:e] == "O-O-O" { // queen side castling
		moveType = Castling
		if pos.SideToMove == White {
			rank, file = SquareE1.Rank(), SquareE1.File()
			to = SquareC1
			target = WhiteKing
		} else {
			rank, file = SquareE8.Rank(), SquareE8.File()
			to = SquareC8
			target = BlackKing
		}
	} else { // all other moves
		// Get the piece.
		if ('a' <= s[b] && s[b] <= 'h') || s[b] == 'x' {
			target = ColorFigure(pos.SideToMove, Pawn)
		} else {
			if fig := symbolToFigure[rune(s[b])]; fig == NoFigure {
				return Move(0), errorUnknownFigure
			} else {
				target = ColorFigure(pos.SideToMove, fig)
			}
			b++
		}

		// Skip e.p. when enpassant.
		if e-4 > b && s[e-4:e] == "e.p." {
			e -= 4
		}

		// Check pawn promotion.
		if e-1 < b {
			return Move(0), errorWrongLength
		}
		if !('1' <= s[e-1] && s[e-1] <= '8') {
			// Not a rank, but a promotion.
			if target.Figure() != Pawn {
				return Move(0), errorBadPromotion
			}
			if fig := symbolToFigure[rune(s[e-1])]; fig == NoFigure {
				return Move(0), errorUnknownFigure
			} else {
				moveType = Promotion
				target = ColorFigure(pos.SideToMove, fig)
			}
			e--
			if e-1 >= b && s[e-1] == '=' {
				// Sometimes = is inserted before promotion figure.
				e--
			}
		}

		// Handle destination square.
		if e-2 < b {
			return Move(0), errorWrongLength
		}
		var err error
		to, err = SquareFromString(s[e-2 : e])
		if err != nil {
			return Move(0), err
		}
		if target.Figure() == Pawn && pos.IsEnpassantSquare(to) {
			moveType = Enpassant
			capture = ColorFigure(pos.SideToMove.Opposite(), Pawn)
		} else {
			capture = pos.Get(to)
		}
		e -= 2

		// Ignore 'x' (capture) or '-' (no capture) if present.
		if e-1 >= b && (s[e-1] == 'x' || s[e-1] == '-') {
			e--
		}

		// Parse disambiguation.
		if e-b > 2 {
			return Move(0), errorBadDisambiguation
		}
		for ; b < e; b++ {
			switch {
			case 'a' <= s[b] && s[b] <= 'h':
				file = int(s[b] - 'a')
			case '1' <= s[b] && s[b] <= '8':
				rank = int(s[b] - '1')
			default:
				return Move(0), errorBadDisambiguation
			}
		}
	}

	// Loop through all moves and find out one that matches.
	var moves []Move
	if moveType == Promotion {
		pos.GenerateFigureMoves(Pawn, All, &moves)
	} else {
		pos.GenerateFigureMoves(target.Figure(), All, &moves)
	}
	for _, pm := range moves {
		if pm.MoveType() != moveType || pm.Capture() != capture {
			continue
		}
		if pm.To() != to || pm.Target() != target {
			continue
		}
		if rank != -1 && pm.From().Rank() != rank {
			continue
		}
		if file != -1 && pm.From().File() != file {
			continue
		}
		return pm, nil
	}
	return Move(0), errorNoSuchMove
}

// UCIToMove parses a move given in UCI format.
// s can be "a2a4" or "h7h8Q" for pawn promotion.
func (pos *Position) UCIToMove(s string) (Move, error) {
	if len(s) < 4 {
		return NullMove, fmt.Errorf("%s is too short", s)
	}

	from, err := SquareFromString(s[0:2])
	if err != nil {
		return NullMove, err
	}
	to, err := SquareFromString(s[2:4])
	if err != nil {
		return NullMove, err
	}

	moveType := Normal
	capt := pos.Get(to)
	target := pos.Get(from)

	pi := pos.Get(from)
	if pi.Figure() == Pawn && pos.IsEnpassantSquare(to) {
		moveType = Enpassant
		capt = ColorFigure(pos.SideToMove.Opposite(), Pawn)
	}
	if pi == WhiteKing && from == SquareE1 && (to == SquareC1 || to == SquareG1) {
		moveType = Castling
	}
	if pi == BlackKing && from == SquareE8 && (to == SquareC8 || to == SquareG8) {
		moveType = Castling
	}
	if pi.Figure() == Pawn && (to.Rank() == 0 || to.Rank() == 7) {
		if len(s) != 5 {
			return NullMove, fmt.Errorf("%s doesn't have a promotion piece", s)
		}
		moveType = Promotion
		target = ColorFigure(pos.SideToMove, symbolToFigure[rune(s[4])])
	} else {
		if len(s) != 4 {
			return NullMove, fmt.Errorf("%s move is too long", s)
		}
	}

	move := MakeMove(moveType, from, to, capt, target)
	if !pos.IsPseudoLegal(move) {
		return NullMove, fmt.Errorf("%s is not a valid move", s)
	}
	return move, nil
}
