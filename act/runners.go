package act

import (
	"reflect"

	"github.com/kjkrol/ceremony/cond"
	"github.com/kjkrol/ceremony/host"
	"github.com/kjkrol/ceremony/store"
)

type every[P any] struct {
	label string
	react func(host.Tick, P)
}

func (e *every[P]) Hosted() any { return e }

func (e *every[P]) ForRole(role cond.Role) Rule { return forRole(e, role) }

func (e *every[P]) String() string { return e.label }

func (*every[P]) BindColumns(store.Columns) {}

func (e *every[P]) RunEach(t host.Tick, _ store.Row, n int, keep func(i int) bool, about func(i int) P) {
	for i := range n {
		if keep == nil || keep(i) {
			e.react(t, about(i))
		}
	}
}

func (e *every[P]) RunOnce(t host.Tick, about P) { e.react(t, about) }

// column is a tag of a family an entity must carry for the rule to fire on it, read off the
// family's column of the pass's query
type column struct {
	family reflect.Type // the family's component type, tag.Tags[F]
	bit    uint8
	col    store.Col
}

type eachWith[P any] struct {
	label string
	conds []column
	react func(host.Tick, P)
}

func newEachWith[P any](label string, react func(host.Tick, P), conds ...column) *eachWith[P] {
	return &eachWith[P]{label: label, conds: conds, react: react}
}

func (e *eachWith[P]) Hosted() any { return e }

func (e *eachWith[P]) ForRole(role cond.Role) Rule { return forRole(e, role) }

func (e *eachWith[P]) String() string { return e.label }

func (e *eachWith[P]) BindColumns(cols store.Columns) {
	for i := range e.conds {
		e.conds[i].col = cols.Column(e.conds[i].family)
	}
}

func (e *eachWith[P]) RunEach(t host.Tick, row store.Row, n int, keep func(i int) bool, about func(i int) P) {
	for _, c := range e.conds {
		if row.Columns[c.col.Idx] == nil {
			return
		}
	}
	for i := range n {
		if (keep == nil || keep(i)) && e.holds(row, i) {
			e.react(t, about(i))
		}
	}
}

func (e *eachWith[P]) holds(row store.Row, i int) bool {
	for _, c := range e.conds {
		if *(*uint64)(row.At(c.col, i))&(1<<c.bit) == 0 {
			return false
		}
	}
	return true
}

type played[P any] struct {
	*eachWith[P]
	roles uint64
}

func (p *played[P]) Hosted() any { return p }

func (p *played[P]) ForRole(role cond.Role) Rule { return forRole(p, role) }

func (p *played[P]) RunOnce(t host.Tick, about P) { p.react(t, about) }

func (p *played[P]) OnceRoles() uint64 { return p.roles }

type pair[P any] struct {
	label       string
	self, other host.Side
	within      []host.Side
	react       func(host.Tick, P)
}

func (p *pair[P]) Hosted() any { return p }

func (p *pair[P]) ForRole(role cond.Role) Rule { return forRole(p, role) }

func (p *pair[P]) String() string { return p.label }

func (p *pair[P]) PairSides() (self, other host.Side) { return p.self, p.other }

func (p *pair[P]) PairWithin() []host.Side { return p.within }

func (p *pair[P]) RunPair(t host.Tick, about P) { p.react(t, about) }
