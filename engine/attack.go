// attack.go generates move bitboards for all pieces.
// zurichess uses magic bitboards to generate sliding pieces moves.
// A very good description by Pradyumna Kannan can be read at:
// http://www.pradu.us/old/Nov27_2008/Buzz/research/magic/Bitboards.pdf
//
// TODO: move magic generation into an internal package.

package engine

import (
	"fmt"
	"math"
	"math/rand"
)

var (
	// bbPawnAttack contains pawn's attack tables.
	bbPawnAttack [64]Bitboard
	// bbKnightAttack contains knight's attack tables.
	bbKnightAttack [64]Bitboard
	// bbKingAttack contains king's attack tables (excluding castling).
	bbKingAttack [64]Bitboard
	// bbSuperAttack contains queen piece's attack tables. This queen can jump.
	bbSuperAttack [64]Bitboard

	rookMagic    [64]magicInfo
	rookDeltas   = [][2]int{{-1, +0}, {+1, +0}, {+0, -1}, {+0, +1}}
	bishopMagic  [64]magicInfo
	bishopDeltas = [][2]int{{-1, +1}, {+1, +1}, {+1, -1}, {-1, -1}}
)

func init() {
	initBbPawnAttack()
	initBbKnightAttack()
	initBbKingAttack()
	initBbSuperAttack()
	initRookMagic()
	initBishopMagic()
}

func initJumpAttack(jump [][2]int, attack []Bitboard) {
	for r := 0; r < 8; r++ {
		for f := 0; f < 8; f++ {
			bb := Bitboard(0)
			for _, d := range jump {
				r0, f0 := r+d[0], f+d[1]
				if 0 > r0 || r0 >= 8 || 0 > f0 || f0 >= 8 {
					continue
				}
				bb |= RankFile(r0, f0).Bitboard()
			}
			attack[RankFile(r, f)] = bb
		}
	}
}

func initBbPawnAttack() {
	pawnJump := [][2]int{
		{-1, -1}, {-1, +1}, {+1, +1}, {+1, -1},
	}
	initJumpAttack(pawnJump, bbPawnAttack[:])
}

func initBbKnightAttack() {
	knightJump := [][2]int{
		{-2, -1}, {-2, +1}, {+2, -1}, {+2, +1},
		{-1, -2}, {-1, +2}, {+1, -2}, {+1, +2},
	}
	initJumpAttack(knightJump, bbKnightAttack[:])
}

func initBbKingAttack() {
	kingJump := [][2]int{
		{-1, -1}, {-1, +0}, {-1, +1}, {+0, +1},
		{+1, +1}, {+1, +0}, {+1, -1}, {+0, -1},
	}
	initJumpAttack(kingJump, bbKingAttack[:])
}

func initBbSuperAttack() {
	for sq := SquareMinValue; sq <= SquareMaxValue; sq++ {
		bbSuperAttack[sq] = slidingAttack(sq, rookDeltas, BbEmpty) | slidingAttack(sq, bishopDeltas, BbEmpty)
	}
}

func slidingAttack(sq Square, deltas [][2]int, occupancy Bitboard) Bitboard {
	r, f := sq.Rank(), sq.File()
	bb := Bitboard(0)
	for _, d := range deltas {
		r0, f0 := r, f
		for {
			r0, f0 = r0+d[0], f0+d[1]
			if 0 > r0 || r0 >= 8 || 0 > f0 || f0 >= 8 {
				// Stop when outside of the board.
				break
			}
			sq0 := RankFile(r0, f0)
			bb |= sq0.Bitboard()
			if occupancy&sq0.Bitboard() != 0 {
				// Stop when a piece was hit.
				break
			}
		}
	}
	return bb
}

func spell(magic uint64, shift uint, bb Bitboard) uint {
	mul := magic * uint64(bb)
	return uint(uint32(mul>>32^mul) >> shift)
}

type magicInfo struct {
	store []Bitboard // attack boards of size 1<<(64-shift)
	mask  Bitboard   // square's mask.
	magic uint64     // magic multiplier
	shift uint       // shift bits to index store
	pad   [2]uint64  // padding so the structure has 32 bytes.
}

func (mi *magicInfo) Attack(ref Bitboard) Bitboard {
	return mi.store[spell(mi.magic, mi.shift, ref&mi.mask)]
}

type wizard struct {
	// Sliding deltas.
	Deltas        [][2]int
	MinShift      uint // Which shifts to search.
	MaxShift      uint
	MaxNumEntries uint // How much to search.
	Rand          *rand.Rand

	numMagicTests uint
	magics        [64]uint64
	shifts        [64]uint // Number of bits for indexes.

	store     []Bitboard // Temporary store to check hash collisions.
	reference []Bitboard
	occupancy []Bitboard
}

func (wiz *wizard) tryMagicNumber(mi *magicInfo, sq Square, magic uint64, shift uint) bool {
	wiz.numMagicTests++

	// Clear store.
	if len(wiz.store) < 1<<shift {
		wiz.store = make([]Bitboard, 1<<shift)
	}
	for j := range wiz.store[:1<<shift] {
		wiz.store[j] = 0
	}

	// Verify that magic gives a perfect hash.
	for i, bb := range wiz.reference {
		index := spell(magic, 32-shift, bb)
		if wiz.store[index] != 0 && wiz.store[index] != wiz.occupancy[i] {
			return false
		}
		wiz.store[index] = wiz.occupancy[i]
	}

	// Perfect hash, store it.
	wiz.magics[sq] = magic
	wiz.shifts[sq] = shift

	mi.store = make([]Bitboard, 1<<shift)
	copy(mi.store, wiz.store)
	mi.mask = wiz.mask(sq)
	mi.magic = magic
	mi.shift = 32 - shift
	return true
}

// randMagic returns a random magic number
func (wiz *wizard) randMagic() uint64 {
	r := uint64(wiz.Rand.Int63())
	r &= uint64(wiz.Rand.Int63())
	r &= uint64(wiz.Rand.Int63())
	return r<<6 + 1
}

// mask is the attack set on empty board minus the border.
func (wiz *wizard) mask(sq Square) Bitboard {
	// Compute border. Trick source: stockfish.
	border := (BbRank1 | BbRank8) & ^RankBb(sq.Rank())
	border |= (BbFileA | BbFileH) & ^FileBb(sq.File())
	return ^border & slidingAttack(sq, wiz.Deltas, BbEmpty)
}

// prepare computes reference and occupancy tables for a square.
func (wiz *wizard) prepare(sq Square) {
	wiz.reference = wiz.reference[:0]
	wiz.occupancy = wiz.occupancy[:0]

	// Carry-Rippler trick to enumerate all subsets of mask.
	for mask, subset := wiz.mask(sq), Bitboard(0); ; {
		attack := slidingAttack(sq, wiz.Deltas, subset)
		wiz.reference = append(wiz.reference, subset)
		wiz.occupancy = append(wiz.occupancy, attack)
		subset = (subset - mask) & mask
		if subset == 0 {
			break
		}
	}
}

func (wiz *wizard) searchMagic(sq Square, mi *magicInfo) {
	if wiz.shifts[sq] != 0 && wiz.shifts[sq] <= wiz.MinShift {
		// Don't search if shift is low enough.
		return
	}

	// Try magic numbers with small shifts.
	wiz.prepare(sq)
	mask := wiz.mask(sq)
	for i := 0; i < 100 || wiz.shifts[sq] == 0; i++ {
		// Pick a smaller shift than current best.
		var shift uint
		if wiz.shifts[sq] == 0 {
			shift = wiz.MaxShift
		} else {
			shift = wiz.shifts[sq] - 1
		}

		// Pick a good magic and test whether it gives a perfect hash.
		var magic uint64
		for popcnt(uint64(mask)*magic) < 6 {
			magic = wiz.randMagic()
		}
		wiz.tryMagicNumber(mi, sq, magic, shift)
	}
}

// SearchMagic finds new magics.
func (wiz *wizard) SearchMagics(mi []magicInfo) {
	numEntries := uint(math.MaxUint32)
	minShift := uint(math.MaxUint32)
	for numEntries > wiz.MaxNumEntries {
		numEntries = 0
		for sq := SquareMinValue; sq <= SquareMaxValue; sq++ {
			wiz.searchMagic(sq, &mi[sq])
			numEntries += 1 << wiz.shifts[sq]
			if minShift > wiz.shifts[sq] {
				minShift = wiz.shifts[sq]
			}
		}
	}
}

func (wiz *wizard) SetMagic(mi []magicInfo, sq Square, magic uint64, shift uint) {
	wiz.prepare(sq)
	if !wiz.tryMagicNumber(&mi[sq], sq, magic, shift) {
		panic(fmt.Sprintf("invalid magic: sq=%v magic=%d shift=%d", sq, magic, shift))
	}
}

func initRookMagic() {
	wiz := &wizard{
		Deltas:        rookDeltas,
		MinShift:      10,
		MaxShift:      13,
		MaxNumEntries: 130000,
		Rand:          rand.New(rand.NewSource(1)),
	}

	// A set of known good magics for rook.
	// Finding good rook magics is slow, so we just use some precomputed values.
	// For readability reasons, do not make an array.
	wiz.SetMagic(rookMagic[:], SquareA1, 36028952711532673, 12)
	wiz.SetMagic(rookMagic[:], SquareA2, 5066692388487169, 11)
	wiz.SetMagic(rookMagic[:], SquareA3, 4631389266822304769, 11)
	wiz.SetMagic(rookMagic[:], SquareA4, 10450310413697025, 11)
	wiz.SetMagic(rookMagic[:], SquareA5, 140737496752193, 11)
	wiz.SetMagic(rookMagic[:], SquareA6, 4755801345016995841, 11)
	wiz.SetMagic(rookMagic[:], SquareA7, 2310346608845258881, 11)
	wiz.SetMagic(rookMagic[:], SquareA8, 1153273486052196353, 12)
	wiz.SetMagic(rookMagic[:], SquareB1, 14411536674683101313, 11)
	wiz.SetMagic(rookMagic[:], SquareB2, 360288245069774977, 10)
	wiz.SetMagic(rookMagic[:], SquareB3, 9304436831221219585, 10)
	wiz.SetMagic(rookMagic[:], SquareB4, 90107726679507201, 10)
	wiz.SetMagic(rookMagic[:], SquareB5, 23081233739161857, 10)
	wiz.SetMagic(rookMagic[:], SquareB6, 17610976739329, 10)
	wiz.SetMagic(rookMagic[:], SquareB7, 9007201406419201, 10)
	wiz.SetMagic(rookMagic[:], SquareB8, 846729215754241, 11)
	wiz.SetMagic(rookMagic[:], SquareC1, 576496005395513857, 11)
	wiz.SetMagic(rookMagic[:], SquareC2, 2355383154875302401, 10)
	wiz.SetMagic(rookMagic[:], SquareC3, 9263904435128516865, 10)
	wiz.SetMagic(rookMagic[:], SquareC4, 9223653580555165697, 10)
	wiz.SetMagic(rookMagic[:], SquareC5, 216208542045048897, 10)
	wiz.SetMagic(rookMagic[:], SquareC6, 2667820173397917761, 10)
	wiz.SetMagic(rookMagic[:], SquareC7, 360428707682197761, 10)
	wiz.SetMagic(rookMagic[:], SquareC8, 4611695089401765889, 11)
	wiz.SetMagic(rookMagic[:], SquareD1, 4604372721729, 11)
	wiz.SetMagic(rookMagic[:], SquareD2, 9304436898871644161, 10)
	wiz.SetMagic(rookMagic[:], SquareD3, 596726951168704769, 10)
	wiz.SetMagic(rookMagic[:], SquareD4, 5190691178076966913, 10)
	wiz.SetMagic(rookMagic[:], SquareD5, 4655469687738433, 10)
	wiz.SetMagic(rookMagic[:], SquareD6, 5764660368316567553, 10)
	wiz.SetMagic(rookMagic[:], SquareD7, 2452350872031592705, 10)
	wiz.SetMagic(rookMagic[:], SquareD8, 1153211792858550273, 11)
	wiz.SetMagic(rookMagic[:], SquareE1, 36031546200687617, 11)
	wiz.SetMagic(rookMagic[:], SquareE2, 144115499663886337, 10)
	wiz.SetMagic(rookMagic[:], SquareE3, 288388705826635841, 10)
	wiz.SetMagic(rookMagic[:], SquareE4, 74380329532524545, 10)
	wiz.SetMagic(rookMagic[:], SquareE5, 4910190248417298433, 10)
	wiz.SetMagic(rookMagic[:], SquareE6, 2251851487527425, 10)
	wiz.SetMagic(rookMagic[:], SquareE7, 7881299415531649, 10)
	wiz.SetMagic(rookMagic[:], SquareE8, 54342271281408001, 11)
	wiz.SetMagic(rookMagic[:], SquareF1, 36033197213089793, 11)
	wiz.SetMagic(rookMagic[:], SquareF2, 108086941350626369, 10)
	wiz.SetMagic(rookMagic[:], SquareF3, 1298162592589676609, 10)
	wiz.SetMagic(rookMagic[:], SquareF4, 9269586743957521409, 10)
	wiz.SetMagic(rookMagic[:], SquareF5, 140754676613633, 10)
	wiz.SetMagic(rookMagic[:], SquareF6, 8859435012, 10)
	wiz.SetMagic(rookMagic[:], SquareF7, 105622918137857, 10)
	wiz.SetMagic(rookMagic[:], SquareF8, 93452063091195905, 11)
	wiz.SetMagic(rookMagic[:], SquareG1, 3848292811265, 11)
	wiz.SetMagic(rookMagic[:], SquareG2, 9441796687501985793, 10)
	wiz.SetMagic(rookMagic[:], SquareG3, 668793341028205569, 10)
	wiz.SetMagic(rookMagic[:], SquareG4, 3503805114303512577, 10)
	wiz.SetMagic(rookMagic[:], SquareG5, 1441856117960359937, 10)
	wiz.SetMagic(rookMagic[:], SquareG6, 648529410319974401, 10)
	wiz.SetMagic(rookMagic[:], SquareG7, 13979322776982393857, 10)
	wiz.SetMagic(rookMagic[:], SquareG8, 13835060872780858369, 11)
	wiz.SetMagic(rookMagic[:], SquareH1, 4539788820801, 12)
	wiz.SetMagic(rookMagic[:], SquareH2, 2359886214407946241, 11)
	wiz.SetMagic(rookMagic[:], SquareH3, 27041389040664577, 11)
	wiz.SetMagic(rookMagic[:], SquareH4, 159429253169153, 11)
	wiz.SetMagic(rookMagic[:], SquareH5, 4613955963706147841, 11)
	wiz.SetMagic(rookMagic[:], SquareH6, 4611686019534716929, 11)
	wiz.SetMagic(rookMagic[:], SquareH7, 27025995845339137, 11)
	wiz.SetMagic(rookMagic[:], SquareH8, 633464726504577, 12)

	// Enable the next line to find new magics.
	// wiz.SearchMagics(rookMagic[:])
}

func initBishopMagic() {
	wiz := &wizard{
		Deltas:        bishopDeltas,
		MinShift:      5,
		MaxShift:      9,
		MaxNumEntries: 6000,
		Rand:          rand.New(rand.NewSource(1)),
	}

	// Bishop magics, unlike rook magics are easy to find.
	wiz.SearchMagics(bishopMagic[:])
}

// KnightMobility returns all squares a knight can reach from sq.
func KnightMobility(sq Square) Bitboard {
	return bbKnightAttack[sq]
}

// BishopMobility returns the squares a bishop can reach from sq given all pieces.
func BishopMobility(sq Square, all Bitboard) Bitboard {
	return bishopMagic[sq].Attack(all)
}

// RookMobility returns the squares a rook can reach from sq given all pieces.
func RookMobility(sq Square, all Bitboard) Bitboard {
	return rookMagic[sq].Attack(all)
}

// QueenMobility returns the squares a queen can reach from sq given all pieces.
func QueenMobility(sq Square, all Bitboard) Bitboard {
	return rookMagic[sq].Attack(all) | bishopMagic[sq].Attack(all)
}

// KingMobility returns all squares a king can reach from sq.
// Doesn't include castling.
func KingMobility(sq Square) Bitboard {
	return bbKingAttack[sq]
}
