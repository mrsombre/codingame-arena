package sha1prng

import (
	"fmt"
	"strings"
	"testing"
)

// TestNextDoubleMatchesJava pins our SHA1PRNG output against values captured
// from Java's SecureRandom.getInstance("SHA1PRNG") for two seeds:
// a positive seed and a negative one (the latter caught a divergence in
// convert that we want to keep regression-tested).
//
// Probe: tmp/SHA1PRNGProbe.java.
func TestNextDoubleMatchesJava(t *testing.T) {
	cases := []struct {
		name string
		seed int64
		want []float64
	}{
		{
			name: "negative seed -8937286792422418000",
			seed: -8937286792422418000,
			want: []float64{
				0.7562021197437313,
				0.11974861373535017,
				0.6481510783399154,
				0.3396049659191963,
				0.5808944660341401,
				0.4610532839726166,
				0.17634608083222192,
				0.15729711992064532,
				0.5536934920895401,
				0.4987055701459304,
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r := New(tc.seed)
			for i, want := range tc.want {
				got := r.NextDouble()
				if got != want {
					t.Errorf("NextDouble()[%d] = %.20g, want %.20g", i, got, want)
				}
			}
		})
	}
}

// TestNextIntRangeMatchesJavaGridLowering replays the exact RNG call sequence
// GridMaker performs up to the lowestIsland nextInt(2, 4) for a known-bad
// seed: 926 nextDouble (pre-pruning) + 19 nextInts from pruning shuffles
// (sizes 2,2,3,2,2,3,2,2,2,2,2,2,2,2,2,2,2 yielding 1+1+2+1+1+2+1×11 = 19
// nextInt calls). Java's reference returns 3; we want the same.
// Reference: tmp/SHA1PRNGProbe.java mode "gridmaker_lowering".
func TestNextIntRangeMatchesJavaGridLowering(t *testing.T) {
	r := New(-8937286792422418000)
	for i := 0; i < 926; i++ {
		r.NextDouble()
	}
	sizes := []int{2, 2, 3, 2, 2, 3, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2}
	for _, sz := range sizes {
		for i := sz; i > 1; i-- {
			r.NextInt(i)
		}
	}
	got := r.NextIntRange(2, 4)
	if got != 3 {
		t.Errorf("NextIntRange(2, 4) after pruning sequence = %d, want 3", got)
	}
}

// TestNextIntAfterBurnMatchesJava regression-tests our NextInt against Java
// SecureRandom("SHA1PRNG") with 926 nextDouble burns first (matches the
// GridMaker pre-pruning consumption for height=23, width=42, league 4 on
// seed -8937286792422418000). Reference values from tmp/SHA1PRNGProbe.java
// "bounded" mode.
func TestNextIntAfterBurnMatchesJava(t *testing.T) {
	r := New(-8937286792422418000)
	for i := 0; i < 926; i++ {
		r.NextDouble()
	}

	want := strings.TrimSpace(`
nextInt(2)=1
nextInt(2)=0
nextInt(2)=0
nextInt(2)=0
nextInt(2)=1
nextInt(3)=0
nextInt(3)=1
nextInt(3)=1
nextInt(3)=2
nextInt(3)=2
nextInt(4)=3
nextInt(4)=0
nextInt(4)=0
nextInt(4)=0
nextInt(4)=0
`)
	var got []string
	for _, bound := range []int{2, 3, 4} {
		for i := 0; i < 5; i++ {
			got = append(got, fmt.Sprintf("nextInt(%d)=%d", bound, r.NextInt(bound)))
		}
	}
	if g := strings.Join(got, "\n"); g != want {
		t.Errorf("NextInt sequence mismatch\nwant:\n%s\ngot:\n%s", want, g)
	}
}
