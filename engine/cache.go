// cache.go implements a simple generic cache.
// We distinguish keys based on their hash assuming
// that collisions are very rare and not that important.

package engine

const (
	murmurMultiplier = uint64(0xc6a4a7935bd1e995)
	murmurShift      = uint(51)
)

var (
	murmurSeed = [ColorArraySize]uint64{
		0x77a166129ab66e91,
		0x4f4863d5038ea3a3,
		0xe14ec7e648a4068b,
	}
)

// murmuxMix function mixes two integers k&h.
//
// murmurMix is based on MurmurHash2 https://sites.google.com/site/murmurhash/
// which is on public domain.
//
// A hash can be constructed like this:
//
//     hash := murmurSeed[us]
//     hash = murmurMix(hash, n1)
//     hash = murmurMix(hash, n2)
//     hash = murmurMix(hash, n3)
func murmurMix(k, h uint64) uint64 {
	h ^= k
	h *= murmurMultiplier
	h ^= h >> murmurShift
	return h
}

// cacheEntry is a cache entry.
type cacheEntry struct {
	lock uint64
	eval Eval
}

// cache implements a fixed size cache.
type cache struct {
	table []cacheEntry
	hash  func(*Position, Color) uint64
	comp  func(*Position, Color) Eval
}

// newCache creates a new cache of size 1<<bits.
func newCache(bits uint, hash func(*Position, Color) uint64, comp func(*Position, Color) Eval) *cache {
	return &cache{
		table: make([]cacheEntry, 1<<bits),
		hash:  hash,
		comp:  comp,
	}
}

// put puts a new entry in the cache.
func (c *cache) put(lock uint64, eval Eval) {
	indx := lock & uint64(len(c.table)-1)
	c.table[indx] = cacheEntry{lock: lock, eval: eval}
}

// get gets an entry from the cache.
func (c *cache) get(lock uint64) (Eval, bool) {
	indx := lock & uint64(len(c.table)-1)
	return c.table[indx].eval, c.table[indx].lock == lock
}

// load evaluates position, using the cache if possible.
func (c *cache) load(pos *Position, us Color) Eval {
	if disableCache {
		return c.comp(pos, us)
	}
	h := c.hash(pos, us)
	if e, ok := c.get(h); ok {
		return e
	}
	e := c.comp(pos, us)
	c.put(h, e)
	return e
}
