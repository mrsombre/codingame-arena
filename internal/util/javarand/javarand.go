// Package javarand reproduces java.util.Random so Go code can match referee
// RNG sequences seeded from the Java engine.
package javarand

const (
	multiplier = int64(0x5DEECE66D)
	addend     = int64(0xB)
	mask       = int64((1 << 48) - 1)
)

// Random reproduces java.util.Random for referee parity.
type Random struct {
	seed int64
}

func New(seed int64) *Random {
	r := &Random{}
	r.SetSeed(seed)
	return r
}

func (r *Random) SetSeed(seed int64) {
	r.seed = (seed ^ multiplier) & mask
}

func (r *Random) next(bits int) int32 {
	r.seed = (r.seed*multiplier + addend) & mask
	return int32(uint64(r.seed) >> (48 - bits))
}

func (r *Random) NextInt(bound int) int {
	if bound <= 0 {
		panic("bound must be positive")
	}
	if bound&(bound-1) == 0 {
		return int((int64(bound) * int64(r.next(31))) >> 31)
	}
	for {
		bits := int(r.next(31))
		val := bits % bound
		if bits-val+(bound-1) >= 0 {
			return val
		}
	}
}

func (r *Random) NextIntRange(origin, bound int) int {
	if origin >= bound {
		panic("origin must be less than bound")
	}
	return origin + r.NextInt(bound-origin)
}

func (r *Random) NextDouble() float64 {
	return float64((int64(r.next(26))<<27)+int64(r.next(27))) / float64(int64(1)<<53)
}
