package sha1prng

import (
	"math"
	"strconv"
	"strings"
	"testing"
)

// Reference values captured from Java 17 SecureRandom.getInstance("SHA1PRNG")
// via .tmp/JavaHashProbe.java in "rng" mode. Floats are recorded as Java's
// Float.toHexString output so no decimal rounding is involved.

func parseHexFloats(t *testing.T, s string) []float32 {
	t.Helper()
	var out []float32
	for _, f := range strings.Fields(s) {
		v, err := strconv.ParseFloat(f, 32)
		if err != nil {
			t.Fatalf("bad hex float %q: %v", f, err)
		}
		out = append(out, float32(v))
	}
	return out
}

func TestNextFloatMatchesJava(t *testing.T) {
	cases := []struct {
		seed int64
		want string
	}{
		{-8937286792422418000, "0x1.832cecp-1 0x1.14d77p-2 0x1.77d878p-2 0x1.4fb0aep-1 0x1.bbc13p-2 0x1.052f68p-3 0x1.d48c88p-3 0x1.64a45cp-2 0x1.5bc164p-2 0x1.e0f174p-1"},
		{468706172918629800, "0x1.3ae05p-3 0x1.45b52cp-2 0x1.a5b36cp-2 0x1.b64d9p-1 0x1.db5b7p-3 0x1.5bf25cp-2 0x1.1fc468p-3 0x1.7ec9d8p-3 0x1.418f4cp-2 0x1.304fc6p-1"},
	}
	for _, tc := range cases {
		t.Run(strconv.FormatInt(tc.seed, 10), func(t *testing.T) {
			r := New(tc.seed)
			for i, want := range parseHexFloats(t, tc.want) {
				if got := r.NextFloat(); got != want {
					t.Errorf("NextFloat()[%d] = %v (%#x), want %v (%#x)", i, got, math.Float32bits(got), want, math.Float32bits(want))
				}
			}
		})
	}
}

func TestNextFloatBoundMatchesJava(t *testing.T) {
	cases := []struct {
		seed int64
		want string
	}{
		{-8937286792422418000, "0x1.f1120ap8 0x1.567072p9 0x1.5b7d08p9 0x1.f90d7ap6 0x1.1c631p9 0x1.e60ce8p8 0x1.54b35ap7 0x1.aec238p8 0x1.14e9bp4 0x1.02f38cp8"},
		{468706172918629800, "0x1.09b28ap9 0x1.bb4148p7 0x1.341ba2p7 0x1.656e8p6 0x1.aa3ae2p8 0x1.006ee8p7 0x1.4ba7e2p9 0x1.189234p8 0x1.0135f8p8 0x1.256f08p7"},
	}
	for _, tc := range cases {
		t.Run(strconv.FormatInt(tc.seed, 10), func(t *testing.T) {
			r := New(tc.seed)
			for i := 0; i < 10; i++ {
				r.NextFloat()
			}
			for i, want := range parseHexFloats(t, tc.want) {
				if got := r.NextFloatBound(966.0); got != want {
					t.Errorf("NextFloatBound(966)[%d] = %v (%#x), want %v (%#x)", i, got, math.Float32bits(got), want, math.Float32bits(want))
				}
			}
		})
	}
}

func TestNextBooleanMatchesJava(t *testing.T) {
	cases := []struct {
		seed int64
		want string
	}{
		{-8937286792422418000, "false false true true true true true false false false false false false true false false false false false true"},
		{468706172918629800, "false true false false false false false true true true false true false true true true false true true true"},
	}
	for _, tc := range cases {
		t.Run(strconv.FormatInt(tc.seed, 10), func(t *testing.T) {
			r := New(tc.seed)
			for i := 0; i < 10; i++ {
				r.NextFloat()
				r.NextFloatBound(966.0)
			}
			var got []string
			for i := 0; i < 20; i++ {
				got = append(got, strconv.FormatBool(r.NextBoolean()))
			}
			if g := strings.Join(got, " "); g != tc.want {
				t.Errorf("NextBoolean sequence\nwant: %s\ngot:  %s", tc.want, g)
			}
		})
	}
}

func TestShuffleMatchesJavaCollectionsShuffle(t *testing.T) {
	cases := []struct {
		seed                     int64
		want10, wantLinked7, want2 string
	}{
		{-8937286792422418000, "3 7 8 6 9 2 0 4 1 5", "6 4 0 2 5 3 1", "0 1"},
		{468706172918629800, "1 6 5 9 3 0 2 7 8 4", "0 3 4 6 5 2 1", "0 1"},
	}
	iota := func(n int) []int {
		s := make([]int, n)
		for i := range s {
			s[i] = i
		}
		return s
	}
	join := func(s []int) string {
		parts := make([]string, len(s))
		for i, v := range s {
			parts[i] = strconv.Itoa(v)
		}
		return strings.Join(parts, " ")
	}
	for _, tc := range cases {
		t.Run(strconv.FormatInt(tc.seed, 10), func(t *testing.T) {
			r := New(tc.seed)
			for i := 0; i < 10; i++ {
				r.NextFloat()
				r.NextFloatBound(966.0)
			}
			for i := 0; i < 20; i++ {
				r.NextBoolean()
			}
			for _, step := range []struct {
				n    int
				want string
			}{{10, tc.want10}, {7, tc.wantLinked7}, {2, tc.want2}} {
				s := iota(step.n)
				Shuffle(r, s)
				if got := join(s); got != step.want {
					t.Errorf("Shuffle(%d) = %s, want %s", step.n, got, step.want)
				}
			}
		})
	}
}
