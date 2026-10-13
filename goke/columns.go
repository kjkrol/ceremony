package goke

import (
	"reflect"
	"unsafe"

	"github.com/kjkrol/ceremony/comp"
	"github.com/kjkrol/ceremony/store"
	"github.com/kjkrol/goke/v3"
)

// Columns binds the columns a host or the plans read to one query, as optional columns, and
// hands the engine each block of it as a store.Row: one Columns a query.
type Columns struct {
	qb     *goke.QueryBuilder
	cols   []*goke.OptColumn
	byType map[reflect.Type]store.Col
	row    store.Row
}

// NewColumns is the columns of the query qb builds.
func NewColumns(qb *goke.QueryBuilder) *Columns {
	return &Columns{qb: qb, byType: map[reflect.Type]store.Col{}}
}

// Column is the column of the component type t, made optional in the query at its first.
func (c *Columns) Column(t reflect.Type) store.Col {
	if col, ok := c.byType[t]; ok {
		return col
	}
	oc := goke.NewOptColumn(t)
	c.qb.Optional(oc)
	col := store.Col{Idx: len(c.cols), Size: t.Size()}
	c.cols = append(c.cols, oc)
	c.row.Columns = append(c.row.Columns, nil)
	c.byType[t] = col
	return col
}

// Own has the pass's own optional column col, of T, stand for T in the rows too, so the rules
// over T read the column the pass reads: bound here, before any Column of T.
func (c *Columns) Own[T any](col *goke.OptComp[T]) {
	t := reflect.TypeFor[T]()
	if _, ok := c.byType[t]; ok {
		panic("ceremony/goke: Own after a Column of " + t.String())
	}
	c.qb.Optional(col)
	oc := goke.NewOptColumn(t)
	oc.Idx = col.Index()
	c.byType[t] = store.Col{Idx: len(c.cols), Size: t.Size()}
	c.cols = append(c.cols, oc)
	c.row.Columns = append(c.row.Columns, nil)
}

// Row is the block of entities the cursor stands on, in All mode; in Seek mode the entity's
// values are at the cursor's slot.
func (c *Columns) Row(cur *goke.Cursor) store.Row {
	for k, col := range c.cols {
		c.row.Columns[k] = col.Base(cur)
	}
	return c.row
}

var _ store.Columns = (*Columns)(nil)

// Spawner writes a kind's components onto the entities a factory makes: its Columns go to the
// factory, Write fills one entity of a batch.
type Spawner struct {
	comps []comp.Comp
	cols  []*goke.Column
}

// NewSpawner is the spawner of comps.
func NewSpawner(comps ...comp.Comp) *Spawner {
	s := &Spawner{comps: comps, cols: make([]*goke.Column, len(comps))}
	for k, c := range comps {
		s.cols[k] = goke.NewColumn(c.Type())
	}
	return s
}

// Columns are the factory's columns, one a component.
func (s *Spawner) Columns() []goke.Addable {
	out := make([]goke.Addable, len(s.cols))
	for k, col := range s.cols {
		out[k] = col
	}
	return out
}

// Write fills the i-th entity of the batch the cursor stands on from row.
func (s *Spawner) Write(cur *goke.Cursor, i int, row any) {
	for k, c := range s.comps {
		if base := s.cols[k].Base(cur); base != nil {
			c.Fill(unsafe.Add(base, uintptr(i)*s.cols[k].Size()), row)
		}
	}
}

// LoadComp names a component to a save file being loaded.
func LoadComp(c comp.Comp) goke.CompToken { return goke.LoadCompOf(c.Type()) }
