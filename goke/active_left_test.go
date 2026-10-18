package goke

import (
	"testing"
	"time"

	"github.com/kjkrol/ceremony/effect"
	"github.com/kjkrol/ceremony/tag"
)

func TestActive_LeftIsWhatRemainsOfTheEffect(t *testing.T) {
	next := effect.Changed
	reg := effect.New(func(string) tag.Tag[effect.States] { t := next; next++; return t })
	reg.Define("burn", effect.Spec{effect.Lasts(time.Second)})
	reg.Define("glow", effect.Spec{})
	burn, glow := reg.Named("burn"), reg.Named("glow")
	var a Active
	if _, ok := a.Left(burn); ok {
		t.Fatal("under nothing: Left says it is under burn")
	}
	a.Slots[0] = effectSlot{Kind: burn.ID(), Left: 300 * time.Millisecond, State: slotRunning}
	a.Slots[1] = effectSlot{Kind: glow.ID(), Left: effect.Forever, State: slotRunning}
	if d, ok := a.Left(burn); !ok || d != 300*time.Millisecond {
		t.Errorf("burn: %v, %v; want 300ms left", d, ok)
	}
	if d, ok := a.Left(glow); !ok || d != effect.Forever {
		t.Errorf("glow: %v, %v; want Forever", d, ok)
	}
	var w Wide
	w.Slots[2] = effectSlot{Kind: burn.ID(), Left: time.Second, State: slotPending}
	if d, ok := w.Left(burn); !ok || d != time.Second {
		t.Errorf("wide, burn pending: %v, %v; want a second left", d, ok)
	}
}
