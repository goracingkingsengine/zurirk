package engine

import (
	"testing"
)

func TestRookAttack(t *testing.T) {
	data := []struct {
		sq  Square
		occ Bitboard
		att Bitboard
	}{
		{SquareB3, 0x0020441002800000, 0x0000000002fd0202},
		{SquareF5, 0x002044d022a00000, 0x0020205020000000},
		{SquareD2, 0x002044d022a00000, 0x080808080808f708},
	}

	for _, d := range data {
		actual := rookMagic[d.sq].Attack(d.occ)
		if actual != d.att {
			t.Errorf("expected %d, got %d", d.att, actual)
		}
	}
}

func TestBishopAttack(t *testing.T) {
	data := []struct {
		sq  Square
		occ Bitboard
		att Bitboard
	}{
		{SquareB3, 0x0020441002800000, 0x20100805000508},
		{SquareF5, 0x002044d022a00000, 0x408500050880402},
		{SquareD2, 0x002044d022a00000, 0x22140014},
	}

	for _, d := range data {
		actual := bishopMagic[d.sq].Attack(d.occ)
		if actual != d.att {
			t.Errorf("expected %d, got %d", d.att, actual)
		}
	}
}
