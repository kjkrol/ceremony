package steps

import (
	"fmt"
	"math/bits"
	"reflect"
	"time"

	"github.com/kjkrol/ceremony/command"
	"github.com/kjkrol/ceremony/effect"
	"github.com/kjkrol/ceremony/fact"
	"github.com/kjkrol/ceremony/store"
	"github.com/kjkrol/uid"
)

// Plans runs the plans of every entity with a Mind, once every step of the simulation, over
// the blocks of entities a store's binding hands it.
type Plans struct {
	now      func() time.Duration
	seed     uint64
	effects  effect.Runtime
	commands *command.Carrier
	roles    func(uid.UID64) uint64
	defs     map[uint64]definition // by name hashed
	trees    map[uint64]*tree
	facts    store.Facts
	columns  []binding
	byType   map[reflect.Type]any // *carried[F] by F
	bound    bool
	c        ctx // one for every tick, so a block costs no allocation
}

// NewPlans is the plans on the clock now, drawing Chance from seed, casting fx and giving
// their commands to commands.
func NewPlans(now func() time.Duration, seed uint64, fx effect.Runtime,
	commands *command.Carrier) *Plans {
	return &Plans{now: now, seed: seed, effects: fx, commands: commands}
}

// Roles has the plans find the roles an entity plays with lookup.
func (s *Plans) Roles(lookup func(uid.UID64) uint64) { s.roles = lookup }

type binder interface{ bind(s *Plans) }

// Bind lays every plan out and binds the facts their steps read and write to the store's
// columns. Call it once, as the binding sets up; a plan defined after panics.
func (s *Plans) Bind(facts store.Facts) {
	s.facts, s.bound = facts, true
	s.byType, s.trees = map[reflect.Type]any{}, map[uint64]*tree{}
	for id, d := range s.defs {
		t := layOut(d.name, d.root)
		for _, n := range t.nodes {
			if b, ok := n.exec.(binder); ok {
				b.bind(s)
			}
		}
		s.trees[id] = t
	}
}

// Facts are the types of the facts the plans read and write, for the store to keep.
func (s *Plans) Facts() []reflect.Type {
	out := make([]reflect.Type, 0, len(s.columns))
	for _, b := range s.columns {
		out = append(out, b.typ)
	}
	return out
}

type binding struct {
	typ   reflect.Type
	sweep func(c *ctx)
}

// Tick runs the plans of one block of entities: ids and minds side by side, row what they
// carry, dt the step.
func (s *Plans) Tick(row store.Row, ids []uid.UID64, minds []store.Mind, dt time.Duration) {
	c := &s.c
	*c = ctx{sys: s, row: row, now: s.now(), pass: Pass{Dt: dt}}
	for i, id := range ids {
		m := &minds[i]
		t := s.trees[m.Plan]
		if t == nil {
			panic(fmt.Sprintf("rule: entity %v runs a plan no kind gave (%x)", id, m.Plan))
		}
		c.i, c.id, c.me, c.entity, c.mind, c.tree = i, id, id, true, m, t
		c.prev, c.next = m.Running, store.StepSet{}
		c.run(0)
		for w := range c.prev {
			for halted := c.prev[w] &^ c.next[w]; halted != 0; halted &= halted - 1 {
				at := w*64 + bits.TrailingZeros64(halted)
				t.nodes[at].exec.halt(c, at)
			}
		}
		m.Running = c.next
		for _, b := range s.columns {
			b.sweep(c)
		}
	}
	c.row, c.mind, c.tree = store.Row{}, nil, nil
}

type ctx struct {
	sys        *Plans // nil for a rule
	instant    bool
	pass       Pass
	payload    any // a pointer to a rule's moment
	subject    uid.UID64
	about      bool
	names      bool // the moment names a subject: an Aimed command fails while it names nobody
	now        time.Duration
	row        store.Row
	i          int
	id         uid.UID64 // whom the steps act on now
	me         uid.UID64 // the entity the rule or the plan is about
	other      uid.UID64 // the other a step acting on others turned to
	hasOther   bool
	entity     bool // false for a rule of a clock.Moment: nobody to act for
	mind       *store.Mind
	tree       *tree
	prev, next store.StepSet
	hops       uint8        // of the message the branch stands under
	hopsOf     reflect.Type // its type
	// the first of the others a rule's moment met; toFirst is keepFirst, bound once, as a
	// closure made at every call would escape to the heap
	first    uid.UID64
	hasFirst bool
	toFirst  func(uid.UID64)
}

func (c *ctx) run(at int) Status {
	n := &c.tree.nodes[at]
	if !c.wasRunning(at) {
		n.exec.enter(c, at)
	}
	s := n.exec.tick(c, at, n.kids)
	if s == Running {
		add(&c.next, at)
	}
	return s
}

func (c *ctx) keepFirst(id uid.UID64) {
	if !c.hasFirst {
		c.first = id
		c.hasFirst = true
	}
}

func (c *ctx) wasRunning(at int) bool { return has(&c.prev, at) }

func has(n *store.StepSet, at int) bool { return n[at>>6]&(1<<(at&63)) != 0 }
func add(n *store.StepSet, at int)      { n[at>>6] |= 1 << (at & 63) }

type carried[F any] struct{ col store.Col }

// false, and nothing given, when to has no mind to take it up
func (f *carried[F]) putOn(c *ctx, to uid.UID64, v F) bool {
	p := c.sys.facts.Stage(to, f.col)
	if p == nil {
		return false
	}
	*(*F)(p) = v
	return true
}

func (f *carried[F]) drop(c *ctx) {
	if f.of(c) != nil {
		c.sys.facts.Drop(c.id, f.col)
	}
}

func (f *carried[F]) sweep(c *ctx) {
	if v := f.of(c); v != nil {
		if m, ok := any(*v).(fact.Sent); ok && c.now-m.Sent().At > fact.AskLife {
			c.sys.facts.Drop(c.id, f.col)
		}
	}
}

func (f *carried[F]) of(c *ctx) *F { return (*F)(c.row.At(f.col, c.i)) }

func carriedOf[F any](s *Plans) *carried[F] {
	key := reflect.TypeFor[F]()
	if f, ok := s.byType[key]; ok {
		return f.(*carried[F])
	}
	f := &carried[F]{col: s.facts.Column(key)}
	s.byType[key] = f
	s.columns = append(s.columns, binding{typ: key, sweep: f.sweep})
	return f
}

// a fact about nobody leaves the subject from further out
func (c *ctx) under(f any, at int) Status {
	s, ok := f.(fact.Subject)
	if !ok {
		return c.run(at)
	}
	who, about := s.Subject()
	if !about {
		return c.run(at)
	}
	outer, was, hops, hopsOf := c.subject, c.about, c.hops, c.hopsOf
	c.subject, c.about = who, true
	if m, ok := f.(fact.Sent); ok {
		c.hops, c.hopsOf = m.Sent().Hops, reflect.TypeOf(f)
	}
	st := c.run(at)
	c.subject, c.about, c.hops, c.hopsOf = outer, was, hops, hopsOf
	return st
}
