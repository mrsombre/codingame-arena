// Package sha1prng reproduces Java's SecureRandom.getInstance("SHA1PRNG")
// with an explicit setSeed — the RNG CodinGame's Java engine uses for map
// generation. Seeds are bit-compatible with the referee so Go reproduces the
// same grids.
package sha1prng

import (
	"crypto/sha1"
	"encoding/binary"
)

const digestSize = 20

type Random struct {
	state     []byte
	remainder []byte
	remCount  int
}

func New(seed int64) *Random {
	r := &Random{}
	r.setSeed(longToLittleEndian(seed))
	return r
}

func (r *Random) NextDouble() float64 {
	hi := int64(r.next(26))
	lo := int64(r.next(27))
	return float64(hi<<27+lo) / float64(int64(1)<<53)
}

func (r *Random) NextInt(bound int) int {
	if bound <= 0 {
		panic("bound must be positive")
	}
	b := int32(bound)
	m := b - 1
	bits := r.next(31)
	if b&m == 0 {
		return int(int64(b) * int64(bits) >> 31)
	}
	val := bits % b
	for bits-val+m < 0 {
		bits = r.next(31)
		val = bits % b
	}
	return int(val)
}

func (r *Random) NextIntRange(origin, bound int) int {
	if origin >= bound {
		panic("origin must be less than bound")
	}
	return origin + r.NextInt(bound-origin)
}

func (r *Random) setSeed(seed []byte) {
	if len(r.state) != 0 {
		buf := make([]byte, 0, len(r.state)+len(seed))
		buf = append(buf, r.state...)
		buf = append(buf, seed...)
		sum := sha1.Sum(buf)
		r.state = sum[:]
	} else {
		sum := sha1.Sum(seed)
		r.state = sum[:]
	}
	r.remCount = 0
}

func (r *Random) next(numBits int) int32 {
	numBytes := (numBits + 7) / 8
	buf := make([]byte, numBytes)
	r.nextBytes(buf)

	next := 0
	for _, b := range buf {
		next = (next << 8) + int(b)
	}
	return int32(uint32(next) >> uint(numBytes*8-numBits))
}

func (r *Random) nextBytes(result []byte) {
	index := 0
	output := r.remainder

	if len(r.state) == 0 {
		r.setSeed(make([]byte, 8))
	}

	if r.remCount > 0 {
		todo := min(len(result)-index, digestSize-r.remCount)
		rpos := r.remCount
		for i := 0; i < todo; i++ {
			result[index+i] = output[rpos]
			output[rpos] = 0
			rpos++
		}
		r.remCount += todo
		index += todo
	}

	for index < len(result) {
		sum := sha1.Sum(r.state)
		output = make([]byte, digestSize)
		copy(output, sum[:])
		updateState(r.state, output)

		todo := min(len(result)-index, digestSize)
		for i := 0; i < todo; i++ {
			result[index] = output[i]
			output[i] = 0
			index++
		}
		r.remCount += todo
	}

	r.remainder = output
	r.remCount %= digestSize
}

func updateState(state, output []byte) {
	last := 1
	changed := false

	for i := range state {
		v := int(int8(state[i])) + int(int8(output[i])) + last
		t := byte(v)
		if state[i] != t {
			changed = true
		}
		state[i] = t
		last = v >> 8
	}

	if !changed {
		state[0]++
	}
}

func longToLittleEndian(seed int64) []byte {
	buf := make([]byte, 8)
	binary.LittleEndian.PutUint64(buf, uint64(seed))
	return buf
}
