package plan

import (
	"fmt"
	"reflect"
	"time"

	"github.com/kjkrol/ceremony/act"
	"github.com/kjkrol/ceremony/command"
	"github.com/kjkrol/ceremony/comp"
	"github.com/kjkrol/ceremony/effect"
	"github.com/kjkrol/ceremony/internal/steps"
	"github.com/kjkrol/ceremony/store"
	"github.com/kjkrol/uid"
)

// Planned is what New makes: the plan's mind as a kind's component, registered on the Plans of
// the world the kind is defined in.
type Planned struct {
	comp.Template[store.Mind]
	p *planned
}

type planned struct {
	name    string
	root    steps.Step
	defined map[*Plans]bool
}

// New is the plan named name: root run in every step of the simulation, given to a kind in
// its Spec; a loop at its top says its Name. A name defined twice in one world panics.
func New(name string, root Part) Planned {
	if name == "" {
		panic("plan: a plan needs a name")
	}
	return Planned{
		Template: comp.Const(store.Mind{Plan: steps.PlanID(name)}),
		p:        &planned{name: name, root: act.TopNode(root.Lasting()), defined: map[*Plans]bool{}},
	}
}

// Name is the plan's.
func (p Planned) Name() string { return p.p.name }

// Plans runs the plans of every entity with a store.Mind, once every step of the simulation:
// a store's binding binds it and ticks it over the blocks of entities it walks.
type Plans struct{ s *steps.Plans }

// NewPlans is the plans on the clock now, drawing Chance from seed, casting fx and giving
// their commands to commands.
func NewPlans(now func() time.Duration, seed uint64, fx effect.Runtime,
	commands *command.Carrier) *Plans {
	return &Plans{s: steps.NewPlans(now, seed, fx, commands)}
}

// Register defines p on the plans, once; a name defined twice, or after Bind, panics.
func (ps *Plans) Register(p Planned) {
	if p.p.defined[ps] {
		return
	}
	if id := ps.s.Define(p.p.name, p.p.root); id != steps.PlanID(p.p.name) {
		panic(fmt.Sprintf("plan: the plan %q got another id than its name's", p.p.name))
	}
	p.p.defined[ps] = true
}

// Roles has the plans find the roles an entity plays with lookup.
func (ps *Plans) Roles(lookup func(uid.UID64) uint64) { ps.s.Roles(lookup) }

// Bind lays every plan out and binds the facts their steps read and write to the store's
// columns. The binding calls it once, as it sets up; a plan registered after panics.
func (ps *Plans) Bind(facts store.Facts) { ps.s.Bind(facts) }

// Facts are the types of the facts the plans read and write, for the store to keep.
func (ps *Plans) Facts() []reflect.Type { return ps.s.Facts() }

// Tick runs the plans of one block of entities: ids and minds side by side, row what they
// carry, dt the step.
func (ps *Plans) Tick(row store.Row, ids []uid.UID64, minds []store.Mind, dt time.Duration) {
	ps.s.Tick(row, ids, minds, dt)
}
