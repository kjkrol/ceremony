package act

import (
	"github.com/kjkrol/ceremony/cond"
	"github.com/kjkrol/ceremony/host"
	"github.com/kjkrol/goke/v3"
)

type every[P any] struct {
	label string
	react func(host.Tick, P)
}

func (e *every[P]) Hosted() any { return e }

func (e *every[P]) ForRole(role cond.Role) Rule { return forRole(e, role) }

func (e *every[P]) String() string { return e.label }

func (*every[P]) BindColumns(*host.Columns) {}

func (e *every[P]) RunEach(t host.Tick, cursor *goke.Cursor, keep func(i int) bool, about func(i int) P) {
	for i := range cursor.IDs {
		if keep == nil || keep(i) {
			e.react(t, about(i))
		}
	}
}

func (e *every[P]) RunOnce(t host.Tick, about P) { e.react(t, about) }

type state interface {
	bind(cols *host.Columns)
	present(cursor *goke.Cursor) bool
	at(cursor *goke.Cursor, i int) any
}

func stateOf[T any]() state { return &stateCol[T]{} }

type stateCol[T any] struct{ comp *goke.OptComp[T] }

func (s *stateCol[T]) bind(cols *host.Columns)           { s.comp = cols.Of[T]() }
func (s *stateCol[T]) present(cursor *goke.Cursor) bool  { return s.comp.Present(cursor) }
func (s *stateCol[T]) at(cursor *goke.Cursor, i int) any { return &s.comp.Slice(cursor)[i] }

type column struct {
	mk    func() state
	s     state
	holds func(state any) bool // nil: carrying it is enough
}

type eachWith[P any] struct {
	label string
	conds []column
	react func(host.Tick, P)
}

func newEachWith[P any](label string, react func(host.Tick, P), conds ...column) *eachWith[P] {
	for i := range conds {
		conds[i].s = conds[i].mk()
	}
	return &eachWith[P]{label: label, conds: conds, react: react}
}

func (e *eachWith[P]) Hosted() any { return e }

func (e *eachWith[P]) ForRole(role cond.Role) Rule { return forRole(e, role) }

func (e *eachWith[P]) String() string { return e.label }

func (e *eachWith[P]) BindColumns(cols *host.Columns) {
	for _, c := range e.conds {
		c.s.bind(cols)
	}
}

func (e *eachWith[P]) RunEach(t host.Tick, cursor *goke.Cursor, keep func(i int) bool, about func(i int) P) {
	for _, c := range e.conds {
		if !c.s.present(cursor) {
			return
		}
	}
	for i := range cursor.IDs {
		if (keep == nil || keep(i)) && e.holds(cursor, i) {
			e.react(t, about(i))
		}
	}
}

func (e *eachWith[P]) holds(cursor *goke.Cursor, i int) bool {
	for _, c := range e.conds {
		if c.holds != nil && !c.holds(c.s.at(cursor, i)) {
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
