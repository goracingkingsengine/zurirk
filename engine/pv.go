package engine

const (
	pvTableSize = 1 << 13
	pvTableMask = pvTableSize - 1
)

// TODO: Unexport pvEntry fields.
type pvEntry struct {
	// lock is used to handled hash conflicts.
	// Normally set to position's Zobrist key.
	lock uint64
	// When was the move added.
	birth uint32
	// move on pricipal variation for this position.
	move Move
}

// pvTable is like hash table, but only to keep principal variation.
//
// The additional table to store the PV was suggested by Robert Hyatt. See
//
// * http://www.talkchess.com/forum/viewtopic.php?topic_view=threads&p=369163&t=35982
// * http://www.talkchess.com/forum/viewtopic.php?t=36099
//
// During alpha-beta search entries that are on principal variation,
// are exact nodes, i.e. their score lies exactly between alpha and beta.
type pvTable struct {
	table []pvEntry
	timer uint32
}

// newPvTable returns a new pvTable.
func newPvTable() pvTable {
	return pvTable{
		table: make([]pvEntry, pvTableSize),
		timer: 0,
	}
}

// Put inserts a new entry.  Ignores NullMoves.
func (pv *pvTable) Put(pos *Position, move Move) {
	if move == NullMove {
		return
	}

	// Based on pos.Zobrist() two entries are looked up.
	// If any of the two entries in the table matches
	// current position, then that one is replaced.
	// Otherwise, the older is replaced.

	entry1 := &pv.table[uint32(pos.Zobrist())&pvTableMask]
	entry2 := &pv.table[uint32(pos.Zobrist()>>32)&pvTableMask]
	zobrist := pos.Zobrist()

	var entry *pvEntry
	if entry1.lock == zobrist {
		entry = entry1
	} else if entry2.lock == zobrist {
		entry = entry2
	} else if entry1.birth <= entry2.birth {
		entry = entry1
	} else {
		entry = entry2
	}

	pv.timer++
	*entry = pvEntry{
		lock:  pos.Zobrist(),
		move:  move,
		birth: pv.timer,
	}
}

// TODO: Lookup move in transposition table if none is available.
func (pv *pvTable) get(pos *Position) Move {
	entry1 := &pv.table[uint32(pos.Zobrist())&pvTableMask]
	entry2 := &pv.table[uint32(pos.Zobrist()>>32)&pvTableMask]
	zobrist := pos.Zobrist()

	var entry *pvEntry
	if entry1.lock == zobrist {
		entry = entry1
	}
	if entry2.lock == zobrist {
		entry = entry2
	}
	if entry == nil {
		return NullMove
	}

	return entry.move
}

// Get returns the principal variation.
func (pv *pvTable) Get(pos *Position) []Move {
	seen := make(map[uint64]bool)
	var moves []Move

	// Extract the moves by following the position.
	next := pv.get(pos)
	for next != NullMove && !seen[pos.Zobrist()] {
		seen[pos.Zobrist()] = true
		moves = append(moves, next)
		pos.DoMove(next)
		next = pv.get(pos)
	}

	// Undo all moves, so we get back to the initial state.
	for i := len(moves) - 1; i >= 0; i-- {
		pos.UndoMove()
	}
	return moves
}
