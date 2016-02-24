// material.go implements position evaluation.

package engine

import (
	"fmt"
)

const (
	KnownWinScore  int32 = 25000000       // KnownWinScore is strictly greater than all evaluation scores (mate not included).
	KnownLossScore int32 = -KnownWinScore // KnownLossScore is strictly smaller than all evaluation scores (mated not included).
	MateScore      int32 = 30000000       // MateScore - N is mate in N plies.
	MatedScore     int32 = -MateScore     // MatedScore + N is mated in N plies.
	InfinityScore  int32 = 32000000       // InfinityScore is possible score. -InfinityScore is the minimum possible score.
)

var (
	// Weights stores all evaluation parameters under one array for easy handling.
	//
	// Zurichess' evaluation is a very simple neural network with no hidden layers,
	// and one output node y = W_m * x * (1-p) + W_e * x * p where W_m are
	// middle game weights, W_e are endgame weights, x is input, p is phase between
	// middle game and end game, and y is the score.
	// The network has |x| = len(Weights) inputs corresponding to features
	// extracted from the position. These features are symmetrical wrt to colors.
	// The network is trained using the Texel's Tuning Method
	// https://chessprogramming.wikispaces.com/Texel%27s+Tuning+Method.
	Weights = [94]Score{
		{M: 1034, E: 5770}, {M: 5363, E: 9844}, {M: 39652, E: 54153}, {M: 42277, E: 58849}, {M: 57185, E: 103947},
		{M: 140637, E: 189061}, {M: 4799, E: 7873}, {M: 9625, E: 9558}, {M: 950, E: 2925}, {M: 1112, E: 1908},
		{M: 806, E: 1167}, {M: 732, E: 824}, {M: 168, E: 1149}, {M: -879, E: -359}, {M: 3495, E: 7396},
		{M: 2193, E: 7557}, {M: 1909, E: 7559}, {M: 3903, E: 3354}, {M: 3372, E: 6143}, {M: 7773, E: 5680},
		{M: 6441, E: 4512}, {M: 2974, E: 2896}, {M: 3912, E: 6372}, {M: 2689, E: 6273}, {M: 3266, E: 4799},
		{M: 3581, E: 4578}, {M: 4765, E: 6213}, {M: 5273, E: 5606}, {M: 5775, E: 4043}, {M: 3817, E: 4274},
		{M: 3708, E: 8782}, {M: 2391, E: 7627}, {M: 5072, E: 4626}, {M: 6109, E: 3746}, {M: 5668, E: 5198},
		{M: 3913, E: 5131}, {M: 2830, E: 5977}, {M: 2266, E: 5967}, {M: 3516, E: 10438}, {M: 3637, E: 8738},
		{M: 4903, E: 5959}, {M: 5655, E: 3593}, {M: 5049, E: 5557}, {M: 5400, E: 4573}, {M: 3630, E: 7749},
		{M: 2604, E: 7455}, {M: 5493, E: 12869}, {M: 5021, E: 10574}, {M: 8042, E: 6544}, {M: 10390, E: -1256},
		{M: 11098, E: -2344}, {M: 12808, E: 4315}, {M: 8494, E: 9675}, {M: 7990, E: 9444}, {M: 13836, E: 17481},
		{M: 12537, E: 16982}, {M: 11116, E: 10810}, {M: 15238, E: 3620}, {M: 10331, E: 2338}, {M: 6943, E: 8458},
		{M: -835, E: 14771}, {M: -1276, E: 18329}, {M: 7371, E: 5198}, {M: 256, E: 1926}, {M: -53, E: 2965},
		{M: -254, E: 6546}, {M: 2463, E: 10465}, {M: 5507, E: 19296}, {M: 11056, E: 20099}, {M: 8034, E: 5202},
		{M: 4857, E: -3126}, {M: 3065, E: 3432}, {M: -137, E: 6127}, {M: -2620, E: 8577}, {M: -9391, E: 12415},
		{M: -3313, E: 12592}, {M: 7738, E: 8987}, {M: 18783, E: -215}, {M: -526, E: 755}, {M: 6310, E: 5426},
		{M: 5263, E: 7710}, {M: -2482, E: 10646}, {M: 2399, E: 8982}, {M: -607, E: 9555}, {M: 7854, E: 5619},
		{M: 5386, E: 402}, {M: 1228, E: 866}, {M: -991, E: 178}, {M: -1070, E: -1129}, {M: 2183, E: 362},
		{M: -2259, E: -681}, {M: 3854, E: 9184}, {M: 4472, E: 890}, {M: 1300, E: 1524},
	}

	// Named chunks of Weights
	wFigure             [FigureArraySize]Score
	wMobility           [FigureArraySize]Score
	wPawn               [48]Score
	wPassedPawn         [8]Score
	wKingRank           [8]Score
	wKingFile           [8]Score
	wConnectedPawn      Score
	wDoublePawn         Score
	wIsolatedPawn       Score
	wPawnThreat         Score
	wKingShelter        Score
	wBishopPair         Score
	wRookOnOpenFile     Score
	wRookOnHalfOpenFile Score

	// Evaluation caches.
	pawnsAndShelterCache *cache
)

const ()

func init() {
	// Initialize caches.
	pawnsAndShelterCache = newCache(9, hashPawnsAndShelter, evaluatePawnsAndShelter)
	initWeights()

	slice := func(w []Score, out []Score) []Score {
		copy(out, w)
		return w[len(out):]
	}
	entry := func(w []Score, out *Score) []Score {
		*out = w[0]
		return w[1:]
	}

	w := Weights[:]
	w = slice(w, wFigure[:])
	w = slice(w, wMobility[:])
	w = slice(w, wPawn[:])
	w = slice(w, wPassedPawn[:])
	w = slice(w, wKingRank[:])
	w = slice(w, wKingFile[:])
	w = entry(w, &wConnectedPawn)
	w = entry(w, &wDoublePawn)
	w = entry(w, &wIsolatedPawn)
	w = entry(w, &wPawnThreat)
	w = entry(w, &wKingShelter)
	w = entry(w, &wBishopPair)
	w = entry(w, &wRookOnOpenFile)
	w = entry(w, &wRookOnHalfOpenFile)

	if len(w) != 0 {
		panic(fmt.Sprintf("not all weights used, left with %d out of %d", len(w), len(Weights)))
	}
}

func hashPawnsAndShelter(pos *Position, us Color) uint64 {
	h := murmurSeed[us]
	h = murmurMix(h, uint64(pos.ByPiece(us, Pawn)))
	h = murmurMix(h, uint64(pos.ByPiece(us.Opposite(), Pawn)))
	h = murmurMix(h, uint64(pos.ByPiece(us, King)))
	if pos.ByPiece(us.Opposite(), Queen) != 0 {
		// Mixes in something to signal queen's presence.
		h = murmurMix(h, murmurSeed[NoColor])
	}
	return h
}

func evaluatePawnsAndShelter(pos *Position, us Color) Eval {
	var eval Eval
	eval.Merge(evaluatePawns(pos, us))
	eval.Merge(evaluateShelter(pos, us))
	return eval
}

func evaluatePawns(pos *Position, us Color) Eval {
	var eval Eval
	ours := pos.ByPiece(us, Pawn)
	theirs := pos.ByPiece(us.Opposite(), Pawn)

	// From white's POV (P - white pawn, p - black pawn).
	// block   wings
	// ....... .....
	// .....P. .....
	// .....x. .....
	// ..p..x. .....
	// .xxx.x. .xPx.
	// .xxx.x. .....
	// .xxx.x. .....
	// .xxx.x. .....
	block := East(theirs) | theirs | West(theirs)
	wings := East(ours) | West(ours)
	double := Bitboard(0)
	if us == White {
		block = SouthSpan(block) | SouthSpan(ours)
		double = ours & South(ours)
	} else /* if us == Black */ {
		block = NorthSpan(block) | NorthSpan(ours)
		double = ours & North(ours)
	}

	isolated := ours &^ Fill(wings)                           // no pawn on the adjacent files
	connected := ours & (North(wings) | wings | South(wings)) // has neighbouring pawns
	passed := ours &^ block                                   // no pawn env front and no enemy on the adjacent files

	for bb := ours; bb != 0; {
		sq := bb.Pop()
		povSq := sq.POV(us)
		rank := povSq.Rank()

		eval.Add(wFigure[Pawn])
		eval.Add(wPawn[povSq-8])

		if passed.Has(sq) {
			eval.Add(wPassedPawn[rank])
		}
		if connected.Has(sq) {
			eval.Add(wConnectedPawn)
		}
		if double.Has(sq) {
			eval.Add(wDoublePawn)
		}
		if isolated.Has(sq) {
			eval.Add(wIsolatedPawn)
		}
	}

	return eval
}

func evaluateShelter(pos *Position, us Color) Eval {
	var eval Eval
	pawns := pos.ByPiece(us, Pawn)
	king := pos.ByPiece(us, King)

	sq := king.AsSquare().POV(us)
	eval.Add(wKingFile[sq.File()])
	eval.Add(wKingRank[sq.Rank()])

	if pos.ByPiece(us.Opposite(), Queen) != 0 {
		king = ForwardSpan(us, king)
		file := sq.File()
		if file > 0 && West(king)&pawns == 0 {
			eval.Add(wKingShelter)
		}
		if king&pawns == 0 {
			eval.AddN(wKingShelter, 2)
		}
		if file < 7 && East(king)&pawns == 0 {
			eval.Add(wKingShelter)
		}
	}
	return eval
}

// evaluateSide evaluates position for a single side.
func evaluateSide(pos *Position, us Color, eval *Eval) {
	eval.Merge(pawnsAndShelterCache.load(pos, us))
	all := pos.ByColor[White] | pos.ByColor[Black]
	them := us.Opposite()

	// Pawn
	mobility := Forward(us, pos.ByPiece(us, Pawn)) &^ all
	eval.AddN(wMobility[Pawn], mobility.Count())
	mobility = pos.PawnThreats(us) & pos.ByColor[us.Opposite()]
	eval.AddN(wPawnThreat, mobility.Count())

	// Knight
	excl := pos.ByPiece(us, Pawn) | pos.PawnThreats(them)
	for bb := pos.ByPiece(us, Knight); bb > 0; {
		sq := bb.Pop()
		eval.Add(wFigure[Knight])
		mobility := KnightMobility(sq) &^ excl
		eval.AddN(wMobility[Knight], mobility.Count())
	}
	// Bishop
	numBishops := int32(0)
	for bb := pos.ByPiece(us, Bishop); bb > 0; {
		sq := bb.Pop()
		eval.Add(wFigure[Bishop])
		mobility := BishopMobility(sq, all) &^ excl
		eval.AddN(wMobility[Bishop], mobility.Count())
		numBishops++
	}
	eval.AddN(wBishopPair, numBishops/2)

	// Rook
	for bb := pos.ByPiece(us, Rook); bb > 0; {
		sq := bb.Pop()
		eval.Add(wFigure[Rook])
		mobility := RookMobility(sq, all) &^ excl
		eval.AddN(wMobility[Rook], mobility.Count())

		// Evaluate rook on open and semi open files.
		// https://chessprogramming.wikispaces.com/Rook+on+Open+File
		f := FileBb(sq.File())
		if pos.ByPiece(us, Pawn)&f == 0 {
			if pos.ByPiece(them, Pawn)&f == 0 {
				eval.Add(wRookOnOpenFile)
			} else {
				eval.Add(wRookOnHalfOpenFile)
			}
		}
	}
	// Queen
	for bb := pos.ByPiece(us, Queen); bb > 0; {
		sq := bb.Pop()
		eval.Add(wFigure[Queen])
		mobility := QueenMobility(sq, all) &^ excl
		eval.AddN(wMobility[Queen], mobility.Count())
	}

	// King, each side has one.
	{
		sq := pos.ByPiece(us, King).AsSquare()
		mobility := KingMobility(sq) &^ excl
		eval.AddN(wMobility[King], mobility.Count())
	}
}

// evaluatePosition evalues position.
func EvaluatePosition(pos *Position) Eval {
	var eval Eval
	evaluateSide(pos, Black, &eval)
	eval.Neg()
	evaluateSide(pos, White, &eval)
	return eval
}

// Evaluate evaluates position from White's POV.
func Evaluate(pos *Position) int32 {
	eval := EvaluatePosition(pos)
	score := eval.Feed(Phase(pos))
	if KnownLossScore >= score || score >= KnownWinScore {
		panic(fmt.Sprintf("score %d should be between %d and %d",
			score, KnownLossScore, KnownWinScore))
	}
	return score
}

// ScaleToCentiPawn scales the score returned by Evaluate
// such that one pawn ~= 100.
func ScaleToCentiPawn(score int32) int32 {
	return (score + 64) / 128
}

// Phase computes the progress of the game.
// 0 is opening, 256 is late end game.
func Phase(pos *Position) int32 {
	total := int32(4*1 + 4*1 + 4*2 + 2*4)
	curr := total
	curr -= pos.ByFigure[Knight].Count() * 1
	curr -= pos.ByFigure[Bishop].Count() * 1
	curr -= pos.ByFigure[Rook].Count() * 2
	curr -= pos.ByFigure[Queen].Count() * 4
	return (curr*256 + total/2) / total
}
