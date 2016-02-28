package engine

import (
	"fmt"
	"log"
	"strconv"
)

const (
	// Quiet indicates no capture, no castling, no promotion.
	Quiet int = 1 << iota
	// Tactical indicates castling and underpromotions (including captures).
	Tactical
	// Violent indicates captures and queen promotions.
	Violent
	// All indicates all moves.
	All = Quiet | Tactical | Violent
)

var (
	// FENStartPos is the FEN string of the starting position.
	FENStartPos = "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"

	// Which castle rights are lost when pieces are moved.
	lostCastleRights [64]Castle
)

func init() {
	lostCastleRights[SquareA1] = WhiteOOO
	lostCastleRights[SquareE1] = WhiteOOO | WhiteOO
	lostCastleRights[SquareH1] = WhiteOO
	lostCastleRights[SquareA8] = BlackOOO
	lostCastleRights[SquareE8] = BlackOOO | BlackOO
	lostCastleRights[SquareH8] = BlackOO
}

type state struct {
	Zobrist         uint64    // Zobrist key
	Move            Move      // last move played.
	HalfmoveClock   int       // last ply when a pawn was moved or a capture was made.
	EnpassantSquare [2]Square // en passant square (polyglot, fen). If none, then SquareA1.
	CastlingAbility Castle    // remaining castling rights.
}

// Position represents the chess board and keeps track of the move history.
type Position struct {
	ByFigure   [FigureArraySize]Bitboard // bitboards of square occupancy by figure.
	ByColor    [ColorArraySize]Bitboard  // bitboards of square occupancy by color.
	SideToMove Color                     // which side is to move. SideToMove is updated by DoMove and UndoMove.
	Ply        int                       // current ply

	fullmoveCounter int     // fullmove counter, incremented after black move
	states          []state // a state for each Ply
	curr            *state  // current state
}

///////////////////////////////////////////////////
// NEW
func (pos *Position) IsOnBaseRank(color Color) bool {
	kingPos := pos.ByFigure[King] & pos.ByColor[color]
	if (BbRank8 & kingPos) != 0 {
		return true
	}
	return false
}

func (pos *Position) PrintBoard() {
	/*fmt.Printf("  0123456701234567012345670123456701234567012345670123456701234567\n")
	for j:=0 ; j<FigureArraySize; j++ {
		fmt.Printf("%d %064b\n",j,pos.ByFigure[j])
	}*/
	var mask uint64=1 << uint(SquareArraySize-1)
	buff:=""
	for i:=0; i<SquareArraySize; i++ {
		found:=false
		var j=0
		for ; j<FigureArraySize; j++ {
			if uint64(pos.ByFigure[j]) & mask != 0 {
				found=true
				break
			}
		}
		if(!found){
			buff="☐"+buff
		} else {
			if uint64(pos.ByColor[Black]) & mask != 0 {
				buff=pieceToSymbolStr(Piece(2*j))+buff
			} else {
				buff=pieceToSymbolStr(Piece(2*j+1))+buff
			}
		}
		if (i%8) == 7 {
			fmt.Printf("%s\n",buff)
			buff=""
		}
		mask=mask >> 1
	}
	fmt.Printf("bw %v bb %v\n",pos.IsOnBaseRank(White),pos.IsOnBaseRank(Black))
}
///////////////////////////////////////////////////

// NewPosition returns a new position.
func NewPosition() *Position {
	pos := &Position{
		fullmoveCounter: 1,
		states:          make([]state, 1, 4),
	}
	pos.curr = &pos.states[pos.Ply]
	return pos
}

// String returns position in FEN format.
// For table format use PrettyPrint.
func (pos *Position) String() string {
	s := FormatPiecePlacement(pos)
	s += " " + FormatSideToMove(pos)
	s += " " + FormatCastlingAbility(pos)
	s += " " + FormatEnpassantSquare(pos)
	s += " " + strconv.Itoa(pos.curr.HalfmoveClock)
	s += " " + strconv.Itoa(pos.fullmoveCounter)
	return s
}

// prev returns state at previous ply.
func (pos *Position) prev() *state {
	return &pos.states[len(pos.states)-1]
}

///////////////////////////////////////////////////
// NEW
func (pos *Position) GetNoStates() int {
	return len(pos.states)
}
///////////////////////////////////////////////////

// popState pops one ply.
func (pos *Position) popState() {
	len := len(pos.states) - 1
	pos.states = pos.states[:len]
	pos.curr = &pos.states[len-1]
	pos.Ply--
}

// pushState adds one ply.
func (pos *Position) pushState() {
	len := len(pos.states)
	pos.states = append(pos.states, pos.states[len-1])
	pos.curr = &pos.states[len]
	pos.Ply++
}

func (pos *Position) FullmoveCounter() int {
	return pos.fullmoveCounter
}

func (pos *Position) SetFullmoveCounter(n int) {
	pos.fullmoveCounter = n
}

func (pos *Position) HalfmoveClock() int {
	return pos.curr.HalfmoveClock
}

func (pos *Position) SetHalfmoveClock(n int) {
	pos.curr.HalfmoveClock = n
}

// IsEnpassantSquare returns true if sq is the en passant square.
func (pos *Position) IsEnpassantSquare(sq Square) bool {
	return sq != SquareA1 && sq == pos.EnpassantSquare()
}

// EnpassantSquare returns the en passant square.
func (pos *Position) EnpassantSquare() Square {
	return pos.curr.EnpassantSquare[1]
}

// CastlingAbility returns kings' castling ability.
func (pos *Position) CastlingAbility() Castle {
	return pos.curr.CastlingAbility
}

// LastMove returns the last move played, if any.
func (pos *Position) LastMove() Move {
	return pos.curr.Move
}

// Zobrist returns the zobrist key of the position.
// The returned value is equal to polyglot book key
// (http://hgm.nubati.net/book_format.html).
func (pos *Position) Zobrist() uint64 {
	return pos.curr.Zobrist
}

// NumNonPawns returns the number of minor and major pieces.
func (pos *Position) NumNonPawns(col Color) int {
	return int((pos.ByColor[col] &^ pos.ByFigure[Pawn] &^ pos.ByFigure[King]).Count())
}

// HasNonPawns returns whether col has at least some minor or major pieces.
func (pos *Position) HasNonPawns(col Color) bool {
	return pos.ByColor[col]&^pos.ByFigure[Pawn]&^pos.ByFigure[King] != 0
}

// IsPseudoLegal returns true if m is a pseudo legal move for pos.
// It returns true iff m can be executed even if own king is in check
// after the move. NullMove is not a valid move.
func (pos *Position) IsPseudoLegal(m Move) bool {
	if m == NullMove ||
		m.SideToMove() != pos.SideToMove ||
		pos.Get(m.From()) != m.Piece() ||
		pos.Get(m.CaptureSquare()) != m.Capture() ||
		m.Piece().Color() == m.Capture().Color() {
		return false
	}

	if m.Piece().Figure() == Pawn {
		// Pawn move is tested above. Promotion is always correct.
		if m.MoveType() == Enpassant && !pos.IsEnpassantSquare(m.To()) {
			return false
		}
		if BbPawnStartRank.Has(m.From()) && BbPawnDoubleRank.Has(m.To()) && !pos.IsEmpty((m.From()+m.To())/2) {
			return false
		}
		return true
	}
	if m.Piece().Figure() == Knight {
		// Knight move is tested above. Knight jumps around.
		return true
	}

	// Quick test of queen's attack on an empty board.
	sq := m.From()
	to := m.To().Bitboard()
	if bbSuperAttack[sq]&to == 0 {
		return false
	}

	all := pos.ByColor[White] | pos.ByColor[Black]

	switch m.Piece().Figure() {
	case Pawn: // handled aove
		panic("unreachable")
	case Knight: // handled above
		panic("unreachable")
	case Bishop:
		return to&BishopMobility(sq, all) != 0
	case Rook:
		return to&RookMobility(sq, all) != 0
	case Queen:
		return to&QueenMobility(sq, all) != 0
	case King:
		if m.MoveType() == Normal {
			return to&bbKingAttack[sq] != 0
		}

		// m.MoveType() == Castling
		if m.SideToMove() == White && m.To() == SquareG1 {
			if pos.CastlingAbility()&WhiteOO == 0 ||
				!pos.IsEmpty(SquareF1) || !pos.IsEmpty(SquareG1) {
				return false
			}
		}
		if m.SideToMove() == White && m.To() == SquareC1 {
			if pos.CastlingAbility()&WhiteOOO == 0 ||
				!pos.IsEmpty(SquareB1) ||
				!pos.IsEmpty(SquareC1) ||
				!pos.IsEmpty(SquareD1) {
				return false
			}
		}
		if m.SideToMove() == Black && m.To() == SquareG8 {
			if pos.CastlingAbility()&BlackOO == 0 ||
				!pos.IsEmpty(SquareF8) ||
				!pos.IsEmpty(SquareG8) {
				return false
			}
		}
		if m.SideToMove() == Black && m.To() == SquareC8 {
			if pos.CastlingAbility()&BlackOOO == 0 ||
				!pos.IsEmpty(SquareB8) ||
				!pos.IsEmpty(SquareC8) ||
				!pos.IsEmpty(SquareD8) {
				return false
			}
		}
		rook, start, end := CastlingRook(m.To())
		if pos.Get(start) != rook {
			return false
		}
		them := m.SideToMove().Opposite()
		if pos.GetAttacker(m.From(), them) != NoFigure ||
			pos.GetAttacker(end, them) != NoFigure ||
			pos.GetAttacker(m.To(), them) != NoFigure {
			return false
		}
	default:
		panic("unreachable")
	}

	return true
}

// Verify check the validity of the position.
// Mostly used for debugging purposes.
func (pos *Position) Verify() error {
	if bb := pos.ByColor[White] & pos.ByColor[Black]; bb != 0 {
		sq := bb.Pop()
		return fmt.Errorf("Square %v is both White and Black", sq)
	}
	// Check that there is at most one king.
	// Catches castling issues.
	for col := ColorMinValue; col <= ColorMaxValue; col++ {
		bb := pos.ByPiece(col, King)
		sq := bb.Pop()
		if bb != 0 {
			sq2 := bb.Pop()
			return fmt.Errorf("More than one King for %v at %v and %v", col, sq, sq2)
		}
	}

	// Verifies that pieces have the right color.
	for col := ColorMinValue; col <= ColorMaxValue; col++ {
		for bb := pos.ByColor[col]; bb != 0; {
			sq := bb.Pop()
			pi := pos.Get(sq)
			if pi.Color() != col {
				return fmt.Errorf("Expected color %v, got %v", col, pi)
			}
		}
	}

	// Verifies that no two pieces sit on the same cell.
	for pi1 := PieceMinValue; pi1 <= PieceMaxValue; pi1++ {
		for pi2 := pi1 + 1; pi2 <= PieceMaxValue; pi2++ {
			if pos.ByPiece(pi1.Color(), pi1.Figure())&pos.ByPiece(pi2.Color(), pi2.Figure()) != 0 {
				return fmt.Errorf("%v and %v overlap", pi1, pi2)
			}
		}
	}

	// Verifies that en passant square is empty.
	if sq := pos.curr.EnpassantSquare[0]; sq != SquareA1 && !pos.IsEmpty(sq) {
		return fmt.Errorf("Expected empty en passant square %v, got %v", sq, pos.Get(sq))
	}

	return nil
}

// SetCastlingAbility sets the side to move, correctly updating the Zobrist key.
func (pos *Position) SetCastlingAbility(castle Castle) {
	if pos.curr.CastlingAbility == castle {
		return
	}

	pos.curr.Zobrist ^= zobristCastle[pos.curr.CastlingAbility]
	pos.curr.CastlingAbility = castle
	pos.curr.Zobrist ^= zobristCastle[pos.curr.CastlingAbility]
}

// SetSideToMove sets the side to move, correctly updating the Zobrist key.
func (pos *Position) SetSideToMove(col Color) {
	pos.curr.Zobrist ^= zobristColor[pos.SideToMove]
	pos.SideToMove = col
	pos.curr.Zobrist ^= zobristColor[pos.SideToMove]
}

// SetEnpassantSquare sets the en passant square correctly updating the Zobrist key.
func (pos *Position) SetEnpassantSquare(sq Square) {
	if sq == pos.curr.EnpassantSquare[1] {
		// In the trivial case both values are SquareA1
		// and zobrist value doesn't change.
		return
	}

	pos.curr.Zobrist ^= zobristEnpassant[pos.curr.EnpassantSquare[0]]
	pos.curr.EnpassantSquare[0] = sq
	pos.curr.EnpassantSquare[1] = sq

	if sq != SquareA1 {
		// In polyglot the hash key for en passant is updated only if
		// an en passant capture is possible next move. In other words
		// if there is an enemy pawn next to the end square of the move.
		var theirs Bitboard
		if sq.Rank() == 2 { // White
			theirs, sq = pos.ByPiece(Black, Pawn), RankFile(3, sq.File())
		} else if sq.Rank() == 5 { // Black
			theirs, sq = pos.ByPiece(White, Pawn), RankFile(4, sq.File())
		} else {
			panic("bad en passant square")
		}

		if (sq.File() == 0 || !theirs.Has(sq-1)) && (sq.File() == 7 || !theirs.Has(sq+1)) {
			pos.curr.EnpassantSquare[0] = SquareA1
		}
	}

	pos.curr.Zobrist ^= zobristEnpassant[pos.curr.EnpassantSquare[0]]
}

// ByPiece is a shortcut for ByColor[col]&ByFigure[fig].
func (pos *Position) ByPiece(col Color, fig Figure) Bitboard {
	return pos.ByColor[col] & pos.ByFigure[fig]
}

// Put puts a piece on the board.
// Does nothing if pi is NoPiece. Does not validate input.
func (pos *Position) Put(sq Square, pi Piece) {
	if pi != NoPiece {
		pos.curr.Zobrist ^= zobristPiece[pi][sq]
		bb := sq.Bitboard()
		pos.ByColor[pi.Color()] |= bb
		pos.ByFigure[pi.Figure()] |= bb
	}
}

// Remove removes a piece from the table.
// Does nothing if pi is NoPiece. Does not validate input.
func (pos *Position) Remove(sq Square, pi Piece) {
	if pi != NoPiece {
		pos.curr.Zobrist ^= zobristPiece[pi][sq]
		bb := ^sq.Bitboard()
		pos.ByColor[pi.Color()] &= bb
		pos.ByFigure[pi.Figure()] &= bb
	}
}

// IsEmpty returns true if there is no piece at sq.
func (pos *Position) IsEmpty(sq Square) bool {
	return !(pos.ByColor[White] | pos.ByColor[Black]).Has(sq)
}

// Get returns the piece at sq.
func (pos *Position) Get(sq Square) Piece {
	var col Color
	if pos.ByColor[White].Has(sq) {
		col = White
	} else if pos.ByColor[Black].Has(sq) {
		col = Black
	} else {
		return NoPiece
	}

	for fig := FigureMinValue; fig <= FigureMaxValue; fig++ {
		if pos.ByFigure[fig].Has(sq) {
			return ColorFigure(col, fig)
		}
	}
	panic("unreachable: square has color, but no figure")
}

// PawnThreats returns the set of squares threatened by side's pawns.
func (pos *Position) PawnThreats(side Color) Bitboard {
	pawns := Forward(side, pos.ByPiece(side, Pawn))
	return West(pawns) | East(pawns)
}

///////////////////////////////////////////////////
// OLD
// HasLegalMoves returns true if current side has any legal moves.
// This function is very expensive.
/*func (pos *Position) HasLegalMoves() bool {
	var moves []Move
	pos.GenerateMoves(All, &moves)
	us := pos.SideToMove

	for _, m := range moves {
		pos.DoMove(m)
		checked := pos.IsChecked(us)
		pos.UndoMove()

		if !checked {
			return true
		}
	}

	return false
}*/
///////////////////////////////////////////////////

///////////////////////////////////////////////////
// NEW
// HasLegalMoves returns true if current side has any legal moves.
// This function is very expensive.
// ! Now we use GetLegalMoves for this.
func (pos *Position) HasLegalMoves() bool {
	numMoves := len(pos.GetLegalMoves(GET_FIRST))
	return numMoves > 0
}

// generate all legal moves
func (pos *Position) GetLegalMoves(getfirst bool) []Move {
	var moves []Move
	var legalMoves=[]Move{}
	pos.GenerateMoves(All, &moves)
	us := pos.SideToMove
	them := us.Opposite()

	for _, m := range moves {
		pos.DoMove(m)
		checked := pos.IsChecked(us)
		// In Racing Kings any move that gives local check is also illegal.
		if Variant == VARIANT_Racing_Kings {
			checkedThem := pos.IsCheckedLocal(them)
			checked=checked||checkedThem
		}
		pos.UndoMove()

		if !checked {
			if getfirst {	
				return []Move{m}
			} else {
				legalMoves=append(legalMoves,m)
			}
		}
	}

	return legalMoves
}

func (pos *Position) PrintLegalMoves() {
	moves := pos.GetLegalMoves(GET_ALL)
	for i, move := range moves {
		fmt.Printf("%2d %6s ",i+1,move.LAN())
		if ((i%8)==7) && (i!=(len(moves)-1)) {
			fmt.Printf("\n")
		}
	}
	fmt.Printf("\n")
}
///////////////////////////////////////////////////

// InsufficientMaterial returns true if the position is theoretical draw.
func (pos *Position) InsufficientMaterial() bool {
	///////////////////////////////////////////////////
	// NEW
	if Variant == VARIANT_Racing_Kings {
		if pos.IsOnBaseRank(White) && pos.IsOnBaseRank(Black) {
			// Both kings on base rank is draw.
			return true
		}
		// No other insufficient material condition for Racking Kings.
		return false
	}
	///////////////////////////////////////////////////

	// K vs K is draw.
	noKings := (pos.ByColor[White] | pos.ByColor[Black]) &^ pos.ByFigure[King]
	if noKings == 0 {
		return true
	}
	// KN vs K is theoretical draw.
	if noKings == pos.ByFigure[Knight] && pos.ByFigure[Knight].CountMax2() == 1 {
		return true
	}
	// KB* vs KB* is theoretical draw if all bishops are on the same square color.
	if bishops := pos.ByFigure[Bishop]; noKings == bishops {
		if bishops&BbWhiteSquares == bishops ||
			bishops&BbBlackSquares == bishops {
			return true
		}
	}
	return false
}

// ThreeFoldRepetition returns whether current position was seen three times already.
// Returns minimum between 3 and the actual number of repetitions.
func (pos *Position) ThreeFoldRepetition() int {
	c, z := 0, pos.Zobrist()
	for i := 0; i < len(pos.states) && i <= pos.curr.HalfmoveClock; i += 2 {
		if pos.states[len(pos.states)-1-i].Zobrist == z {
			if c++; c == 3 {
				break
			}
		}
	}
	return c
}

// FiftyMoveRule returns True if 50 moves (on each side) were made
// without any capture of pawn move.
//
// If FiftyMoveRule returns true, the position is a draw.
func (pos *Position) FiftyMoveRule() bool {
	return pos.curr.HalfmoveClock >= 100
}

// Is the side in local check. Only makes a difference in Racing Kings, where reaching the base rank is global check.
func (pos *Position) IsCheckedLocal(side Color) bool {
	kingSq := pos.ByPiece(side, King).AsSquare()
	return pos.GetAttacker(kingSq, side.Opposite()) != NoFigure
}

// IsChecked returns true if side's king is checked.
func (pos *Position) IsChecked(side Color) bool {
	///////////////////////////////////////////////////
	// NEW
	// Check Racing Kings global checks.
	if Variant == VARIANT_Racing_Kings {
		onbb := pos.IsOnBaseRank(Black)
		onbw := pos.IsOnBaseRank(White)
		if onbb && onbw {
			// If both kings on base rank, there is no global check.
		} else if (side==White) && onbb {
			// If black reached the base rank white is always in check.
			return true
		} else if (side==Black) && onbw {
			// If white reached the base rank and black is not on base rank, black is in check.
			if !pos.IsOnBaseRank(Black) {
				return true
			}
		}
		// If no automatic check is true, then return the normal check.
	}
	///////////////////////////////////////////////////
	return pos.IsCheckedLocal(side)
}

// PrettyPrint pretty prints the current position to log.
func (pos *Position) PrettyPrint() {
	log.Println("zobrist =", pos.Zobrist())
	log.Println("fen =", pos.String())
	for r := 7; r >= 0; r-- {
		line := ""
		for f := 0; f < 8; f++ {
			sq := RankFile(r, f)
			if pos.IsEnpassantSquare(sq) {
				line += ","
			} else {
				line += string(pieceToSymbol[pos.Get(sq)])
			}
		}
		if r == 7 && pos.SideToMove == Black {
			line += " *"
		}
		if r == 0 && pos.SideToMove == White {
			line += " *"
		}
		log.Println(line)
	}

}

// DoMove executes a legal move.
func (pos *Position) DoMove(move Move) {
	pos.pushState()
	curr := pos.curr
	curr.Move = move

	// Update castling rights.
	pi := move.Piece()
	if pi != NoPiece { // nullmove cannot change castling ability
		pos.SetCastlingAbility(curr.CastlingAbility &^ lostCastleRights[move.From()] &^ lostCastleRights[move.To()])
	}
	// update fullmove counter.
	if pos.SideToMove == Black {
		pos.fullmoveCounter++
	}
	// Update halfmove clock.
	curr.HalfmoveClock++
	if pi.Figure() == Pawn || move.Capture() != NoPiece {
		curr.HalfmoveClock = 0
	}
	// Set Enpassant square for capturing.
	if pi.Figure() == Pawn && move.From().Rank()^move.To().Rank() == 2 {
		pos.SetEnpassantSquare((move.From() + move.To()) / 2)
	} else if pos.EnpassantSquare() != SquareA1 {
		pos.SetEnpassantSquare(SquareA1)
	}
	// Move rook on castling.
	if move.MoveType() == Castling {
		rook, start, end := CastlingRook(move.To())
		pos.Remove(start, rook)
		pos.Put(end, rook)
	}

	// Update the pieces on the chess board.
	pos.Remove(move.From(), pi)
	pos.Remove(move.CaptureSquare(), move.Capture())
	pos.Put(move.To(), move.Target())
	pos.SetSideToMove(pos.SideToMove.Opposite())
}

// UndoMove takes back the last move.
func (pos *Position) UndoMove() {
	move := pos.LastMove()
	pos.SetSideToMove(pos.SideToMove.Opposite())
	// CastlingAbility and EnpassantSquare are restored by pos.popState().
	// pos.SetCastlingAbility(pos.prev().CastlingAbility)
	// pos.SetEnpassantSquare(pos.prev().EnpassantSquare[1])

	// Modify the chess board.
	pi := move.Piece()
	pos.Put(move.From(), pi)
	pos.Remove(move.To(), move.Target())
	pos.Put(move.CaptureSquare(), move.Capture())

	// Move rook on castling.
	if move.MoveType() == Castling {
		rook, start, end := CastlingRook(move.To())
		pos.Put(start, rook)
		pos.Remove(end, rook)
	}

	if pos.SideToMove == Black {
		pos.fullmoveCounter--
	}

	pos.popState()
}

func (pos *Position) genPawnPromotions(kind int, moves *[]Move) {
	if kind&(Violent|Tactical) == 0 {
		return
	}

	// Minimum and maximum promotion pieces.
	// Tactical -> Knight - Rook
	// Violent -> Queen
	pMin, pMax := Queen, Rook
	if kind&Violent != 0 {
		pMax = Queen
	}
	if kind&Tactical != 0 {
		pMin = Knight
	}

	us := pos.SideToMove
	them := us.Opposite()

	// Get the pawns that can be promoted.
	all := pos.ByColor[White] | pos.ByColor[Black]
	ours := pos.ByPiece(us, Pawn)
	theirs := pos.ByColor[them] // their pieces

	forward := Square(0)
	if us == White {
		ours &= BbRank7
		forward = RankFile(+1, 0)
	} else {
		ours &= BbRank2
		forward = RankFile(-1, 0)
	}

	for ours != 0 {
		from := ours.Pop()
		to := from + forward

		if !all.Has(to) { // advance front
			for p := pMin; p <= pMax; p++ {
				*moves = append(*moves, MakeMove(Promotion, from, to, NoPiece, ColorFigure(us, p)))
			}
		}
		if to.File() != 0 && theirs.Has(to-1) { // take west
			capt := pos.Get(to - 1)
			for p := pMin; p <= pMax; p++ {
				*moves = append(*moves, MakeMove(Promotion, from, to-1, capt, ColorFigure(us, p)))
			}
		}
		if to.File() != 7 && theirs.Has(to+1) { // take east
			capt := pos.Get(to + 1)
			for p := pMin; p <= pMax; p++ {
				*moves = append(*moves, MakeMove(Promotion, from, to+1, capt, ColorFigure(us, p)))
			}
		}
	}
}

// genPawnAdvanceMoves moves pawns one square.
// Does not generate promotions.
func (pos *Position) genPawnAdvanceMoves(kind int, moves *[]Move) {
	if kind&Quiet == 0 {
		return
	}

	ours := pos.ByPiece(pos.SideToMove, Pawn)
	occu := pos.ByColor[White] | pos.ByColor[Black]
	pawn := ColorFigure(pos.SideToMove, Pawn)

	var forward Square
	if pos.SideToMove == White {
		ours = ours &^ South(occu) &^ BbRank7
		forward = RankFile(+1, 0)
	} else {
		ours = ours &^ North(occu) &^ BbRank2
		forward = RankFile(-1, 0)
	}

	for ours != 0 {
		from := ours.Pop()
		to := from + forward
		*moves = append(*moves, MakeMove(Normal, from, to, NoPiece, pawn))
	}
}

// genPawnDoubleAdvanceMoves moves pawns two square.
func (pos *Position) genPawnDoubleAdvanceMoves(kind int, moves *[]Move) {
	if kind&Quiet == 0 {
		return
	}

	ours := pos.ByPiece(pos.SideToMove, Pawn)
	occu := pos.ByColor[White] | pos.ByColor[Black]
	pawn := ColorFigure(pos.SideToMove, Pawn)

	var forward Square
	if pos.SideToMove == White {
		ours &= RankBb(1) &^ South(occu) &^ South(South(occu))
		forward = RankFile(+2, 0)
	} else {
		ours &= RankBb(6) &^ North(occu) &^ North(North(occu))
		forward = RankFile(-2, 0)
	}

	for ours != 0 {
		from := ours.Pop()
		to := from + forward
		*moves = append(*moves, MakeMove(Normal, from, to, NoPiece, pawn))
	}
}

func (pos *Position) pawnCapture(to Square) (MoveType, Piece) {
	if pos.IsEnpassantSquare(to) {
		return Enpassant, ColorFigure(pos.SideToMove.Opposite(), Pawn)
	}
	return Normal, pos.Get(to)
}

// Generate pawn attacks moves.
// Does not generate promotions.
func (pos *Position) genPawnAttackMoves(kind int, moves *[]Move) {
	if kind&Violent == 0 {
		return
	}

	theirs := pos.ByColor[pos.SideToMove.Opposite()]
	if pos.curr.EnpassantSquare[0] != SquareA1 {
		theirs |= pos.curr.EnpassantSquare[0].Bitboard()
	}

	forward := 0
	pawn := ColorFigure(pos.SideToMove, Pawn)
	ours := pos.ByPiece(pos.SideToMove, Pawn)
	if pos.SideToMove == White {
		ours = ours &^ BbRank7
		theirs = South(theirs)
		forward = +1
	} else {
		ours = ours &^ BbRank2
		theirs = North(theirs)
		forward = -1
	}

	// Left
	att := RankFile(forward, -1)
	for bbl := ours & East(theirs); bbl > 0; {
		from := bbl.Pop()
		to := from + att
		mt, capt := pos.pawnCapture(to)
		*moves = append(*moves, MakeMove(mt, from, to, capt, pawn))
	}

	// Right
	att = RankFile(forward, +1)
	for bbr := ours & West(theirs); bbr > 0; {
		from := bbr.Pop()
		to := from + att
		mt, capt := pos.pawnCapture(to)
		*moves = append(*moves, MakeMove(mt, from, to, capt, pawn))
	}
}

func (pos *Position) genBitboardMoves(pi Piece, from Square, att Bitboard, moves *[]Move) {
	for att != 0 {
		to := att.Pop()
		*moves = append(*moves, MakeMove(Normal, from, to, pos.Get(to), pi))
	}
}

func (pos *Position) getMask(kind int) Bitboard {
	mask := Bitboard(0)
	if kind&Violent != 0 {
		// Generate all attacks.
		// Promotions are handled specially.
		mask |= pos.ByColor[pos.SideToMove.Opposite()]
	}
	if kind&Quiet != 0 {
		// Generate all non-attacks.
		mask |= ^(pos.ByColor[White] | pos.ByColor[Black])
	}
	// Tactical is handled specially.
	return mask
}

func (pos *Position) genKnightMoves(mask Bitboard, moves *[]Move) {
	pi := ColorFigure(pos.SideToMove, Knight)
	for bb := pos.ByPiece(pos.SideToMove, Knight); bb != 0; {
		from := bb.Pop()
		att := bbKnightAttack[from] & mask
		pos.genBitboardMoves(pi, from, att, moves)
	}
}

func (pos *Position) genBishopMoves(fig Figure, mask Bitboard, moves *[]Move) {
	pi := ColorFigure(pos.SideToMove, fig)
	ref := pos.ByColor[White] | pos.ByColor[Black]
	for bb := pos.ByPiece(pos.SideToMove, fig); bb != 0; {
		from := bb.Pop()
		att := bishopMagic[from].Attack(ref) & mask
		pos.genBitboardMoves(pi, from, att, moves)
	}
}

func (pos *Position) genRookMoves(fig Figure, mask Bitboard, moves *[]Move) {
	pi := ColorFigure(pos.SideToMove, fig)
	ref := pos.ByColor[White] | pos.ByColor[Black]
	for bb := pos.ByPiece(pos.SideToMove, fig); bb != 0; {
		from := bb.Pop()
		att := rookMagic[from].Attack(ref) & mask
		pos.genBitboardMoves(pi, from, att, moves)
	}
}

func (pos *Position) genKingMovesNear(mask Bitboard, moves *[]Move) {
	pi := ColorFigure(pos.SideToMove, King)
	from := pos.ByPiece(pos.SideToMove, King).AsSquare()
	att := bbKingAttack[from] & mask
	pos.genBitboardMoves(pi, from, att, moves)
}

func (pos *Position) genKingCastles(kind int, moves *[]Move) {
	if kind&Tactical == 0 {
		return
	}

	rank := pos.SideToMove.KingHomeRank()
	oo, ooo := WhiteOO, WhiteOOO
	if pos.SideToMove == Black {
		oo, ooo = BlackOO, BlackOOO
	}

	// Castle king side.
	if pos.curr.CastlingAbility&oo != 0 {
		r5 := RankFile(rank, 5)
		r6 := RankFile(rank, 6)
		if !pos.IsEmpty(r5) || !pos.IsEmpty(r6) {
			goto EndCastleOO
		}

		r4 := RankFile(rank, 4)
		other := pos.SideToMove.Opposite()
		if pos.GetAttacker(r4, other) != NoFigure ||
			pos.GetAttacker(r5, other) != NoFigure ||
			pos.GetAttacker(r6, other) != NoFigure {
			goto EndCastleOO
		}

		*moves = append(*moves, MakeMove(Castling, r4, r6, NoPiece, ColorFigure(pos.SideToMove, King)))
	}
EndCastleOO:

	// Castle queen side.
	if pos.curr.CastlingAbility&ooo != 0 {
		r3 := RankFile(rank, 3)
		r2 := RankFile(rank, 2)
		r1 := RankFile(rank, 1)
		if !pos.IsEmpty(r3) || !pos.IsEmpty(r2) || !pos.IsEmpty(r1) {
			goto EndCastleOOO
		}

		r4 := RankFile(rank, 4)
		other := pos.SideToMove.Opposite()
		if pos.GetAttacker(r4, other) != NoFigure ||
			pos.GetAttacker(r3, other) != NoFigure ||
			pos.GetAttacker(r2, other) != NoFigure {
			goto EndCastleOOO
		}

		*moves = append(*moves, MakeMove(Castling, r4, r2, NoPiece, ColorFigure(pos.SideToMove, King)))
	}
EndCastleOOO:
}

// GetAttacker returns the smallest figure of color them that attacks sq.
func (pos *Position) GetAttacker(sq Square, them Color) Figure {
	enemy := pos.ByColor[them]
	// Pawn
	if enemy&bbPawnAttack[sq]&pos.ByFigure[Pawn] != 0 {
		if att := sq.Bitboard() & pos.PawnThreats(them); att != 0 {
			return Pawn
		}
	}
	// Knight
	if enemy&bbKnightAttack[sq]&pos.ByFigure[Knight] != 0 {
		return Knight
	}
	// Quick test of queen's attack on an empty board.
	// Exclude pawns and knights because they were already tested.
	enemy &^= pos.ByFigure[Pawn]
	enemy &^= pos.ByFigure[Knight]
	if enemy&bbSuperAttack[sq] == 0 {
		return NoFigure
	}
	// Bishop
	all := pos.ByColor[White] | pos.ByColor[Black]
	bishop := BishopMobility(sq, all)
	if enemy&pos.ByFigure[Bishop]&bishop != 0 {
		return Bishop
	}
	// Rook
	rook := RookMobility(sq, all)
	if enemy&pos.ByFigure[Rook]&rook != 0 {
		return Rook
	}
	// Queen
	if enemy&pos.ByFigure[Queen]&(bishop|rook) != 0 {
		return Queen
	}
	// King.
	if enemy&bbKingAttack[sq]&pos.ByFigure[King] != 0 {
		return King
	}
	return NoFigure
}

// GenerateMoves appends to moves all moves valid from pos.
// The generated moves are pseudo-legal, i.e. they can leave the king in check.
// kind is a combination of Quiet, Tactical or Violent.
func (pos *Position) GenerateMoves(kind int, moves *[]Move) {
	mask := pos.getMask(kind)
	// Order of the moves is important because the last quiet
	// moves will be reduced less.  Current order was produced
	// by testing 20 random orders and picking the best.
	pos.genKingMovesNear(mask, moves)
	pos.genPawnDoubleAdvanceMoves(kind, moves)
	pos.genRookMoves(Rook, mask, moves)
	pos.genBishopMoves(Queen, mask, moves)
	pos.genPawnAttackMoves(kind, moves)
	pos.genPawnAdvanceMoves(kind, moves)
	pos.genPawnPromotions(kind, moves)
	pos.genKnightMoves(mask, moves)
	pos.genBishopMoves(Bishop, mask, moves)
	pos.genKingCastles(kind, moves)
	pos.genRookMoves(Queen, mask, moves)
}

// GenerateFigureMoves generate moves for a given figure.
// The generated moves are pseudo-legal, i.e. they can leave the king in check.
// kind is a combination of Quiet, Tactical or Violent.
func (pos *Position) GenerateFigureMoves(fig Figure, kind int, moves *[]Move) {
	mask := pos.getMask(kind)
	switch fig {
	case Pawn:
		pos.genPawnAdvanceMoves(kind, moves)
		pos.genPawnAttackMoves(kind, moves)
		pos.genPawnDoubleAdvanceMoves(kind, moves)
		pos.genPawnPromotions(kind, moves)
	case Knight:
		pos.genKnightMoves(mask, moves)
	case Bishop:
		pos.genBishopMoves(Bishop, mask, moves)
	case Rook:
		pos.genRookMoves(Rook, mask, moves)
	case Queen:
		pos.genBishopMoves(Queen, mask, moves)
		pos.genRookMoves(Queen, mask, moves)
	case King:
		pos.genKingMovesNear(mask, moves)
		pos.genKingCastles(kind, moves)
	}
}
