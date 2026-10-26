package host_test

import (
	"testing"

	"github.com/kjkrol/ceremony/host"
)

func TestPairRules_For_IsTrueForTheFirstSideAndItsNarrowing(t *testing.T) {
	h, m := hostWithin[int](t, []any{pairWithin(hunter, hunted, []host.Side{host.SideOf(alpha)},
		func(host.Tick, int) {})},
		as(hunter).in(alpha), as(hunter), as(hunted).in(alpha), as(hunted))
	for i, want := range []bool{true, false, false, false} {
		if got := h.For(m[i]); got != want {
			t.Errorf("For(%d) = %v, want %v: only the hunter in alpha has a rule", i, got, want)
		}
	}
	if h.For(host.Marks{}) {
		t.Error("For(nothing) = true, want false: the rule needs the hunter's tag")
	}
	if (&host.PairRules[int]{}).For(m[0]) {
		t.Error("an empty host is for somebody")
	}
}
