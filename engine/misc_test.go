package engine

import "testing"

func TestLogN(t *testing.T) {
	for e := uint(0); e < 64; e++ {
		n := uint64(1) << e
		actual := logN(n)
		if actual != e {
			t.Errorf("expected logN(%d) == %d, got %d", n, e, actual)
		}
	}
}
