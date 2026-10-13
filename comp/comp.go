package comp

import (
	"reflect"
	"unsafe"
)

// Comp is one component of a kind's Spec — made with Const, Load, Tagged or Without, never
// implemented by a game. Its methods are the spawning store's way in.
type Comp interface {
	// Type is the component type the Comp stands for.
	Type() reflect.Type
	// Row is the row type a Load reads, nil for a component that reads none.
	Row() reflect.Type
	// Fill writes the component's value for the entity row describes into dst, a value of
	// Type the store laid out.
	Fill(dst unsafe.Pointer, row any)
}

// Template yields one component value per spawned entity.
type Template[T any] struct {
	value func(row any) T
	row   reflect.Type
}

var _ Comp = Template[struct{}]{}

// Const is a component whose value is the same for every entity of the kind.
func Const[T any](v T) Template[T] {
	return Template[T]{value: func(any) T { return v }}
}

// Load is a component whose value is read from each entity's own row.
func Load[P, T any](load func(row P) T) Template[T] {
	return Template[T]{value: func(row any) T { return load(row.(P)) }, row: reflect.TypeFor[P]()}
}

// Resolve is the value for the entity row describes.
func (t Template[T]) Resolve(row any) T { return t.value(row) }

// Type is T.
func (t Template[T]) Type() reflect.Type { return reflect.TypeFor[T]() }

// Row is the row type a Load reads, nil for a Const.
func (t Template[T]) Row() reflect.Type { return t.row }

// Fill writes the value for row into dst; nothing for a T of no size.
func (t Template[T]) Fill(dst unsafe.Pointer, row any) {
	v := t.value(row)
	if unsafe.Sizeof(v) > 0 {
		*(*T)(dst) = v
	}
}
