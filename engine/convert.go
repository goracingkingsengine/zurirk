// Methods here are exported because they are used also by the notation package.

package engine

import (
	"fmt"
	"strconv"
)

type castleInfo struct {
	Castle Castle
	Piece  [2]Piece
	Square [2]Square
}

var (
	itoa               = "0123456789" // shortcut for Itoa
	colorToSymbol      = "?bw"
	pieceToSymbol      = ".?pPnNbBrRqQkK"
	symbolToCastleInfo = map[rune]castleInfo{
		'K': castleInfo{
			Castle: WhiteOO,
			Piece:  [2]Piece{WhiteKing, WhiteRook},
			Square: [2]Square{SquareE1, SquareH1},
		},
		'k': castleInfo{
			Castle: BlackOO,
			Piece:  [2]Piece{BlackKing, BlackRook},
			Square: [2]Square{SquareE8, SquareH8},
		},
		'Q': castleInfo{
			Castle: WhiteOOO,
			Piece:  [2]Piece{WhiteKing, WhiteRook},
			Square: [2]Square{SquareE1, SquareA1},
		},
		'q': castleInfo{
			Castle: BlackOOO,
			Piece:  [2]Piece{BlackKing, BlackRook},
			Square: [2]Square{SquareE8, SquareA8},
		},
	}
	symbolToColor = map[string]Color{
		"w": White,
		"b": Black,
	}
	symbolToPiece = map[rune]Piece{
		'p': BlackPawn,
		'n': BlackKnight,
		'b': BlackBishop,
		'r': BlackRook,
		'q': BlackQueen,
		'k': BlackKing,

		'P': WhitePawn,
		'N': WhiteKnight,
		'B': WhiteBishop,
		'R': WhiteRook,
		'Q': WhiteQueen,
		'K': WhiteKing,
	}
)

// PositionFromFEN parses fen and returns the position.
//
// fen must contain the position using Forsyth–Edwards Notation
// http://en.wikipedia.org/wiki/Forsyth%E2%80%93Edwards_Notation
//
// Rejects FEN with only four fields,
// i.e. no full move counter or have move number.
func PositionFromFEN(fen string) (*Position, error) {
	// Split fen into 6 fields.
	// Same as string.Fields() but creates much less garbage.
	// The optimization is important when a huge number of positions
	// need to be evaluated.
	f, p := [6]string{}, 0
	for i := 0; i < len(fen); {
		// Find the start and end of the token.
		for ; i < len(fen) && fen[i] == ' '; i++ {
		}
		start := i
		for ; i < len(fen) && fen[i] != ' '; i++ {
		}
		limit := i

		if start == limit {
			continue
		}
		if p >= len(f) {
			return nil, fmt.Errorf("fen has too many fields")
		}
		f[p] = fen[start:limit]
		p++
	}
	if p < len(f) {
		return nil, fmt.Errorf("fen has too few fields")
	}

	// Parse each field.
	pos := NewPosition()
	if err := ParsePiecePlacement(f[0], pos); err != nil {
		return nil, err
	}
	if err := ParseSideToMove(f[1], pos); err != nil {
		return nil, err
	}
	if err := ParseCastlingAbility(f[2], pos); err != nil {
		return nil, err
	}
	if err := ParseEnpassantSquare(f[3], pos); err != nil {
		return nil, err
	}
	var err error
	if pos.curr.HalfmoveClock, err = strconv.Atoi(f[4]); err != nil {
		return nil, err
	}
	if pos.fullmoveCounter, err = strconv.Atoi(f[5]); err != nil {
		return nil, err
	}
	pos.Ply = (pos.fullmoveCounter - 1) * 2
	if pos.SideToMove == Black {
		pos.Ply++
	}
	return pos, nil
}

// ParsePiecePlacement parse pieces from str (FEN like) into pos.
func ParsePiecePlacement(str string, pos *Position) error {
	r, f := 0, 0
	for _, p := range str {
		if p == '/' {
			if r == 7 {
				return fmt.Errorf("expected 8 ranks")
			}
			if f != 8 {
				return fmt.Errorf("expected 8 squares per rank, got %d", f)
			}
			r, f = r+1, 0
			continue
		}

		if '1' <= p && p <= '8' {
			f += int(p) - int('0')
			continue
		}
		pi := symbolToPiece[p]
		if pi == NoPiece {
			return fmt.Errorf("expected rank or number, got %s", string(p))
		}
		if f >= 8 {
			return fmt.Errorf("rank %d too long (%d cells)", 8-r, f)
		}

		// 7-r because FEN describes the table from 8th rank.
		pos.Put(RankFile(7-r, f), pi)
		f++

	}

	if f < 8 {
		return fmt.Errorf("rank %d too short (%d cells)", r+1, f)
	}
	return nil
}

// FormatPiecePlacement converts a position to FEN piece placement.
func FormatPiecePlacement(pos *Position) string {
	s := ""
	for r := 7; r >= 0; r-- {
		space := 0
		for f := 0; f < 8; f++ {
			sq := RankFile(r, f)
			pi := pos.Get(sq)
			if pi == NoPiece {
				space++
			} else {
				if space != 0 {
					s += itoa[space:][:1]
					space = 0
				}
				s += pieceToSymbol[pi:][:1]
			}
		}

		if space != 0 {
			s += itoa[space:][:1]
		}
		if r != 0 {
			s += "/"
		}
	}
	return s
}

// ParseEnpassantSquare parses the en passant square from str.
func ParseEnpassantSquare(str string, pos *Position) error {
	if str[:1] == "-" {
		pos.SetEnpassantSquare(SquareA1)
		return nil
	}
	sq, err := SquareFromString(str)
	if err != nil {
		return err
	}
	pos.SetEnpassantSquare(sq)
	return nil
}

// FormatEnpassantSquare converts position's castling ability to string.
func FormatEnpassantSquare(pos *Position) string {
	if pos.EnpassantSquare() != SquareA1 {
		return pos.EnpassantSquare().String()
	}
	return "-"
}

// ParseSideToMove sets side to move for pos from str.
func ParseSideToMove(str string, pos *Position) error {
	if col, ok := symbolToColor[str]; ok {
		pos.SetSideToMove(col)
		return nil
	}
	return fmt.Errorf("invalid color %s", str)
}

// FormatSideToMove returns "w" for white to play or "b" for black to play.
func FormatSideToMove(pos *Position) string {
	return colorToSymbol[pos.SideToMove:][:1]
}

// ParseCastlingAbility sets castling ability for pos from str.
func ParseCastlingAbility(str string, pos *Position) error {
	if str == "-" {
		pos.SetCastlingAbility(NoCastle)
		return nil
	}

	ability := NoCastle
	for _, p := range str {
		info, ok := symbolToCastleInfo[p]
		if !ok {
			return fmt.Errorf("invalid castling ability %s", str)
		}
		ability |= info.Castle
		for i := 0; i < 2; i++ {
			if info.Piece[i] != pos.Get(info.Square[i]) {
				return fmt.Errorf("expected %v at %v, got %v",
					info.Piece[i], info.Square[i], pos.Get(info.Square[i]))
			}
		}
	}
	pos.SetCastlingAbility(ability)
	return nil
}

// FormatCastlingAbility returns a string specifying the castling ability
// using standard FEN format.
func FormatCastlingAbility(pos *Position) string {
	return pos.CastlingAbility().String()
}
