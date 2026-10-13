package host_test

import (
	"reflect"
	"testing"
	"time"

	"github.com/kjkrol/ceremony/host"
	"github.com/kjkrol/ceremony/store"
	"github.com/kjkrol/ceremony/tag"
	"github.com/kjkrol/goke/v3"

	cgoke "github.com/kjkrol/ceremony/goke"
)

type moment struct{ i int }

func TestRules_TriggersOverOneComponentShareItsColumn(t *testing.T) {
	h := &host.Rules[moment]{}
	seen := [2]int{}
	for k := range seen {
		if err := h.Add(eachRule(func(_ host.Tick, b *hostBody, _ moment) { seen[k] += b.N })); err != nil {
			t.Fatal(err)
		}
	}
	ecs := goke.New()
	ecs.Setup(goke.SystemFn{OnInit: func(si *goke.SysInit) {
		var bodies goke.Comp[hostBody]
		f := si.NewFactory(&bodies)
		f.Create(3)
		f.Next()
		for i := range bodies.Slice(&f.Cursor) {
			bodies.Slice(&f.Cursor)[i].N = 1
		}
		qb := si.NewQueryBuilder()
		cols := cgoke.NewColumns(qb)
		h.Bind(cols)
		if own := cols.Column(reflect.TypeFor[hostBody]()); own.Idx != 0 {
			t.Errorf("the two rules' column of hostBody is at %d; want one column, the first", own.Idx)
		}
		q := qb.Build()
		for q.All(); q.Next(); {
			cur := q.Cursor()
			h.Run(host.Tick{Dt: time.Millisecond}, cols.Row(cur), len(cur.IDs), func(i int) moment { return moment{i: i} })
		}
	}})
	if seen != [2]int{3, 3} {
		t.Errorf("the rules saw %v; want both rules on all 3", seen)
	}
}

func TestRules_RunWhereSkipsWhatKeepLeavesOut(t *testing.T) {
	h := &host.Rules[moment]{}
	var each, every []int
	if err := h.Add(eachRule(func(_ host.Tick, _ *hostBody, m moment) { each = append(each, m.i) })); err != nil {
		t.Fatal(err)
	}
	if err := h.Add(everyRule(func(_ host.Tick, m moment) { every = append(every, m.i) })); err != nil {
		t.Fatal(err)
	}
	ecs := goke.New()
	ecs.Setup(goke.SystemFn{OnInit: func(si *goke.SysInit) {
		var bodies goke.Comp[hostBody]
		f := si.NewFactory(&bodies)
		f.Create(4)
		f.Next()
		qb := si.NewQueryBuilder()
		cols := cgoke.NewColumns(qb)
		h.Bind(cols)
		q := qb.Build()
		for q.All(); q.Next(); {
			cur := q.Cursor()
			h.RunWhere(host.Tick{}, cols.Row(cur), len(cur.IDs), func(i int) bool { return i%2 == 1 }, func(i int) moment { return moment{i: i} })
		}
	}})
	if len(each) != 2 || each[0] != 1 || each[1] != 3 || len(every) != 2 || every[0] != 1 || every[1] != 3 {
		t.Errorf("Each saw %v, Every %v; want both [1 3]", each, every)
	}
}

func eachRule[T, P any](react func(host.Tick, *T, P)) any { return &eachOf[T, P]{react: react} }

func everyRule[P any](react func(host.Tick, P)) any { return &everyOf[P]{react: react} }

func pairRule[FA, FB, P any](a tag.Tag[FA], b tag.Tag[FB], react func(host.Tick, P)) any {
	return &pairOf[P]{a: host.SideOf(a), b: host.SideOf(b), react: react}
}

type eachOf[T, P any] struct {
	col   store.Col
	react func(host.Tick, *T, P)
}

func (e *eachOf[T, P]) BindColumns(cols store.Columns) { e.col = cols.Column(reflect.TypeFor[T]()) }

func (e *eachOf[T, P]) RunEach(t host.Tick, row store.Row, n int, keep func(i int) bool, about func(i int) P) {
	if row.Columns[e.col.Idx] == nil {
		return
	}
	for i := range n {
		if keep == nil || keep(i) {
			e.react(t, (*T)(row.At(e.col, i)), about(i))
		}
	}
}

type everyOf[P any] struct{ react func(host.Tick, P) }

func (*everyOf[P]) BindColumns(store.Columns) {}

func (e *everyOf[P]) RunEach(t host.Tick, _ store.Row, n int, keep func(i int) bool, about func(i int) P) {
	for i := range n {
		if keep == nil || keep(i) {
			e.react(t, about(i))
		}
	}
}

func (e *everyOf[P]) RunOnce(t host.Tick, about P) { e.react(t, about) }

type pairOf[P any] struct {
	a, b  host.Side
	react func(host.Tick, P)
}

func (p *pairOf[P]) PairSides() (self, other host.Side) { return p.a, p.b }

func (p *pairOf[P]) RunPair(t host.Tick, pair P) { p.react(t, pair) }
