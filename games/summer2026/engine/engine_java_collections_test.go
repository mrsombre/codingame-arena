package engine

import (
	"math"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Every expectation below was captured from OpenJDK 17 with a standalone
// probe that runs the same inputs through java.util.Arrays.sort and
// java.util.PriorityQueue.

func TestJavaListSortMatchesArraysSortOnSmallArrays(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"3,1,2", "1,2,3"},
		{"5,4,3,2,1", "1,2,3,4,5"},
		{"1,1,2,2", "1,1,2,2"},
		{"2,1,4,3,6,5", "1,2,3,4,5,6"},
		{"9", "9"},
		{"1,2,3,4,5,6,7,8", "1,2,3,4,5,6,7,8"},
	}
	for _, tc := range cases {
		t.Run(tc.in, func(t *testing.T) {
			a := strings.Split(tc.in, ",")
			javaListSort(a, func(x, y string) int {
				return compareInt(atoi(t, x), atoi(t, y))
			})
			assert.Equal(t, tc.want, strings.Join(a, ","))
		})
	}
}

// A comparator that reports every pair as greater is not a valid ordering, but
// Arrays.sort still has one defined answer for it: the initial ascending run
// covers the whole array, so nothing moves. TrainBFS depends on this.
func TestJavaListSortLeavesAnAlwaysGreaterComparatorUntouched(t *testing.T) {
	a := []string{"d", "c", "b", "a"}
	javaListSort(a, func(_, _ string) int { return 4 })

	assert.Equal(t, []string{"d", "c", "b", "a"}, a)
}

func TestJavaListSortPanicsOnceTimSortWouldMerge(t *testing.T) {
	a := make([]int, javaTimSortMinMerge)
	assert.Panics(t, func() { javaListSort(a, compareInt) })
}

// The AbstractAStar comparator: f ascending, then the tie-break value. The
// order among entries equal on both is a heap artefact, which is exactly why
// the poll sequence is pinned against Java rather than merely checked for
// being non-decreasing.
func TestJavaPriorityQueuePollOrderMatchesJava(t *testing.T) {
	type node struct {
		name string
		f    float64
		tie  int
	}
	q := newJavaPriorityQueue(func(a, b node) int {
		if c := javaCompareDouble(a.f, b.f); c != 0 {
			return c
		}
		return compareInt(a.tie, b.tie)
	})

	nodes := []node{
		{"a", 5.0, 0}, {"b", 3.0, 2}, {"c", 5.0, 1},
		{"d", 3.0, 0}, {"e", 4.0, 3}, {"f", 3.0, 2},
		{"g", 5.0, 0}, {"h", 1.0, 9}, {"i", 4.0, 3},
	}

	var out []string
	for _, n := range nodes {
		q.Add(n)
		if len(out) < 3 && q.Len()%4 == 0 {
			polled, ok := q.Poll()
			assert.True(t, ok)
			out = append(out, "poll:"+polled.name)
		}
	}
	for q.Len() > 0 {
		polled, _ := q.Poll()
		out = append(out, polled.name)
	}

	assert.Equal(t,
		[]string{"poll:d", "poll:b", "poll:f", "h", "e", "i", "g", "a", "c"},
		out,
	)
}

func TestJavaPriorityQueuePollOnEmptyReportsNotOk(t *testing.T) {
	q := newJavaPriorityQueue(compareInt)
	_, ok := q.Poll()

	assert.False(t, ok)
}

func TestJavaCompareDoubleOrdersSignedZeroAndNaNLikeJava(t *testing.T) {
	nan := math.NaN()

	assert.Equal(t, -1, javaCompareDouble(math.Copysign(0, -1), 0))
	assert.Equal(t, 1, javaCompareDouble(0, math.Copysign(0, -1)))
	assert.Equal(t, 0, javaCompareDouble(1.5, 1.5))
	assert.Equal(t, 1, javaCompareDouble(nan, math.Inf(1)))
	assert.Equal(t, 0, javaCompareDouble(nan, nan))
}

func atoi(t *testing.T, s string) int {
	t.Helper()
	v, err := strconv.Atoi(s)
	if err != nil {
		t.Fatalf("not an integer: %q", s)
	}
	return v
}
