package steps

import (
	"fmt"
	"strings"
	"time"

	"github.com/kjkrol/ceremony/command"
	"github.com/kjkrol/ceremony/effect"
	"github.com/kjkrol/ceremony/fact"
	"github.com/kjkrol/ceremony/host"
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
	Places   func(moment any, p host.Places, each func(uid.UID64))
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
	r.c.toFirst = r.c.keepFirst
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
	return composite{kids: []Step{step}, sign: "forother", make: func() exec {
		f := &forOther{}
		f.each = f.on
		return f
	}}
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

// NewPlaces runs step on each of the places at the host tells for the moment, in place of the
// entity.
func NewPlaces(at host.Places, step Step) Step {
	return composite{kids: []Step{step}, sign: strings.ToLower(at.String()),
		make: func() exec {
			a := &around{at: at}
			a.each = a.on
			return a
		}}
}

type met interface{ Whom(each func(uid.UID64)) }

type around struct {
	basic
	at  host.Places
	c   *ctx
	kid int
	st  Status
	// on, bound once: a closure made at every tick would escape to the heap
	each func(uid.UID64)
}

func (*around) instant() {}

func (a *around) refuses(payload any) string {
	if a.at.Placing.Takes(payload) {
		return ""
	}
	return fmt.Sprintf("needs a moment with places %s", a.at.Placing)
}

func (a *around) tick(c *ctx, _ int, kids []int) Status {
	tell := c.pass.Places
	if !c.entity || tell == nil {
		return Failure
	}
	self := c.id
	a.c = c
	a.kid = kids[0]
	a.st = Failure
	tell(c.payload, a.at, a.each)
	c.id = self
	return a.st
}

func (a *around) on(place uid.UID64) {
	a.c.id = place
	if a.c.run(a.kid) == Success {
		a.st = Success
	}
}

type forOther struct {
	basic
	c   *ctx
	kid int
	st  Status
	// on, bound once: a closure made at every tick would escape to the heap
	each func(uid.UID64)
}

func (*forOther) instant() {}

// in a plan the other is the subject of the fact the branch stands on
func (f *forOther) tick(c *ctx, _ int, kids []int) Status {
	if !c.entity {
		return Failure
	}
	self, other, had := c.id, c.other, c.hasOther
	f.c = c
	f.kid = kids[0]
	f.st = Failure
	switch {
	case !c.instant:
		if c.about {
			f.on(c.subject)
		}
	default:
		switch m := c.payload.(type) {
		case met:
			m.Whom(f.each)
		case fact.Subject:
			if id, ok := m.Subject(); ok {
				f.on(id)
			}
		}
	}
	c.id, c.other, c.hasOther = self, other, had
	return f.st
}

func (f *forOther) on(id uid.UID64) {
	f.c.id = id
	f.c.other = id
	f.c.hasOther = true
	if f.c.run(f.kid) == Success {
		f.st = Success
	}
}
