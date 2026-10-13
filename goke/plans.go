package goke

import (
	"reflect"
	"time"
	"unsafe"

	"github.com/kjkrol/ceremony/plan"
	"github.com/kjkrol/ceremony/store"
	"github.com/kjkrol/goke/v3"
	"github.com/kjkrol/uid"
)

// Plans runs a plan.Plans over goke: a system walking every entity with a store.Mind once a
// step, the facts the plans read bound to its query, the facts they write staged in its own
// command buffer.
type Plans struct {
	p       *plan.Plans
	si      *goke.SysInit
	cols    *Columns
	ids     []goke.CompID // by store.Col.Idx
	query   *goke.Query
	mind    goke.Comp[store.Mind]
	minds   *goke.Query
	anyMind goke.Comp[store.Mind]
	cb      *goke.CmdBuf
}

// NewPlans is the system running p.
func NewPlans(p *plan.Plans) *Plans { return &Plans{p: p} }

var (
	_ goke.System = (*Plans)(nil)
	_ store.Facts = (*Plans)(nil)
)

// Init binds the plans; a plan registered after panics.
func (s *Plans) Init(si *goke.SysInit) {
	s.si = si
	qb := si.NewQueryBuilder(&s.mind)
	s.cols = NewColumns(qb)
	s.p.Bind(s)
	s.query = qb.Build()
	s.minds = si.NewQueryBuilder(&s.anyMind).Build()
}

// Update runs the plans of every entity with a mind.
func (s *Plans) Update(cb *goke.CmdBuf, dt time.Duration) {
	s.cb = cb
	for s.query.All(); s.query.Next(); {
		cur := s.query.Cursor()
		s.p.Tick(s.cols.Row(cur), cur.IDs, s.mind.Slice(cur), dt)
	}
}

// Column is store.Columns: the fact type t registered and bound to the query.
func (s *Plans) Column(t reflect.Type) store.Col {
	col := s.cols.Column(t)
	for len(s.ids) <= col.Idx {
		s.ids = append(s.ids, 0)
	}
	s.ids[col.Idx] = s.si.RegCompOf(t)
	return col
}

// Stage is store.Facts: the staged slot of a fact for id, nil for an entity with no mind.
func (s *Plans) Stage(id uid.UID64, col store.Col) unsafe.Pointer {
	if !s.minds.Seek(id) {
		return nil
	}
	return s.cb.Reserve(id, s.ids[col.Idx])
}

// Drop is store.Facts: the fact of col taken off id at the next Sync.
func (s *Plans) Drop(id uid.UID64, col store.Col) { s.cb.RemoveCompOne(id, s.ids[col.Idx]) }

// LoadComps lists what the plans save: the minds and every fact they read or write.
func (s *Plans) LoadComps() []goke.CompToken {
	out := []goke.CompToken{goke.LoadComp[store.Mind]()}
	for _, t := range s.p.Facts() {
		out = append(out, goke.LoadCompOf(t))
	}
	return out
}
