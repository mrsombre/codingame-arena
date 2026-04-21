package javarand

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Reference values for seed=123 captured from OpenJDK 17's java.util.Random.
// Probe: tmp/JavaRandProbe.java.
var (
	javaSeed123NextInt      = []int32{-1188957731, 1018954901, -39088943, 1295249578, 1087885590}
	javaSeed123NextInt10    = []int{2, 0, 6, 9, 5, 7, 4, 7, 5, 3}
	javaSeed123NextInt16    = []int{11, 3, 15, 4, 4, 9, 9, 4, 12, 9}
	javaSeed123NextInt100   = []int{82, 50, 76, 89, 95, 57, 34, 37, 85, 53}
	javaSeed123NextDouble   = []float64{0.7231742029971469, 0.9908988967772393, 0.25329310557439133, 0.6088003703785169, 0.8058695140834087}
)

func TestNext32MatchesJava(t *testing.T) {
	r := New(123)
	got := make([]int32, len(javaSeed123NextInt))
	for i := range got {
		got[i] = r.next(32)
	}
	assert.Equal(t, javaSeed123NextInt, got)
}

func TestNextIntBoundedMatchesJava(t *testing.T) {
	cases := []struct {
		name  string
		bound int
		want  []int
	}{
		{"bound=10", 10, javaSeed123NextInt10},
		{"bound=16 (power of 2 fast path)", 16, javaSeed123NextInt16},
		{"bound=100 (rejection loop)", 100, javaSeed123NextInt100},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r := New(123)
			got := make([]int, len(tc.want))
			for i := range got {
				got[i] = r.NextInt(tc.bound)
			}
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestNextDoubleMatchesJava(t *testing.T) {
	r := New(123)
	got := make([]float64, len(javaSeed123NextDouble))
	for i := range got {
		got[i] = r.NextDouble()
	}
	assert.Equal(t, javaSeed123NextDouble, got)
}

func TestNextIntRange(t *testing.T) {
	r := New(42)
	for i := 0; i < 1000; i++ {
		v := r.NextIntRange(5, 10)
		assert.GreaterOrEqual(t, v, 5)
		assert.Less(t, v, 10)
	}
}

func TestSetSeedResetsSequence(t *testing.T) {
	r := New(123)
	first := make([]float64, 3)
	for i := range first {
		first[i] = r.NextDouble()
	}
	// Confirms the sequence matches the Java reference, not just self-consistent.
	assert.Equal(t, javaSeed123NextDouble[:3], first)

	r.SetSeed(123)
	second := make([]float64, 3)
	for i := range second {
		second[i] = r.NextDouble()
	}
	assert.Equal(t, first, second)
}
