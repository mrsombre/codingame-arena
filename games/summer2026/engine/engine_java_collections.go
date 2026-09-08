// Package engine
// Ports of two java.util behaviours the pathfinders depend on for parity.
// They have no Java counterpart file in the game source; they reproduce the
// JDK itself.
package engine

import (
	"math"
	"slices"
)

// javaInt is an int usable as a javahash key. Integer.hashCode() returns the
// value itself, so HashSet<Integer> ordering falls straight out of it.
type javaInt int

func (v javaInt) JavaHash() int32 { return int32(v) }

// javaTimSortMinMerge is TimSort.MIN_MERGE. Below it, Arrays.sort runs a
// single countRunAndMakeAscending + binarySort pass and never merges, which
// is the only path this engine takes: the sorted lists are neighbour lists of
// at most four cells.
const javaTimSortMinMerge = 32

// javaListSort reproduces java.util.Arrays.sort(T[], Comparator), which is
// what Stream.sorted(cmp).toList() runs underneath. A comparator that
// violates the Comparable contract still has a well-defined result here, and
// TrainBFS ships one, so the algorithm has to be reproduced rather than
// approximated by any correct sort.
func javaListSort[T any](a []T, cmp func(x, y T) int) {
	if len(a) < 2 {
		return
	}
	if len(a) >= javaTimSortMinMerge {
		panic("javaListSort: input reaches TimSort's merge path, which is not ported")
	}
	javaBinarySort(a, javaCountRunAndMakeAscending(a, cmp), cmp)
}

// javaCountRunAndMakeAscending measures the initial run, reversing it in
// place if it is strictly descending, and returns its length.
func javaCountRunAndMakeAscending[T any](a []T, cmp func(x, y T) int) int {
	hi := len(a)
	runHi := 1
	if runHi == hi {
		return 1
	}

	first := runHi
	runHi++
	if cmp(a[first], a[0]) < 0 {
		for runHi < hi && cmp(a[runHi], a[runHi-1]) < 0 {
			runHi++
		}
		slices.Reverse(a[:runHi])
	} else {
		for runHi < hi && cmp(a[runHi], a[runHi-1]) >= 0 {
			runHi++
		}
	}
	return runHi
}

// javaBinarySort insertion-sorts a[start:] into the already-sorted a[:start].
// The binary search takes the rightmost position among equal elements, which
// is what makes the sort stable.
func javaBinarySort[T any](a []T, start int, cmp func(x, y T) int) {
	if start == 0 {
		start = 1
	}
	for ; start < len(a); start++ {
		pivot := a[start]
		left, right := 0, start
		for left < right {
			mid := int(uint(left+right) >> 1)
			if cmp(pivot, a[mid]) < 0 {
				right = mid
			} else {
				left = mid + 1
			}
		}
		copy(a[left+1:start+1], a[left:start])
		a[left] = pivot
	}
}

// javaPriorityQueue reproduces java.util.PriorityQueue's binary heap.
// Ordering among elements the comparator calls equal is an artefact of the
// sift sequence rather than a guarantee, and AbstractAStar's choice of path
// depends on it, so the heap operations are ported rather than delegated to
// container/heap.
type javaPriorityQueue[T any] struct {
	es  []T
	cmp func(x, y T) int
}

func newJavaPriorityQueue[T any](cmp func(x, y T) int) *javaPriorityQueue[T] {
	return &javaPriorityQueue[T]{cmp: cmp}
}

func (q *javaPriorityQueue[T]) Len() int { return len(q.es) }

// Add is PriorityQueue.offer: place at the end, then sift up.
func (q *javaPriorityQueue[T]) Add(x T) {
	var zero T
	q.es = append(q.es, zero)

	k := len(q.es) - 1
	for k > 0 {
		parent := (k - 1) >> 1
		e := q.es[parent]
		if q.cmp(x, e) >= 0 {
			break
		}
		q.es[k] = e
		k = parent
	}
	q.es[k] = x
}

// Poll is PriorityQueue.poll: take the head, then sift the last element down
// from the root.
func (q *javaPriorityQueue[T]) Poll() (T, bool) {
	if len(q.es) == 0 {
		var zero T
		return zero, false
	}

	result := q.es[0]
	n := len(q.es) - 1
	x := q.es[n]
	q.es = q.es[:n]
	if n > 0 {
		q.siftDown(0, x, n)
	}
	return result, true
}

func (q *javaPriorityQueue[T]) siftDown(k int, x T, n int) {
	half := n >> 1
	for k < half {
		child := k*2 + 1
		c := q.es[child]
		if right := child + 1; right < n && q.cmp(c, q.es[right]) > 0 {
			child = right
			c = q.es[child]
		}
		if q.cmp(x, c) <= 0 {
			break
		}
		q.es[k] = c
		k = child
	}
	q.es[k] = x
}

// javaCompareDouble is Double.compare, which orders -0.0 below 0.0 and NaN
// above everything. Go's `<` and `>` do neither.
func javaCompareDouble(a, b float64) int {
	switch {
	case a < b:
		return -1
	case a > b:
		return 1
	}
	// Equal, or one of them is NaN: fall through to the bit comparison Java
	// uses, which distinguishes -0.0 from 0.0 and sorts NaN last.
	ab, bb := javaDoubleToLongBits(a), javaDoubleToLongBits(b)
	switch {
	case ab == bb:
		return 0
	case ab < bb:
		return -1
	}
	return 1
}

// javaDoubleToLongBits collapses every NaN to the canonical one, as
// Double.doubleToLongBits does.
func javaDoubleToLongBits(v float64) int64 {
	if math.IsNaN(v) {
		return 0x7ff8000000000000
	}
	return int64(math.Float64bits(v))
}
