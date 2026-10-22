package steps_test

import (
	"testing"
	"time"

	"github.com/kjkrol/ceremony/command"
	"github.com/kjkrol/ceremony/internal/steps"
	"github.com/kjkrol/ceremony/store"
	"github.com/kjkrol/uid"
)

type (
	aside struct{ Of uid.UID64 }
	halt  struct {
		Long uint64
		Why  string
	}
)

func (aside) Command() {}
func (halt) Command()  {}

func (a *aside) Aim(who uid.UID64) { a.Of = who }

var _ steps.Aimed = (*aside)(nil)

type moment struct{}

func TestOrder_GivesWithoutAllocating(t *testing.T) {
	var carrier command.Carrier
	var asides command.Queue[aside]
	var halts command.Queue[halt]
	if err := carrier.Carry(&asides, &halts); err != nil {
		t.Fatal(err)
	}
	const me, other = uid.UID64(1 << 40), uid.UID64(1<<40 + 1)
	aimed := steps.NewInstant("aimed", steps.NewOrder(aside{}), &moment{})
	plain := steps.NewInstant("plain", steps.NewOrder(halt{Long: 1 << 20, Why: "rest"}), &moment{})
	p := steps.Pass{Commands: &carrier}
	aimed.Fire(p, me, true, other, true)
	aimed.Fire(p, me, true, 0, false)
	var of []uid.UID64
	asides.Drain(func(i command.Issued[aside]) { of = append(of, i.Command.Of) })
	if len(of) != 2 || of[0] != other || of[1] != 0 {
		t.Errorf("aimed at %v, want [%v 0]: one aimed before leaves nothing behind", of, other)
	}
	run := func() {
		aimed.Fire(p, me, true, other, true)
		plain.Fire(p, me, true, 0, false)
		asides.Drain(func(command.Issued[aside]) {})
		halts.Drain(func(command.Issued[halt]) {})
	}
	run()
	if n := testing.AllocsPerRun(100, run); n != 0 {
		t.Errorf("a rule giving two commands allocated %v times", n)
	}

	plans := steps.NewPlans(func() time.Duration { return 0 }, 0, nil, &carrier)
	id := plans.Define("plain", steps.NewThen("", steps.NewOrder(halt{Long: 1 << 20, Why: "rest"}),
		steps.NewIdle()))
	plans.Bind(nil)
	ids := []uid.UID64{me, other}
	minds := []store.Mind{{Plan: id}, {Plan: id}}
	tick := func() {
		for i := range minds {
			minds[i].Running = store.StepSet{}
		}
		plans.Tick(store.Row{}, ids, minds, time.Millisecond)
		halts.Drain(func(command.Issued[halt]) {})
	}
	tick()
	if n := testing.AllocsPerRun(100, tick); n != 0 {
		t.Errorf("a plan giving a command to two entities allocated %v times", n)
	}
}
