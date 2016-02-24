// see.go implements static exchange evaluation.

package engine

// piece bonuses when calulating the see.
// The values are fixed to approximatively the figure bonus in mid game.
var seeBonus = [FigureArraySize]int32{0, 55, 325, 341, 454, 1110, 20000}

func seeScore(m Move) int32 {
	score := seeBonus[m.Capture().Figure()]
	if m.MoveType() == Promotion {
		score -= seeBonus[Pawn]
		score += seeBonus[m.Target().Figure()]
	}
	return score
}

// seeSign return true if see(m) < 0.
func seeSign(pos *Position, m Move) bool {
	if m.Piece().Figure() <= m.Capture().Figure() {
		// Even if m.Piece() is captured, we are still positive.
		return false
	}
	return see(pos, m) < 0
}

// see returns the static exchange evaluation for m, where is
// the last move executed.
//
// https://chessprogramming.wikispaces.com/Static+Exchange+Evaluation
// https://chessprogramming.wikispaces.com/SEE+-+The+Swap+Algorithm
//
// The implementation here is optimized for the common case when there
// isn't any capture following the move. The score returned is based
// on some fixed values for figures, different from the ones
// defined in material.go.
func see(pos *Position, m Move) int32 {
	us := pos.SideToMove
	sq := m.To()
	bb := sq.Bitboard()
	target := m.Target() // piece in position
	bb27 := bb &^ (BbRank1 | BbRank8)
	bb18 := bb & (BbRank1 | BbRank8)

	// Occupancy tables as if moves are executed.
	var occ [ColorArraySize]Bitboard
	occ[White] = pos.ByColor[White]
	occ[Black] = pos.ByColor[Black]
	all := occ[White] | occ[Black]

	// Adjust score for move.
	score := seeScore(m)
	tmp := [16]int32{score}
	gain := tmp[:1]

	for score >= 0 {
		// Try every figure in order of value.
		var fig Figure                  // attacking figure
		var att Bitboard                // attackers
		var pawn, bishop, rook Bitboard // mobilies for our figures

		ours := occ[us]
		mt := Normal

		// Pawn attacks.
		pawn = Backward(us, West(bb27)|East(bb27))
		if att = pawn & ours & pos.ByFigure[Pawn]; att != 0 {
			fig = Pawn
			goto makeMove
		}

		if att = bbKnightAttack[sq] & ours & pos.ByFigure[Knight]; att != 0 {
			fig = Knight
			goto makeMove
		}

		if bbSuperAttack[sq]&ours == 0 {
			// No other figure can attack sq so we give up early.
			break
		}

		bishop = BishopMobility(sq, all)
		if att = bishop & ours & pos.ByFigure[Bishop]; att != 0 {
			fig = Bishop
			goto makeMove
		}

		rook = RookMobility(sq, all)
		if att = rook & ours & pos.ByFigure[Rook]; att != 0 {
			fig = Rook
			goto makeMove
		}

		// Pawn promotions are considered queens minus the pawn.
		pawn = Backward(us, West(bb18)|East(bb18))
		if att = pawn & ours & pos.ByFigure[Pawn]; att != 0 {
			fig, mt = Queen, Promotion
			goto makeMove
		}

		if att = (rook | bishop) & ours & pos.ByFigure[Queen]; att != 0 {
			fig = Queen
			goto makeMove
		}

		if att = bbKingAttack[sq] & ours & pos.ByFigure[King]; att != 0 {
			fig = King
			goto makeMove
		}

		// No attack found.
		break

	makeMove:
		// Make a new pseudo-legal move of the smallest attacker.
		from := att.LSB()
		attacker := ColorFigure(us, fig)
		m := MakeMove(mt, from.AsSquare(), sq, target, attacker)
		target = attacker // attacker becomes the new target

		// Update score.
		score = seeScore(m) - score
		gain = append(gain, score)

		// Update occupancy tables for executing the move.
		occ[us] = occ[us] &^ from
		all = all &^ from

		// Switch sides.
		us = us.Opposite()
	}

	for i := len(gain) - 2; i >= 0; i-- {
		if -gain[i+1] < gain[i] {
			gain[i] = -gain[i+1]
		}
	}
	return gain[0]
}
