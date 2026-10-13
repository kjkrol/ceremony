package ceremony_test

import (
	"errors"
	"fmt"
	"slices"
	"strings"
	"testing"

	"github.com/kjkrol/ceremony"
	"github.com/kjkrol/ceremony/act"
	"github.com/kjkrol/ceremony/cond"
	"github.com/kjkrol/ceremony/host"
	"github.com/kjkrol/ceremony/tag"
	"github.com/kjkrol/goke/v3"
	"github.com/kjkrol/uid"
)

type roles struct{}

const (
	hunter tag.Tag[roles] = iota
	hunted
)

type token struct{ N int }

var (
	hunterPart = ceremony.NewPart("hunter", 1)
	huntedPart = ceremony.NewPart("hunted", 2)
)

type poke struct{ self uid.UID64 }

func (p poke) Who() uid.UID64 { return p.self }

type glance struct{ self, other uid.UID64 }

func (g glance) Who() uid.UID64            { return g.self }
func (g glance) Whom(each func(uid.UID64)) { each(g.other) }

type dawn struct{ world uid.UID64 }

func (d dawn) Who() uid.UID64 { return d.world }

type noted struct{}

func (noted) Command() {}

func noteOf[P any](name string, when cond.Cond) act.Rule {
	if when == nil {
		when = cond.Self.If(cond.Always)
	}
	return act.Branch[P]{Name: name, When: when, Then: act.Self.Do(noted{})}
}

func narrowed(r act.Rule) act.Rule { return ceremony.NewPart("hookon scout", 0).Obeys(r).Rules()[0] }

type keeper struct {
	name  string
	adds  host.Host
	asked int
	took  []string
}

func (h *keeper) Add(r any) error {
	h.asked++
	if err := h.adds.Add(r); err != nil {
		return fmt.Errorf("%w in %s", err, h.name)
	}
	h.took = append(h.took, fmt.Sprint(r))
	return nil
}

func names(rules ...act.Rule) []string {
	out := make([]string, len(rules))
	for i, r := range rules {
		out[i] = fmt.Sprint(r)
	}
	return out
}

type broken struct{ asked int }

var errBroken = errors.New("broken keeper")

func (b *broken) Add(any) error {
	b.asked++
	return errBroken
}

func TestDeliver_HandsEachRuleToTheHostOfItsMoment(t *testing.T) {
	pokes := &keeper{name: "pokes", adds: &host.Rules[poke]{}}
	glances := &keeper{name: "glances", adds: &host.PairRules[glance]{}}
	dawns := &keeper{name: "dawns", adds: &host.StepRules[dawn]{}}
	among := []host.Host{glances, dawns, pokes}

	wake := noteOf[dawn]("wake", nil)
	spot := noteOf[poke]("spot", cond.Self.If(hunterPart))
	stare := noteOf[glance]("stare", cond.Self.If(hunterPart).And(cond.Other.If(huntedPart)))
	drown := noteOf[poke]("drown", nil)
	scouting := narrowed(drown)
	scoutStare := narrowed(stare)

	if err := ceremony.Deliver(among, wake, spot, stare, drown, scouting, scoutStare); err != nil {
		t.Fatalf("Deliver = %v", err)
	}
	for _, c := range []struct {
		h     *keeper
		took  []string
		asked int
	}{
		{glances, names(stare, scoutStare), 6},
		{dawns, names(wake), 4},
		{pokes, names(spot, drown, scouting), 3},
	} {
		if !slices.Equal(c.h.took, c.took) {
			t.Errorf("%s took %v; want %v", c.h.name, c.h.took, c.took)
		}
		if c.h.asked != c.asked {
			t.Errorf("%s was asked %d times; want %d", c.h.name, c.h.asked, c.asked)
		}
	}
}

func TestDeliver_TheFirstHostToTakeTheRuleHasIt(t *testing.T) {
	first := &keeper{name: "first", adds: &host.Rules[poke]{}}
	second := &keeper{name: "second", adds: &host.Rules[poke]{}}
	drown := noteOf[poke]("drown", nil)

	if err := ceremony.Deliver([]host.Host{first, second}, drown); err != nil {
		t.Fatalf("Deliver = %v", err)
	}
	if !slices.Equal(first.took, names(drown)) || second.asked != 0 {
		t.Errorf("first took %v, second was asked %d times; want first to take it, second never asked", first.took, second.asked)
	}
}

func TestDeliver_StopsAtAnotherError(t *testing.T) {
	b := &broken{}
	pokes := &keeper{name: "pokes", adds: &host.Rules[poke]{}}
	drown, spot := noteOf[poke]("drown", nil), noteOf[poke]("spot", cond.Self.If(hunterPart))

	err := ceremony.Deliver([]host.Host{b, pokes}, drown, spot)
	if !errors.Is(err, errBroken) || errors.Is(err, host.ErrUnhosted) {
		t.Errorf("Deliver = %v; want the broken keeper's error, not host.ErrUnhosted", err)
	}
	if b.asked != 1 || pokes.asked != 0 {
		t.Errorf("the broken keeper was asked %d times, the next %d; want once and never", b.asked, pokes.asked)
	}
}

func TestDeliver_StopsAtAHostAlreadyBuilt(t *testing.T) {
	bound := &host.Rules[poke]{}
	goke.New().Setup(goke.SystemFn{OnInit: func(si *goke.SysInit) {
		var tokens goke.Comp[token]
		bound.Bind(si.NewQueryBuilder(&tokens))
	}})
	built := &keeper{name: "built", adds: bound}
	later := &keeper{name: "later", adds: &host.Rules[poke]{}}

	err := ceremony.Deliver([]host.Host{built, later}, noteOf[poke]("drown", nil))
	if !errors.Is(err, host.ErrHostBuilt) {
		t.Errorf("Deliver = %v; want host.ErrHostBuilt", err)
	}
	if later.asked != 0 {
		t.Errorf("the later keeper was asked %d times; want never", later.asked)
	}
}

func TestDeliver_RefusesARuleNoHostTakes(t *testing.T) {
	pokes := &keeper{name: "pokes", adds: &host.Rules[poke]{}}
	dawns := &keeper{name: "dawns", adds: &host.StepRules[dawn]{}}
	stare := noteOf[glance]("stare", cond.Self.If(hunterPart).And(cond.Other.If(huntedPart)))

	for _, c := range []struct {
		name  string
		among []host.Host
		r     act.Rule
	}{
		{"a moment nobody hosts", []host.Host{pokes, dawns}, stare},
		{"no hosts", nil, stare},
	} {
		t.Run(c.name, func(t *testing.T) {
			err := ceremony.Deliver(c.among, c.r)
			if !errors.Is(err, host.ErrUnhosted) {
				t.Fatalf("Deliver = %v; want host.ErrUnhosted", err)
			}
			if !strings.Contains(err.Error(), c.r.String()) {
				t.Errorf("Deliver = %q; want it naming the rule %s", err, c.r)
			}
		})
	}
	if len(pokes.took) != 0 || len(dawns.took) != 0 {
		t.Errorf("pokes took %v, dawns %v; want nothing taken", pokes.took, dawns.took)
	}
}

func TestDeliver_ARoleIsDeliveredRuleByRule(t *testing.T) {
	pokes := &keeper{name: "pokes", adds: &host.Rules[poke]{}}
	dawns := &keeper{name: "dawns", adds: &host.StepRules[dawn]{}}
	role := ceremony.NewPart("hosts early", 1).
		Obeys(noteOf[dawn]("wake", nil), noteOf[poke]("drown", nil))

	if err := ceremony.Deliver([]host.Host{pokes, dawns}, role); err != nil {
		t.Fatalf("Deliver = %v", err)
	}
	if len(dawns.took) != 1 || len(pokes.took) != 1 {
		t.Errorf("dawns took %v, pokes %v; want one rule of the role each", dawns.took, pokes.took)
	}
}

func TestDeliver_NoRulesIsNothing(t *testing.T) {
	if err := ceremony.Deliver(nil); err != nil {
		t.Errorf("Deliver(nil) = %v; want nil", err)
	}
}
