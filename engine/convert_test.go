package engine

import (
	"testing"
)

func TestPositionFromFENAndBack(t *testing.T) {
	for _, d := range testFENs {
		pos, err := PositionFromFEN(d)
		if err != nil {
			t.Errorf("%s failed with %v", d, err)
		} else if fen := pos.String(); d != fen {
			t.Errorf("expected %s, got %s", d, fen)
		}
	}
}

func BenchmarkPositionFromFEN(b *testing.B) {
	for i := 0; i < b.N; i++ {
		for _, d := range testFENs {
			PositionFromFEN(d)
		}
	}
}
