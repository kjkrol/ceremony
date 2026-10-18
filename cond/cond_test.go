package cond

import (
	"strings"
	"testing"

	"github.com/kjkrol/ceremony/host"
	"github.com/kjkrol/ceremony/internal/steps"
	"github.com/kjkrol/ceremony/tag"
)

type part tag.Tag[Roles]

func (p part) Tag() tag.Tag[Roles] { return tag.Tag[Roles](p) }

type poke struct{}

func TestAndAndOr_FlattenTheirOwnKind(t *testing.T) {
	a, b, c := part(1), part(2), part(3)
	both := Self.If(a).And(Self.If(b)).And(Self.If(c))
	if j, ok := both.(junction); !ok || !j.all || len(j.parts) != 3 {
		t.Errorf("And twice = %#v; want one junction of three", both)
	}
	either := Self.If(a).Or(Self.If(b)).Or(Self.If(c))
	if j, ok := either.(junction); !ok || j.all || len(j.parts) != 3 {
		t.Errorf("Or twice = %#v; want one junction of three", either)
	}
	mixed := Self.If(a).And(Self.If(b)).Or(Self.If(c))
	if j, ok := mixed.(junction); !ok || j.all || len(j.parts) != 2 {
		t.Errorf("And then Or = %#v; want an or of two, the and inside", mixed)
	}
}

func TestLower_TakesTheRolesAtTheTopAndGuardsTheRest(t *testing.T) {
	a, b := part(1), part(2)
	word := NewWord("poked", func(poke) bool { return true })
	top := Self.If(a).And(Other.If(b)).And(Self.If(word)).(steps.Lowered).Lower(steps.AtTop, "top")
	if len(top.SelfRoles) != 1 || top.SelfRoles[0] != 1 || top.OtherRole != 2 {
		t.Errorf("at the top: self roles %v, other %d; want [1] and 2", top.SelfRoles,
			top.OtherRole)
	}
	if top.Guard == nil || len(top.Words) != 1 {
		t.Errorf("at the top the word alone guards; got %v of %v", top.Guard, top.Words)
	}
	inner := Self.If(a).And(Other.If(b)).(steps.Lowered).Lower(steps.Inside, "inner")
	if len(inner.SelfRoles) != 0 || inner.OtherRole != -1 || inner.Guard == nil {
		t.Errorf("inside: roles %v, %d, guard %v; want both guards", inner.SelfRoles,
			inner.OtherRole, inner.Guard)
	}
	under := Self.If(a).Or(Other.If(b)).(steps.Lowered).Lower(steps.AtTop, "or")
	if len(under.SelfRoles) != 0 || under.OtherRole != -1 {
		t.Errorf("under Or: roles %v, %d; want none, the roles are guards", under.SelfRoles,
			under.OtherRole)
	}
}

var (
	round = host.NewPlacing("Round", func(any) bool { return true })
	front = host.NewPlacing("Front", func(any) bool { return true })
)

func at(p *host.Placing, n int) Subject { return At(host.Places{Placing: p, N: n}) }

func TestLower_CountsThePlacesOfOneReach(t *testing.T) {
	a := part(1)
	low := at(round, 2).If(a).And(at(round, 2).If(Not(a))).(steps.Lowered).
		Lower(steps.AtTop, "places")
	if low.At.N != 2 || low.Places != 2 {
		t.Errorf("reach %d, places %d; want 2 and 2", low.At.N, low.Places)
	}
	defer func() {
		if msg, _ := recover().(string); !strings.Contains(msg, "one reach") {
			t.Errorf("panic %q; want one saying one reach", msg)
		}
	}()
	at(round, 1).If(a).And(at(round, 2).If(a)).(steps.Lowered).Lower(steps.AtTop, "reaches")
}

func TestLower_CountsThePlacesOfOneKind(t *testing.T) {
	a := part(1)
	low := at(front, 3).If(a).(steps.Lowered).Lower(steps.AtTop, "front")
	if low.At.Placing != front || low.At.N != 3 || low.Places != 1 {
		t.Errorf("places %v, %d of them; want Front(3) and 1", low.At, low.Places)
	}
	defer func() {
		if msg, _ := recover().(string); !strings.Contains(msg, "one kind") {
			t.Errorf("panic %q; want one saying one kind", msg)
		}
	}()
	at(round, 1).If(a).And(at(front, 1).If(a)).(steps.Lowered).Lower(steps.AtTop, "kinds")
}

type reply struct{ Yes bool }

func TestIf_AFactComesInAnUntilAlone(t *testing.T) {
	yes := NewWord("yes", func(r reply) bool { return r.Yes })
	low := Self.If(reply{}).And(Self.If(yes)).(steps.Lowered).Lower(steps.AsUntil, "until")
	if low.Came != 1 || len(low.Words) != 1 || low.Guard == nil {
		t.Errorf("Until of a fact and its word: came %d, words %d; want 1 and 1", low.Came,
			len(low.Words))
	}
	for name, c := range map[string]struct {
		c    Cond
		as   steps.LowerAs
		want string
	}{
		"in a When":    {Self.If(reply{}), steps.Inside, "a loop's type"},
		"at the top":   {Self.If(reply{}), steps.AtTop, "a loop's type"},
		"of the other": {Other.If(reply{}), steps.AsUntil, "Self alone"},
		"under Not":    {Self.If(Not(reply{})), steps.AsUntil, "no Not of it"},
	} {
		t.Run(name, func(t *testing.T) {
			defer func() {
				if msg, _ := recover().(string); !strings.Contains(msg, c.want) {
					t.Errorf("panic %q; want one saying %q", msg, c.want)
				}
			}()
			c.c.(steps.Lowered).Lower(c.as, name)
		})
	}
}

func (reply) Fact() {}
