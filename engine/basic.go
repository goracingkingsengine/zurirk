// basic.go defines basic types.
//
//go:generate stringer -type Figure
//go:generate stringer -type Color
//go:generate stringer -type Piece
//go:generate stringer -type MoveType

package engine

import (
	"fmt"
)

var (
	figureToSymbol = map[Figure]string{
		Knight: "N",
		Bishop: "B",
		Rook:   "R",
		Queen:  "Q",
		King:   "K",
	}

	// pov xor mask indexed by color.
	povMask = [ColorArraySize]Square{0x00, 0x38, 0x00}
)

// Square identifies the location on the board.
type Square uint8

const (
	SquareA1 = Square(iota)
	SquareB1
	SquareC1
	SquareD1
	SquareE1
	SquareF1
	SquareG1
	SquareH1
	SquareA2
	SquareB2
	SquareC2
	SquareD2
	SquareE2
	SquareF2
	SquareG2
	SquareH2
	SquareA3
	SquareB3
	SquareC3
	SquareD3
	SquareE3
	SquareF3
	SquareG3
	SquareH3
	SquareA4
	SquareB4
	SquareC4
	SquareD4
	SquareE4
	SquareF4
	SquareG4
	SquareH4
	SquareA5
	SquareB5
	SquareC5
	SquareD5
	SquareE5
	SquareF5
	SquareG5
	SquareH5
	SquareA6
	SquareB6
	SquareC6
	SquareD6
	SquareE6
	SquareF6
	SquareG6
	SquareH6
	SquareA7
	SquareB7
	SquareC7
	SquareD7
	SquareE7
	SquareF7
	SquareG7
	SquareH7
	SquareA8
	SquareB8
	SquareC8
	SquareD8
	SquareE8
	SquareF8
	SquareG8
	SquareH8

	SquareArraySize = int(iota)
	SquareMinValue  = SquareA1
	SquareMaxValue  = SquareH8
)

// RankFile returns a square with rank r and file f.
// r and f should be between 0 and 7.
func RankFile(r, f int) Square {
	return Square(r*8 + f)
}

// SquareFromString parses a square from a string.
// The string has standard chess format [a-h][1-8].
func SquareFromString(s string) (Square, error) {
	if len(s) != 2 {
		return SquareA1, fmt.Errorf("invalid square %s", s)
	}

	f, r := -1, -1
	if 'a' <= s[0] && s[0] <= 'h' {
		f = int(s[0] - 'a')
	}
	if 'A' <= s[0] && s[0] <= 'H' {
		f = int(s[0] - 'A')
	}
	if '1' <= s[1] && s[1] <= '8' {
		r = int(s[1] - '1')
	}
	if f == -1 || r == -1 {
		return SquareA1, fmt.Errorf("invalid square %s", s)
	}

	return RankFile(r, f), nil
}

// Bitboard returns a bitboard that has sq set.
func (sq Square) Bitboard() Bitboard {
	return 1 << uint(sq)
}

func (sq Square) Relative(dr, df int) Square {
	return sq + Square(dr*8+df)
}

// Rank returns a number from 0 to 7 representing the rank of the square.
func (sq Square) Rank() int {
	return int(sq / 8)
}

// File returns a number from 0 to 7 representing the file of the square.
func (sq Square) File() int {
	return int(sq % 8)
}

// POV returns the square from col's point of view.
// That is for Black the rank is flipped, file stays the same.
// Useful in evaluation based on king's or pawns' positions.
func (sq Square) POV(col Color) Square {
	return sq ^ povMask[col]
}

func (sq Square) String() string {
	return string([]byte{
		uint8(sq.File() + 'a'),
		uint8(sq.Rank() + '1'),
	})
}

// Figure represents a piece without a color.
type Figure uint

const (
	NoFigure Figure = iota
	Pawn
	Knight
	Bishop
	Rook
	Queen
	King

	FigureArraySize = int(iota)
	FigureMinValue  = Pawn
	FigureMaxValue  = King
)

// Color represents a side.
type Color uint

const (
	NoColor Color = iota
	Black
	White

	ColorArraySize = int(iota)
	ColorMinValue  = Black
	ColorMaxValue  = White
)

var (
	kingHomeRank = [ColorArraySize]int{0, 7, 0}
)

// Opposite returns the reversed color.
// Result is undefined if c is not White or Black.
func (c Color) Opposite() Color {
	return White + Black - c
}

// KingHomeRank return king's rank on starting position.
// Result is undefined if c is not White or Black.
func (c Color) KingHomeRank() int {
	return kingHomeRank[c]
}

// Piece is a figure owned by one side.
type Piece uint8

// Piece constants must stay in sync with ColorFigure
// The order of pieces must match Polyglot format:
// http://hgm.nubati.net/book_format.html
const (
	NoPiece Piece = iota
	_
	BlackPawn
	WhitePawn
	BlackKnight
	WhiteKnight
	BlackBishop
	WhiteBishop
	BlackRook
	WhiteRook
	BlackQueen
	WhiteQueen
	BlackKing
	WhiteKing

	PieceArraySize = int(iota)
	PieceMinValue  = BlackPawn
	PieceMaxValue  = WhiteKing
)

// ColorFigure returns a piece with col and fig.
func ColorFigure(col Color, fig Figure) Piece {
	return Piece(fig<<1) + Piece(col>>1)
}

// Color returns piece's color.
func (pi Piece) Color() Color {
	return Color(21844 >> pi & 3)
}

// Figure returns piece's figure.
func (pi Piece) Figure() Figure {
	return Figure(pi) >> 1
}

// Bitboard is a set representing the 8x8 chess board squares.
type Bitboard uint64

const (
	BbEmpty          Bitboard = 0x0000000000000000
	BbFull           Bitboard = 0xffffffffffffffff
	BbBorder         Bitboard = 0xff818181818181ff
	BbPawnStartRank  Bitboard = 0x00ff00000000ff00
	BbPawnDoubleRank Bitboard = 0x000000ffff000000
	BbBlackSquares   Bitboard = 0xaa55aa552a55aa55
	BbWhiteSquares   Bitboard = 0xd5aa55aad5aa55aa
)

const (
	BbFileA Bitboard = 0x101010101010101 << iota
	BbFileB
	BbFileC
	BbFileD
	BbFileE
	BbFileF
	BbFileG
	BbFileH
)

const (
	BbRank1 Bitboard = 0x0000000000000FF << (8 * iota)
	BbRank2
	BbRank3
	BbRank4
	BbRank5
	BbRank6
	BbRank7
	BbRank8
)

// RankBb returns a bitboard with all bits on rank set.
func RankBb(rank int) Bitboard {
	return BbRank1 << uint(8*rank)
}

// FileBb returns a bitboard with all bits on file set.
func FileBb(file int) Bitboard {
	return BbFileA << uint(file)
}

// AdjacentFilesBb returns a bitboard with all bits set on adjacent files.
func AdjacentFilesBb(file int) Bitboard {
	var bb Bitboard
	if file > 0 {
		bb |= FileBb(file - 1)
	}
	if file < 7 {
		bb |= FileBb(file + 1)
	}
	return bb
}

// North shifts all squares one rank up.
func North(bb Bitboard) Bitboard {
	return bb << 8
}

// South shifts all squares one rank down.
func South(bb Bitboard) Bitboard {
	return bb >> 8
}

// East shifts all squares one file right.
func East(bb Bitboard) Bitboard {
	return bb &^ BbFileH << 1
}

// West shifts all squares one file left.
func West(bb Bitboard) Bitboard {
	return bb &^ BbFileA >> 1
}

// Forward returns bb shifted one rank forward wrt color.
func Forward(col Color, bb Bitboard) Bitboard {
	if col == White {
		return bb << 8
	}
	if col == Black {
		return bb >> 8
	}
	return bb
}

// Backward returns bb shifted one rank backward wrt color.
func Backward(col Color, bb Bitboard) Bitboard {
	if col == White {
		return bb >> 8
	}
	if col == Black {
		return bb << 8
	}
	return bb
}

// Fill returns a bitboard with all files with squares filled.
func Fill(bb Bitboard) Bitboard {
	return NorthFill(bb) | SouthFill(bb)
}

// ForwardSpan computes forward span wrt to color.
func ForwardSpan(col Color, bb Bitboard) Bitboard {
	if col == White {
		return NorthSpan(bb)
	}
	if col == Black {
		return SouthSpan(bb)
	}
	return bb
}

// BackwardSpan computes backward span wrt to color.
func BackwardSpan(col Color, bb Bitboard) Bitboard {
	if col == White {
		return SouthSpan(bb)
	}
	if col == Black {
		return NorthSpan(bb)
	}
	return bb
}

// NorthFill returns a bitboard with all north bits set.
func NorthFill(bb Bitboard) Bitboard {
	bb |= (bb << 8)
	bb |= (bb << 16)
	bb |= (bb << 24)
	return bb
}

// NorthSpan is like NorthFill shifted on up.
func NorthSpan(bb Bitboard) Bitboard {
	return NorthFill(North(bb))
}

// SouthFill returns a bitboard with all south bits set.
func SouthFill(bb Bitboard) Bitboard {
	bb |= (bb >> 8)
	bb |= (bb >> 16)
	bb |= (bb >> 24)
	return bb
}

// SouthSpan is like SouthFill shifted on up.
func SouthSpan(bb Bitboard) Bitboard {
	return SouthFill(South(bb))
}

// Has returns bb if sq is occupied in bitboard.
func (bb Bitboard) Has(sq Square) bool {
	return bb>>sq&1 != 0
}

// AsSquare returns the occupied square if the bitboard has a single piece.
// If the board has more then one piece the result is undefined.
func (bb Bitboard) AsSquare() Square {
	// same as logN(bb)
	return Square(debrujin64[bb*debrujinMul>>debrujinShift])
}

// LSB picks a square in the board.
// Returns empty board for empty board.
func (bb Bitboard) LSB() Bitboard {
	return bb & (-bb)
}

// CountMax2 is equivalent to, but faster than, min(bb.Count(), 2).
func (bb Bitboard) CountMax2() int32 {
	if bb == 0 {
		return 0
	}
	if bb&(bb-1) == 0 {
		return 1
	}
	return 2
}

// Count counts number of squares set in bb.
func (bb Bitboard) Count() int32 {
	// same as popcnt.
	// Code adapted from https://chessprogramming.wikispaces.com/Population+Count.
	bb = bb - ((bb >> 1) & k1)
	bb = (bb & k2) + ((bb >> 2) & k2)
	bb = (bb + (bb >> 4)) & k4
	bb = (bb * kf) >> 56
	return int32(bb)
}

// Pop pops a set square from the bitboard.
func (bb *Bitboard) Pop() Square {
	sq := *bb & (-*bb)
	*bb -= sq
	// same as logN(sq)
	return Square(debrujin64[sq*debrujinMul>>debrujinShift])
}

// MoveType defines the move type.
type MoveType uint8

const (
	NoMove    MoveType = iota // no move or null move
	Normal                    // regular move
	Promotion                 // pawn is promoted. Move.Promotion() gives the new piece
	Castling                  // king castles
	Enpassant                 // pawn takes enpassant
)

const (
	// NullMove is a move that does nothing. Has value to 0.
	NullMove = Move(0)
)

// Move stores a position dependent move.
//
// Bit representation
//   00.00.00.ff - from
//   00.00.ff.00 - to
//   00.0f.00.00 - move type
//   00.f0.00.00 - target
//   0f.00.00.00 - capture
//   f0.00.00.00 - piece
type Move uint32

// MakeMove constructs a move.
func MakeMove(moveType MoveType, from, to Square, capture, target Piece) Move {
	piece := target
	if moveType == Promotion {
		piece = ColorFigure(target.Color(), Pawn)
	}

	return Move(from)<<0 +
		Move(to)<<8 +
		Move(moveType)<<16 +
		Move(target)<<20 +
		Move(capture)<<24 +
		Move(piece)<<28
}

// From returns the starting square.
func (m Move) From() Square {
	return Square(m >> 0 & 0xff)
}

// To returns the destination square.
func (m Move) To() Square {
	return Square(m >> 8 & 0xff)
}

// MoveType returns the move type.
func (m Move) MoveType() MoveType {
	return MoveType(m >> 16 & 0xf)
}

// SideToMove returns which player is moving.
func (m Move) SideToMove() Color {
	return m.Piece().Color()
}

// CaptureSquare returns the captured piece square.
// If no piece is captured, the result is the destination square.
func (m Move) CaptureSquare() Square {
	if m.MoveType() != Enpassant {
		return m.To()
	}
	return m.From()&0x38 + m.To()&0x7
}

// Capture returns the captured pieces.
func (m Move) Capture() Piece {
	return Piece(m >> 24 & 0xf)
}

// Target returns the piece on the to square after the move is executed.
func (m Move) Target() Piece {
	return Piece(m >> 20 & 0xf)
}

// Piece returns the piece moved.
func (m Move) Piece() Piece {
	return Piece(m >> 28 & 0xf)
}

// Promotion returns the promoted piece if any.
func (m Move) Promotion() Piece {
	if m.MoveType() != Promotion {
		return NoPiece
	}
	return m.Target()
}

// IsViolent returns true if the move can change the position's score significantly.
// TODO: IsViolent should be in sync with GenerateViolentMoves.
func (m Move) IsViolent() bool {
	return m.Capture() != NoPiece || m.MoveType() == Promotion
}

// IsQuiet returns true if the move is quiet.
// In particular Castling is not quiet and not violent.
func (m Move) IsQuiet() bool {
	return m.MoveType() == Normal && m.Capture() == NoPiece
}

// UCI converts a move to UCI format.
// The protocol specification at http://wbec-ridderkerk.nl/html/UCIProtocol.html
// incorrectly states that this is the long algebraic notation (LAN).
func (m Move) UCI() string {
	return m.From().String() + m.To().String() + figureToSymbol[m.Promotion().Figure()]
}

// LAN converts a move to Long Algebraic Notation.
// http://en.wikipedia.org/wiki/Algebraic_notation_%28chess%29#Long_algebraic_notation
// E.g. a2-a3, b7-b8Q, Nb1xc3
func (m Move) LAN() string {
	r := figureToSymbol[m.Piece().Figure()] + m.From().String()
	if m.Capture() != NoPiece {
		r += "x"
	} else {
		r += "-"
	}
	r += m.To().String() + figureToSymbol[m.Promotion().Figure()]
	return r
}

func (m Move) String() string {
	return m.LAN()
}

// Castle represents the castling rights mask.
type Castle uint8

const (
	WhiteOO  Castle = 1 << iota // WhiteOO indicates that White can castle on King side.
	WhiteOOO                    // WhiteOOO indicates that White can castle on Queen side.
	BlackOO                     // BlackOO indicates that Black can castle on King side.
	BlackOOO                    // BlackOOO indicates that Black can castle on Queen side.

	NoCastle  Castle = 0                                       // NoCastle indicates no castling rights.
	AnyCastle Castle = WhiteOO | WhiteOOO | BlackOO | BlackOOO // AnyCastle indicates all castling rights.

	CastleArraySize = int(AnyCastle + 1)
	CastleMinValue  = NoCastle
	CastleMaxValue  = AnyCastle
)

var castleToString = [...]string{
	"-", "K", "Q", "KQ", "k", "Kk", "Qk", "KQk", "q", "Kq", "Qq", "KQq", "kq", "Kkq", "Qkq", "KQkq",
}

func (c Castle) String() string {
	if c < NoCastle || c > AnyCastle {
		return fmt.Sprintf("Castle(%d)", c)
	}
	return castleToString[c]
}

// CastlingRook returns the rook moved during castling
// together with starting and stopping squares.
func CastlingRook(kingEnd Square) (Piece, Square, Square) {
	// Explanation how rookStart works for king on E1.
	// if kingEnd == C1 == b010, then rookStart == A1 == b000
	// if kingEnd == G1 == b110, then rookStart == H1 == b111
	// So bit 3 will set bit 2 and bit 1.
	//
	// Explanation how rookEnd works for king on E1.
	// if kingEnd == C1 == b010, then rookEnd == D1 == b011
	// if kingEnd == G1 == b110, then rookEnd == F1 == b101
	// So bit 3 will invert bit 2. bit 1 is always set.
	piece := Piece(Rook<<1) + (1 - Piece(kingEnd>>5))
	rookStart := kingEnd&^3 | (kingEnd & 4 >> 1) | (kingEnd & 4 >> 2)
	rookEnd := kingEnd ^ (kingEnd & 4 >> 1) | 1
	return piece, rookStart, rookEnd
}
