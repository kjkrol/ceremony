package goke_test

import (
	"reflect"
	"testing"

	"github.com/kjkrol/ceremony/comp"
	cgoke "github.com/kjkrol/ceremony/goke"
	"github.com/kjkrol/ceremony/store"
	"github.com/kjkrol/goke/v3"
)

type (
	health struct{ HP int32 }
	armour struct{ N int32 }
	row    struct{ hp int32 }
)

func TestColumns_HandsTheEngineTheArraysOfTheChunk(t *testing.T) {
	ecs := goke.New()
	var hp goke.Comp[health]
	var own goke.OptComp[armour]
	var cols *cgoke.Columns
	var q *goke.Query
	var hpCol, armCol store.Col
	ecs.Setup(goke.SystemFn{OnInit: func(si *goke.SysInit) {
		f := si.NewFactory(&hp)
		f.Create(3)
		for f.Next() {
			for i := range hp.Slice(&f.Cursor) {
				hp.Slice(&f.Cursor)[i].HP = int32(10 + i)
			}
		}
		qb := si.NewQueryBuilder()
		cols = cgoke.NewColumns(qb)
		cols.Own(&own)
		armCol = cols.Column(reflect.TypeFor[armour]())
		hpCol = cols.Column(reflect.TypeFor[health]())
		if again := cols.Column(reflect.TypeFor[health]()); again != hpCol {
			t.Errorf("a second Column of health is %v, want the first %v", again, hpCol)
		}
		q = qb.Build()
	}})
	seen := 0
	for q.All(); q.Next(); {
		cur := q.Cursor()
		r := cols.Row(cur)
		if r.At(armCol, 0) != nil || own.Present(cur) {
			t.Error("the absent own column is present")
		}
		for i := range cur.IDs {
			if got := (*health)(r.At(hpCol, i)); got.HP != int32(10+i) {
				t.Errorf("entity %d read %v", i, *got)
			}
			seen++
		}
	}
	if seen != 3 {
		t.Fatalf("walked %d entities, want 3", seen)
	}
}

func TestSpawner_FillsEveryComponentOfTheBatch(t *testing.T) {
	ecs := goke.New()
	sp := cgoke.NewSpawner(
		comp.Load(func(r row) health { return health{HP: r.hp} }),
		comp.Const(armour{N: 2}),
	)
	var hp goke.Comp[health]
	var arm goke.Comp[armour]
	var q *goke.Query
	ecs.Setup(goke.SystemFn{OnInit: func(si *goke.SysInit) {
		f := si.NewFactory(sp.Columns()...)
		f.Create(2)
		for f.Next() {
			for i := range f.IDs {
				sp.Write(&f.Cursor, i, row{hp: int32(5 + i)})
			}
		}
		q = si.NewQueryBuilder(&hp, &arm).Build()
	}})
	n := 0
	for q.All(); q.Next(); {
		cur := q.Cursor()
		for i := range cur.IDs {
			if hp.Slice(cur)[i].HP != int32(5+i) || arm.Slice(cur)[i].N != 2 {
				t.Errorf("entity %d: %v %v", i, hp.Slice(cur)[i], arm.Slice(cur)[i])
			}
			n++
		}
	}
	if n != 2 {
		t.Fatalf("spawned %d, want 2", n)
	}
	if cgoke.LoadComp(comp.Const(armour{})).Name != goke.LoadComp[armour]().Name {
		t.Error("LoadComp names the component otherwise than goke.LoadComp")
	}
}
