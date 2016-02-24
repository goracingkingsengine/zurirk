// move_ordering generates and orders moves for an engine.
// Generation is done in several phases and many times
// actual generation or sorting can be eliminated.

package engine

const (
	// Move generation states.

	msHash          = iota // return hash move
	msGenViolent           // generate violent moves
	msReturnViolent        // return violent moves in order
	msGenKiller            // generate killer moves
	msReturnKiller         // return killer moves  in order
	msGenRest              // generate remaining moves
	msReturnRest           // return remaining moves in order
	msDone                 // all moves returned
)

var (
	// mvvlva values based on one pawn = 10.
	mvvlvaBonus = [...]int16{0, 10, 40, 45, 68, 145, 256}
)

// mvvlva computes Most Valuable Victim / Least Valuable Aggressor
// https://chessprogramming.wikispaces.com/MVV-LVA
func mvvlva(m Move) int16 {
	a := int(m.Target().Figure())
	v := int(m.Capture().Figure())
	return int16(mvvlvaBonus[v]*64 - mvvlvaBonus[a])
}

// movesStack is a stack of moves.
type moveStack struct {
	moves []Move  // list of moves
	order []int16 // weight of each move for comparison

	kind   int     // violent or all
	state  int     // current generation state
	hash   Move    // hash move
	killer [4]Move // killer moves
}

// stack is a stack of plies (movesStack).
type stack struct {
	position *Position
	moves    []moveStack
}

// Reset clear the stack for a new position.
func (st *stack) Reset(pos *Position) {
	st.position = pos
	st.moves = st.moves[:0]
}

// get returns the moveStack for current ply.
// allocates memory if necessary.
func (st *stack) get() *moveStack {
	for len(st.moves) <= st.position.Ply {
		st.moves = append(st.moves, moveStack{
			moves: make([]Move, 0, 4),
			order: make([]int16, 0, 4),
		})
	}
	return &st.moves[st.position.Ply]
}

// GenerateMoves generates all moves of kind.
func (st *stack) GenerateMoves(kind int, hash Move) {
	ms := st.get()
	ms.moves = ms.moves[:0] // clear the array, but keep the backing memory
	ms.order = ms.order[:0]
	ms.kind = kind
	ms.state = msHash
	ms.hash = hash
	// ms.killer = ms.killer // keep killers
}

// generateMoves generates all moves.
func (st *stack) generateMoves(kind int) {
	ms := &st.moves[st.position.Ply]
	if len(ms.moves) != 0 || len(ms.order) != 0 {
		panic("expected no moves")
	}
	if ms.kind&kind == 0 {
		return
	}
	st.position.GenerateMoves(ms.kind&kind, &ms.moves)
	for _, m := range ms.moves {
		ms.order = append(ms.order, mvvlva(m))
	}
}

// moveBest moves best move to front.
func (st *stack) moveBest() {
	ms := &st.moves[st.position.Ply]
	if len(ms.moves) == 0 {
		return
	}

	bi := 0
	for i := range ms.moves {
		if ms.order[i] > ms.order[bi] {
			bi = i
		}
	}

	last := len(ms.moves) - 1
	ms.moves[bi], ms.moves[last] = ms.moves[last], ms.moves[bi]
	ms.order[bi], ms.order[last] = ms.order[last], ms.order[bi]
}

// popFront pops the move from the front
func (st *stack) popFront() Move {
	ms := &st.moves[st.position.Ply]
	if len(ms.moves) == 0 {
		return NullMove
	}

	last := len(ms.moves) - 1
	move := ms.moves[last]
	ms.moves = ms.moves[:last]
	ms.order = ms.order[:last]
	return move
}

// Pop pops a new move.
// Returns NullMove if there are no moves.
// Moves are generated in several phases:
//	first the hash move,
//      then the violent moves,
//      then the killer moves,
//      then the tactical and quiet moves.
func (st *stack) PopMove() Move {
	ms := &st.moves[st.position.Ply]
	for {
		switch ms.state {
		// Return the hash move.
		case msHash:
			// Return the hash move directly without generating the pseudo legal moves.
			ms.state = msGenViolent
			if st.position.IsPseudoLegal(ms.hash) {
				return ms.hash
			}

		// Return the violent moves.
		case msGenViolent:
			ms.state = msReturnViolent
			st.generateMoves(Violent)

		case msReturnViolent:
			// Most positions have only very violent moves so
			// it doesn't make sense to sort given that captures have a high
			// chance to fail high. We just pop the moves in order of score.
			st.moveBest()
			if m := st.popFront(); m == NullMove {
				if ms.kind&(Tactical|Quiet) == 0 {
					// Optimization: skip remaining steps if no Tactical or Quiet moves
					// were requested (e.g. in quiescence search).
					ms.state = msDone
				} else {
					ms.state = msGenKiller
				}
			} else if m == ms.hash {
				break
			} else if m != NullMove {
				return m
			}

		// Return killer moves.
		// NB: Not all killer moves are valid.
		case msGenKiller:
			ms.state = msReturnKiller
			for i := len(ms.killer) - 1; i >= 0; i-- {
				if m := ms.killer[i]; m != NullMove {
					ms.moves = append(ms.moves, ms.killer[i])
					ms.order = append(ms.order, -int16(i))
				}
			}

		case msReturnKiller:
			if m := st.popFront(); m == NullMove {
				ms.state = msGenRest
			} else if m == ms.hash {
				break
			} else if st.position.IsPseudoLegal(m) {
				return m
			}

		// Return the quiet and tactical moves in the order they were generated.
		case msGenRest:
			ms.state = msReturnRest
			st.generateMoves(Tactical | Quiet)

		case msReturnRest:
			if m := st.popFront(); m == NullMove {
				ms.state = msDone
			} else if m == ms.hash || st.IsKiller(m) {
				break
			} else {
				return m
			}

		case msDone:
			// Just in case another move is requested.
			return NullMove
		}

	}
}

// IsKiller returns true if m is a killer move for currenty ply.
func (st *stack) IsKiller(m Move) bool {
	ms := &st.moves[st.position.Ply]
	return m == ms.killer[0] || m == ms.killer[1] || m == ms.killer[2] || m == ms.killer[3]
}

// SaveKiller saves a killer move, m.
func (st *stack) SaveKiller(m Move) {
	ms := &st.moves[st.position.Ply]
	if !m.IsViolent() {
		// Move the newly found killer first.
		if m == ms.killer[0] {
			// do nothing
		} else if m == ms.killer[1] {
			ms.killer[1] = ms.killer[0]
			ms.killer[0] = m
		} else if m == ms.killer[2] {
			ms.killer[2] = ms.killer[1]
			ms.killer[1] = ms.killer[0]
			ms.killer[0] = m
		} else {
			ms.killer[3] = ms.killer[2]
			ms.killer[2] = ms.killer[1]
			ms.killer[1] = ms.killer[0]
			ms.killer[0] = m
		}
	}
}
