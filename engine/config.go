///////////////////////////////////////////////////
// NEW
package engine

const TEST                   = true

//var Variant int              = VARIANT_Standard
var Variant int              = VARIANT_Racing_Kings

var START_FENS = [...]string{
		"rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1",
		"8/8/8/8/8/8/krbnNBRK/qrbnNBRQ w - - 0 1",
	}

var RK_PIECE_VALUES = []int32{
	0,
	0,
	300,
	325,
	500,
	700,
}

var KING_ADVANCE_VALUE int32 = 250

var USE_UNICODE_SYMBOLS      = true

const GET_FIRST              = true
const GET_ALL                = false
///////////////////////////////////////////////////