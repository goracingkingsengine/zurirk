// Package engine implements board, move generation and position searching.
//
// The package can be used as a general library for chess tool writing and
// provides the core functionality for the zurichess chess engine.
//
// Position (basic.go, position.go) uses:
//
//   * Bitboards for representation - https://chessprogramming.wikispaces.com/Bitboards
//   * Magic bitboards for sliding move generation - https://chessprogramming.wikispaces.com/Magic+Bitboards
//
// Search (engine.go) features implemented are:
//
//   * Aspiration window - https://chessprogramming.wikispaces.com/Aspiration+Windows
//   * Check extension - https://chessprogramming.wikispaces.com/Check+Extensions
//   * Fail soft - https://chessprogramming.wikispaces.com/Fail-Soft
//   * Futility Pruning - https://chessprogramming.wikispaces.com/Futility+pruning
//   * History leaf pruning - https://chessprogramming.wikispaces.com/History+Leaf+Pruning
//   * Killer move heuristic - https://chessprogramming.wikispaces.com/Killer+Heuristic
//   * Late move redution (LMR) - https://chessprogramming.wikispaces.com/Late+Move+Reductions
//   * Mate distance pruning - https://chessprogramming.wikispaces.com/Mate+Distance+Pruning
//   * Negamax framework - http://chessprogramming.wikispaces.com/Alpha-Beta#Implementation-Negamax%20Framework
//   * Null move prunning (NMP) - https://chessprogramming.wikispaces.com/Null+Move+Pruning
//   * Principal variation search (PVS) - https://chessprogramming.wikispaces.com/Principal+Variation+Search
//   * Quiescence search - https://chessprogramming.wikispaces.com/Quiescence+Search.
//   * Static Single Evaluation - https://chessprogramming.wikispaces.com/Static+Exchange+Evaluation
//   * Zobrist hashing - https://chessprogramming.wikispaces.com/Zobrist+Hashing
//
// Move ordering (move_ordering.go) consists of:
//
//   * Hash move heuristic
//   * Captures sorted by MVVLVA - https://chessprogramming.wikispaces.com/MVV-LVA
//   * Killer moves - https://chessprogramming.wikispaces.com/Killer+Move
//
// Evaluation (material.go) function is quite basic and consists of:
//
//   * Material and mobility
//   * Piece square tables for pawns and king. Other figures did not improve the eval.
//   * King shelter (only in mid game)
//   * Pawn structure: connected, isolated, double, passed. Evaluation is cached (see pawn_table.go).
//   * Phased eval between mid game and end game.
//
package engine

import(
	"fmt"
)

///////////////////////////////////////////////////
// NEW
const (
	VARIANT_Standard         = iota
	VARIANT_Racing_Kings
	)

var FigureToName = [...]string{".","Pawn","Knight","Bishop","Rook","Queen","King"}

const VARIANT_CURRENT        = -1
///////////////////////////////////////////////////

const (
	CheckDepthExtension    int32 = 1 // how much to extend search in case of checks
	NullMoveDepthLimit     int32 = 1 // disable null-move below this limit
	NullMoveDepthReduction int32 = 1 // default null-move depth reduction. Can reduce more in some situations.
	PVSDepthLimit          int32 = 0 // do not do PVS below and including this limit
	LMRDepthLimit          int32 = 3 // do not do LMR below and including this limit
	FutilityDepthLimit     int32 = 3 // maximum depth to do futility pruning.

	initialAspirationWindow = 21  // ~a quarter of a pawn
	futilityMargin          = 150 // ~one and a halfpawn
	checkpointStep          = 10000
)

var (
	// scoreMultiplier is used to compute the score from side
	// to move POV from given the score from white POV.
	scoreMultiplier = [ColorArraySize]int32{0, -1, 1}
)

// Options keeps engine's options.
type Options struct {
	AnalyseMode bool // true to display info strings
}

// Stats stores some basic stats of the search.
//
// Statistics are reset every iteration of the iterative deepening search.
type Stats struct {
	CacheHit  uint64 // number of times the position was found transposition table
	CacheMiss uint64 // number of times the position was not found in the transposition table
	Nodes     uint64 // number of nodes searched
	Depth     int32  // depth search
	SelDepth  int32  // maximum depth reached on PV (doesn't include the hash moves)
}

// CacheHitRatio returns the ration of hits over total number of lookups.
func (s *Stats) CacheHitRatio() float32 {
	return float32(s.CacheHit) / float32(s.CacheHit+s.CacheMiss)
}

// Logger logs search progress.
type Logger interface {
	// BeginSearch signals a new search is started.
	BeginSearch()
	// EndSearch signals end of search.
	EndSearch()
	// PrintPV logs the principal variation after
	// iterative deepening completed one depth.
	PrintPV(stats Stats, score int32, pv []Move)
}

// NulLogger is a logger that does nothing.
type NulLogger struct {
}

func (nl *NulLogger) BeginSearch() {
}

func (nl *NulLogger) EndSearch() {
}

func (nl *NulLogger) PrintPV(stats Stats, score int32, pv []Move) {
}

// historyEntry keeps counts of how well move performed in the past.
type historyEntry struct {
	counter [2]int
	move    Move
}

// historyTable is a hash table that contains history of moves.
//
// old moves are automatically evicted when new moves are inserted
// so this cache is approx. LRU.
type historyTable []historyEntry

func newHistoryTable() historyTable {
	return make([]historyEntry, 1024)
}

// historyHash hashes the move and returns an index into the history table.
func historyHash(m Move) uint32 {
	// This is a murmur inspired hash so upper bits are better
	// mixed than the lower bits. The hash multiplier was chosen
	// to minimize the number of misses.
	h := uint32(m) * 438650727
	return (h + (h << 17)) >> 22
}

// get returns counters for m, i.e. pair of (bad, good)
//
// TODO: consider returning only if the move is good or bad.
func (ht historyTable) get(m Move) (int, int) {
	h := historyHash(m)
	if ht[h].move != m {
		return 0, 0
	}
	return ht[h].counter[0], ht[h].counter[1]
}

// inc increments the counters for m.
//
// Evicts an old move if necessary.
// Counters start from 1 so probability is correctly estimated. TODO: insert reference.
func (ht historyTable) inc(m Move, good bool) {
	h := historyHash(m)
	if ht[h].move != m {
		ht[h] = historyEntry{
			counter: [2]int{1, 1},
			move:    m,
		}
	}
	if good {
		ht[h].counter[1]++
	} else {
		ht[h].counter[0]++
	}
}

// Engine implements the logic to search the best move for a position.
type Engine struct {
	Options  Options   // engine options
	Log      Logger    // logger
	Stats    Stats     // search statistics
	Position *Position // current Position

	rootPly int          // position's ply at the start of the search
	stack   stack        // stack of moves
	pvTable pvTable      // principal variation table
	history historyTable // keeps history of moves

	timeControl *TimeControl
	stopped     bool
	checkpoint  uint64
}

// NewEngine creates a new engine to search for pos.
// If pos is nil then the start position is used.
func NewEngine(pos *Position, log Logger, options Options) *Engine {
	if log == nil {
		log = &NulLogger{}
	}
	eng := &Engine{
		Options: options,
		Log:     log,
		pvTable: newPvTable(),
		history: newHistoryTable(),
	}
	eng.SetPosition(pos)
	return eng
}

///////////////////////////////////////////////////
// NEW
func FigureNameToFigure(figureString string) Figure {
	switch figureString {
		case "Pawn" : return Pawn
		case "Knight" : return Knight
		case "Bishop" : return Bishop
		case "Rook" : return Rook
		case "Queen" : return Queen
		case "King" : return King
	}
	return NoFigure
}

func PrintPieceValues() {
	if Variant == VARIANT_Racing_Kings {
		for i:=Knight; i<King ; i++ {
			fmt.Printf("%s %d\n",FigureToName[i],RK_PIECE_VALUES[i])
		}
		fmt.Printf("King Advance %d\n",KING_ADVANCE_VALUE)
	}
}

func IsTest() bool {
	return TEST
}

func SetUseUnicodeSymbols(useUnicode bool) {
	USE_UNICODE_SYMBOLS=useUnicode
}

func (eng *Engine) SetVariant(setVariant int) {
	if(setVariant<0) {
		setVariant=Variant
	}
	Variant=setVariant
	pos, _ := PositionFromFEN(START_FENS[Variant])
	eng.SetPosition(pos)
}

func (eng *Engine) PrintBoard() {
	if eng.Position != nil {
		eng.Position.PrintBoard()
	} else {
		fmt.Printf("position missing\n")
	}
}
///////////////////////////////////////////////////

// SetPosition sets current position.
// If pos is nil, the starting position is set.
func (eng *Engine) SetPosition(pos *Position) {
	if pos != nil {
		eng.Position = pos
	} else {
		eng.Position, _ = PositionFromFEN(FENStartPos)
	}
}

// DoMove executes a move.
func (eng *Engine) DoMove(move Move) {
	eng.Position.DoMove(move)
}

// UndoMove undoes the last move.
func (eng *Engine) UndoMove() {
	eng.Position.UndoMove()
}

// Score evaluates current position from current player's POV.
func (eng *Engine) Score() int32 {
	score := Evaluate(eng.Position)
	score = ScaleToCentiPawn(score)
	return scoreMultiplier[eng.Position.SideToMove] * score
}

// endPosition determines whether the current position is an end game.
// Returns score and a bool if the game has ended.
func (eng *Engine) endPosition() (int32, bool) {
	pos := eng.Position // shortcut
	// Trivial cases when kings are missing.
	if pos.ByPiece(White, King) == 0 && pos.ByPiece(Black, King) == 0 {
		return 0, true
	}
	if pos.ByPiece(White, King) == 0 {
		return scoreMultiplier[pos.SideToMove] * (MatedScore + eng.ply()), true
	}
	if pos.ByPiece(Black, King) == 0 {
		return scoreMultiplier[pos.SideToMove] * (MateScore - eng.ply()), true
	}
	// Neither side cannot mate.
	if pos.InsufficientMaterial() {
		return 0, true
	}
	// Fifty full moves without a capture or a pawn move.
	if pos.FiftyMoveRule() {
		return 0, true
	}
	// Repetition is a draw.
	// At root we need to continue searching even if we saw two repetitions already,
	// however we can prune deeper search only at two repetitions.
	if r := pos.ThreeFoldRepetition(); eng.ply() > 0 && r >= 2 || r >= 3 {
		return 0, true
	}
	return 0, false
}

// retrieveHash gets from GlobalHashTable the current position.
func (eng *Engine) retrieveHash() hashEntry {
	entry := GlobalHashTable.get(eng.Position)

	if entry.kind == noEntry {
		eng.Stats.CacheMiss++
		return hashEntry{}
	}
	if entry.move != NullMove && !eng.Position.IsPseudoLegal(entry.move) {
		eng.Stats.CacheMiss++
		return hashEntry{}
	}

	// Return mate score relative to root.
	// The score was adjusted relative to position before the
	// hash table was updated.
	if entry.score < KnownLossScore {
		if entry.kind == exact {
			entry.score += eng.ply()
		}
	} else if entry.score > KnownWinScore {
		if entry.kind == exact {
			entry.score -= eng.ply()
		}
	}

	eng.Stats.CacheHit++
	return entry
}

// updateHash updates GlobalHashTable with the current position.
func (eng *Engine) updateHash(α, β, depth, score int32, move Move) {
	kind := exact
	if score <= α {
		kind = failedLow
	} else if score >= β {
		kind = failedHigh
	}

	// Save the mate score relative to the current position.
	// When retrieving from hash the score will be adjusted relative to root.
	if score < KnownLossScore {
		if kind == exact {
			score -= eng.ply()
		} else if kind == failedLow {
			score = KnownLossScore
		} else {
			return
		}
	} else if score > KnownWinScore {
		if kind == exact {
			score += eng.ply()
		} else if kind == failedHigh {
			score = KnownWinScore
		} else {
			return
		}
	}

	GlobalHashTable.put(eng.Position, hashEntry{
		kind:  kind,
		score: score,
		depth: int8(depth),
		move:  move,
	})
}

// searchQuiescence evaluates the position after solving all captures.
//
// This is a very limited search which considers only violent moves.
// Checks are not considered. In fact it assumes that the move
// ordering will always put the king capture first.
func (eng *Engine) searchQuiescence(α, β int32) int32 {
	eng.Stats.Nodes++
	if score, done := eng.endPosition(); done {
		return score
	}

	// Stand pat.
	// TODO: Some suggest to not stand pat when in check.
	// However, I did several tests and handling checks in quiescence
	// doesn't help at all.
	static := eng.Score()
	if static >= β {
		return static
	}
	localα := α
	if static > localα {
		localα = static
	}

	pos := eng.Position
	us := pos.SideToMove
	inCheck := pos.IsChecked(us)

	var bestMove Move
	eng.stack.GenerateMoves(Violent, NullMove)
	for move := eng.stack.PopMove(); move != NullMove; move = eng.stack.PopMove() {
		// Prune futile moves that would anyway result in a stand-pat
		// at that next depth.
		if !inCheck && isFutile(pos, static, localα, futilityMargin, move) {
			// TODO: should it update localα?
			continue
		}

		// Discard illegal or losing captures.
		eng.DoMove(move)
		if eng.Position.IsChecked(us) ||
			!inCheck && move.MoveType() == Normal && seeSign(pos, move) {
			eng.UndoMove()
			continue
		}

		///////////////////////////////////////////////////
		// NEW
		// In Racing Kings avoid captures that give check.
		if Variant == VARIANT_Racing_Kings {
			if eng.Position.IsCheckedLocal(us.Opposite()) {
				eng.UndoMove()
				continue
			}
		}
		///////////////////////////////////////////////////

		score := -eng.searchQuiescence(-β, -localα)
		eng.UndoMove()

		if score >= β {
			return score
		}
		if score > localα {
			localα = score
			bestMove = move
		}
	}

	if α < localα && localα < β {
		eng.pvTable.Put(eng.Position, bestMove)
	}
	return localα
}

// tryMove makes a move and descends on the search tree.
//
// α, β represent lower and upper bounds.
// ply is the move number (increasing).
// depth is the remaining depth (decreasing)
// lmr is how much to reduce a late move. Implies non-null move.
// nullWindow indicates whether to scout first. Implies non-null move.
// move is the move to execute. Can be NullMove.
//
// Returns the score from the deeper search.
func (eng *Engine) tryMove(α, β, depth, lmr int32, nullWindow bool, move Move) int32 {
	depth--

	score := α + 1
	if lmr > 0 { // reduce late moves
		score = -eng.searchTree(-α-1, -α, depth-lmr)
	}

	if score > α { // if late move reduction is disabled or has failed
		if nullWindow {
			score = -eng.searchTree(-α-1, -α, depth)
			if α < score && score < β {
				score = -eng.searchTree(-β, -α, depth)
			}
		} else {
			score = -eng.searchTree(-β, -α, depth)
		}
	}

	eng.UndoMove()
	return score
}

// ply returns the ply from the beginning of the search.
func (eng *Engine) ply() int32 {
	return int32(eng.Position.Ply - eng.rootPly)
}

// passed returns true if a passed pawn appears or disappears.
//
// TODO: The heuristic is incomplete and doesn't handled discovered passed pawns.
func passed(pos *Position, m Move) bool {
	if m.Piece().Figure() == Pawn {
		// Checks no pawns are in front on its and adjacent files.
		bb := m.To().Bitboard()
		bb = West(bb) | bb | East(bb)
		pawns := pos.ByFigure[Pawn] &^ m.To().Bitboard() &^ m.From().Bitboard()
		if ForwardSpan(m.SideToMove(), bb)&pawns == 0 {
			return true
		}
	}
	if m.Capture().Figure() == Pawn {
		// Checks no pawns are in front on its and adjacent files.
		bb := m.To().Bitboard()
		bb = West(bb) | bb | East(bb)
		pawns := pos.ByFigure[Pawn] &^ m.To().Bitboard() &^ m.From().Bitboard()
		if BackwardSpan(m.SideToMove(), bb)&pawns == 0 {
			return true
		}
	}
	return false
}

// searchTree implements searchTree framework.
//
// searchTree fails soft, i.e. the score returned can be outside the bounds.
//
// α, β represent lower and upper bounds.
// depth is the search depth (decreasing)
//
// Returns the score of the current position up to depth (modulo reductions/extensions).
// The returned score is from current player's POV.
//
// Invariants:
//   If score <= α then the search failed low and the score is an upper bound.
//   else if score >= β then the search failed high and the score is a lower bound.
//   else score is exact.
//
// Assuming this is a maximizing nodes, failing high means that a
// minimizing ancestor node already has a better alternative.
func (eng *Engine) searchTree(α, β, depth int32) int32 {
	ply := eng.ply()
	pvNode := α+1 < β
	pos := eng.Position
	us := pos.SideToMove

	// Update statistics.
	eng.Stats.Nodes++
	if !eng.stopped && eng.Stats.Nodes >= eng.checkpoint {
		eng.checkpoint = eng.Stats.Nodes + checkpointStep
		if eng.timeControl.Stopped() {
			eng.stopped = true
		}
	}
	if eng.stopped {
		return α
	}
	if pvNode && ply > eng.Stats.SelDepth {
		eng.Stats.SelDepth = eng.ply()
	}

	// Verify that this is not already an endgame.
	if score, done := eng.endPosition(); done {
		return score
	}

	// Mate pruning: If an ancestor already has a mate in ply moves then
	// the search will always fail low so we return the lowest wining score.
	if MateScore-ply <= α {
		return KnownWinScore
	}

	// Check the transposition table.
	entry := eng.retrieveHash()
	hash := entry.move
	if entry.kind != noEntry && depth <= int32(entry.depth) {
		if entry.kind == exact {
			// Simply return if the score is exact.
			// Update principal variation table if possible.
			if α < entry.score && entry.score < β {
				eng.pvTable.Put(pos, hash)
			}
			return entry.score
		}
		if entry.kind == failedLow && entry.score <= α {
			// Previously the move failed low so the actual score
			// is at most entry.score. If that's lower than α
			// this will also fail low.
			return entry.score
		}
		if entry.kind == failedHigh && entry.score >= β {
			// Previously the move failed high so the actual score
			// is at least entry.score. If that's higher than β
			// this will also fail high.
			return entry.score
		}
	}

	// Stop searching when the maximum search depth is reached.
	if depth <= 0 {
		// Depth can be < 0 due to aggressive LMR.
		score := eng.searchQuiescence(α, β)
		eng.updateHash(α, β, depth, score, NullMove)
		return score
	}

	sideIsChecked := pos.IsChecked(us)

	// Do a null move. If the null move fails high then the current
	// position is too good, so opponent will not play it.
	// Verification that we are not in check is done by tryMove
	// which bails out if after the null move we are still in check.
	if depth > NullMoveDepthLimit && // not very close to leafs
		!sideIsChecked && // nullmove is illegal when in check
		pos.HasNonPawns(us) && // at least one minor/major piece.
		KnownLossScore < α && β < KnownWinScore { // disable in lost or won positions

		reduction := NullMoveDepthReduction
		if pos.NumNonPawns(us) >= 3 {
			// Reduce more when there are three minor/major pieces.
			reduction++
		}

		eng.DoMove(NullMove)
		score := eng.tryMove(β-1, β, depth-reduction, 0, false, NullMove)
		if score >= β {
			return score
		}
	}

	bestMove, bestScore := NullMove, -InfinityScore

	// Futility and history pruning at frontier nodes.
	// Based on Deep Futility Pruning http://home.hccnet.nl/h.g.muller/deepfut.html
	// Based on History Leaf Pruning https://chessprogramming.wikispaces.com/History+Leaf+Pruning
	static := int32(0)
	allowLeafsPruning := false
	if depth <= FutilityDepthLimit && // enable when close to the frontier
		!sideIsChecked && // disable in check
		!pvNode && // disable in pv nodes
		KnownLossScore < α && β < KnownWinScore { // disable when searching for a mate
		allowLeafsPruning = true
		static = eng.Score()
	}

	// Principal variation search: search with a null window if there is already a good move.
	nullWindow := false // updated once alpha is improved
	// Late move reduction: search best moves with full depth, reduce remaining moves.
	allowLateMove := !sideIsChecked && depth > LMRDepthLimit

	// dropped true if not all moves were searched.
	// Mate cannot be declared unless all moves were tested.
	dropped := false
	numQuiet := int32(0)
	localα := α

	eng.stack.GenerateMoves(All, hash)
	for move := eng.stack.PopMove(); move != NullMove; move = eng.stack.PopMove() {
		critical := move == hash || eng.stack.IsKiller(move)
		if move.IsQuiet() {
			numQuiet++ // TODO: Move from here.
		}

		newDepth := depth
		eng.DoMove(move)

		// Skip illegal moves that leave the king in check.
		if pos.IsChecked(us) {
			eng.UndoMove()
			continue
		}

		///////////////////////////////////////////////////
		// NEW
		// In Racing Kings skip moves that give check.
		if Variant == VARIANT_Racing_Kings {
			if pos.IsCheckedLocal(us.Opposite()) {
				eng.UndoMove()
				continue
			}
		}
		///////////////////////////////////////////////////

		// Extend the search when our move gives check.
		// However do not extend if we can just take the undefended piece.
		// See discussion: http://www.talkchess.com/forum/viewtopic.php?t=56361
		// When the move gives check, history pruning and futility pruning are also disabled.
		givesCheck := pos.IsChecked(us.Opposite())
		if givesCheck {
			if pos.GetAttacker(move.To(), us.Opposite()) == NoFigure ||
				pos.GetAttacker(move.To(), us) != NoFigure {
				newDepth += CheckDepthExtension
			}
		}

		// Reduce late quiet moves and bad captures.
		// TODO: Do not compute see when in check.
		lmr := int32(0)
		if allowLateMove && !givesCheck && !critical {
			if move.IsQuiet() {
				// Reduce quiet moves more at high depths and after many quiet moves.
				// Large numQuiet means it's likely not a CUT node.
				// Large depth means reductions are less risky.
				lmr = 1 + min(depth, numQuiet)/5
			} else if seeSign(pos, move) {
				// Bad captures (SEE<0) can be reduced, too.
				lmr = 1
			}
		}

		// Prune moves close to frontier.
		if allowLeafsPruning && !givesCheck && !critical {
			// Prune quiet moves that performed bad historically.
			if bad, good := eng.history.get(move); bad > 16*good && (move.IsQuiet() || seeSign(pos, move)) {
				dropped = true
				eng.UndoMove()
				continue
			}
			// Prune moves that do not raise alphas.
			if isFutile(pos, static, localα, depth*futilityMargin, move) {
				bestScore = max(bestScore, static)
				dropped = true
				eng.UndoMove()
				continue
			}
		}

		score := eng.tryMove(localα, β, newDepth, lmr, nullWindow, move)
		if allowLeafsPruning && !givesCheck { // Update history scores.
			eng.history.inc(move, score > α)
		}
		if score >= β { // Fail high, cut node.
			eng.stack.SaveKiller(move)
			eng.updateHash(α, β, depth, score, move)
			return score
		}
		if score > bestScore {
			nullWindow = true
			bestMove, bestScore = move, score
			localα = max(localα, score)
		}
	}

	if !dropped {
		// If no move was found then the game is over.
		if bestMove == NullMove {
			if sideIsChecked {
				bestScore = MatedScore + ply
			} else {
				bestScore = 0
			}
		}
		// Update hash and principal variation tables.
		eng.updateHash(α, β, depth, bestScore, bestMove)
		if α < bestScore && bestScore < β {
			eng.pvTable.Put(pos, bestMove)
		}
	}

	return bestScore
}

// search starts the search up to depth depth.
// The returned score is from current side to move POV.
// estimated is the score from previous depths.
func (eng *Engine) search(depth, estimated int32) int32 {
	// This method only implements aspiration windows.
	//
	// The gradual widening algorithm is the one used by RobboLito
	// and Stockfish and it is explained here:
	// http://www.talkchess.com/forum/viewtopic.php?topic_view=threads&p=499768&t=46624
	γ, δ := estimated, int32(initialAspirationWindow)
	α, β := max(γ-δ, -InfinityScore), min(γ+δ, InfinityScore)
	score := estimated

	if depth < 4 {
		// Disable aspiration window for very low search depths.
		// This wastes a lot of time when for tunning.
		α = -InfinityScore
		β = +InfinityScore
	}

	for !eng.stopped {
		// At root a non-null move is required, cannot prune based on null-move.
		score = eng.searchTree(α, β, depth)
		if score <= α {
			α = max(α-δ, -InfinityScore)
			δ += δ / 2
		} else if score >= β {
			β = min(β+δ, InfinityScore)
			δ += δ / 2
		} else {
			return score
		}
	}

	return score
}

// Play evaluates current position.
//
// Returns the principal variation, that is
//	moves[0] is the best move found and
//	moves[1] is the pondering move.
//
// If no move was found because the game has finished
// then an empty pv is returned.
//
// Time control, tc, should already be started.
func (eng *Engine) Play(tc *TimeControl) (moves []Move) {
	eng.Log.BeginSearch()
	eng.Stats = Stats{Depth: -1}

	eng.rootPly = eng.Position.Ply
	eng.timeControl = tc
	eng.stopped = false
	eng.checkpoint = checkpointStep
	eng.stack.Reset(eng.Position)

	score := int32(0)
	for depth := int32(0); depth < 64; depth++ {
		if !tc.NextDepth(depth) {
			// Stop if tc control says we are done.
			// Search at least one depth, otherwise a move cannot be returned.
			break
		}

		eng.Stats.Depth = depth
		score = eng.search(depth, score)

		if !eng.stopped {
			// if eng has not been stopped then this is a legit pv.
			moves = eng.pvTable.Get(eng.Position)
			eng.Log.PrintPV(eng.Stats, score, moves)
		}
	}

	eng.Log.EndSearch()
	return moves
}

// isFutile return true if m cannot raise current static
// evaluation above α. This is just an heuristic and mistakes
// can happen.
func isFutile(pos *Position, static, α, margin int32, m Move) bool {
	if m.MoveType() == Promotion {
		// Promotion and passed pawns can increase static evaluation
		// by more than futilityMargin.
		return false
	}
	f := m.Capture().Figure()
	δ := ScaleToCentiPawn(max(wFigure[f].M, wFigure[f].E))
	return static+δ+margin < α && !passed(pos, m)
}
