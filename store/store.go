package store

import (
	"reflect"
	"unsafe"

	"github.com/kjkrol/uid"
)

// Row is a block of entities as the store hands it to the engine: for every column bound, the
// array of its values over the block, nil where the block carries none.
type Row struct {
	Columns []unsafe.Pointer // by Col.Idx
}

// Col is one column of a Row: its place and the size of one value.
type Col struct {
	Idx  int
	Size uintptr
}

// At is the value of col for the i-th entity of the row, nil where the row carries none.
func (r Row) At(col Col, i int) unsafe.Pointer {
	base := r.Columns[col.Idx]
	if base == nil {
		return nil
	}
	return unsafe.Add(base, uintptr(i)*col.Size)
}

// Columns is what hands the engine the columns it reads, as a pass is bound: Column registers
// the component type t with the store and says where the Row carries it.
type Columns interface {
	Column(t reflect.Type) Col
}

// Facts is how a store gives the engine the facts it reads and writes: a fact is a value of
// its type on an entity, one at most.
type Facts interface {
	Columns
	// Stage is where a fact of col for id is written, in the store's own staging, landing as
	// the store syncs; nil where id has no mind to take it.
	Stage(id uid.UID64, col Col) unsafe.Pointer
	// Drop takes the fact of col off id as the store syncs.
	Drop(id uid.UID64, col Col)
}
