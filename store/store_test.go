package store_test

import (
	"testing"
	"unsafe"

	"github.com/kjkrol/ceremony/store"
)

type fact struct{ A, B int32 }

func TestRow_IndexesTheArraysTheStoreHandsOver(t *testing.T) {
	facts := []fact{{1, 2}, {3, 4}, {5, 6}}
	tags := []uint64{7, 8, 9}
	row := store.Row{Columns: []unsafe.Pointer{unsafe.Pointer(&facts[0]), nil, unsafe.Pointer(&tags[0])}}
	factCol := store.Col{Idx: 0, Size: unsafe.Sizeof(fact{})}
	absent := store.Col{Idx: 1, Size: 8}
	tagCol := store.Col{Idx: 2, Size: 8}
	for i := range facts {
		if got := *(*fact)(row.At(factCol, i)); got != facts[i] {
			t.Errorf("fact %d read %v", i, got)
		}
		if got := *(*uint64)(row.At(tagCol, i)); got != tags[i] {
			t.Errorf("tag %d read %v", i, got)
		}
		if row.At(absent, i) != nil {
			t.Errorf("an absent column has a value at %d", i)
		}
	}
	*(*fact)(row.At(factCol, 1)) = fact{30, 40}
	if facts[1] != (fact{30, 40}) {
		t.Error("a write through the row did not reach the array")
	}
}
