// Package javahash reproduces java.util.HashMap iteration order for keys with
// a known Java hashCode. CodinGame referees sometimes pick an element from a
// HashSet by index after `random.nextInt(size)`, which ties map generation to
// bucket ordering; matching the RNG stream alone is not enough for seed
// parity. This package is the missing half: a set whose Values() come back in
// exactly the order Java would iterate them.
//
// Reproduced from OpenJDK 17 HashMap: the `h ^ (h >>> 16)` spread, the lazily
// allocated 16-slot table, the 0.75 load factor, tail-append chaining, and
// the resize split of each bucket into a low and a high list preserving
// relative order. Treeification is not reproduced; see Add.
package javahash

// Hashable is any comparable key that can report its Java hashCode().
type Hashable interface {
	comparable
	JavaHash() int32
}

const (
	defaultCapacity    = 16
	loadFactor         = 0.75
	treeifyThreshold   = 8
	minTreeifyCapacity = 64
)

type node[K Hashable] struct {
	hash int32
	key  K
	next *node[K]
}

// Set is an insertion-ordered set with java.util.HashSet iteration order.
// The zero value is not usable; call NewSet.
type Set[K Hashable] struct {
	table     []*node[K]
	size      int
	threshold int
}

func NewSet[K Hashable]() *Set[K] {
	return &Set[K]{}
}

// spread mirrors HashMap.hash(): xor the high half of hashCode into the low
// half so small tables still see the high bits.
func spread(h int32) int32 {
	return h ^ int32(uint32(h)>>16)
}

func (s *Set[K]) Len() int { return s.size }

func (s *Set[K]) Contains(key K) bool {
	if len(s.table) == 0 {
		return false
	}
	h := spread(key.JavaHash())
	for e := s.table[int(h)&(len(s.table)-1)]; e != nil; e = e.next {
		if e.hash == h && e.key == key {
			return true
		}
	}
	return false
}

// Add inserts key and reports whether the set changed. It mirrors
// HashMap.putVal, including the resize that Java performs instead of
// treeifying when the table is still smaller than 64 slots. A chain that
// would treeify on a table of 64 or more slots panics, because treeified
// buckets iterate in an order this package does not reproduce.
func (s *Set[K]) Add(key K) bool {
	if len(s.table) == 0 {
		s.resize()
	}
	h := spread(key.JavaHash())
	i := int(h) & (len(s.table) - 1)
	if s.table[i] == nil {
		s.table[i] = &node[K]{hash: h, key: key}
	} else {
		p := s.table[i]
		binCount := 0
		for {
			if p.hash == h && p.key == key {
				return false
			}
			if p.next == nil {
				p.next = &node[K]{hash: h, key: key}
				if binCount >= treeifyThreshold-1 {
					s.treeifyBin()
				}
				break
			}
			p = p.next
			binCount++
		}
	}
	s.size++
	if s.size > s.threshold {
		s.resize()
	}
	return true
}

// Values returns the keys in Java iteration order: bucket by bucket, chain
// order within each bucket.
func (s *Set[K]) Values() []K {
	out := make([]K, 0, s.size)
	for _, e := range s.table {
		for ; e != nil; e = e.next {
			out = append(out, e.key)
		}
	}
	return out
}

func (s *Set[K]) treeifyBin() {
	if len(s.table) < minTreeifyCapacity {
		s.resize()
		return
	}
	panic("javahash: bucket reached the treeify threshold; treeified iteration order is not reproduced")
}

// resize mirrors HashMap.resize: allocate the initial table or double it,
// splitting every chain into the nodes that stay at index j and those that
// move to j+oldCap, each list keeping its relative order.
func (s *Set[K]) resize() {
	oldCap := len(s.table)
	newCap := defaultCapacity
	if oldCap > 0 {
		newCap = oldCap * 2
	}
	s.threshold = int(float64(newCap) * loadFactor)
	newTab := make([]*node[K], newCap)
	for j, e := range s.table {
		if e == nil {
			continue
		}
		if e.next == nil {
			newTab[int(e.hash)&(newCap-1)] = e
			continue
		}
		var loHead, loTail, hiHead, hiTail *node[K]
		for ; e != nil; e = e.next {
			if int(e.hash)&oldCap == 0 {
				if loTail == nil {
					loHead = e
				} else {
					loTail.next = e
				}
				loTail = e
			} else {
				if hiTail == nil {
					hiHead = e
				} else {
					hiTail.next = e
				}
				hiTail = e
			}
		}
		if loTail != nil {
			loTail.next = nil
			newTab[j] = loHead
		}
		if hiTail != nil {
			hiTail.next = nil
			newTab[j+oldCap] = hiHead
		}
	}
	s.table = newTab
}
