package javahash

import (
	"fmt"
	"strconv"
	"strings"
	"testing"
)

// Reference orders captured from Java 17 java.util.HashSet via
// .tmp/JavaHashProbe.java in "hash" mode. The probe's Coord hashes as
// 31*(31+x)+y, the same formula the Back Track King engine uses.

type coord struct{ x, y int32 }

func (c coord) JavaHash() int32 { return 31*(31+c.x) + c.y }

func (c coord) String() string { return fmt.Sprintf("%d,%d", c.x, c.y) }

func parseCoords(t *testing.T, s string) []coord {
	t.Helper()
	var out []coord
	for _, f := range strings.Fields(s) {
		xy := strings.Split(f, ",")
		x, err1 := strconv.Atoi(xy[0])
		y, err2 := strconv.Atoi(xy[1])
		if err1 != nil || err2 != nil {
			t.Fatalf("bad coord %q", f)
		}
		out = append(out, coord{int32(x), int32(y)})
	}
	return out
}

func join(cs []coord) string {
	parts := make([]string, len(cs))
	for i, c := range cs {
		parts[i] = c.String()
	}
	return strings.Join(parts, " ")
}

func TestSmallSetIteratesInJavaBucketOrder(t *testing.T) {
	s := NewSet[coord]()
	for _, c := range parseCoords(t, "5,5 0,0 3,1 1,3 41,22") {
		s.Add(c)
	}
	const want = "5,5 0,0 1,3 41,22 3,1"
	if got := join(s.Values()); got != want {
		t.Errorf("Values() = %s, want %s", got, want)
	}
}

func TestDuplicateAddIsNoOp(t *testing.T) {
	s := NewSet[coord]()
	for _, c := range parseCoords(t, "5,5 0,0 3,1 1,3 41,22") {
		s.Add(c)
	}
	if s.Add(coord{0, 0}) {
		t.Error("Add of an existing key reported a change")
	}
	const want = "5,5 0,0 1,3 41,22 3,1"
	if got := join(s.Values()); got != want || s.Len() != 5 {
		t.Errorf("Values() = %s (len %d), want %s (len 5)", got, s.Len(), want)
	}
}

// Row-major scan of a 12x12 block, checked one insert before and one after
// each resize boundary (13th, 25th, 49th, 97th insert) and at the end.
func TestScanOrderAcrossResizeBoundaries(t *testing.T) {
	want := map[int]string{
		12:  "1,0 0,0 11,0 10,0 9,0 8,0 7,0 6,0 5,0 4,0 3,0 2,0",
		13:  "1,0 0,0 0,1 11,0 10,0 9,0 8,0 7,0 6,0 5,0 4,0 3,0 2,0",
		24:  "1,0 2,1 0,0 1,1 0,1 11,0 10,0 11,1 9,0 10,1 8,0 9,1 7,0 8,1 6,0 7,1 5,0 6,1 4,0 5,1 3,0 4,1 2,0 3,1",
		25:  "2,1 0,0 0,1 0,2 11,0 11,1 9,0 9,1 7,0 7,1 5,0 5,1 3,0 3,1 1,0 1,1 10,0 10,1 8,0 8,1 6,0 6,1 4,0 4,1 2,0",
		48:  "2,1 4,3 0,0 2,2 0,1 2,3 0,2 0,3 11,0 11,1 9,0 11,2 9,1 11,3 7,0 9,2 7,1 9,3 5,0 7,2 5,1 7,3 3,0 5,2 3,1 5,3 1,0 3,2 1,1 3,3 1,2 1,3 10,0 10,1 8,0 10,2 8,1 10,3 6,0 8,2 6,1 8,3 4,0 6,2 4,1 6,3 2,0 4,2",
		49:  "2,1 2,2 2,3 11,0 11,1 11,2 11,3 7,0 7,1 7,2 7,3 3,0 3,1 3,2 3,3 8,0 8,1 8,2 8,3 4,0 4,1 4,2 4,3 0,0 0,1 0,2 0,3 0,4 9,0 9,1 9,2 9,3 5,0 5,1 5,2 5,3 1,0 1,1 1,2 1,3 10,0 10,1 10,2 10,3 6,0 6,1 6,2 6,3 2,0",
		96:  "2,1 6,5 2,2 6,6 2,3 6,7 2,4 2,5 2,6 2,7 11,0 11,1 11,2 11,3 7,0 11,4 7,1 11,5 7,2 11,6 7,3 11,7 3,0 7,4 3,1 7,5 3,2 7,6 3,3 7,7 3,4 3,5 3,6 3,7 8,0 8,1 8,2 8,3 4,0 8,4 4,1 8,5 4,2 8,6 4,3 8,7 0,0 4,4 0,1 4,5 0,2 4,6 0,3 4,7 0,4 0,5 0,6 0,7 9,0 9,1 9,2 9,3 5,0 9,4 5,1 9,5 5,2 9,6 5,3 9,7 1,0 5,4 1,1 5,5 1,2 5,6 1,3 5,7 1,4 1,5 1,6 1,7 10,0 10,1 10,2 10,3 6,0 10,4 6,1 10,5 6,2 10,6 6,3 10,7 2,0 6,4",
		97:  "2,1 2,2 2,3 2,4 2,5 2,6 2,7 11,0 11,1 11,2 11,3 11,4 11,5 11,6 11,7 3,0 3,1 3,2 3,3 3,4 3,5 3,6 3,7 4,0 4,1 4,2 4,3 4,4 4,5 4,6 4,7 5,0 5,1 5,2 5,3 5,4 5,5 5,6 5,7 6,0 6,1 6,2 6,3 6,4 6,5 6,6 6,7 7,0 7,1 7,2 7,3 7,4 7,5 7,6 7,7 8,0 8,1 8,2 8,3 8,4 8,5 8,6 8,7 0,0 0,1 0,2 0,3 0,4 0,5 0,6 0,7 0,8 9,0 9,1 9,2 9,3 9,4 9,5 9,6 9,7 1,0 1,1 1,2 1,3 1,4 1,5 1,6 1,7 10,0 10,1 10,2 10,3 10,4 10,5 10,6 10,7 2,0",
		144: "2,1 10,9 2,2 10,10 2,3 10,11 2,4 2,5 2,6 2,7 2,8 2,9 2,10 2,11 11,0 11,1 11,2 11,3 11,4 11,5 11,6 11,7 3,0 11,8 3,1 11,9 3,2 11,10 3,3 11,11 3,4 3,5 3,6 3,7 3,8 3,9 3,10 3,11 4,0 4,1 4,2 4,3 4,4 4,5 4,6 4,7 4,8 4,9 4,10 4,11 5,0 5,1 5,2 5,3 5,4 5,5 5,6 5,7 5,8 5,9 5,10 5,11 6,0 6,1 6,2 6,3 6,4 6,5 6,6 6,7 6,8 6,9 6,10 6,11 7,0 7,1 7,2 7,3 7,4 7,5 7,6 7,7 7,8 7,9 7,10 7,11 8,0 8,1 8,2 8,3 8,4 8,5 8,6 8,7 0,0 8,8 0,1 8,9 0,2 8,10 0,3 8,11 0,4 0,5 0,6 0,7 0,8 0,9 0,10 0,11 9,0 9,1 9,2 9,3 9,4 9,5 9,6 9,7 1,0 9,8 1,1 9,9 1,2 9,10 1,3 9,11 1,4 1,5 1,6 1,7 1,8 1,9 1,10 1,11 10,0 10,1 10,2 10,3 10,4 10,5 10,6 10,7 2,0 10,8",
	}
	s := NewSet[coord]()
	n := 0
	for y := int32(0); y < 12; y++ {
		for x := int32(0); x < 12; x++ {
			s.Add(coord{x, y})
			n++
			if w, ok := want[n]; ok {
				if got := join(s.Values()); got != w {
					t.Errorf("after %d inserts:\n got %s\nwant %s", n, got, w)
				}
			}
		}
	}
}

// The neighbour set of a random-walk blob on a 42x23 grid, the shape
// GridMaker.getAvailableNeighbours builds during region and mountain growth.
func TestBlobNeighbourOrderMatchesJava(t *testing.T) {
	const w, h = 42, 23
	blob := parseCoords(t, "20,11 20,10 20,9 19,11 20,8 21,9 21,11 21,8 18,11 19,10 22,8 20,7 17,11 17,10 18,12 18,10 23,8 18,9 18,13 18,8 21,7 19,7 19,8 17,9 21,10 19,9 24,8")
	in := map[coord]bool{}
	for _, c := range blob {
		in[c] = true
	}
	s := NewSet[coord]()
	for _, c := range blob {
		for _, d := range []coord{{0, -1}, {1, 0}, {0, 1}, {-1, 0}} {
			n := coord{c.x + d.x, c.y + d.y}
			if n.x < 0 || n.y < 0 || n.x >= w || n.y >= h || in[n] {
				continue
			}
			s.Add(n)
		}
	}
	const want = "24,7 25,8 23,7 22,7 21,6 24,9 20,6 23,9 22,9 19,6 22,10 22,11 18,7 21,12 17,8 20,12 19,12 16,9 16,10 19,13 17,12 16,11 18,14 17,13"
	if got := join(s.Values()); got != want {
		t.Errorf("Values() = %s\nwant %s", got, want)
	}
}

func TestNegativeCoordinatesOrderMatchesJava(t *testing.T) {
	s := NewSet[coord]()
	for i := int32(-6); i <= 6; i++ {
		s.Add(coord{i, -i * 3})
	}
	const want = "0,0 -1,3 -2,6 6,-18 -3,9 5,-15 -4,12 4,-12 -5,15 3,-9 -6,18 2,-6 1,-3"
	if got := join(s.Values()); got != want {
		t.Errorf("Values() = %s\nwant %s", got, want)
	}
}

type fixedHash int32

func (f fixedHash) JavaHash() int32 { return 0 }

func TestContainsReportsMembership(t *testing.T) {
	s := NewSet[coord]()
	s.Add(coord{3, 4})
	if !s.Contains(coord{3, 4}) || s.Contains(coord{4, 3}) {
		t.Error("Contains disagrees with Add")
	}
}

// Nine keys sharing one bucket: while the table is under 64 slots Java
// resizes instead of treeifying, and the ninth collision keeps the table
// growing until it reaches 64. The collision after that would treeify, and
// this package refuses rather than diverge silently.
func TestTreeifyThresholdPanics(t *testing.T) {
	s := NewSet[fixedHash]()
	for i := 0; i < treeifyThreshold; i++ {
		s.Add(fixedHash(i))
	}
	// Ninth insert on a 16-slot table doubles to 32, tenth doubles to 64,
	// eleventh would treeify and must panic mid-Add, leaving added at 2.
	added := 0
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic at the treeify threshold")
		}
		if added != 2 {
			t.Errorf("panicked during insert %d after the threshold, want 3", added+1)
		}
	}()
	for i := treeifyThreshold; ; i++ {
		s.Add(fixedHash(i))
		added++
	}
}

// Mountain growth collects neighbours of a blob that is only 2..8 cells,
// so the set never leaves the initial 16-slot table.
func TestMountainNeighbourOrderMatchesJava(t *testing.T) {
	const w, h = 42, 23
	cases := []struct{ blob, want string }{
		{"9,5 10,5", "10,4 11,5 9,4 10,6 9,6 8,5"},
		{"10,6 10,7 9,6", "10,5 11,6 11,7 9,5 10,8 9,7 8,6"},
		{"11,7 10,7 9,7 11,6", "9,8 8,7 11,5 12,6 12,7 10,6 11,8 9,6 10,8"},
		{"12,8 12,7 12,9 11,8 11,7", "12,6 13,7 13,8 11,6 13,9 10,7 12,10 11,9 10,8"},
		{"13,9 14,9 14,8 14,7 14,10 14,6", "14,5 15,6 15,7 15,8 13,6 15,9 13,7 13,8 15,10 13,10 12,9 14,11"},
		{"14,10 14,11 15,10 14,12 15,11 13,10 14,13", "13,12 14,14 13,13 15,9 16,10 14,9 16,11 13,9 15,12 13,11 12,10 15,13"},
		{"15,11 14,11 14,10 15,10 16,10 15,9 16,11 16,12", "16,9 17,10 15,8 17,11 14,9 17,12 15,12 13,10 16,13 14,12 13,11"},
	}
	for _, tc := range cases {
		blob := parseCoords(t, tc.blob)
		t.Run(strconv.Itoa(len(blob)), func(t *testing.T) {
			in := map[coord]bool{}
			for _, c := range blob {
				in[c] = true
			}
			s := NewSet[coord]()
			for _, c := range blob {
				for _, d := range []coord{{0, -1}, {1, 0}, {0, 1}, {-1, 0}} {
					n := coord{c.x + d.x, c.y + d.y}
					if n.x < 0 || n.y < 0 || n.x >= w || n.y >= h || in[n] {
						continue
					}
					s.Add(n)
				}
			}
			if got := join(s.Values()); got != tc.want {
				t.Errorf("Values() = %s\nwant %s", got, tc.want)
			}
		})
	}
}
