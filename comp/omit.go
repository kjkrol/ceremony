package comp

import (
	"reflect"
	"unsafe"
)

// Without leaves the roster's default T out of a kind's Spec — a unit nothing pushes leaves out
// collision.Physics. It is a marker kind.Template.Spec takes out; it never reaches a spawn.
func Without[T any]() Comp { return omit{reflect.TypeFor[T]()} }

// Omitted reports whether c is a Without marker, and of which type.
func Omitted(c Comp) (reflect.Type, bool) {
	o, ok := c.(omit)
	return o.t, ok
}

type omit struct{ t reflect.Type }

func (o omit) Type() reflect.Type     { return o.t }
func (omit) Row() reflect.Type        { return nil }
func (omit) Fill(unsafe.Pointer, any) {}
