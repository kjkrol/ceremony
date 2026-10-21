package steps_test

import (
	"slices"
	"testing"

	"github.com/kjkrol/ceremony/command"
	"github.com/kjkrol/ceremony/host"
	"github.com/kjkrol/ceremony/internal/steps"
	"github.com/kjkrol/uid"
)

type pair struct {
	me    uid.UID64
	other uid.UID64
}

func (p pair) Who() uid.UID64            { return p.me }
func (p pair) Whom(each func(uid.UID64)) { each(p.other) }

var _ host.Met = pair{}

const (
	one    = uid.UID64(1 << 41)
	two    = uid.UID64(1<<41 + 1)
	beyond = uid.UID64(1<<41 + 2)
)

// gives fires root on payload, says by whom it gave a halt and how many times a firing allocates
func gives(t *testing.T, payload any, root steps.Step, p steps.Pass) ([]uid.UID64, float64) {
	t.Helper()
	var carrier command.Carrier
	var halts command.Queue[halt]
	if err := carrier.Carry(&halts); err != nil {
		t.Fatal(err)
	}
	p.Commands = &carrier
	r := steps.NewInstant(t.Name(), root, payload)
	r.Fire(p, one, true, 0, false)
	var by []uid.UID64
	halts.Drain(func(i command.Issued[halt]) { by = append(by, i.Entity) })
	n := testing.AllocsPerRun(100, func() {
		r.Fire(p, one, true, 0, false)
		halts.Drain(func(command.Issued[halt]) {})
	})
	return by, n
}

func TestForOther_RunsWithoutAllocating(t *testing.T) {
	by, n := gives(t, &pair{me: one, other: two}, steps.NewForOther(steps.NewOrder(halt{})),
		steps.Pass{})
	if !slices.Equal(by, []uid.UID64{two}) {
		t.Errorf("given by %v, want the other %v", by, two)
	}
	if n != 0 {
		t.Errorf("a step on the other of a pair allocated %v times", n)
	}
}

func TestPlaces_RunWithoutAllocating(t *testing.T) {
	at := host.Places{Placing: host.NewPlacing("round", func(any) bool { return true }), N: 1}
	p := steps.Pass{Places: func(_ any, _ host.Places, each func(uid.UID64)) {
		each(two)
		each(beyond)
	}}
	by, n := gives(t, &moment{}, steps.NewPlaces(at, steps.NewOrder(halt{})), p)
	if !slices.Equal(by, []uid.UID64{two, beyond}) {
		t.Errorf("given by %v, want the places %v", by, []uid.UID64{two, beyond})
	}
	if n != 0 {
		t.Errorf("a step on the places allocated %v times", n)
	}
}

func TestOfSubject_ResolvesWithoutAllocating(t *testing.T) {
	asked := steps.NewBranch[any]("", steps.GuardPlaying(0, steps.OfSubject),
		steps.NewOrder(halt{}), nil)
	p := steps.Pass{Roles: func(id uid.UID64) uint64 {
		if id == two {
			return 1
		}
		return 0
	}}
	by, n := gives(t, &pair{me: one, other: two}, asked, p)
	if !slices.Equal(by, []uid.UID64{one}) {
		t.Errorf("given by %v, want %v: the other plays the role", by, one)
	}
	if n != 0 {
		t.Errorf("asking of the other of a pair allocated %v times", n)
	}
	if by, _ := gives(t, &pair{me: one, other: beyond}, asked, p); len(by) != 0 {
		t.Errorf("given by %v: the other plays no role", by)
	}
}
