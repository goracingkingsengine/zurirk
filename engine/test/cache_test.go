package engine

import "testing"

const (
	c1 = uint64(3080512559332270987)
	c2 = uint64(1670079002898303149)
)

func TestMurmurMixSwap(t *testing.T) {
	h1 := murmurSeed[NoFigure]
	h1 = murmurMix(h1, c1)
	h1 = murmurMix(h1, c2)

	h2 := murmurSeed[NoFigure]
	h2 = murmurMix(h2, c2)
	h2 = murmurMix(h2, c1)

	if h1 == h2 {
		t.Errorf("murmurMix(c1, c2) == murmurMix(c2, c1) (%d, %d), wanted different", h1, h2)
	}
}

func TestCachePutGet(t *testing.T) {
	h1 := murmurSeed[NoFigure]
	h1 = murmurMix(h1, c1)
	h1 = murmurMix(h1, c2)

	e := Eval{1, 2}
	c := newCache(6, nil, nil)
	c.put(h1, e)
	if got, ok := c.get(h1); !ok {
		t.Errorf("entry not in the cache, expecting a git")
	} else if e != got {
		t.Errorf("got get(%d) == %v, wanted %v", h1, got, e)
	}

	h2 := murmurSeed[NoFigure]
	h2 = murmurMix(h2, c2)
	h2 = murmurMix(h2, c1)
	if _, ok := c.get(h2); ok {
		t.Errorf("entry in the cache, expecting a miss")
	}
}
