package steps

import (
	"fmt"
	"strings"
	"time"

	"github.com/kjkrol/ceremony/command"
	"github.com/kjkrol/ceremony/effect"
	"github.com/kjkrol/ceremony/fact"
	"github.com/kjkrol/ceremony/store"
	"github.com/kjkrol/uid"
)

// Pass is what a rule is told of the pass it fires in.
type Pass struct {
	Commands *command.Carrier
	Effects  effect.Runtime
	Dt       time.Duration
	Time     time.Duration // game time at the step's end
	Seed     uint64
	Around   func(moment any, rings int, each func(uid.UID64))
	Ahead    func(moment any, steps int, each func(uid.UID64))
	Roles    func(id uid.UID64) uint64 // a bit a role
}

func (c *ctx) rolesOf(id uid.UID64) uint64 {
	roles := c.pass.Roles
	if c.sys != nil {
		roles = c.sys.roles
	}
	if roles == nil {
		return 0
	}
	return roles(id)
}

func (c *ctx) effects() effect.Runtime {
	if c.sys != nil {
		return c.sys.effects
	}
	return c.pass.Effects
}

type instantExec interface{ instant() }

// refuses says why not on the moment payload points to, empty where it runs
type momentExec interface{ refuses(payload any) string }

type placed interface{ Placed() }

type headed interface{ Headed() }

// Placing is which places a rule turns to: those round the entity, or those ahead of it.
type Placing uint8

const (
	// AroundPlaces are the places within n rings of where the entity stands.
	AroundPlaces Placing = iota + 1
	// AheadPlaces are the places on the cells one to n steps ahead of the entity.
	AheadPlaces
)

// String says Around or Ahead.
func (p Placing) String() string {
	if p == AheadPlaces {
		return "Ahead"
	}
	return "Around"
}

// Instant is a rule laid out for running at a moment, every step of it done within the pass.
type Instant struct {
	tree *tree
	mind store.Mind
	c    ctx
}

// NewInstant lays root out as the rule named name, run on the moment payload points to. It
// panics on a step that lasts over ticks or one the moment cannot run.
func NewInstant(name string, root Step, payload any) *Instant {
	r := &Instant{tree: layOut(name, root)}
	for i, n := range r.tree.nodes {
		if _, ok := n.exec.(instantExec); !ok {
			panic(fmt.Sprintf("rule: %q: %s lasts over ticks — it belongs to a plan", name, r.tree.signs[i]))
		}
		if m, ok := n.exec.(momentExec); ok {
			if why := m.refuses(payload); why != "" {
				panic(fmt.Sprintf("rule: %q: %s %s, not %T", name, r.tree.signs[i], why, payload))
			}
		}
	}
	_, names := payload.(fact.Subject)
	r.c = ctx{instant: true, tree: r.tree, mind: &r.mind, payload: payload, names: names}
	return r
}

// Fire runs the rule once for id (entity false: for nobody), an Aimed command of it at subject
// when about.
func (r *Instant) Fire(p Pass, id uid.UID64, entity bool, subject uid.UID64, about bool) {
	c := &r.c
	c.pass, c.prev, c.next = p, store.StepSet{}, store.StepSet{}
	c.id, c.me, c.entity, c.subject, c.about = id, id, entity, subject, about
	c.run(0)
}

// NewForOther runs step on each of the others the moment met, or on its subject, in place of
// the entity.
func NewForOther(step Step) Step {
	return composite{kids: []Step{step}, sign: "forother", make: func() exec { return forOther{} }}
}

// NewSame runs step on the entity itself, inside a step acting on others.
func NewSame(step Step) Step {
	return composite{kids: []Step{step}, sign: "self", make: func() exec { return same{} }}
}

type same struct{ basic }

func (same) instant() {}

func (same) tick(c *ctx, _ int, kids []int) Status {
	was := c.id
	c.id = c.me
	st := c.run(kids[0])
	c.id = was
	return st
}

// NewPlaces runs step on each place within n rings of where the entity stands, or on the cells
// n steps ahead of it, in place of it.
func NewPlaces(placing Placing, n int, step Step) Step {
	sign := fmt.Sprintf("%s(%d)", strings.ToLower(placing.String()), n)
	return composite{kids: []Step{step}, sign: sign,
		make: func() exec { return around{placing: placing, n: n} }}
}

type met interface{ Whom(each func(uid.UID64)) }

type around struct {
	basic
	placing Placing
	n       int
}

func (around) instant() {}

func (a around) refuses(payload any) string {
	if a.placing == AheadPlaces {
		if _, ok := payload.(headed); ok {
			return ""
		}
		return "needs a moment that is Headed"
	}
	if _, ok := payload.(placed); ok {
		return ""
	}
	return "needs a moment that is Placed"
}

func (a around) tick(c *ctx, _ int, kids []int) Status {
	tell := c.pass.Around
	if a.placing == AheadPlaces {
		tell = c.pass.Ahead
	}
	if !c.entity || tell == nil {
		return Failure
	}
	self, st := c.id, Failure
	tell(c.payload, a.n, func(place uid.UID64) {
		c.id = place
		if c.run(kids[0]) == Success {
			st = Success
		}
	})
	c.id = self
	return st
}

type forOther struct{ basic }

func (forOther) instant() {}

// in a plan the other is the subject of the fact the branch stands on
func (forOther) tick(c *ctx, _ int, kids []int) Status {
	if !c.entity {
		return Failure
	}
	self, other, had, st := c.id, c.other, c.hasOther, Failure
	on := func(id uid.UID64) {
		c.id, c.other, c.hasOther = id, id, true
		if c.run(kids[0]) == Success {
			st = Success
		}
	}
	switch {
	case !c.instant:
		if c.about {
			on(c.subject)
		}
	default:
		switch m := c.payload.(type) {
		case met:
			m.Whom(on)
		case fact.Subject:
			if id, ok := m.Subject(); ok {
				on(id)
			}
		}
	}
	c.id, c.other, c.hasOther = self, other, had
	return st
}
